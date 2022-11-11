package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"user_balance_service/internal/apperror"
	"user_balance_service/pkg/logging"
	"user_balance_service/pkg/postgresql"
)

type Repository struct {
	client postgresql.Client
	logger *logging.Logger
	ReservationRepository
	HistoryRepository
	BalanceRepository
	ReportRepository
	BalanceChanger
}

func NewRepository(c *pgxpool.Pool, l *logging.Logger) *Repository {
	return &Repository{
		client:                c,
		logger:                l,
		HistoryRepository:     *NewHistoryRepository(c, l),
		BalanceRepository:     *NewBalanceRepository(c, l),
		ReportRepository:      *NewReportRepository(c, l),
		ReservationRepository: *NewReservationRepository(c, l),
		BalanceChanger:        *NewBalanceChanger(c, l),
	}
}

func PgxErrorLog(err error, l *logging.Logger) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		pgErr = err.(*pgconn.PgError)
		if pgErr.Code == "23514" && pgErr.ConstraintName == "balance_balance_check" {
			return toDBError(NotEnoughMoney)
		}
		newErr := fmt.Errorf("Code: %s, Message: %s, Where: %s, Detail: %s, SQLState: %s", pgErr.Code, pgErr.Message, pgErr.Where, pgErr.Detail, pgErr.SQLState())
		l.Error(newErr)
		return newErr
	}
	return err
}

func toDBError(err error) error {
	return apperror.NewAppError(err, "DB Error", err.Error())
}

func (r *Repository) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	return r.client.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
}

func (r *Repository) RollbackTransaction(ctx context.Context, tx pgx.Tx) {
	err := tx.Rollback(ctx)
	if err != nil {
		r.logger.Errorf("transaction rollback failed")
	}
}

func (r *Repository) CommitTransaction(ctx context.Context, tx pgx.Tx) {
	err := tx.Commit(ctx)
	if err != nil {
		r.logger.Errorf("transaction commit failed")
	}
}
