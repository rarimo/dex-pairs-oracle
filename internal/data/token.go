package data

type Token struct {
	Address  string `json:"address"`
	ChainID  int64  `json:"chain_id"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals int64  `json:"decimals"`
	LogoURI  string `json:"logo_uri"`
}
