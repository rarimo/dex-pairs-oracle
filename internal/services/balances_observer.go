package services

import (
	"context"
	"math/big"
	"time"

	"gitlab.com/rarimo/dex-pairs-oracle/internal/chains"

	"gitlab.com/rarimo/dex-pairs-oracle/pkg/ethamounts"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	tokenmanager "gitlab.com/rarimo/rarimo-core/x/tokenmanager/types"

	"gitlab.com/distributed_lab/kit/pgdb"

	"gitlab.com/distributed_lab/logan/v3/errors"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/running"

	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"

	"gitlab.com/rarimo/dex-pairs-oracle/internal/config"
)

func RunBalancesObserver(ctx context.Context, cfg config.Config) {
	observer := balancesObserver{
		log:             cfg.Log(),
		storage:         cfg.NewStorage(),
		chains:          cfg.ChainsCfg(),
		ethAmounter:     ethamounts.NewProvider(cfg.ChainsCfg()),
		observePageSize: cfg.BalancesObserver().PageSize,
	}

	running.WithBackOff(ctx, observer.log, "balances_observer",
		observer.runOnce,
		cfg.BalancesObserver().Interval,
		2*cfg.BalancesObserver().Interval,
		5*cfg.BalancesObserver().Interval)

}

type EthAmounter interface {
	// TODO handle native token here as well
	Amount(ctx context.Context, chainID int64, token common.Address, account common.Address) (amount *big.Int, blockNumber *big.Int, err error)
}

type balancesObserver struct {
	log     *logan.Entry
	storage data.Storage
	chains  *chains.Config

	ethClients  map[int64]*ethclient.Client // map[chain_id]client
	ethAmounter EthAmounter

	observePageSize uint64
}

func (b balancesObserver) runOnce(ctx context.Context) error {
	cursor := hexutil.MustDecode("0x0000000000000000000000000000000000000000")

	running.UntilSuccess(ctx, b.log, "run_once_balances_observer", func(ctx context.Context) (bool, error) {
		balances, err := b.storage.BalanceQ().SelectCtx(ctx, data.BalancesSelector{
			TokenCursor: cursor,
			PageSize:    b.observePageSize,
			Sort: pgdb.Sorts{
				"token",
			},
		})

		if err != nil {
			return false, errors.Wrap(err, "failed to select balances")
		}

		if len(balances) == 0 {
			return true, nil
		}

		for i := 0; i < len(balances); i++ {
			chain := b.chains.Find(balances[i].ChainID)
			if chain == nil {
				return false, errors.From(errors.New("chain not found"), logan.F{
					"chain_id": balances[i].ChainID,
				})
			}

			switch chain.Type {
			case tokenmanager.NetworkType_EVM:
				amount, block, err := b.ethAmounter.Amount(ctx,
					balances[i].ChainID,
					common.BytesToAddress(balances[i].Token),
					common.BytesToAddress(balances[i].AccountAddress))
				if err != nil {
					return false, errors.Wrap(err, "failed to get token balance", logan.F{
						"token":   hexutil.Encode(balances[i].Token),
						"account": hexutil.Encode(balances[i].AccountAddress),
						"block":   block.String(),
					})
				}

				balances[i].LastKnownBlock = block.Int64()
				balances[i].Amount = data.Int256{Int: amount}
			default: // solana, near etc.
				b.log.WithField("chain_id", balances[i].ChainID).Debug("chain type not supported")
				continue
			}
		}

		err = b.storage.BalanceQ().UpsertBatchCtx(ctx, balances...)
		if err != nil {
			return false, errors.Wrap(err, "failed to upsert balances")
		}

		cursor = balances[len(balances)-1].Token

		return false, nil
	}, 1*time.Second, 5*time.Second)

	return nil
}
