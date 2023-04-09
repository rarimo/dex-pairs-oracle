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
	InsertCtx(ctx context.Context, balance *Balance) error
}

type GorpMigrationQ interface{}
