package api

import (
	"context"
	"time"

	"github.com/go-chi/chi/v5"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/config"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/service/api/handlers"
)

func Run(ctx context.Context, cfg config.Config) {
	r := chi.NewRouter()

	r.Use(
		ape.RecoverMiddleware(cfg.Log()),
		ape.LoganMiddleware(cfg.Log()),
		ape.CtxMiddleware(
			handlers.CtxLog(cfg.Log()),
			handlers.CtxConfig(cfg),
		),
	)
	r.Route("/dexoracle", func(r chi.Router) {
		//r.Get("/tokens", handlers.ListSupportedToken)
		r.Get("/chains", handlers.ListSupportedChain)
		r.Route("/chains", func(r chi.Router) {
			r.Get("/", handlers.ListSupportedChain)
			r.Route("/{chain-id}", func(r chi.Router) {
				//r.Get("/", handlers.GetChain)
				r.Get("/{account_address}/balances", handlers.ListBalances)
			})
		})
	})

	ape.Serve(ctx, r, cfg, ape.ServeOpts{
		ShutdownTimeout: 20 * time.Second,
	})
}
