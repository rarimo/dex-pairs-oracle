package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"gitlab.com/rarimo/dex-pairs-oracle/internal/services"

	"gitlab.com/distributed_lab/logan/v3/errors"

	"github.com/alecthomas/kingpin"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/config"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/services/api"
)

func Run(args []string) {
	log := logan.New()

	defer func() {
		if rvr := recover(); rvr != nil {
			log.WithRecover(rvr).Error("app panicked")
		}
	}()

	cfg := config.New(kv.MustFromEnv())
	log = cfg.Log()

	app := kingpin.New("dex-pairs-oracle", "")

	runCmd := app.Command("run", "run command")
	apiCmd := runCmd.Command("api", "run api")
	balancesObserverCmd := runCmd.Command("balances_observer", "run balances observer")

	allCmd := app.Command("all", "run all services")

	migrateCmd := app.Command("migrate", "migrate command")
	migrateUpCmd := migrateCmd.Command("up", "migrate db up")
	migrateDownCmd := migrateCmd.Command("down", "migrate db down")

	// custom commands go here...

	cmd, err := app.Parse(args[1:])
	if err != nil {
		panic(errors.Wrap(err, "failed to parse args"))
	}

	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	run := func(f func(context.Context, config.Config)) {
		wg.Add(1)

		go func() {
			defer wg.Done()
			defer func() {
				if rvr := recover(); rvr != nil {
					logan.New().WithRecover(rvr).Error("service panicked")
				}
			}()

			f(ctx, cfg)
		}()
	}

	switch cmd {
	case apiCmd.FullCommand():
		cfg.Log().Info("starting API")
		run(api.Run)
	case balancesObserverCmd.FullCommand():
		cfg.Log().Info("starting balances observer")
		run(services.RunBalancesObserver)
	case allCmd.FullCommand():
		cfg.Log().Info("starting all services")
		run(api.Run)
		run(services.RunBalancesObserver)
	case migrateUpCmd.FullCommand():
		if err := MigrateUp(cfg); err != nil {
			panic(errors.Wrap(err, "failed to migrate up"))
		}
	case migrateDownCmd.FullCommand():
		if err := MigrateDown(cfg); err != nil {
			panic(errors.Wrap(err, "failed to migrate down"))
		}
	default:
		panic(fmt.Errorf("unknown command %s", cmd))
	}

	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT)

	wgch := make(chan struct{})
	go func() {
		wg.Wait()
		close(wgch)
	}()

	select {
	case <-wgch:
		cfg.Log().Warn("all services stopped")
	case <-gracefulStop:
		cfg.Log().Info("received signal to stop")
		cancel()
		<-wgch
	}
}
