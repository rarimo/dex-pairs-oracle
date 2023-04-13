package pools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/config"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"
	redisdata "gitlab.com/rarimo/dex-pairs-oracle/internal/data/redis"
)

type TokensListProvider struct {
	log      *logan.Entry
	httpc    *http.Client
	rawRedis *redis.Client
	store    data.RedisStore
}

func NewTokensListProvider(log *logan.Entry, client *http.Client, redisClient *redis.Client) *TokensListProvider {
	return &TokensListProvider{
		log:   log,
		httpc: client,
		store: redisdata.NewStore(redisClient),
	}
}

func (t *TokensListProvider) Init(ctx context.Context, chains config.ChainsConfig) error {
	for _, c := range chains.Chains {
		urls := make(map[string]struct{})

		for _, url := range c.TokensInfo.ListURL {
			urls[url.String()] = struct{}{}
		}

		storedURLs, err := t.store.TokenLists().GetURLs(ctx, c.ID)
		if err != nil {
			return errors.Wrap(err, "failed to get stored token list urls", logan.F{
				"chain_id": c.ID,
			})
		}

		for _, url := range storedURLs {
			urls[url] = struct{}{}
		}

		knownURLsSlice := make([]string, 0, len(urls))
		for url := range urls {
			knownURLsSlice = append(knownURLsSlice, url)
		}

		err = t.store.TokenLists().PutURLs(ctx, c.ID, knownURLsSlice...)
		if err != nil {
			return errors.Wrap(err, "failed to add token list urls", logan.F{
				"chain_id": c.ID,
				"urls":     knownURLsSlice,
			})
		}
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

		lists[i] = list
	}

	return lists, nil
}

func (t *TokensListProvider) LastKnownList(ctx context.Context, chainID int64, name string) (*data.VersionedTokenList, error) {
	tokenList, err := t.store.TokenLists().Get(ctx, chainID, name)
	if err != nil {
		if errors.Cause(err) == redis.Nil {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get token list", logan.F{
			"chain_id": chainID,
			"name":     name,
		})
	}

	return tokenList, nil
}

func makeChainsTokenListKey(chainID int64) string {
	return fmt.Sprintf("chain_tokens_urls:%d", chainID)
}

func makeChainsTokenListVersionKey(chainID int64) string {
	return fmt.Sprintf("chain_tokens_urls_version:%d", chainID)
}
