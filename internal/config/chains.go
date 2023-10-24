package config

import (
	"github.com/rarimo/dex-pairs-oracle/internal/chains"
	"github.com/rarimo/dex-pairs-oracle/pkg/ethamounts"
	"gitlab.com/distributed_lab/logan/v3"

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

		if err := validateChains(cfg); err != nil {
			panic(errors.Wrap(err, "failed to validate "+cfgName))
		}

		for i := 0; i < len(cfg.Chains); i++ {
			cfg.Chains[i].BalanceProvider, err = ethamounts.NewMultiProvider(cfg.Chains[i].RPCUrl)
			if err != nil {
				panic(errors.Wrap(err, "failed to create balance provider", logan.F{
					"chain": cfg.Chains[i].Name,
				}))
			}
		}

		return &cfg
	}).(*chains.Config)
}

func validateChains(cfg chains.Config) error {
chainsLoop:
	for _, chain := range cfg.Chains {
		if chain.NativeSymbol == "" {
			return errors.From(errors.New("native symbol is required"), logan.F{
				"chain": chain.Name,
			})
		}

		for _, tokenInfo := range chain.TokensInfo.Tokens {
			if tokenInfo.Symbol == chain.NativeSymbol {
				continue chainsLoop // native token is configured
			}
		}

		return errors.From(errors.New("native token is not configured in tokens"), logan.F{
			"chain": chain.Name,
		})
	}

	return nil
}
