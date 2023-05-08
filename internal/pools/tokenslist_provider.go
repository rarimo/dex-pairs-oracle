package pools

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"gitlab.com/rarimo/dex-pairs-oracle/internal/chains"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/config"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"
	redisdata "gitlab.com/rarimo/dex-pairs-oracle/internal/data/redis"
)

type TokensListProvider struct {
	log   *logan.Entry
	httpc *http.Client
	store data.RedisStore
}

func NewTokensListProvider(ctx context.Context, cfg config.Config) (*TokensListProvider, error) {
	p := TokensListProvider{
		log:   cfg.Log(),
		httpc: http.DefaultClient,
		store: cfg.RedisStore(),
	}

	return &p, p.init(ctx, cfg.ChainsCfg())
}

func (t *TokensListProvider) init(ctx context.Context, chains *chains.Config) error {
	for _, c := range chains.Chains {
		configURLs := make(map[string]data.TokenListVersion)

		for _, url := range c.TokensInfo.ListURL {
			configURLs[url.String()] = data.TokenListVersion{}
		}

		storedURLs, err := t.store.TokenLists().GetURLs(ctx, c.ID)
		if err != nil {
			return errors.Wrap(err, "failed to get stored token list urls", logan.F{
				"chain_id": c.ID,
			})
		}

		deleteURLs := make([]string, 0)

		for _, storedURL := range storedURLs {
			if _, ok := configURLs[storedURL]; !ok {
				deleteURLs = append(deleteURLs, storedURL)
				continue
			}

			version, err := t.store.TokenLists().GetVersion(ctx, storedURL)
			if err != nil {
				return errors.Wrap(err, "failed to get stored token list version", logan.F{
					"chain_id": c.ID,
					"url":      storedURL,
				})
			}

			if version != nil {
				configURLs[storedURL] = *version
			}
		}

		running.UntilSuccess(ctx, t.log, "init", func(ctx context.Context) (bool, error) {
			err = t.store.TokenLists().DeleteURLs(ctx, c.ID, deleteURLs...)
			if err != nil {
				if err == redisdata.ErrTxFailed {
					return false, nil
				}

				return false, errors.Wrap(err, "failed to delete token list urls", logan.F{
					"chain_id": c.ID,
					"urls":     deleteURLs,
				})
			}

			err = t.store.TokenLists().PutURLs(ctx, c.ID, configURLs)
			if err != nil {
				if err == redisdata.ErrTxFailed {
					return false, nil
				}

				return false, errors.Wrap(err, "failed to add token list urls", logan.F{
					"chain_id": c.ID,
					"urls":     configURLs,
				})
			}

			return true, nil
		}, 1*time.Second, 1*time.Second)
	}

	return nil
}

func (t *TokensListProvider) LiveLists(ctx context.Context, chainID int64) ([]VersionedTokenList, error) {
	urls, err := t.store.TokenLists().GetURLs(ctx, chainID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get token list urls", logan.F{
			"chain_id": chainID,
		})
	}

	lists := make([]VersionedTokenList, len(urls))

	for i, url := range urls {
		resp, err := t.httpc.Get(url)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get token list", logan.F{
				"chain_id": chainID,
				"url":      url,
			})
		}

		var list VersionedTokenList
		if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
			return nil, errors.Wrap(err, "failed to decode token list", logan.F{
				"chain_id": chainID,
				"url":      url,
			})
		}

		list.URI = url
		lists[i] = list
	}

	return lists, nil
}
