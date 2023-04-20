package chains

import (
	"net/url"

	"github.com/ethereum/go-ethereum/common"
	tokenmanager "gitlab.com/rarimo/rarimo-core/x/tokenmanager/types"
)

type SwapContractVersion string

const (
	SwapContractVersionTraderJoe   SwapContractVersion = "TraderJoe"
	SwapContractVersionQuickSwap   SwapContractVersion = "QuickSwap"
	SwapContractVersionPancakeSwap SwapContractVersion = "PancakeSwap"
	SwapContractVersionUniswapV3   SwapContractVersion = "UniswapV3"
)

type Kind string

const (
	KindTestnet Kind = "testnet"
	KindMainnet Kind = "mainnet"
)

type Config struct {
	Chains []Chain `fig:"list,required"`
}

func (c Config) FindByName(name string) *Chain {
	for _, chain := range c.Chains {
		if chain.Name == name {
			return &chain
		}
	}
	return nil
}

func (c Config) Find(id int64) *Chain {
	for _, chain := range c.Chains {
		if chain.ID == id {
			return &chain
		}
	}
	return nil
}

type Chain struct {
	ID                  int64                    `fig:"id,required"`
	Name                string                   `fig:"name,required"`
	RPCUrl              *url.URL                 `fig:"rpc_url,required"`
	NativeSymbol        string                   `fig:"native_symbol,required"`
	ExplorerURL         string                   `fig:"explorer_url,required"`
	Type                tokenmanager.NetworkType `fig:"type,required"`
	Kind                Kind                     `fig:"kind,required"`
	IconURL             string                   `fig:"icon_url,required"`
	SwapContractAddr    common.Address           `fig:"swap_contract_address,required"`
	SwapContractVersion SwapContractVersion      `fig:"swap_contract_version,required"`
	TokensInfo          TokensInfo               `fig:"tokens_info"`
}

type TokensInfo struct {
	ListURL []url.URL   `fig:"list_urls"`
	Tokens  []TokenInfo `fig:"tokens"`
}

type TokenInfo struct {
	Name     string `json:"name" fig:"name"`
	Symbol   string `json:"symbol" fig:"symbol"`
	Address  string `json:"address" fig:"address"`
	Decimals int64  `json:"decimals" fig:"decimals"`
	LogoURI  string `json:"logoURI" fig:"logo_uri"`
	Native   bool   `json:"native" fig:"native"`
}
