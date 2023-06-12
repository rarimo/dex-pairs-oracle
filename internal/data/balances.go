package data

import "gitlab.com/distributed_lab/kit/pgdb"

type BalancesSelector struct {
	ChainID        *int64  `json:"chain_id,omitempty"`
	AccountAddress *string `json:"account_address,omitempty"`

	Cursor      int64      `json:"cursor,omitempty"`
	TokenCursor []byte     `json:"token_cursor,omitempty"`
	PageSize    uint64     `json:"page_size,omitempty"`
	Sort        pgdb.Sorts `json:"sort"`
}
