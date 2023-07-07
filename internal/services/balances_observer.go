package services

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"gitlab.com/rarimo/dex-pairs-oracle/internal/chains"

	"gitlab.com/rarimo/dex-pairs-oracle/pkg/ethamounts"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum/go-ethereum/common"

	"gitlab.com/distributed_lab/kit/pgdb"

	"gitlab.com/distributed_lab/logan/v3/errors"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/running"

	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"

	"gitlab.com/rarimo/dex-pairs-oracle/internal/config"
)

func RunBalancesObserver(ctx context.Context, cfg config.Config) {
	var wg sync.WaitGroup

	wg.Add(len(cfg.ChainsCfg().Chains))

	for _, chain := range cfg.ChainsCfg().Chains {
		go func(chain chains.Chain) {
			defer wg.Done()

			observer := balancesObserver{
				log:             cfg.Log(),
				storage:         cfg.NewStorage(),
				chain:           chain,
				ethAmounter:     ethamounts.NewProvider(cfg.ChainsCfg()),
				observePageSize: cfg.BalancesObserver().PageSize,
			}

			running.WithBackOff(ctx, observer.log, fmt.Sprintf("balances_observer:%s", chain.Name),
				observer.runOnce,
				cfg.BalancesObserver().Interval,
				2*cfg.BalancesObserver().Interval,
				5*cfg.BalancesObserver().Interval)
		}(chain)
	}

	wg.Wait()
}

type EthAmounter interface {
	Amount(ctx context.Context, chainID int64, token common.Address, account common.Address) (amount *big.Int, err error)
}

type balancesObserver struct {
	log     *logan.Entry
	storage data.Storage
	chain   chains.Chain

	ethClients  map[int64]*ethclient.Client // map[chain_id]client
	ethAmounter EthAmounter

	observePageSize uint64
}

func (b balancesObserver) runOnce(ctx context.Context) error {
	var cursor int64

	running.UntilSuccess(ctx, b.log, fmt.Sprintf("run_once_balances_observer:%s", b.chain.Name), func(ctx context.Context) (bool, error) {
		balances, err := b.storage.BalanceQ().SelectCtx(ctx, data.BalancesSelector{
			Cursor:   cursor,
			PageSize: b.observePageSize,
			ChainID:  &b.chain.ID,
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

		accountTokens := make(map[common.Address][]common.Address) // map[address]token

		balancesMapping := make(map[string]data.Balance) // map[address:token]balance

		for _, balance := range balances {
			addrTokens, ok := accountTokens[common.BytesToAddress(balance.AccountAddress)]
			if !ok {
				addrTokens = make([]common.Address, 0, len(balances))
			}
			accountTokens[common.BytesToAddress(balance.AccountAddress)] = append(addrTokens, common.BytesToAddress(balance.Token))

			balancesMapping[fmt.Sprintf("%s:%s",
				common.BytesToAddress(balance.AccountAddress).String(),
				common.BytesToAddress(balance.Token).String())] = balance
		}

		updatedBalances := make([]data.Balance, 0, len(balances))
		for addr, tokens := range accountTokens {
			block, amounts, err := b.chain.BalanceProvider.GetMultipleBalances(&bind.CallOpts{Context: ctx}, tokens, addr)
			if err != nil {
				return false, errors.Wrap(err, "failed to fetch amounts")
			}

			now := time.Now()

			for i, token := range tokens {
				balanceKey := fmt.Sprintf("%s:%s", addr.String(), token.String())

				balance := balancesMapping[balanceKey]
				balance.Amount = data.Int256{
					Int: amounts[i],
				}
				balance.LastKnownBlock = block.Int64()
				balance.UpdatedAt = now

				updatedBalances = append(updatedBalances, balance)
			}
		}

		err = b.storage.BalanceQ().UpsertBatchCtx(ctx, updatedBalances...)
		if err != nil {
			return false, errors.Wrap(err, "failed to upsert balances")
		}

		cursor = balances[len(balances)-1].ID

		return false, nil
	}, 1*time.Second, 5*time.Second)

	return nil
}
