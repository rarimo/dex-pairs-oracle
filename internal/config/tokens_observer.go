package config

import (
	"time"

	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"

	"gitlab.com/distributed_lab/figure/v3"
)

type TokensObserverConfig struct {
	Period time.Duration `fig:"period"`
}

func (c *config) TokensObserver() *TokensObserverConfig {
	return c.tokensObserver.Do(func() interface{} {
		var cfg TokensObserverConfig

		yamlName := "tokens_observer"

		err := figure.
			Out(&cfg).
			From(kv.MustGetStringMap(c.getter, yamlName)).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out "+yamlName))
		}

		return &cfg
	}).(*TokensObserverConfig)
}
