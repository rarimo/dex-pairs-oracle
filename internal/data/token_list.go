package data

type TokenListVersion struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

type VersionedTokenList struct {
	Name    string           `json:"name"`
	Version TokenListVersion `json:"version"`
	URI     string           `json:"uri"`
}
