package services

import (
	"context"
	"fmt"

	"gitlab.com/rarimo/dex-pairs-oracle/internal/chains"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/config"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/pools"
)

func RunTokensObserver(ctx context.Context, cfg config.Config) {
	provider, err := pools.NewTokensListProvider(ctx, cfg)
	if err != nil {
		panic(errors.Wrap(err, "failed to create tokens list provider"))
	}

	observer := tokensObserver{
		log:               cfg.Log(),
		chains:            cfg.ChainsCfg(),
		redisStore:        cfg.RedisStore(),
		tokenListProvider: provider,
	}

	running.WithBackOff(ctx, observer.log, "tokens_observer",
		observer.runOnce,
		cfg.TokensObserver().Interval, 2*cfg.TokensObserver().Interval, 5*cfg.TokensObserver().Interval)

}

type tokensObserver struct {
	log               *logan.Entry
	chains            *chains.Config
	tokenListProvider TokenListProvider
	redisStore        data.RedisStore
}

type TokenListProvider interface {
	LiveLists(ctx context.Context, chainID int64) ([]pools.VersionedTokenList, error)
}

func (t *tokensObserver) runOnce(ctx context.Context) error {
	for _, c := range t.chains.Chains {
		preConfiguredTokens := make([]chains.TokenInfo, 0, len(c.TokensInfo.Tokens))
		for _, token := range c.TokensInfo.Tokens {
			existing, err := t.redisStore.Tokens().Get(ctx, token.Address, c.ID)
			if err != nil {
				t.log.WithError(err).WithFields(logan.F{
					"address":  token.Address,
					"chain_id": c.ID,
				}).Error("failed to get token from redis")
			}

			if existing != nil {
				continue
			}

			preConfiguredTokens = append(preConfiguredTokens, token)
		}

		if len(preConfiguredTokens) != 0 {
			err := t.redisStore.Tokens().Put(ctx, c.ID, preConfiguredTokens...)
			if err != nil {
				return errors.Wrap(err, "failed to store pre-configured tokens", logan.F{
					"chain_id": c.ID,
				})
			}
		}
		liveLists, err := t.tokenListProvider.LiveLists(ctx, c.ID)
		if err != nil {
			return errors.Wrap(err, "failed to get live token list", logan.F{
				"chain_id": c.ID,
			})
		}

		for _, live := range liveLists {
			lastKnownVersion, err := t.redisStore.TokenLists().GetVersion(ctx, live.URI, c.ID)
			if err != nil {
				return errors.Wrap(err, "failed to get last known token list", logan.F{
					"chain_id": c.ID,
				})
			}

			if lastKnownVersion != nil {
				if liveIsFresher := isVersionGreater(live.Version, *lastKnownVersion); !liveIsFresher {
					continue
				}
			}

			// in case lastKnownVersion is nil, we need store tokens and version

			newTokens := make([]chains.TokenInfo, 0, len(live.Tokens))

			for _, token := range live.Tokens {
				if token.ChainID != c.ID {
					continue
				}

				stored, err := t.redisStore.Tokens().Get(ctx, token.Address, token.ChainID)
				if err != nil {
					return errors.Wrap(err, "failed to get token", logan.F{
						"chain_id": c.ID,
						"address":  token.Address,
					})
				}

				if stored == nil || stored.Name != token.Name || stored.LogoURI != token.LogoURI {
					newTokens = append(newTokens, token.TokenInfo)
				}
			}

			if err = t.redisStore.Tokens().Put(ctx, c.ID, newTokens[:]...); err != nil {
				return errors.Wrap(err, "failed to store tokens", logan.F{
					"chain_id": c.ID,
					"version":  fmt.Sprintf("%d.%d.%d", live.Version.Major, live.Version.Minor, live.Version.Patch),
				})
			}

			err = t.redisStore.TokenLists().PutURLs(ctx, c.ID, map[string]data.TokenListVersion{
				live.URI: live.Version,
			})

			if err != nil {
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
