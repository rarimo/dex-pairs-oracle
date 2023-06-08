/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

type ChainAttributes struct {
	ExplorerUrl string `json:"explorer_url"`
	Icon        string `json:"icon"`
	// The kind of the chain
	Kind                ChainKind       `json:"kind"`
	Name                string          `json:"name"`
	NativeToken         NativeTokenInfo `json:"native_token"`
	Rpc                 string          `json:"rpc"`
	SwapContractAddress string          `json:"swap_contract_address"`
	SwapContractVersion string          `json:"swap_contract_version"`
	// The type of the chain
	Type ChainType `json:"type"`
}
