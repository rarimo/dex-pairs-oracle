package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/config"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"
	"gitlab.com/rarimo/dex-pairs-oracle/resources"
)

type listSupportedTokensRequest struct {
	Chain config.Chain
}

func newListSupportedTokensRequest(r *http.Request) (*listSupportedTokensRequest, error) {
	chainName := chi.URLParam(r, "chain")

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

func ListSupportedTokens(w http.ResponseWriter, r *http.Request) {
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

func tokenToResource(token data.Token) resources.Token {
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
