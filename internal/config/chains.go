package config

import (
	"gitlab.com/rarimo/dex-pairs-oracle/internal/chains"

	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (c *config) ChainsCfg() *chains.Config {
	return c.chains.Do(func() interface{} {
		var cfg chains.Config

		cfgName := "chains"

		err := figure.
			Out(&cfg).
			With(figure.BaseHooks, figure.EthereumHooks, chains.Hooks).
			From(kv.MustGetStringMap(c.getter, cfgName)).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out "+cfgName))
		}

		return &cfg
	}).(*chains.Config)
}
