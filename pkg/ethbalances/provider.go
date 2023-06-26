package ethbalances

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"gitlab.com/distributed_lab/logan"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"
	"gitlab.com/rarimo/dex-pairs-oracle/pkg/ethamounts"
)

type Provider struct {
	redisstore data.RedisStore
	amounter   *ethamounts.Provider
}

func NewProvider(redisstore data.RedisStore, provider *ethamounts.Provider) *Provider {
	return &Provider{
		redisstore: redisstore,
		amounter:   provider,
	}
}

func (p *Provider) GetBalances(ctx context.Context, address string, chainID int64, cursor string, limit int64) ([]data.Balance, error) {
	tokens, err := p.redisstore.Tokens().Page(ctx, chainID, cursor, limit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get tokens by chain id", logan.F{
			"chain_id": chainID,
		})
	}

	if len(tokens) == 0 {
		return nil, nil
	}

	balances := make([]data.Balance, len(tokens))

	accountAddr := common.HexToAddress(address)

	for i, token := range tokens {
		tokenAddr := common.HexToAddress(token.Address)

		amount, err := p.amounter.Amount(ctx, chainID, tokenAddr, accountAddr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get amount", logan.F{
				"address": address,
				"token":   token,
			})
		}

		now := time.Now()

		balances[i] = data.Balance{
			AccountAddress: accountAddr.Bytes(),
			Token:          tokenAddr.Bytes(),
			ChainID:        chainID,
			Amount:         data.Int256{Int: amount},
			CreatedAt:      now,
			UpdatedAt:      now,
		}
	}

	return balances, nil
}
