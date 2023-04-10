package ethbalances

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"gitlab.com/distributed_lab/logan"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/config"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"
	"gitlab.com/rarimo/dex-pairs-oracle/pkg/ethamounts"
)

var (
	ErrChainNotSupported = errors.New("chain not supported")
)

type Provider struct {
	ethclient    *ethclient.Client
	redisstore   data.RedisStore
	chains       config.ChainsConfig
	amountGetter *ethamounts.Provider
}

func NewProvider(ethclient *ethclient.Client, redisstore data.RedisStore, chains config.ChainsConfig) *Provider {
	return &Provider{
		ethclient:    ethclient,
		redisstore:   redisstore,
		chains:       chains,
		amountGetter: ethamounts.NewProvider(ethclient),
	}
}

func (p *Provider) GetBalances(ctx context.Context, address string, chainID int64) ([]data.Balance, error) {
	chain := p.chains.Find(chainID)
	if chain == nil {
		return nil, ErrChainNotSupported
	}

	block, err := p.ethclient.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get block number", logan.F{
			"chain_id": chainID,
		})
	}

	tokens, err := p.redisstore.Tokens().All(ctx, chainID)
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

		amount, err := p.amountGetter.Amount(ctx, tokenAddr, accountAddr, block.Number())
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
			ChainID:        chain.ID,
			Amount:         data.Int256{Int: amount},
			CreatedAt:      now,
			UpdatedAt:      now,
			LastKnownBlock: block.Number().Int64(),
		}
	}

	return balances, nil
}
