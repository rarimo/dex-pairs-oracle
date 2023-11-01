package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"gitlab.com/distributed_lab/urlval"

	"github.com/go-chi/chi"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rarimo/dex-pairs-oracle/internal/chains"
	"github.com/rarimo/dex-pairs-oracle/resources"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3"
)

type listSupportedTokensRequest struct {
	chain chains.Chain

	Limit     int64  `page:"limit" default:"20"`
	RawCursor string `page:"cursor" default:""`

	tokenCursor []byte
}

func newListSupportedTokensRequest(r *http.Request) (*listSupportedTokensRequest, error) {
	var req listSupportedTokensRequest
	if err := urlval.Decode(r.URL.Query(), &req); err != nil {
		return nil, err
	}

	if req.RawCursor != "" {
		rawCursor, err := hexutil.Decode(req.RawCursor)
		if err != nil {
			return nil, validation.Errors{
				"cursor": fmt.Errorf("invalid cursor: %w", err),
			}
		}
		req.tokenCursor = rawCursor
	}

	if req.Limit < 0 || req.Limit > 500 {
		return nil, validation.Errors{
			"limit": fmt.Errorf("limit should be less than 500"),
		}
	}

	chainName := chi.URLParam(r, "chain_name")
	chain := Config(r).ChainsCfg().FindByName(strings.ToLower(chainName))
	if chain == nil {
		return nil, validation.Errors{
			"chain": fmt.Errorf("chain [%s] is not supported", chainName),
		}
	}
	req.chain = *chain

	return &req, nil
}

func ListSupportedEVMTokens(w http.ResponseWriter, r *http.Request) {
	req, err := newListSupportedTokensRequest(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	tokensCursor := ""
	if len(req.tokenCursor) > 0 {
		tokensCursor = fmt.Sprintf("token:%d:%s", req.chain.ID, hexutil.Encode(req.tokenCursor))
	}

	tokens, err := Config(r).RedisStore().Tokens().Page(r.Context(), req.chain.ID, tokensCursor, req.Limit)
	if err != nil {
		Log(r).WithError(err).WithFields(logan.F{
			"chain_id": req.chain.ID,
		}).Error("failed to get tokens")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	resp := resources.TokenListResponse{
		Data:     make([]resources.Token, len(tokens)),
		Included: resources.Included{},
		Links: &resources.Links{
			Self: fmt.Sprintf("%s?%s", r.URL.Path, urlval.MustEncode(req)),
		},
	}

	if len(tokens) == 0 {
		ape.Render(w, resp)
		return
	}

	req.RawCursor = tokens[len(tokens)-1].Address
	resp.Links.Next = fmt.Sprintf("%s?%s", r.URL.Path, urlval.MustEncode(req))

	for i, token := range tokens {
		resp.Data[i] = tokenToResource(token)
	}

	ape.Render(w, resp)
}

func tokenToResource(token chains.TokenInfo) resources.Token {
	tresource := resources.Token{
		Key: resources.Key{
			ID:   token.Address,
			Type: resources.TOKENS,
		},
		Attributes: resources.TokenAttributes{
			Name:     token.Name,
			Symbol:   token.Symbol,
			Decimals: token.Decimals,
			LogoUri:  token.LogoURI,
		},
	}

	if token.Native {
		tresource.Attributes.Native = &token.Native
	}

	return tresource
}
