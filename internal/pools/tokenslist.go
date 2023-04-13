package pools

import "gitlab.com/rarimo/dex-pairs-oracle/internal/data"

type VersionedTokenList struct {
	Version data.TokenListVersion `json:"version"`
	Name    string                `json:"name"`
	URI     string                `json:"uri"`
	Tokens  []data.Token          `json:"tokens"`
}
