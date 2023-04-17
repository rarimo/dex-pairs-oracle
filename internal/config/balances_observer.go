package config

import (
	"time"

	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type BalancesObserverConfig struct {
	PageSize uint64        `fig:"page_size"`
	Interval time.Duration `fig:"interval"`
}

func (c *config) BalancesObserver() *BalancesObserverConfig {
	return c.balancesObserver.Do(func() interface{} {
		balancesObserver := BalancesObserverConfig{
			PageSize: 100,
			Interval: 1 * time.Minute,
		}

		yamlName := "balances_observer"

		err := figure.
			Out(&balancesObserver).
			From(kv.MustGetStringMap(c.getter, yamlName)).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out "+yamlName))
		}

		return &balancesObserver
	}).(*BalancesObserverConfig)
}
