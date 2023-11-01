package pools

import (
	"github.com/rarimo/dex-pairs-oracle/internal/chains"
	"github.com/rarimo/dex-pairs-oracle/internal/data"
)

type VersionedTokenList struct {
	Version data.TokenListVersion `json:"version"`
	Name    string                `json:"name"`
	URI     string                `json:"uri"`
	Tokens  []FullTokenInfo       `json:"tokens"`
}

type FullTokenInfo struct {
	ChainID int64 `json:"chainId"`
	chains.TokenInfo
}
