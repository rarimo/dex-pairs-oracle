package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/urlval"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/config"
	"gitlab.com/rarimo/dex-pairs-oracle/resources"
	tokenmanager "gitlab.com/rarimo/rarimo-core/x/tokenmanager/types"
)

type listSupportedChainRequest struct {
	Type *tokenmanager.NetworkType `filter:"type"`
	Kind *config.ChainKind         `filter:"kind"`
}

func newListSupportedChainRequest(r *http.Request) (*listSupportedChainRequest, error) {
	var req listSupportedChainRequest

	err := urlval.Decode(r.URL.Query(), &req)
	if err != nil {
		return nil, err
	}

	return &req, validateListSupportedChainRequest(req)
}

func validateListSupportedChainRequest(req listSupportedChainRequest) error {
	return validation.Errors{
		"filter[type]": validation.Validate(req.Type, validation.In(append(resources.SupportedChainTypes(), nil)...)),
		"filter[kind]": validation.Validate(req.Kind, validation.In(config.ChainKindTestnet, config.ChainKindMainnet)),
	}.Filter()
}

func ListSupportedChain(w http.ResponseWriter, r *http.Request) {
	req, err := newListSupportedChainRequest(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	chains := filterChains(Config(r).Chains(), req)

	resp := resources.ChainListResponse{
		Data: make([]resources.Chain, len(chains)),
		Links: &resources.Links{
			Self: r.URL.String(),
		},
		Included: resources.Included{},
		Meta:     json.RawMessage(fmt.Sprintf(`{"total":%d}`, len(Config(r).Chains()))),
	}

	if len(chains) == 0 {
		ape.Render(w, resp)
		return
	}

	for i, chain := range chains {
		resp.Data[i] = chainToResource(chain)
	}

	ape.Render(w, resp)
}

func chainToResource(chain config.Chain) resources.Chain {
	return resources.Chain{
		Key: resources.Key{
			ID:   strconv.FormatInt(chain.ID, 10),
			Type: resources.CHAINS,
		},
		Attributes: resources.ChainAttributes{
			Icon:                chain.IconURL,
			Kind:                chainKindToResource(chain.Kind),
			Name:                chain.Name,
			Rpc:                 chain.RPCUrl.String(),
			SwapContractAddress: chain.SwapContractAddr.String(),
			SwapContractVersion: string(chain.SwapContractVersion),
			Type:                chainTypeToResource(chain.Type),
		},
	}
}

func chainKindToResource(kind config.ChainKind) resources.ChainKind {
	switch kind {
	case config.ChainKindMainnet:
		return resources.ChainKindMainnet
	case config.ChainKindTestnet:
		return resources.ChainKindTestnet
	default:
		panic("unknown chain kind")
	}
}

func chainTypeToResource(typ tokenmanager.NetworkType) resources.ChainType {
	switch typ {
	case tokenmanager.NetworkType_EVM:
		return resources.ChainTypeEVM
	case tokenmanager.NetworkType_Solana:
		return resources.ChainTypeSolana
	case tokenmanager.NetworkType_Near:
		return resources.ChainTypeNearProtocol
	case tokenmanager.NetworkType_Other:
		return resources.ChainTypeOther
	default:
		panic("unknown chain type")
	}
}

func filterChains(chains []config.Chain, req *listSupportedChainRequest) []config.Chain {
	n := 0
	for _, chain := range chains {
		if req.Type != nil && *req.Type != chain.Type {
			continue
		}

		if req.Kind != nil && *req.Kind != chain.Kind {
			continue
		}

		chains[n] = chain
		n++
	}

	return chains[:n]
}
