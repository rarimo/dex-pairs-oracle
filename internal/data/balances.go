package data

import "gitlab.com/distributed_lab/kit/pgdb"

type BalancesSelector struct {
	TokenAddress   *string `json:"token,omitempty"`
	ChainID        *int64  `json:"chain_id,omitempty"`
	AccountAddress *string `json:"account_address,omitempty"`

	PageCursor uint64     `json:"page_number,omitempty"`
	PageSize   uint64     `json:"page_size,omitempty"`
	Sort       pgdb.Sorts `json:"sort"`
}
