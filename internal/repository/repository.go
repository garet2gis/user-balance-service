package repository

import (
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"user_balance_service/pkg/logging"
)

type Repository struct {
	ReservationRepository
	HistoryRepository
	BalanceRepository
	ReportRepository
}

func NewRepository(c *pgxpool.Pool, l *logging.Logger) *Repository {
	return &Repository{
		HistoryRepository:     *NewHistoryRepository(c, l),
		BalanceRepository:     *NewBalanceRepository(c, l),
		ReportRepository:      *NewReportRepository(c, l),
		ReservationRepository: *NewReservationRepository(c, l),
	}
}

func PgxErrorLog(err error, l *logging.Logger) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		pgErr = err.(*pgconn.PgError)
		if pgErr.Code == "23514" && pgErr.ConstraintName == "balance_balance_check" {
			return NotEnoughMoney
		}
		newErr := fmt.Errorf("Code: %s, Message: %s, Where: %s, Detail: %s, SQLState: %s", pgErr.Code, pgErr.Message, pgErr.Where, pgErr.Detail, pgErr.SQLState())
		l.Error(newErr)
		return newErr
	}
	return err
}
