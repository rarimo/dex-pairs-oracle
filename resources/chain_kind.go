package resources

import (
	"encoding/json"
	"strconv"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type ChainKind int32

const (
	ChainKindTestnet ChainKind = iota
	ChainKindMainnet
)

var chainKindIntStr = map[ChainKind]string{
	ChainKindTestnet: "testnet",
	ChainKindMainnet: "mainnet",
}

var chainKindStrInt = map[string]ChainKind{
	"testnet": ChainKindTestnet,
	"mainnet": ChainKindMainnet,
}

func (t ChainKind) String() string {
	return chainKindIntStr[t]
}

func (t ChainKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(Flag{
		Name:  chainKindIntStr[t],
		Value: int32(t),
	})
}

func (t *ChainKind) UnmarshalJSON(b []byte) error {
	var res Flag
	err := json.Unmarshal(b, &res)
	if err != nil {
		return err
	}

	*t = ChainKind(res.Value)
	return nil
}

func SupportedChainKindsText() []string {
	return []string{
		ChainKindTestnet.String(),
		ChainKindMainnet.String(),
	}
}

func SupportedChainKinds() []interface{} {
	return []interface{}{
		ChainKindTestnet,
		ChainKindMainnet,
	}
}

func (t *ChainKind) UnmarshalText(b []byte) error {
	typ, err := strconv.ParseInt(string(b), 0, 0)
	if err != nil {
		return err
	}

	if _, ok := chainKindIntStr[ChainKind(typ)]; !ok {
		return errors.From(errors.New("unsupported value"), logan.F{
			"supported": SupportedChainKinds(),
		})
	}

	*t = ChainKind(typ)
	return nil
}

func (t ChainKind) Int() int {
	return int(t)
}
