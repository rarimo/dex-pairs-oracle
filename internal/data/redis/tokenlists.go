package redis

import (
	"context"

	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"

	"github.com/go-redis/redis/v8"
)

type tokenListsQ struct {
	r *redis.Client
}

func (s *tokenListsQ) Get(ctx context.Context, chainID int64, name string) (*data.VersionedTokenList, error) {
	//TODO implement me
	panic("implement me")
}

func (s *tokenListsQ) PutURLs(ctx context.Context, chainID int64, urls ...string) error {
	//TODO implement me
	panic("implement me")
}

func (s *tokenListsQ) Put(ctx context.Context, chainID int64, tokenList data.VersionedTokenList) error {
	//TODO implement me
	panic("implement me")
}

func (s *tokenListsQ) GetURLs(ctx context.Context, chainID int64) ([]string, error) {
	//TODO implement me
	panic("implement me")
}
