package handlers

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/rarimo/dex-pairs-oracle/pkg/ethbalances"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rarimo/dex-pairs-oracle/internal/chains"
	"github.com/rarimo/dex-pairs-oracle/internal/data"
	"github.com/rarimo/dex-pairs-oracle/pkg/etherrors"
	"github.com/rarimo/dex-pairs-oracle/resources"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/kit/pgdb"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/urlval"
)

type listBalancesRequest struct {
	ChainID        int64
	AccountAddress string

	IncludeChain bool `include:"chain"`
	IncludeToken bool `include:"token"`

	RawCursor   string `page:"cursor" default:""`
	TokenCursor []byte

	PageLimit int64      `page:"limit" default:"15"`
	Sorts     pgdb.Sorts `url:"sort" default:"token"`
}

func newListEvmBalancesAddress(r *http.Request) (*listBalancesRequest, error) {
	chainName := chi.URLParam(r, "chain_name")

	supported := Config(r).ChainsCfg().FindByName(chainName)
	if supported == nil {
		return nil, validation.Errors{
			"chain_name": fmt.Errorf("chain [%s] is not supported", chainName),
		}
	}

	req := listBalancesRequest{
		ChainID:        supported.ID,
		AccountAddress: chi.URLParam(r, "account_address"),
	}

	if err := urlval.Decode(r.URL.Query(), &req); err != nil {
		return nil, err
	}

	if req.PageLimit < 1 || req.PageLimit > 100 {
		return nil, validation.Errors{
			"page[limit]": errors.New("should be in the range [1; 100]"),
		}
	}

	if req.RawCursor != "" {
		req.TokenCursor = hexutil.MustDecode(req.RawCursor)
	}

	return &req, nil
}

type ChainBalancesProvider interface {
	GetBalances(ctx context.Context, address string, tokens []chains.TokenInfo) ([]data.Balance, error)
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

	if len(balances) < int(req.PageLimit) {
		limit := req.PageLimit - int64(len(balances))

		tokenCursor := ""
		if len(balances) != 0 {
			tokenCursor = fmt.Sprintf("token:%d:%s", req.ChainID, hexutil.Encode(balances[len(balances)-1].Token))
		}

		// balances we need rn to fill the page
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

		// balance dummies that will later be populated by balances_observer
		// (inserting them to make furhter requests faster)
		balanceDummies, err := makeBalancesDummies(r, req.ChainID, req.AccountAddress, 5*req.PageLimit)
		if err != nil {
			// we already have all the things we need for the user so
			// just log it here to be aware if something is wrong with db/redis connection
			Log(r).WithError(err).Warn("failed to make balance dummies")
		}

		err = Config(r).Storage().BalanceQ().InsertBatchCtx(r.Context(), balanceDummies...)
		if err != nil {
			// same here
			Log(r).WithError(err).Warn("failed to insert balance dummies")
		}
	}

	if len(balances) == 0 {
		ape.Render(w, resp)
		return
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

// hack to make further requests faster
// most probably many of user's balances are of zero amount
// so we can simply pre-populate them now and
// let the balances observer will make all the rest
// in case user has non-zero balances - balances_observer will update them in a short while
func makeBalancesDummies(r *http.Request, chainID int64, accountAddress string, number int64) ([]data.Balance, error) {

	lastBalance, err := Config(r).Storage().BalanceQ().SelectCtx(r.Context(), data.BalancesSelector{
		AccountAddress: &accountAddress,
		ChainID:        &chainID,
		PageSize:       1,
		Sort: pgdb.Sorts{
			"-token",
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get last balance")
	}

	startCursor := ""
	if len(lastBalance) != 0 {
		Log(r).WithFields(logan.F{
			"account_address": accountAddress,
			"chain_id":        chainID,
		}).Debug("found balances for account, starting from the last one")
		startCursor = fmt.Sprintf("token:%d:%s", chainID, hexutil.Encode(lastBalance[0].Token))
	}

	tokens, err := Config(r).RedisStore().Tokens().Page(r.Context(),
		chainID,
		startCursor,
		number)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get tokens")
	}

	if len(tokens) == 0 {
		Log(r).WithFields(logan.F{
			"account_address": accountAddress,
			"chain_id":        chainID,
			"cursor":          startCursor,
			"limit":           number,
		}).Debug("no tokens left for creating dummies")
		return nil, nil
	}

	now := time.Now()

	balances := make([]data.Balance, 0, len(tokens))
	for _, token := range tokens {
		balances = append(balances, data.Balance{
			AccountAddress: hexutil.MustDecode(accountAddress),
			ChainID:        chainID,
			Token:          hexutil.MustDecode(token.Address),
			Amount:         data.Int256{big.NewInt(0)},
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}

	return balances, nil
}

func fetchAndSaveBalances(r *http.Request, req listBalancesRequest, redisTokenCursor string, count int64) ([]data.Balance, error) {
	tokens, err := Config(r).RedisStore().Tokens().Page(r.Context(), req.ChainID, redisTokenCursor, count)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get tokens by chain id", logan.F{
			"chain_id": req.ChainID,
		})
	}

	chainBalances, err := ethbalances.
		NewProvider(Config(r).RedisStore(), Config(r).ChainsCfg().Find(req.ChainID)).
		GetBalances(r.Context(), req.AccountAddress, tokens)
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

	err = Config(r).Storage().BalanceQ().UpsertBatchCtx(r.Context(), chainBalances...)
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
