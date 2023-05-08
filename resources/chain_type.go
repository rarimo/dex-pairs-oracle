package resources

import (
	"encoding/json"
	"strconv"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarimo/rarimo-core/x/tokenmanager/types"
)

type ChainType types.NetworkType

const (
	ChainTypeEVM          = ChainType(types.NetworkType_EVM)
	ChainTypeSolana       = ChainType(types.NetworkType_Solana)
	ChainTypeNearProtocol = ChainType(types.NetworkType_Near)
	ChainTypeOther        = ChainType(types.NetworkType_Other)
)

var chainTypeIntStr = map[ChainType]string{
	ChainTypeEVM:          "evm",
	ChainTypeSolana:       "solana",
	ChainTypeNearProtocol: "nearprotocol",
	ChainTypeOther:        "other",
}

var chainTypeStrInt = map[string]ChainType{
	"evm":          ChainTypeEVM,
	"solana":       ChainTypeSolana,
	"nearprotocol": ChainTypeNearProtocol,
	"other":        ChainTypeOther,
}

func (t ChainType) String() string {
	return chainTypeIntStr[t]
}

func (t ChainType) MarshalJSON() ([]byte, error) {
	return json.Marshal(Flag{
		Name:  chainTypeIntStr[t],
		Value: int32(t),
	})
}

func (t *ChainType) UnmarshalJSON(b []byte) error {
	var res Flag
	err := json.Unmarshal(b, &res)
	if err != nil {
		return err
	}

	*t = ChainType(res.Value)
	return nil
}

func SupportedChainTypesText() []string {
	return []string{
		ChainTypeEVM.String(),
		ChainTypeSolana.String(),
		ChainTypeNearProtocol.String(),
		ChainTypeOther.String(),
	}
}

func SupportedChainTypes() []interface{} {
	return []interface{}{
		ChainTypeEVM,
		ChainTypeSolana,
		ChainTypeNearProtocol,
		ChainTypeOther,
	}
}

func (t *ChainType) UnmarshalText(b []byte) error {
	typ, err := strconv.ParseInt(string(b), 0, 0)
	if err != nil {
		return err
	}

	if _, ok := chainTypeIntStr[ChainType(typ)]; !ok {
		return errors.From(errors.New("unsupported value"), logan.F{
			"supported": SupportedChainTypes(),
		})
	}

	*t = ChainType(typ)
	return nil
}

func (t ChainType) Int() int {
	return int(t)
}
