// Package pg contains generated code for schema 'public'.
package pg

// Code generated by xo. DO NOT EDIT.

import (
	"context"
	"database/sql"

	"gitlab.com/distributed_lab/kit/pgdb"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarimo/dex-pairs-oracle/internal/data"
)

// Storage is the helper struct for database operations
type Storage struct {
	db *pgdb.DB
}

// New - returns new instance of storage
func New(db *pgdb.DB) *Storage {
	return &Storage{
		db,
	}
}

// DB - returns db used by Storage
func (s *Storage) DB() *pgdb.DB {
	return s.db
}

// Clone - returns new storage with clone of db
func (s *Storage) Clone() data.Storage {
	return New(s.db.Clone())
}

// Transaction begins a transaction on repo.
func (s *Storage) Transaction(tx func() error) error {
	return s.db.Transaction(tx)
} // BalanceQ represents helper struct to access row of 'balances'.
type BalanceQ struct {
	db *pgdb.DB
}

// NewBalanceQ  - creates new instance
func NewBalanceQ(db *pgdb.DB) BalanceQ {
	return BalanceQ{
		db,
	}
}

// BalanceQ  - creates new instance of BalanceQ
func (s Storage) BalanceQ() data.BalanceQ {
	return NewBalanceQ(s.DB())
}

var colsBalance = `id, account_address, token, chain_id, amount, created_at, updated_at, last_known_block`

// InsertCtx inserts a Balance to the database.
func (q BalanceQ) InsertCtx(ctx context.Context, b *data.Balance) error {
	// insert (primary key generated and returned by database)
	sqlstr := `INSERT INTO public.balances (` +
		`account_address, token, chain_id, amount, created_at, updated_at, last_known_block` +
		`) VALUES (` +
		`$1, $2, $3, $4, $5, $6, $7` +
		`) RETURNING id`
		// run

	err := q.db.GetRawContext(ctx, &b.ID, sqlstr, b.AccountAddress, b.Token, b.ChainID, b.Amount, b.CreatedAt, b.UpdatedAt, b.LastKnownBlock)
	if err != nil {
		return errors.Wrap(err, "failed to execute insert")
	}

	return nil
}

// Insert insert a Balance to the database.
func (q BalanceQ) Insert(b *data.Balance) error {
	return q.InsertCtx(context.Background(), b)
}

// UpdateCtx updates a Balance in the database.
func (q BalanceQ) UpdateCtx(ctx context.Context, b *data.Balance) error {
	// update with composite primary key
	sqlstr := `UPDATE public.balances SET ` +
		`account_address = $1, token = $2, chain_id = $3, amount = $4, updated_at = $5, last_known_block = $6 ` +
		`WHERE id = $7`
	// run
	err := q.db.ExecRawContext(ctx, sqlstr, b.AccountAddress, b.Token, b.ChainID, b.Amount, b.UpdatedAt, b.LastKnownBlock, b.ID)
	return errors.Wrap(err, "failed to execute update")
}

// Update updates a Balance in the database.
func (q BalanceQ) Update(b *data.Balance) error {
	return q.UpdateCtx(context.Background(), b)
}

// UpsertCtx performs an upsert for Balance.
func (q BalanceQ) UpsertCtx(ctx context.Context, b *data.Balance) error {
	// upsert
	sqlstr := `INSERT INTO public.balances (` +
		`id, account_address, token, chain_id, amount, created_at, updated_at, last_known_block` +
		`) VALUES (` +
		`$1, $2, $3, $4, $5, $6, $7, $8` +
		`)` +
		` ON CONFLICT (id) DO ` +
		`UPDATE SET ` +
		`account_address = EXCLUDED.account_address, token = EXCLUDED.token, chain_id = EXCLUDED.chain_id, amount = EXCLUDED.amount, updated_at = EXCLUDED.updated_at, last_known_block = EXCLUDED.last_known_block `
	// run
	if err := q.db.ExecRawContext(ctx, sqlstr, b.ID, b.AccountAddress, b.Token, b.ChainID, b.Amount, b.CreatedAt, b.UpdatedAt, b.LastKnownBlock); err != nil {
		return errors.Wrap(err, "failed to execute upsert stmt")
	}
	return nil
}

// Upsert performs an upsert for Balance.
func (q BalanceQ) Upsert(b *data.Balance) error {
	return q.UpsertCtx(context.Background(), b)
}

// DeleteCtx deletes the Balance from the database.
func (q BalanceQ) DeleteCtx(ctx context.Context, b *data.Balance) error {
	// delete with single primary key
	sqlstr := `DELETE FROM public.balances ` +
		`WHERE id = $1`
	// run
	if err := q.db.ExecRawContext(ctx, sqlstr, b.ID); err != nil {
		return errors.Wrap(err, "failed to exec delete stmt")
	}
	return nil
}

// Delete deletes the Balance from the database.
func (q BalanceQ) Delete(b *data.Balance) error {
	return q.DeleteCtx(context.Background(), b)
} // GorpMigrationQ represents helper struct to access row of 'gorp_migrations'.
type GorpMigrationQ struct {
	db *pgdb.DB
}

