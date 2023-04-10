package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gitlab.com/rarimo/dex-pairs-oracle/pkg/ethbalances"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi/v5"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/kit/pgdb"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/urlval"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/config"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"
	"gitlab.com/rarimo/dex-pairs-oracle/resources"
)

type listBalancesRequest struct {
	ChainID        int64
	AccountAddress string

	TokenAddress *string `filter:"token_address"`

	IncludeChain bool `include:"chain"`
	IncludeToken bool `include:"token"`

	PageCursor uint64     `page:"cursor"`
	PageLimit  uint64     `page:"limit" default:"15"`
	Sorts      pgdb.Sorts `url:"sort" default:"-amount"`
}

func newListBalancesAddress(r *http.Request) (*listBalancesRequest, error) {
	chainID, err := strconv.ParseInt(chi.URLParam(r, "chain_id"), 10, 64)
	if err != nil {
		return nil, validation.Errors{
			"chain_id": err,
		}
	}

	accountAddress := chi.URLParam(r, "account_address")

	req := listBalancesRequest{
		ChainID:        chainID,
		AccountAddress: accountAddress,
	}

	if err := urlval.Decode(r.URL.Query(), &req); err != nil {
		return nil, err
	}

	if req.TokenAddress != nil {
		chain, _, err := parseAddress(r, *req.TokenAddress)
		if err != nil {
			return nil, validation.Errors{
				"filter[token_address]": err,
			}
		}

		if chain.ID != req.ChainID {
			return nil, validation.Errors{
				"filter[token_address]": errors.New("chain id should not differ from filter[chain_id]"),
			}
		}
	}

	return &req, nil
}

type ChainBalancesProvider interface {
	GetBalances(ctx context.Context, address string, chainID int64) ([]data.Balance, error)
}

