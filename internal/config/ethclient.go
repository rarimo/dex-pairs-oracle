package config

import (
	"gitlab.com/distributed_lab/kit/kv"

	"github.com/ethereum/go-ethereum/ethclient"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type evmConfig struct {
	RPCClient *ethclient.Client `fig:"rpc,required"`
}

func (c *config) EVM() *evmConfig {
	return c.evm.Do(func() interface{} {
		var evm evmConfig

		err := figure.
			Out(&evm).
			From(kv.MustGetStringMap(c.getter, "evm")).
			With(figure.BaseHooks, figure.EthereumHooks).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out evm"))
		}

		return &evm
	}).(*evmConfig)
}
