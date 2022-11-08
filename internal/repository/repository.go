package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"user_balance_service/pkg/logging"
)

type Repository struct {
	HR *HistoryRepository
	BR *BalanceRepository
	RR *ReportRepository
}

func NewRepository(c *pgxpool.Pool, l *logging.Logger) *Repository {
	return &Repository{
		HR: NewHistoryRepository(c, l),
		BR: NewBalanceRepository(c, l),
		RR: NewReportRepository(c, l),
	}
}
