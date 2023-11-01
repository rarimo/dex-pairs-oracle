package ethbalances

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rarimo/dex-pairs-oracle/internal/chains"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rarimo/dex-pairs-oracle/internal/data"
	"gitlab.com/distributed_lab/logan"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Provider struct {
	redisstore data.RedisStore
	chain      *chains.Chain
}

func NewProvider(redisstore data.RedisStore, chain *chains.Chain) *Provider {
	if chain == nil {
		panic("chain is nil")
	}

	return &Provider{
		redisstore: redisstore,
		chain:      chain,
	}
}

func (p *Provider) GetBalances(ctx context.Context, address string, tokens []chains.TokenInfo) ([]data.Balance, error) {
	if len(tokens) == 0 {
		return nil, nil
	}

	balances := make([]data.Balance, len(tokens))

	accountAddr := common.HexToAddress(address)

	tokenAddrs := make([]common.Address, len(tokens))

	for i, t := range tokens {
		tokenAddrs[i] = common.HexToAddress(t.Address)
	}

	block, amounts, err := p.chain.BalanceProvider.GetMultipleBalances(&bind.CallOpts{Context: ctx}, tokenAddrs, accountAddr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get amounts", logan.F{
			"address": address,
		})
	}

	if len(amounts) != len(tokens) {
		return nil, errors.From(errors.New("amounts and tokens length mismatch"), logan.F{
			"address": address,
			"tokens":  tokens,
			"amounts": amounts,
		})
	}

	now := time.Now()

	for i, amount := range amounts {
		balances[i] = data.Balance{
			AccountAddress: accountAddr.Bytes(),
			Token:          tokenAddrs[i].Bytes(),
			ChainID:        p.chain.ID,
			Amount:         data.Int256{Int: amount},
			CreatedAt:      now,
			UpdatedAt:      now,
			LastKnownBlock: block.Int64(),
		}
	}

	return balances, nil
}
