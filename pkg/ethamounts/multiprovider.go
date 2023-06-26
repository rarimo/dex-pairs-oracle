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

const (
	multicallABI = "[{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"tokens\",\"type\":\"address[]\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"}]"
	bytecode     = "608060405234801561001057600080fd5b506040516105003803806105008339818101604052810190610032919061029c565b60008251905060608167ffffffffffffffff8111801561005157600080fd5b506040519080825280602002602001820160405280156100805781602001602082028036833780820191505090505b50905060005b828110156101c057600085828151811061009c57fe5b60200260200101519050600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161415610110578473ffffffffffffffffffffffffffffffffffffffff16318383815181106100ff57fe5b6020026020010181815250506101b2565b8073ffffffffffffffffffffffffffffffffffffffff166370a08231866040518263ffffffff1660e01b815260040161014991906103bc565b60206040518083038186803b15801561016157600080fd5b505afa158015610175573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061019991906102f0565b8383815181106101a557fe5b6020026020010181815250505b508080600101915050610086565b50606043826040516020016101d69291906103d7565b6040516020818303038152906040529050805160208201f35b6000815190506101fe816104d1565b92915050565b600082601f83011261021557600080fd5b815161022861022382610434565b610407565b9150818183526020840193506020810190508385602084028201111561024d57600080fd5b60005b8381101561027d578161026388826101ef565b845260208401935060208301925050600181019050610250565b5050505092915050565b600081519050610296816104e8565b92915050565b600080604083850312156102af57600080fd5b600083015167ffffffffffffffff8111156102c957600080fd5b6102d585828601610204565b92505060206102e6858286016101ef565b9150509250929050565b60006020828403121561030257600080fd5b600061031084828501610287565b91505092915050565b6000610325838361039e565b60208301905092915050565b61033a81610495565b82525050565b600061034b8261046c565b6103558185610484565b93506103608361045c565b8060005b838110156103915781516103788882610319565b975061038383610477565b925050600181019050610364565b5085935050505092915050565b6103a7816104c7565b82525050565b6103b6816104c7565b82525050565b60006020820190506103d16000830184610331565b92915050565b60006040820190506103ec60008301856103ad565b81810360208301526103fe8184610340565b90509392505050565b6000604051905081810181811067ffffffffffffffff8211171561042a57600080fd5b8060405250919050565b600067ffffffffffffffff82111561044b57600080fd5b602082029050602081019050919050565b6000819050602082019050919050565b600081519050919050565b6000602082019050919050565b600082825260208201905092915050565b60006104a0826104a7565b9050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b6104da81610495565b81146104e557600080fd5b50565b6104f1816104c7565b81146104fc57600080fd5b5056fe"
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
