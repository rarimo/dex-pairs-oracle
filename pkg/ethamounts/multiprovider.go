package ethamounts

import (
	"context"
	"encoding/hex"
	"math/big"
	"net/url"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mitchellh/mapstructure"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type MuliAmountsResponse struct {
	BlockNum *big.Int   `mapstructure:"block_number"`
	Amounts  []*big.Int `mapstructure:"amounts"`
}

type MultiProvider struct {
	ethClient            *ethclient.Client
	bc                   []byte
	abi                  abi.ABI
	returnValuesUnpacker abi.Arguments
}

func NewMultiProvider(rpc *url.URL) (*MultiProvider, error) {
	ethClient, err := ethclient.Dial(rpc.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial rpc")
	}

	parsedABI, err := abi.JSON(strings.NewReader(multicallABI))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse abi")
	}

	bc, err := hex.DecodeString(bytecode)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode bytecode")
	}

	u256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create uint256 type")
	}

	u256ArrType, err := abi.NewType("uint256[]", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create uint256[] type")
	}

	returnValues := abi.Arguments{
		abi.Argument{
			Name: "block_number",
			Type: u256Type,
		},
		abi.Argument{
			Name: "amounts",
			Type: u256ArrType,
		},
	}

	return &MultiProvider{
		bc:                   bc,
		abi:                  parsedABI,
		returnValuesUnpacker: returnValues,
		ethClient:            ethClient,
	}, nil
}

func (m *MultiProvider) Amounts(ctx context.Context, account common.Address, tokens []common.Address) (*big.Int, []*big.Int, error) {
	if len(tokens) == 0 {
		return nil, nil, nil
	}

	inputData, err := m.abi.Constructor.Inputs.Pack(tokens, account)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to pack input data for account", logan.F{
			"account": account.String(),
		})
	}

	result, err := m.ethClient.PendingCallContract(ctx, ethereum.CallMsg{
		Data: append(m.bc[:], inputData[:]...),
	})

	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to call contract")
	}

	values := make(map[string]interface{})

	if err := m.returnValuesUnpacker.UnpackIntoMap(values, result); err != nil {
		return nil, nil, errors.Wrap(err, "failed to unpack return values", logan.F{
			"response_hex": hexutil.Encode(result),
		})
	}

	var resp MuliAmountsResponse
	if err := mapstructure.Decode(values, &resp); err != nil {
		return nil, nil, errors.Wrap(err, "failed to decode response", logan.F{
			"values": values,
		})
	}

	return resp.BlockNum, resp.Amounts, nil
}
