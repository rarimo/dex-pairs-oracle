package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/kit/pgdb"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/urlval"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/chains"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"
	"gitlab.com/rarimo/dex-pairs-oracle/pkg/etherrors"
	"gitlab.com/rarimo/dex-pairs-oracle/resources"
)

type listBalancesRequest struct {
	ChainID        int64
	AccountAddress string

	IncludeChain bool `include:"chain"`
	IncludeToken bool `include:"token"`

	RawCursor   string `page:"cursor" default:"0x0000000000000000000000000000000000000000"`
	TokenCursor []byte

	PageLimit int64      `page:"limit" default:"15"`
	Sorts     pgdb.Sorts `url:"sort" default:"token"`
}

func newListEvmBalancesAddress(r *http.Request) (*listBalancesRequest, error) {
	chainID, err := strconv.ParseInt(chi.URLParam(r, "chain_id"), 10, 64)
	if err != nil {
		return nil, validation.Errors{
			"chain_id": err,
		}
	}

	if supported := Config(r).ChainsCfg().Find(chainID); supported == nil {
		return nil, validation.Errors{
			"chain_id": fmt.Errorf("chain %d is not supported", chainID),
		}
	}

	req := listBalancesRequest{
		ChainID:        chainID,
		AccountAddress: chi.URLParam(r, "account_address"),
	}

	if err := urlval.Decode(r.URL.Query(), &req); err != nil {
		return nil, err
	}

	req.TokenCursor = hexutil.MustDecode(req.RawCursor)

	return &req, nil
}

type ChainBalancesProvider interface {
	GetBalances(ctx context.Context, address string, chainID int64, cursor string, limit int64) ([]data.Balance, error)
}

func ListEVMBalances(w http.ResponseWriter, r *http.Request) {
	req, err := newListEvmBalancesAddress(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	balances, err := Config(r).Storage().BalanceQ().SelectCtx(r.Context(), data.BalancesSelector{
		AccountAddress: &req.AccountAddress,
		ChainID:        &req.ChainID,
		TokenCursor:    req.TokenCursor,
		PageSize:       uint64(req.PageLimit),
		Sort:           req.Sorts,
	})
	if err != nil {
		Log(r).WithError(err).Error("failed to select balances")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	resp := resources.BalanceListResponse{
		Data:     make([]resources.Balance, 0, len(balances)),
		Included: resources.Included{},
		Links: &resources.Links{
			Self: fmt.Sprintf("%s?%s", r.URL.Path, urlval.MustEncode(req)),
		},
	}

	tokenCursor := fmt.Sprintf("token:%d:%s", req.ChainID, hexutil.Encode(req.TokenCursor))

	if len(balances) < int(req.PageLimit) {
		limit := req.PageLimit - int64(len(balances))

		if len(balances) != 0 {
			tokenCursor = fmt.Sprintf("token:%d:%s", req.ChainID, string(balances[len(balances)-1].Token))
		}

		additionalBalances, err := fetchAndSaveBalances(r, *req, tokenCursor, limit)
		if err != nil {
			if cerr := errors.Cause(err); cerr == etherrors.ErrChainNotSupported {
				ape.RenderErr(w, problems.BadRequest(validation.Errors{
					"chain_id": err,
				})...)
				return
			}

			Log(r).WithError(err).WithFields(logan.F{
				"account_address": req.AccountAddress,
				"chain_id":        req.ChainID,
			}).Error("failed to get balances from chain")
			ape.RenderErr(w, problems.InternalError())
			return
		}

		balances = append(balances, additionalBalances...)
	}

	chain := Config(r).ChainsCfg().Find(req.ChainID)

	if req.IncludeChain {
		chainR := chainToResource(*chain)
		resp.Included.Add(&chainR)
	}

	for _, balance := range balances {
		resp.Data = append(resp.Data,
			balanceToResource(balance, *Config(r).ChainsCfg().Find(balance.ChainID)))

		if req.IncludeToken {
			if err := includeToken(r, &resp, balance, *chain); err != nil {
				Log(r).WithError(err).Error("failed to include token")
				ape.RenderErr(w, problems.InternalError())
				return
			}
		}
	}

	req.RawCursor = hexutil.Encode(balances[len(balances)-1].Token)
	resp.Links.Next = fmt.Sprintf("%s?%s", r.URL.Path, urlval.MustEncode(req))

	ape.Render(w, resp)
}

func balanceToResource(balance data.Balance, chain chains.Chain) resources.Balance {
	return resources.Balance{
		Key: resources.Key{
			ID:   fmt.Sprintf("%s:%s:%s", chain.Name, hexutil.Encode(balance.Token), hexutil.Encode(balance.AccountAddress)),
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

func fetchAndSaveBalances(r *http.Request, req listBalancesRequest, redisTokenCursor string, count int64) ([]data.Balance, error) {
	chainBalances, err := BalancesProvider(r).GetBalances(r.Context(), req.AccountAddress, req.ChainID, redisTokenCursor, count)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get balances from chain", logan.F{
			"account_address": req.AccountAddress,
			"chain_id":        req.ChainID,
			"token_cursor":    redisTokenCursor,
		})
	}

	if len(chainBalances) == 0 {
		return nil, nil
	}

	err = Config(r).Storage().BalanceQ().InsertBatchCtx(r.Context(), chainBalances...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert balances")
	}

	return Config(r).Storage().BalanceQ().SelectCtx(r.Context(), data.BalancesSelector{
		AccountAddress: &req.AccountAddress,
		ChainID:        &req.ChainID,

		TokenCursor: req.TokenCursor,
		PageSize:    uint64(count),
		Sort:        req.Sorts,
	})
}

func includeToken(r *http.Request, resp *resources.BalanceListResponse, balance data.Balance, chain chains.Chain) error {
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
