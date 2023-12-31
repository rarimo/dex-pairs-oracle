package ethamounts

import (
	"context"
	"math/big"

	"github.com/rarimo/dex-pairs-oracle/internal/chains"

	abibind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rarimo/dex-pairs-oracle/pkg/ethamounts/bind"
	"github.com/rarimo/dex-pairs-oracle/pkg/etherrors"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Provider struct {
	chains       *chains.Config
	contracts    map[common.Address]*bind.ERC20Caller
	chainClients map[int64]*ethclient.Client
}

func NewProvider(chains *chains.Config) *Provider {
	return &Provider{
		chains:       chains,
		contracts:    make(map[common.Address]*bind.ERC20Caller),
		chainClients: make(map[int64]*ethclient.Client),
	}
}

func (p *Provider) Amount(ctx context.Context, chainID int64, token common.Address, account common.Address) (*big.Int, error) {
	chain := p.chains.Find(chainID)
	if chain == nil {
		return nil, etherrors.ErrChainNotSupported
	}

	ethc, ok := p.chainClients[chainID]
	if !ok {
		dial, err := ethclient.Dial(chain.RPCUrl.String())
		if err != nil {
			return nil, errors.Wrap(err, "failed to dial rpc", logan.F{
				"chain_id": chainID,
			})
		}
		p.chainClients[chainID] = dial
		ethc = dial
	}

	if isZeroAddr(token) {
		balance, err := ethc.BalanceAt(ctx, account, nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get balance", logan.F{
				"account": account.String(),
			})
		}

		return balance, nil
	}

	contract, ok := p.contracts[token]
	if !ok {
		erc20, err := bind.NewERC20Caller(token, ethc)
		if err != nil {
			return nil, errors.Wrap(err, "failed to init erc20 caller", logan.F{
				"token": token.String(),
			})
		}
		p.contracts[token] = erc20

		contract = erc20
	}

	balance, err := contract.BalanceOf(&abibind.CallOpts{
		Context: ctx,
	}, account)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get balance", logan.F{
			"account": account.String(),
			"token":   token.String(),
		})
	}

	return balance, nil
}

func isZeroAddr(addr common.Address) bool {
	for _, b := range addr {
		if b != 0 {
			return false
		}
	}
	return true
}
