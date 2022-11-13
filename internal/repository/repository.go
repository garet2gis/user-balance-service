package repository

import (
	"errors"
	"fmt"
	"github.com/garet2gis/user_balance_service/internal/apperror"
	"github.com/garet2gis/user_balance_service/pkg/logging"
	"github.com/garet2gis/user_balance_service/pkg/postgresql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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
