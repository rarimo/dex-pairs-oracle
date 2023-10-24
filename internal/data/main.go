package data

import (
	"context"

	"github.com/rarimo/dex-pairs-oracle/internal/chains"
)

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
	Get(ctx context.Context, address string, chainID int64) (*chains.TokenInfo, error)
	All(ctx context.Context, chain int64) ([]chains.TokenInfo, error)
	Page(ctx context.Context, chainID int64, cursor string, limit int64) ([]chains.TokenInfo, error)
	Put(ctx context.Context, chainID int64, tokens ...chains.TokenInfo) error
}

type TokenListsQ interface {
	GetVersion(ctx context.Context, url string, chainID int64) (*TokenListVersion, error)
	GetURLs(ctx context.Context, chainID int64) ([]string, error)
	PutURLs(ctx context.Context, chainID int64, urlVersions map[string]TokenListVersion) error
	DeleteURLs(ctx context.Context, chainID int64, urls ...string) error
}
