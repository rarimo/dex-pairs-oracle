package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rarimo/dex-pairs-oracle/internal/chains"
	"github.com/rarimo/dex-pairs-oracle/resources"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3"
)

type listSupportedTokensRequest struct {
	Chain chains.Chain
}

func newListSupportedTokensRequest(r *http.Request) (*listSupportedTokensRequest, error) {
	chainName := chi.URLParam(r, "chain_name")

	chain := Config(r).ChainsCfg().FindByName(strings.ToLower(chainName))
	if chain == nil {
		return nil, validation.Errors{
			"chain": fmt.Errorf("chain [%s] is not supported", chainName),
		}
	}

	return &listSupportedTokensRequest{
		Chain: *chain,
	}, nil
}

func ListSupportedEVMTokens(w http.ResponseWriter, r *http.Request) {
	req, err := newListSupportedTokensRequest(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	tokens, err := Config(r).RedisStore().Tokens().All(r.Context(), req.Chain.ID)
	if err != nil {
		Log(r).WithError(err).WithFields(logan.F{
			"chain_id": req.Chain.ID,
		}).Error("failed to get tokens")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	resp := resources.TokenListResponse{
		Data:     make([]resources.Token, len(tokens)),
		Included: resources.Included{},
		Links: &resources.Links{
			Self: r.URL.String(),
		},
	}

	if len(tokens) == 0 {
		ape.Render(w, resp)
		return
	}

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
