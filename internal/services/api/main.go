package api

import (
	"context"
	"time"

	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/config"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/services/api/handlers"
)

func Run(ctx context.Context, cfg config.Config) {
	r := chi.NewRouter()

	const slowRequestDurationThreshold = time.Second
	ape.DefaultMiddlewares(r, cfg.Log(), slowRequestDurationThreshold)

	r.Use(
		ape.RecoverMiddleware(cfg.Log()),
		ape.LoganMiddleware(cfg.Log()),
		ape.CtxMiddleware(
			handlers.CtxLog(cfg.Log()),
			handlers.CtxConfig(cfg),
			handlers.CtxBalancesProvider(cfg.EthBalancesProvider()),
		),
	)
	r.Route("/dexoracle", func(r chi.Router) {
		r.Route("/chains", func(r chi.Router) {
			r.Get("/", handlers.ListSupportedChain)
			r.Route("/evm", func(r chi.Router) {
				r.Route("/{chain_id}", func(r chi.Router) {
					r.Get("/tokens", handlers.ListSupportedEVMTokens)
					r.Get("/{account_address}/balances", handlers.ListEVMBalances)
				})
			})

		})
	})

	ape.Serve(ctx, r, cfg, ape.ServeOpts{})
}