// NewGorpMigrationQ  - creates new instance
func NewGorpMigrationQ(db *pgdb.DB) GorpMigrationQ {
	return GorpMigrationQ{
		db,
	}
}

// GorpMigrationQ  - creates new instance of GorpMigrationQ
func (s Storage) GorpMigrationQ() data.GorpMigrationQ {
	return NewGorpMigrationQ(s.DB())
}

var colsGorpMigration = `id, applied_at`

// InsertCtx inserts a GorpMigration to the database.
func (q GorpMigrationQ) InsertCtx(ctx context.Context, gm *data.GorpMigration) error {
	// sql insert query, primary key must be provided
	sqlstr := `INSERT INTO public.gorp_migrations (` +
		`id, applied_at` +
		`) VALUES (` +
		`$1, $2` +
		`)`
	// run
	err := q.db.ExecRawContext(ctx, sqlstr, gm.ID, gm.AppliedAt)
	return errors.Wrap(err, "failed to execute insert query")
}

// Insert insert a GorpMigration to the database.
func (q GorpMigrationQ) Insert(gm *data.GorpMigration) error {
	return q.InsertCtx(context.Background(), gm)
}

// UpdateCtx updates a GorpMigration in the database.
func (q GorpMigrationQ) UpdateCtx(ctx context.Context, gm *data.GorpMigration) error {
	// update with composite primary key
	sqlstr := `UPDATE public.gorp_migrations SET ` +
		`applied_at = $1 ` +
		`WHERE id = $2`
	// run
	err := q.db.ExecRawContext(ctx, sqlstr, gm.AppliedAt, gm.ID)
	return errors.Wrap(err, "failed to execute update")
}

// Update updates a GorpMigration in the database.
func (q GorpMigrationQ) Update(gm *data.GorpMigration) error {
	return q.UpdateCtx(context.Background(), gm)
}

// UpsertCtx performs an upsert for GorpMigration.
func (q GorpMigrationQ) UpsertCtx(ctx context.Context, gm *data.GorpMigration) error {
	// upsert
	sqlstr := `INSERT INTO public.gorp_migrations (` +
		`id, applied_at` +
		`) VALUES (` +
		`$1, $2` +
		`)` +
		` ON CONFLICT (id) DO ` +
		`UPDATE SET ` +
		`applied_at = EXCLUDED.applied_at `
	// run
	if err := q.db.ExecRawContext(ctx, sqlstr, gm.ID, gm.AppliedAt); err != nil {
		return errors.Wrap(err, "failed to execute upsert stmt")
	}
	return nil
}

// Upsert performs an upsert for GorpMigration.
func (q GorpMigrationQ) Upsert(gm *data.GorpMigration) error {
	return q.UpsertCtx(context.Background(), gm)
}

// DeleteCtx deletes the GorpMigration from the database.
func (q GorpMigrationQ) DeleteCtx(ctx context.Context, gm *data.GorpMigration) error {
	// delete with single primary key
	sqlstr := `DELETE FROM public.gorp_migrations ` +
		`WHERE id = $1`
	// run
	if err := q.db.ExecRawContext(ctx, sqlstr, gm.ID); err != nil {
		return errors.Wrap(err, "failed to exec delete stmt")
	}
	return nil
}

// Delete deletes the GorpMigration from the database.
func (q GorpMigrationQ) Delete(gm *data.GorpMigration) error {
	return q.DeleteCtx(context.Background(), gm)
}

// BalancesByAccountAddressCtx retrieves a row from 'public.balances' as a Balance.
//
// Generated from index 'balances_account_address_idx'.
func (q BalanceQ) BalancesByAccountAddressCtx(ctx context.Context, accountAddress []byte, isForUpdate bool) ([]data.Balance, error) {
	// query
	sqlstr := `SELECT ` +
		`id, account_address, token, chain_id, amount, created_at, updated_at, last_known_block ` +
		`FROM public.balances ` +
		`WHERE account_address = $1`
	// run
	if isForUpdate {
		sqlstr += " for update"
	}
	var res []data.Balance
	err := q.db.SelectRawContext(ctx, &res, sqlstr, accountAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to exec select")
	}

	return res, nil
}

// BalancesByAccountAddress retrieves a row from 'public.balances' as a Balance.
//
// Generated from index 'balances_account_address_idx'.
func (q BalanceQ) BalancesByAccountAddress(accountAddress []byte, isForUpdate bool) ([]data.Balance, error) {
	return q.BalancesByAccountAddressCtx(context.Background(), accountAddress, isForUpdate)
}

// BalancesByChainIDCtx retrieves a row from 'public.balances' as a Balance.
//
// Generated from index 'balances_chain_id_idx'.
func (q BalanceQ) BalancesByChainIDCtx(ctx context.Context, chainID int64, isForUpdate bool) ([]data.Balance, error) {
	// query
	sqlstr := `SELECT ` +
		`id, account_address, token, chain_id, amount, created_at, updated_at, last_known_block ` +
		`FROM public.balances ` +
		`WHERE chain_id = $1`
	// run
	if isForUpdate {
		sqlstr += " for update"
	}
	var res []data.Balance
	err := q.db.SelectRawContext(ctx, &res, sqlstr, chainID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to exec select")
	}

	return res, nil
}

