package config

import (
	"fmt"
	"net/url"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cast"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan"
	"gitlab.com/distributed_lab/logan/v3/errors"
	tokenmanager "gitlab.com/rarimo/rarimo-core/x/tokenmanager/types"
	"golang.org/x/exp/slices"
)

type SwapContractVersion string

const (
	SwapContractVersionTraderJoe   SwapContractVersion = "TraderJoe"
	SwapContractVersionQuickSwap   SwapContractVersion = "QuickSwap"
	SwapContractVersionPancakeSwap SwapContractVersion = "PancakeSwap"
	SwapContractVersionUniswapV3   SwapContractVersion = "UniswapV3"
)

type ChainKind string

const (
	ChainKindTestnet ChainKind = "testnet"
	ChainKindMainnet ChainKind = "mainnet"
)

type ChainsConfig struct {
	Chains []Chain `fig:"chains,required"`
}

func (c ChainsConfig) Find(id int64) *Chain {
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
	Kind                ChainKind                `fig:"kind,required"`
	IconURL             string                   `fig:"icon_url,required"`
	SwapContractAddr    common.Address           `fig:"swap_contract_address,required"`
	SwapContractVersion SwapContractVersion      `fig:"swap_contract_version,required"`
	TokensInfo          ChainTokensInfo          `fig:"tokens_info"`
}

type ChainTokensInfo struct {
	ListURL []url.URL   `fig:"list_urls"`
	Tokens  []TokenInfo `fig:"tokens"`
}

type TokenInfo struct {
	Name     string `json:"name" fig:"name"`
	Symbol   string `json:"symbol" fig:"symbol"`
	ChainID  int64  `json:"chainId" fig:"chain_id"`
	Address  string `json:"address" fig:"address"`
	Decimals int64  `json:"decimals" fig:"decimals"`
	LogoURI  string `json:"logoURI" fig:"logo_uri"`
}

func (c *config) Chains() *ChainsConfig {
	return c.chains.Do(func() interface{} {
		var cfg ChainsConfig

		cfgName := "chains"

		err := figure.
			Out(&cfg).
			With(figure.BaseHooks, figure.EthereumHooks, chainsHooks).
			From(kv.MustGetStringMap(c.getter, cfgName)).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out "+cfgName))
		}

		return cfg
	}).(*ChainsConfig)
}

var (
	tokenHooks = figure.Hooks{
		"[]url.URL": func(value interface{}) (reflect.Value, error) {
			raw, err := cast.ToStringSliceE(value)
			if err != nil {
				return reflect.Value{}, errors.Wrap(err, "failed to parse string slice")
			}

			urls := make([]url.URL, len(raw))

			for i, rawURL := range raw {
				u, err := url.Parse(rawURL)
				if err != nil {
					return reflect.Value{}, errors.Wrap(err, "failed to parse url", logan.F{
						"index": i,
						"raw":   rawURL,
					})
				}

				urls[i] = *u
			}

			return reflect.ValueOf(urls), nil
		},
		"[]config.TokenInfo": func(value interface{}) (reflect.Value, error) {
			raw, err := cast.ToSliceE(value)
			if err != nil {
				return reflect.Value{}, errors.Wrap(err, "failed to parse slice")
			}

			result := make([]TokenInfo, len(raw))

			for i, tokenMap := range raw {
				tokenCfg, err := cast.ToStringMapE(tokenMap)
				if err != nil {
					return reflect.Value{}, errors.Wrap(err, "failed to parse token cfg", logan.F{
						"index": i,
					})
				}

				var token TokenInfo
				if err := figure.Out(&token).From(tokenCfg).Please(); err != nil {
					return reflect.Value{}, errors.Wrap(err, "failed to parse token cfg", logan.F{
						"index": i,
					})
				}

				result[i] = token
			}

			return reflect.ValueOf(result), nil
		},
	}
	chainsHooks = figure.Hooks{
		"tokenmanager.NetworkType": figure.BaseHooks["int32"],
		"config.SwapContractVersion": func(value interface{}) (reflect.Value, error) {
			supported := []string{
				string(SwapContractVersionTraderJoe),
				string(SwapContractVersionQuickSwap),
				string(SwapContractVersionPancakeSwap),
				string(SwapContractVersionUniswapV3),
			}

			result, err := cast.ToStringE(value)
			if err != nil {
				return reflect.Value{}, errors.Wrap(err, "failed to parse string")
			}

			if !slices.Contains(supported, result) {
				return reflect.Value{}, fmt.Errorf("supported values are: %v", supported)
			}

			return reflect.ValueOf(SwapContractVersion(result)), nil
		},
		"config.ChainKind": func(value interface{}) (reflect.Value, error) {
			supported := []string{
				string(ChainKindTestnet),
				string(ChainKindMainnet),
			}

			result, err := cast.ToStringE(value)
			if err != nil {
				return reflect.Value{}, errors.Wrap(err, "failed to parse string")
			}

			if !slices.Contains(supported, result) {
				return reflect.Value{}, fmt.Errorf("supported values are: %v", supported)
			}

			return reflect.ValueOf(ChainKind(result)), nil
		},
		"config.ChainTokensInfo": func(value interface{}) (reflect.Value, error) {
			cfgMap, err := cast.ToStringMapE(value)
			if err != nil {
				return reflect.Value{}, errors.Wrap(err, "failed to parse map[string]interface{}")
			}

			var cfg ChainTokensInfo
			err = figure.
				Out(&cfg).
				With(figure.BaseHooks, tokenHooks).
				From(cfgMap).
				Please()
			if err != nil {
				return reflect.Value{}, errors.Wrap(err, "failed to figure out chain tokens")
			}

			return reflect.ValueOf(cfg), nil
		},
	}
)
