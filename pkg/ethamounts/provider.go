package ethamounts

import (
	"context"
	"math/big"

	abibind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarimo/dex-pairs-oracle/pkg/ethamounts/bind"
)

type Provider struct {
	contracts map[common.Address]*bind.ERC20Caller
	ethclient *ethclient.Client
}

func NewProvider(ethclient *ethclient.Client) *Provider {
	return &Provider{
		contracts: make(map[common.Address]*bind.ERC20Caller),
		ethclient: ethclient,
	}
}

func (p *Provider) Amount(ctx context.Context, token common.Address, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	if isZeroAddr(token) {
		return p.ethclient.BalanceAt(ctx, account, blockNumber)
	}

	contract := p.contracts[token]
	if contract == nil {
		erc20, err := bind.NewERC20Caller(token, p.ethclient)
		if err != nil {
			return nil, errors.Wrap(err, "failed to init erc20 caller", logan.F{
				"token": token.String(),
			})
		}
		p.contracts[token] = erc20

		contract = erc20
	}

	return contract.BalanceOf(&abibind.CallOpts{
		BlockNumber: blockNumber,
		Context:     ctx,
	}, account)
}

func isZeroAddr(addr common.Address) bool {
	for _, b := range addr {
		if b != 0 {
			return false
		}
	}
	return true
}
