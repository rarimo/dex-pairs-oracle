package config

import (
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/copus"
	"gitlab.com/distributed_lab/kit/copus/types"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/kit/pgdb"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data/pg"
	"gitlab.com/rarimo/dex-pairs-oracle/pkg/rd"
)

type Config interface {
	comfig.Logger
	types.Copuser
	comfig.Listenerer
	pgdb.Databaser
	rd.Rediser

	ChainsCfg() *ChainsConfig
	NewStorage() data.Storage
	RedisStore() data.RedisStore
}

type config struct {
	comfig.Logger
	types.Copuser
	comfig.Listenerer
	pgdb.Databaser
	rd.Rediser

	chains comfig.Once

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

func (c *config) RedisStore() data.RedisStore {
	return redisdata.NewStore(c.RedisClient())
}
