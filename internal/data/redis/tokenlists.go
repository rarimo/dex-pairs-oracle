package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"

	"github.com/go-redis/redis/v8"
)

type tokenListsQ struct {
	r *redis.Client
}

var (
	ErrTxFailed = errors.New("tx failed")
)

func (s *tokenListsQ) GetVersion(ctx context.Context, url string, chainID int64) (*data.TokenListVersion, error) {
	rawVersion, err := s.r.Get(ctx, makeChainsTokenListVersionKey(url, chainID)).Result()
	if err != nil {
		if errors.Cause(err) == redis.Nil {
			return nil, nil
		}

		return nil, errors.Wrap(err, "failed to get token list version", logan.F{
			"url": url,
		})
	}

	var version data.TokenListVersion
	if err := json.Unmarshal([]byte(rawVersion), &version); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal token list version", logan.F{
			"raw_version": rawVersion,
		})
	}

	return &version, nil
}

func (s *tokenListsQ) GetURLs(ctx context.Context, chainID int64) ([]string, error) {
	return s.r.SMembers(ctx, makeChainsTokenListURLsKey(chainID)).Result()
}

func (s *tokenListsQ) DeleteURLs(ctx context.Context, chainID int64, urls ...string) error {
	if len(urls) == 0 {
		return nil
	}

	urlI := make([]interface{}, 0, len(urls))
	urlVersionKeys := make([]string, 0, len(urls))

	watch := make([]string, 0, 1+len(urls))

	chainURLsListKey := makeChainsTokenListURLsKey(chainID)
	watch = append(watch, chainURLsListKey)

	for _, url := range urls {
		urlI = append(urlI, url)
		urlVersionKeys = append(urlVersionKeys, makeChainsTokenListVersionKey(url, chainID))
		watch = append(watch, makeChainsTokenListVersionKey(url, chainID))
	}

	err := s.r.Watch(ctx, func(tx *redis.Tx) error {
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.SRem(ctx, chainURLsListKey, urlI[:]...)

			pipe.Del(ctx, urlVersionKeys...)

			return nil
		})

		return err
	}, watch...)

	if err != nil {
		if err == redis.TxFailedErr {
			return ErrTxFailed
		}

		return errors.Wrap(err, "failed to delete exec tx", logan.F{
			"urls": urls,
		})
	}

	return nil
}

func (s *tokenListsQ) PutURLs(ctx context.Context, chainID int64, urlVersions map[string]data.TokenListVersion) error {
	if len(urlVersions) == 0 {
		return nil
	}

	newURLs := make([]interface{}, 0, len(urlVersions))

	chainTokensListURLKey := makeChainsTokenListURLsKey(chainID)
	watch := make([]string, 0, 1+len(urlVersions))
	watch = append(watch, chainTokensListURLKey)

	for url := range urlVersions {
		newURLs = append(newURLs, url)
		watch = append(watch, url)
	}

	rawVersions := make([]interface{}, 0, 2*len(urlVersions))
	for url, version := range urlVersions {
		rawVersion, err := json.Marshal(version)
		if err != nil {
			return errors.Wrap(err, "failed to marshal token list version", logan.F{
				"version": version,
			})
		}

		rawVersions = append(rawVersions,
			makeChainsTokenListVersionKey(url, chainID),
			string(rawVersion))
	}

	err := s.r.Watch(ctx, func(tx *redis.Tx) error {
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.SAdd(ctx, chainTokensListURLKey, newURLs...)
			pipe.MSet(ctx, rawVersions...)
			return nil
		})

		return err
	}, watch...)

	if err != nil {
		if err == redis.TxFailedErr {
			return ErrTxFailed
		}

		return errors.Wrap(err, "failed to put token list urls", logan.F{
			"urls":     urlVersions,
			"chain_id": chainID,
		})
	}

	return nil
}

func makeChainsTokenListURLsKey(chainID int64) string {
	return fmt.Sprintf("chain_tokens_list_urls:%d", chainID)
}

func makeChainsTokenListVersionKey(url string, chainID int64) string {
	return fmt.Sprintf("version:%d:%s", chainID, url)
}
