package data

import "context"

//go:generate xo schema "postgres://postgres:postgres@localhost:5432/rarimo_dex_oracle?sslmode=disable" -o ./ --single=schema.xo.go --src templates
//go:generate xo schema "postgres://postgres:postgres@localhost:5432/rarimo_dex_oracle?sslmode=disable" -o pg --single=schema.xo.go --src=pg/templates --go-context=both
//go:generate goimports -w ./

type Storage interface {
	Transaction(func() error) error

	BalanceQ() BalanceQ
}

type BalanceQ interface {
	SelectCtx(ctx context.Context, selector BalancesSelector) ([]Balance, error)
	InsertBatchCtx(ctx context.Context, balances ...Balance) error
	UpsertBatchCtx(ctx context.Context, balances ...Balance) error
}

type GorpMigrationQ interface{}

type RedisStore interface {
	Tokens() TokensQ
	TokenLists() TokenListsQ
}

type TokensQ interface {
	Get(ctx context.Context, address string, chainID int64) (*Token, error)
	All(ctx context.Context, chain int64) ([]Token, error)
	Put(ctx context.Context, tokens ...Token) error
}

type TokenListsQ interface {
	GetURLs(ctx context.Context, chainID int64) ([]string, error)
	Get(ctx context.Context, chainID int64, name string) (*VersionedTokenList, error)

	PutURLs(ctx context.Context, chainID int64, urls ...string) error
	Put(ctx context.Context, chainID int64, tokenList VersionedTokenList) error
}
