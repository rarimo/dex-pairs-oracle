package services

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/config"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/pools"
)

func RunTokensObserver(ctx context.Context, cfg config.Config) {
	provider := pools.NewTokensListProvider(cfg.Log(), http.DefaultClient, cfg.RedisClient())

	if err := provider.Init(ctx, *cfg.ChainsCfg()); err != nil {
		panic(errors.Wrap(err, "failed to init token list provider"))
	}

	observer := tokensObserver{
		log:               cfg.Log(),
		chains:            cfg.ChainsCfg(),
		redisStore:        cfg.RedisStore(),
		tokenListProvider: provider,
	}

	running.WithBackOff(ctx, observer.log, "tokens_observer",
		observer.runOnce,
		cfg.TokensObserver().Period, 2*cfg.TokensObserver().Period, 5*cfg.TokensObserver().Period)

}

type tokensObserver struct {
	log               *logan.Entry
	chains            *config.ChainsConfig
	tokenListProvider TokenListProvider
	redisStore        data.RedisStore
}

type TokenListProvider interface {
	Init(ctx context.Context, chains config.ChainsConfig) error
	LiveLists(ctx context.Context, chainID int64) ([]pools.VersionedTokenList, error)
	//LastKnownList(ctx context.Context, chainID int64, name string) (*data.VersionedTokenList, error)
}

func (t *tokensObserver) runOnce(ctx context.Context) error {
	for _, c := range t.chains.Chains {
		liveLists, err := t.tokenListProvider.LiveLists(ctx, c.ID)
		if err != nil {
			return errors.Wrap(err, "failed to get live token list", logan.F{
				"chain_id": c.ID,
			})
		}

		for _, live := range liveLists {
			lastKnown, err := t.redisStore.TokenLists().Get(ctx, c.ID, live.Name)
			if err != nil {
				return errors.Wrap(err, "failed to get last known token list", logan.F{
					"chain_id": c.ID,
				})
			}

			if lastKnown == nil {
				lastKnown = &data.VersionedTokenList{
					Version: live.Version,
					Name:    live.Name,
					URI:     live.URI,
				}
			}

			if liveIsFresher := isVersionGreater(live.Version, lastKnown.Version); !liveIsFresher {
				continue
			}

			if err = t.redisStore.Tokens().Put(ctx, live.Tokens...); err != nil {
				return errors.Wrap(err, "failed to store tokens", logan.F{
					"chain_id": c.ID,
					"version":  fmt.Sprintf("%d.%d.%d", live.Version.Major, live.Version.Minor, live.Version.Patch),
				})
			}

			if err := t.redisStore.TokenLists().Put(ctx, c.ID, *lastKnown); err != nil {
				return errors.Wrap(err, "failed to store token list", logan.F{
					"chain_id": c.ID,
					"version":  fmt.Sprintf("%d.%d.%d", live.Version.Major, live.Version.Minor, live.Version.Patch),
				})
			}
		}

	}

	return nil
}

// isVersionGreater returns true if versionA is greater than versionB (compares major, minor and patch versions)
func isVersionGreater(versionA, versionB data.TokenListVersion) bool {
	return versionA.Major > versionB.Major ||
		versionA.Minor > versionB.Minor ||
		versionA.Patch > versionB.Patch
}