// BalancesByChainID retrieves a row from 'public.balances' as a Balance.
//
// Generated from index 'balances_chain_id_idx'.
func (q BalanceQ) BalancesByChainID(chainID int64, isForUpdate bool) ([]data.Balance, error) {
	return q.BalancesByChainIDCtx(context.Background(), chainID, isForUpdate)
}

// BalanceByChainIDTokenAccountAddressCtx retrieves a row from 'public.balances' as a Balance.
//
// Generated from index 'balances_chain_id_token_account_address_key'.
func (q BalanceQ) BalanceByChainIDTokenAccountAddressCtx(ctx context.Context, chainID int64, token, accountAddress []byte, isForUpdate bool) (*data.Balance, error) {
	// query
	sqlstr := `SELECT ` +
		`id, account_address, token, chain_id, amount, created_at, updated_at, last_known_block ` +
		`FROM public.balances ` +
		`WHERE chain_id = $1 AND token = $2 AND account_address = $3`
	// run
	if isForUpdate {
		sqlstr += " for update"
	}
	var res data.Balance
	err := q.db.GetRawContext(ctx, &res, sqlstr, chainID, token, accountAddress)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, nil
		}

		return nil, errors.Wrap(err, "failed to exec select")
	}

	return &res, nil
}

// BalanceByChainIDTokenAccountAddress retrieves a row from 'public.balances' as a Balance.
//
// Generated from index 'balances_chain_id_token_account_address_key'.
func (q BalanceQ) BalanceByChainIDTokenAccountAddress(chainID int64, token, accountAddress []byte, isForUpdate bool) (*data.Balance, error) {
	return q.BalanceByChainIDTokenAccountAddressCtx(context.Background(), chainID, token, accountAddress, isForUpdate)
}

// BalanceByIDCtx retrieves a row from 'public.balances' as a Balance.
//
// Generated from index 'balances_pkey'.
func (q BalanceQ) BalanceByIDCtx(ctx context.Context, id int64, isForUpdate bool) (*data.Balance, error) {
	// query
	sqlstr := `SELECT ` +
		`id, account_address, token, chain_id, amount, created_at, updated_at, last_known_block ` +
		`FROM public.balances ` +
		`WHERE id = $1`
	// run
	if isForUpdate {
		sqlstr += " for update"
	}
	var res data.Balance
	err := q.db.GetRawContext(ctx, &res, sqlstr, id)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, nil
		}

		return nil, errors.Wrap(err, "failed to exec select")
	}

	return &res, nil
}

// BalanceByID retrieves a row from 'public.balances' as a Balance.
//
// Generated from index 'balances_pkey'.
func (q BalanceQ) BalanceByID(id int64, isForUpdate bool) (*data.Balance, error) {
	return q.BalanceByIDCtx(context.Background(), id, isForUpdate)
}

// BalancesByTokenCtx retrieves a row from 'public.balances' as a Balance.
//
// Generated from index 'balances_token_idx'.
func (q BalanceQ) BalancesByTokenCtx(ctx context.Context, token []byte, isForUpdate bool) ([]data.Balance, error) {
	// query
	sqlstr := `SELECT ` +
		`id, account_address, token, chain_id, amount, created_at, updated_at, last_known_block ` +
		`FROM public.balances ` +
		`WHERE token = $1`
	// run
	if isForUpdate {
		sqlstr += " for update"
	}
	var res []data.Balance
	err := q.db.SelectRawContext(ctx, &res, sqlstr, token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to exec select")
	}

	return res, nil
}

// BalancesByToken retrieves a row from 'public.balances' as a Balance.
//
// Generated from index 'balances_token_idx'.
func (q BalanceQ) BalancesByToken(token []byte, isForUpdate bool) ([]data.Balance, error) {
	return q.BalancesByTokenCtx(context.Background(), token, isForUpdate)
}

// GorpMigrationByIDCtx retrieves a row from 'public.gorp_migrations' as a GorpMigration.
//
// Generated from index 'gorp_migrations_pkey'.
func (q GorpMigrationQ) GorpMigrationByIDCtx(ctx context.Context, id string, isForUpdate bool) (*data.GorpMigration, error) {
	// query
	sqlstr := `SELECT ` +
		`id, applied_at ` +
		`FROM public.gorp_migrations ` +
		`WHERE id = $1`
	// run
	if isForUpdate {
		sqlstr += " for update"
	}
	var res data.GorpMigration
	err := q.db.GetRawContext(ctx, &res, sqlstr, id)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, nil
		}

		return nil, errors.Wrap(err, "failed to exec select")
	}

	return &res, nil
}

// GorpMigrationByID retrieves a row from 'public.gorp_migrations' as a GorpMigration.
//
// Generated from index 'gorp_migrations_pkey'.
func (q GorpMigrationQ) GorpMigrationByID(id string, isForUpdate bool) (*data.GorpMigration, error) {
	return q.GorpMigrationByIDCtx(context.Background(), id, isForUpdate)
}
