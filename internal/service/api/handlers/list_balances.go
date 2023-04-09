package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/kit/pgdb"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/urlval"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/config"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"
	"gitlab.com/rarimo/dex-pairs-oracle/resources"
)

type listBalancesRequest struct {
	ChainID        *int64  `filter:"chain_id"`
	TokenAddress   *string `filter:"token_address"`
	AccountAddress *string `filter:"account_address"`

	PageCursor uint64     `page:"cursor"`
	PageLimit  uint64     `page:"limit" default:"15"`
	Sorts      pgdb.Sorts `url:"sort" default:"-amount"`
}

func newListBalancesAddress(r *http.Request) (*listBalancesRequest, error) {
	var req listBalancesRequest

	err := urlval.Decode(r.URL.Query(), &req)
	if err != nil {
		return nil, err
	}

	if req.ChainID != nil {
		if req.TokenAddress != nil {
			chain, _, err := parseAddress(r, *req.TokenAddress)
			if err != nil {
				return nil, validation.Errors{
					"filter[token_address]": err,
				}
			}

			if chain.ID != *req.ChainID {
				return nil, validation.Errors{
					"filter[token_address]": errors.New("chain id should not differ from filter[chain_id]"),
				}
			}
		}

		if req.AccountAddress != nil {
			chain, _, err := parseAddress(r, *req.AccountAddress)
			if err != nil {
				return nil, validation.Errors{
					"filter[account_address]": err,
				}
			}

			if chain.ID != *req.ChainID {
				return nil, validation.Errors{
					"filter[account_address]": errors.New("chain id should not differ from filter[chain_id]"),
				}
			}
		}
	}

	return &req, nil
}

func ListBalances(w http.ResponseWriter, r *http.Request) {
	req, err := newListBalancesAddress(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	balances, err := Storage(r).BalanceQ().SelectCtx(r.Context(), data.BalancesSelector{
		TokenAddress:   req.TokenAddress,
		AccountAddress: req.AccountAddress,
		ChainID:        req.ChainID,
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

	if len(balances) == 0 {
		ape.Render(w, resp)
		return
	}

	for i, balance := range balances {
		resp.Data[i] = balanceToResource(balance, *Config(r).Chains().Find(balance.ChainID))
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

func parseAddress(r *http.Request, raw string) (*config.Chain, string, error) {
	split := strings.Split(raw, ":")

	if len(split) != 2 {
		return nil, "", errors.New("invalid address")
	}

	chainID, err := strconv.ParseInt(split[0], 10, 64)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to parse chain id")
	}

	chain := Config(r).Chains().Find(chainID)
	if chain == nil {
		return nil, "", errors.New("unsupported chain")
	}

	return chain, split[1], nil
}
