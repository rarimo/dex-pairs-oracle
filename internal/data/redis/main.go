package redis

import (
	"github.com/go-redis/redis/v8"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"
)

type Store struct {
	raw *redis.Client
}

func (s *Store) Tokens() data.TokensQ {
	return &tokensQ{r: s.raw}
}

func (s *Store) TokenLists() data.TokenListsQ {
	return &tokenListsQ{r: s.raw}
}

func NewStore(raw *redis.Client) *Store {
	return &Store{raw: raw}
}
