package handlers

import (
	"context"
	"net/http"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/config"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"
)

type ctxKey int

const (
	logCtxKey ctxKey = iota
	configCtxKey
	storageCtxKey
	balancesProviderCtxKey
)

func CtxLog(entry *logan.Entry) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, logCtxKey, entry)
	}
}

func Log(r *http.Request) *logan.Entry {
	return r.Context().Value(logCtxKey).(*logan.Entry)
}

func CtxConfig(cfg config.Config) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, configCtxKey, cfg)
	}
}

func Config(r *http.Request) config.Config {
	return r.Context().Value(configCtxKey).(config.Config)
}

func CtxStorage(storage data.Storage) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, storageCtxKey, storage)
	}
}

func Storage(r *http.Request) data.Storage {
	return r.Context().Value(storageCtxKey).(data.Storage)
}

func CtxBalancesProvider(provider ChainBalancesProvider) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, balancesProviderCtxKey, provider)
	}
}

func BalancesProvider(r *http.Request) ChainBalancesProvider {
	return r.Context().Value(balancesProviderCtxKey).(ChainBalancesProvider)
}
