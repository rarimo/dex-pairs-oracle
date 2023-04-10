package services

import (
	"context"
	"math/big"
	"time"

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

var (
	EthZeroAddr = common.Address{}
)

func RunBalancesObserver(ctx context.Context, cfg config.Config) {
	observer := balancesObserver{
		log:             cfg.Log(),
		storage:         cfg.NewStorage(),
		chains:          cfg.ChainsCfg(),
		ethclient:       cfg.EVM().RPCClient,
		ethAmounter:     ethamounts.NewProvider(cfg.EVM().RPCClient),
		observePageSize: cfg.BalancesObserver().PageSize,
	}

	running.WithBackOff(ctx, observer.log, "balances_observer",
		observer.runOnce,
		cfg.BalancesObserver().Period,
		2*cfg.BalancesObserver().Period,
		5*cfg.BalancesObserver().Period)

}

type EthAmounter interface {
	// TODO handle native token here as well
	Amount(ctx context.Context, token common.Address, account common.Address, blockNumber *big.Int) (*big.Int, error)
}

type balancesObserver struct {
	log     *logan.Entry
	storage data.Storage
	chains  *config.ChainsConfig

	ethclient   *ethclient.Client
	ethAmounter EthAmounter

	observePageSize uint64
}

func (b balancesObserver) runOnce(ctx context.Context) error {
	cursor := uint64(0) // it's okay i guess (in case of big number of balances should be changed to proper cursor from redis)

	running.UntilSuccess(ctx, b.log, "run_once", func(ctx context.Context) (bool, error) {
		block, err := b.ethclient.BlockByNumber(ctx, nil)
		if err != nil {
			return false, errors.Wrap(err, "failed to get latest block number")
		}

		balances, err := b.storage.BalanceQ().SelectCtx(ctx, data.BalancesSelector{
			PageCursor: cursor,
			PageSize:   b.observePageSize,
			Sort: pgdb.Sorts{
				"id",
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
				amount, err := b.ethAmounter.Amount(ctx,
					common.BytesToAddress(balances[i].Token),
					common.BytesToAddress(balances[i].AccountAddress),
					block.Number())
				if err != nil {
					return false, errors.Wrap(err, "failed to get token balance", logan.F{
						"token":   hexutil.Encode(balances[i].Token),
						"account": hexutil.Encode(balances[i].AccountAddress),
						"block":   block.Number(),
					})
				}

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

		cursor = uint64(balances[len(balances)-1].ID)

		return false, nil
	}, 1*time.Second, 5*time.Second)

	return nil
}
