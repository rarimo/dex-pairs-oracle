package data

type Token struct {
	Address  string `json:"address"`
	ChainID  int64  `json:"chainId"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals int64  `json:"decimals"`
	LogoURI  string `json:"logoURI"`
	Native   bool   `json:"native"`

	Cursor string `json:"cursor"`
}
