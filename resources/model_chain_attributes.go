/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

type ChainAttributes struct {
	Icon string `json:"icon"`
	// The kind of the chain
	Kind                ChainKind `json:"kind"`
	Name                string    `json:"name"`
	Rpc                 string    `json:"rpc"`
	SwapContractAddress string    `json:"swap_contract_address"`
	SwapContractVersion string    `json:"swap_contract_version"`
	// The type of the chain
	Type ChainType `json:"type"`
}
