package chains

import (
	"fmt"
	"net/url"
	"reflect"

	"github.com/spf13/cast"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"golang.org/x/exp/slices"
)

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
		"[]chains.TokenInfo": func(value interface{}) (reflect.Value, error) {
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
	Hooks = figure.Hooks{
		"tokenmanager.NetworkType": figure.BaseHooks["int32"],
		"chains.SwapContractVersion": func(value interface{}) (reflect.Value, error) {
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
		"chains.Kind": func(value interface{}) (reflect.Value, error) {
			supported := []string{
				string(KindTestnet),
				string(KindMainnet),
			}

			result, err := cast.ToStringE(value)
			if err != nil {
				return reflect.Value{}, errors.Wrap(err, "failed to parse string")
			}

			if !slices.Contains(supported, result) {
				return reflect.Value{}, fmt.Errorf("supported values are: %v", supported)
			}

			return reflect.ValueOf(Kind(result)), nil
		},
		"chains.TokensInfo": func(value interface{}) (reflect.Value, error) {
			cfgMap, err := cast.ToStringMapE(value)
			if err != nil {
				return reflect.Value{}, errors.Wrap(err, "failed to parse map[string]interface{}")
			}

			var cfg TokensInfo
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
