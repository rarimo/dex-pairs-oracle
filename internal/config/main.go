package config

import (
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/copus"
	"gitlab.com/distributed_lab/kit/copus/types"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/kit/pgdb"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/chains"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data/pg"
	redisdata "gitlab.com/rarimo/dex-pairs-oracle/internal/data/redis"
	"gitlab.com/rarimo/dex-pairs-oracle/pkg/ethamounts"
	"gitlab.com/rarimo/dex-pairs-oracle/pkg/ethbalances"
	"gitlab.com/rarimo/dex-pairs-oracle/pkg/rd"
)

type Config interface {
	comfig.Logger
	types.Copuser
	comfig.Listenerer
	pgdb.Databaser
	rd.Rediser

	ChainsCfg() *chains.Config
	NewStorage() data.Storage
	Storage() data.Storage
	RedisStore() data.RedisStore
	EVM() *evmConfig
	BalancesObserver() *BalancesObserverConfig
	TokensObserver() *TokensObserverConfig
	EthBalancesProvider() *ethbalances.Provider
}

type config struct {
	comfig.Logger
	types.Copuser
	comfig.Listenerer
	pgdb.Databaser
	rd.Rediser

	chains              comfig.Once
	evm                 comfig.Once
	balancesObserver    comfig.Once
	tokensObserver      comfig.Once
	ethBalancesProvider comfig.Once

	getter kv.Getter
}

func New(getter kv.Getter) Config {
	return &config{
		getter:     getter,
		Databaser:  pgdb.NewDatabaser(getter),
		Copuser:    copus.NewCopuser(getter),
		Listenerer: comfig.NewListenerer(getter),
		Logger:     comfig.NewLogger(getter, comfig.LoggerOpts{}),
		Rediser:    rd.NewRediser(getter),
	}
}

func (c *config) NewStorage() data.Storage {
	return pg.New(c.DB().Clone())
}

func (c *config) Storage() data.Storage {
	return pg.New(c.DB())
}

func (c *config) RedisStore() data.RedisStore {
	return redisdata.NewStore(c.RedisClient())
}

func (c *config) EthBalancesProvider() *ethbalances.Provider {
	return c.ethBalancesProvider.Do(func() interface{} {
		return ethbalances.NewProvider(c.RedisStore(), ethamounts.NewProvider(c.ChainsCfg()))
	}).(*ethbalances.Provider)
}
