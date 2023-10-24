package pg

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"gitlab.com/distributed_lab/kit/pgdb"

	"gitlab.com/distributed_lab/logan/v3/errors"

	"github.com/Masterminds/squirrel"

	"github.com/rarimo/dex-pairs-oracle/internal/data"
)

func (q BalanceQ) InsertBatchCtx(ctx context.Context, balances ...data.Balance) error {
	if len(balances) == 0 {
		return nil
	}

	stmt := squirrel.Insert("public.balances").
		Columns(
			"account_address", "token", "chain_id",
			"amount", "created_at", "updated_at",
			"last_known_block")

	for _, balance := range balances {
		stmt = stmt.Values(
			balance.AccountAddress, balance.Token, balance.ChainID,
			balance.Amount, balance.CreatedAt, balance.UpdatedAt,
			balance.LastKnownBlock)
	}

	return q.db.ExecContext(ctx, stmt)
}

func (q BalanceQ) UpsertBatchCtx(ctx context.Context, balances ...data.Balance) error {
	if len(balances) == 0 {
		return nil
	}

	stmt := squirrel.Insert("public.balances").
		Columns(
			"account_address", "token", "chain_id",
			"amount", "created_at", "updated_at",
			"last_known_block")

	for _, balance := range balances {
		stmt = stmt.Values(
			balance.AccountAddress, balance.Token, balance.ChainID,
			balance.Amount, balance.CreatedAt, balance.UpdatedAt,
			balance.LastKnownBlock)
	}

	// mitigating conflict on index problems in case balances get re-submitted
	stmt = stmt.Suffix(
		`ON CONFLICT(account_address,token,chain_id) DO ` +
			`UPDATE SET ` +
			`amount = EXCLUDED.amount, updated_at = EXCLUDED.updated_at, last_known_block = EXCLUDED.last_known_block`)

	return q.db.ExecContext(ctx, stmt)
}

func (q BalanceQ) SelectCtx(ctx context.Context, selector data.BalancesSelector) ([]data.Balance, error) {
	stmt := applyBalancesSelector(
		squirrel.Select("*").From("public.balances"),
		selector)

	var balances []data.Balance

	if err := q.db.SelectContext(ctx, &balances, stmt); err != nil {
		return nil, errors.Wrap(err, "failed to select balances")
	}

	return balances, nil
}

func applyBalancesSelector(stmt squirrel.SelectBuilder, selector data.BalancesSelector) squirrel.SelectBuilder {
	if selector.ChainID != nil {
		stmt = stmt.Where(squirrel.Eq{"chain_id": selector.ChainID})
	}

	if selector.AccountAddress != nil {
		stmt = stmt.Where(squirrel.Eq{"account_address": hexutil.MustDecode(*selector.AccountAddress)})
	}

	stmt = applyBalancesPagination(stmt, selector.Sort, selector.Cursor, selector.TokenCursor, selector.PageSize)

	return stmt
}

func applyBalancesPagination(stmt squirrel.SelectBuilder, sorts pgdb.Sorts, idCursor int64, tokenCursor []byte, limit uint64) squirrel.SelectBuilder {
	stmt = stmt.Limit(limit)

	if len(sorts) == 0 {
		sorts = pgdb.Sorts{"id"}
	}

	stmtSorts := pgdb.Sorts{"token"}
	for _, sort := range sorts {
		if sort == "token" {
			continue
		}

		if sort == "-token" {
			stmtSorts[0] = "-token"
			continue
		}

		stmtSorts = append(stmtSorts, sort)
	}

	stmt = sorts.ApplyTo(stmt, map[string]string{ // TODO move it kinda closer to the handler's request model
		"token":  "token",
		"id":     "id",
		"time":   "created_at",
		"amount": "amount",
	})

	comp := ">" // default to ascending order

	if sortDesc := strings.HasPrefix(string(sorts[0]), "-"); sortDesc {
		comp = "<"
	}

	switch {
	case len(tokenCursor) != 0:
		stmt = stmt.Where(fmt.Sprintf("token %s ?", comp), tokenCursor)
	case idCursor != 0:
		stmt = stmt.Where(fmt.Sprintf("id %s ?", comp), idCursor)
	}

	return stmt
}
