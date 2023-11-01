package config

import (
	"github.com/rarimo/dex-pairs-oracle/internal/chains"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/logan/v3"

	"github.com/ethereum/go-ethereum/ethclient"
	ethamountsbind "github.com/rarimo/dex-pairs-oracle/pkg/ethamounts/bind"
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
			chainEthClient, err := ethclient.Dial(cfg.Chains[i].RPCUrl.String())
			if err != nil {
				panic(errors.Wrap(err, "failed to dial rpc", logan.F{
					"chain": cfg.Chains[i].Name,
				}))
			}

			cfg.Chains[i].BalanceProvider, err = ethamountsbind.NewMultiBalanceGetter(cfg.Chains[i].MultiBalanceGetterContractAddr, chainEthClient)
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
