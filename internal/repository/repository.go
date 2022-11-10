package repository

import (
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
