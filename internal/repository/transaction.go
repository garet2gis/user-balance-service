package repository

import (
	"context"
	"github.com/garet2gis/user_balance_service/pkg/logging"
	"github.com/garet2gis/user_balance_service/pkg/postgresql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionHelper struct {
	client postgresql.Client
	logger *logging.Logger
}

func NewTransactionHelper(c *pgxpool.Pool, l *logging.Logger) *TransactionHelper {
	return &TransactionHelper{
		client: c,
		logger: l,
	}
}

func (r *TransactionHelper) beginTransaction(ctx context.Context) (pgx.Tx, error) {
	return r.client.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
}

func (r *TransactionHelper) rollbackTransaction(ctx context.Context, tx pgx.Tx) {
	err := tx.Rollback(ctx)
	if err != nil {
		r.logger.Errorf("transaction rollback failed")
	}
}

func (r *TransactionHelper) commitTransaction(ctx context.Context, tx pgx.Tx) {
	err := tx.Commit(ctx)
	if err != nil {
		r.logger.Errorf("transaction commit failed")
	}
}