func ListBalances(w http.ResponseWriter, r *http.Request) {
	req, err := newListBalancesAddress(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	balances, err := Storage(r).BalanceQ().SelectCtx(r.Context(), data.BalancesSelector{
		TokenAddress:   req.TokenAddress,
		AccountAddress: &req.AccountAddress,
		ChainID:        &req.ChainID,

		PageCursor: req.PageCursor,
		PageSize:   req.PageLimit,
		Sort:       req.Sorts,
	})
	if err != nil {
		Log(r).WithError(err).Error("failed to select balances")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	resp := resources.BalanceListResponse{
		Data:     make([]resources.Balance, len(balances)),
		Included: resources.Included{},
		Links: &resources.Links{
			Self: fmt.Sprintf("%s?%s", r.URL.Path, urlval.MustEncode(req)),
		},
	}

	chain := Config(r).ChainsCfg().Find(req.ChainID)

	if len(balances) != 0 {
		// returning cached balances in case any exist

		for i, balance := range balances {
			resp.Data[i] = balanceToResource(balance, *Config(r).ChainsCfg().Find(balance.ChainID))
			if req.IncludeToken {
				if err := includeToken(r, &resp, balance, *chain); err != nil {
					Log(r).WithError(err).Error("failed to include token")
					ape.RenderErr(w, problems.InternalError())
					return
				}
			}
		}

		if req.IncludeChain {
			chainR := chainToResource(*chain)
			resp.Included.Add(&chainR)
		}

		req.PageCursor = uint64(balances[len(balances)-1].ID)
		resp.Links.Next = fmt.Sprintf("%s?%s", r.URL.Path, urlval.MustEncode(req))

		ape.Render(w, resp)

		return
	}

	chainBalances, err := BalancesProvider(r).GetBalances(r.Context(), req.AccountAddress, req.ChainID)
	if err != nil {
		if errors.Cause(err) == ethbalances.ErrChainNotSupported {
			ape.RenderErr(w, problems.BadRequest(validation.Errors{
				"chain_id": err,
			})...)
			return
		}

		Log(r).WithError(err).WithFields(logan.F{
			"account_address": req.AccountAddress,
			"chain_id":        req.ChainID,
			"token":           req.TokenAddress,
		}).Error("failed to get balances from chain")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if len(chainBalances) == 0 {
		ape.Render(w, resp)
		return
	}

	err = Config(r).NewStorage().BalanceQ().InsertBatchCtx(r.Context(), chainBalances...)
	if err != nil {
		Log(r).WithError(err).Error("failed to insert balances")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	balances, err = Storage(r).BalanceQ().SelectCtx(r.Context(), data.BalancesSelector{
		TokenAddress:   req.TokenAddress,
		AccountAddress: &req.AccountAddress,
		ChainID:        &req.ChainID,

		PageCursor: req.PageCursor,
		PageSize:   req.PageLimit,
		Sort:       req.Sorts,
	})
	if err != nil {
		Log(r).WithError(err).Error("failed to select balances")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if req.IncludeChain {
		chainR := chainToResource(*chain)
		resp.Included.Add(&chainR)
	}

	for i, balance := range balances {
		resp.Data[i] = balanceToResource(balance, *Config(r).ChainsCfg().Find(balance.ChainID))

		if req.IncludeToken {
			if err := includeToken(r, &resp, balance, *chain); err != nil {
				Log(r).WithError(err).Error("failed to include token")
				ape.RenderErr(w, problems.InternalError())
				return
			}
		}
	}

	req.PageCursor = uint64(balances[len(balances)-1].ID)
	resp.Links.Next = fmt.Sprintf("%s?%s", r.URL.Path, urlval.MustEncode(req))

	ape.Render(w, resp)
}

func balanceToResource(balance data.Balance, chain config.Chain) resources.Balance {
	return resources.Balance{
		Key: resources.Key{
			ID:   strconv.FormatInt(balance.ID, 10),
			Type: resources.BALANCES,
		},
		Attributes: resources.BalanceAttributes{
			Amount: balance.Amount.String(),
		},
		Relationships: resources.BalanceRelationships{
			Chain: resources.Relation{
				Data: &resources.Key{
					ID:   strconv.FormatInt(balance.ChainID, 10),
					Type: resources.CHAINS,
				},
			},
			Owner: resources.Relation{
				Data: &resources.Key{
					ID:   fmt.Sprintf("%s:%s", chain.Name, hexutil.Encode(balance.AccountAddress)),
					Type: resources.ACCOUNTS,
				},
			},
			Token: resources.Relation{
				Data: &resources.Key{
					ID:   fmt.Sprintf("%s:%s", chain.Name, hexutil.Encode(balance.Token)),
					Type: resources.TOKENS,
				},
			},
		},
	}
}

func includeToken(r *http.Request, resp *resources.BalanceListResponse, balance data.Balance, chain config.Chain) error {
	t, err := Config(r).RedisStore().Tokens().Get(
		r.Context(),
		hexutil.Encode(balance.Token),
		balance.ChainID)
	if err != nil {
		return errors.Wrap(err, "failed to get token from redis")
	}

	if t == nil {
		Log(r).WithFields(logan.F{
			"token_address": hexutil.Encode(balance.Token),
			"chain_id":      balance.ChainID,
		}).Warn("token not found in redis")
		return nil
	}

	resp.Included.Add(&resources.Token{
		Key: resources.Key{
			ID:   fmt.Sprintf("%s:%s", chain.Name, t.Address),
			Type: resources.TOKENS,
		},
		Attributes: resources.TokenAttributes{
			Decimals: t.Decimals,
			LogoUri:  t.LogoURI,
			Name:     t.Name,
			Symbol:   t.Symbol,
		},
	})

	return nil
}

func parseAddress(r *http.Request, raw string) (*config.Chain, string, error) {
	split := strings.Split(raw, ":")

	if len(split) != 2 {
		return nil, "", errors.New("invalid address")
	}

	chainID, err := strconv.ParseInt(split[0], 10, 64)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to parse chain id")
	}

	chain := Config(r).ChainsCfg().Find(chainID)
	if chain == nil {
		return nil, "", errors.New("unsupported chain")
	}

	return chain, split[1], nil
}
