package pg

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"gitlab.com/distributed_lab/kit/pgdb"

	"gitlab.com/distributed_lab/logan/v3/errors"

	"github.com/Masterminds/squirrel"

	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"
)

func (q BalanceQ) InsertBatchCtx(ctx context.Context, balances ...data.Balance) error {
	stmt := squirrel.Insert("public.balances").
		Columns("account_address", "token", "chain_id",
			"amount", "created_at", "updated_at")

	for _, balance := range balances {
		stmt = stmt.
			Values(balance.AccountAddress, balance.Token, balance.ChainID,
				balance.Amount, balance.CreatedAt, balance.UpdatedAt)
	}

	return q.db.ExecContext(ctx, stmt)
}

func (q BalanceQ) UpsertBatchCtx(ctx context.Context, balances ...data.Balance) error {
	stmt := squirrel.Insert("public.balances").
		Columns("account_address", "token", "chain_id",
			"amount", "created_at", "updated_at")

	for _, balance := range balances {
		stmt = stmt.
			Values(balance.AccountAddress, balance.Token, balance.ChainID,
				balance.Amount, balance.CreatedAt, balance.UpdatedAt)
	}

	// mitigating conflict on index problems in case balances get re-submitted
	stmt = stmt.Suffix(
		`ON CONFLICT(id) DO ` +
			`UPDATE SET ` +
			`amount = EXCLUDED.amount, updated_at = EXCLUDED.updated_at`)

	return q.db.ExecContext(ctx, stmt)
}

func (q BalanceQ) SelectCtx(ctx context.Context, selector data.BalancesSelector) ([]data.Balance, error) {
	stmt := applyBalancesSelector(
		squirrel.Select("*").From("public.balances"),
		selector)

	var balances []data.Balance

	if err := q.db.SelectContext(ctx, &balances, stmt); err != nil {
		return nil, errors.Wrap(err, "failed to select transfers")
	}

	return balances, nil
}

func applyBalancesSelector(stmt squirrel.SelectBuilder, selector data.BalancesSelector) squirrel.SelectBuilder {
	if selector.ChainID != nil {
		stmt = stmt.Where(squirrel.Eq{"chain": selector.ChainID})
	}

	if selector.AccountAddress != nil {
		stmt = stmt.Where(squirrel.Eq{"account_address": hexutil.MustDecode(*selector.AccountAddress)})
	}

	if selector.TokenAddress != nil {
		stmt = stmt.Where(squirrel.Eq{"token": hexutil.MustDecode(*selector.TokenAddress)})
	}

	return stmt
}

func applyTransfersPagination(stmt squirrel.SelectBuilder, sorts pgdb.Sorts, cursor, limit uint64) squirrel.SelectBuilder {
	stmt = stmt.Limit(limit)

	if len(sorts) == 0 {
		sorts = pgdb.Sorts{"-time"}
	}

	stmt = sorts.ApplyTo(stmt, map[string]string{
		"id":   "id",
		"time": "rarimo_tx_timestamp",
	})

	if cursor != 0 {
		comp := ">" // default to ascending order
		if sortDesc := strings.HasPrefix(string(sorts[0]), "-"); sortDesc {
			comp = "<"
		}

		stmt = stmt.Where(fmt.Sprintf("id %s ?", comp), cursor)
	}

	return stmt
}
