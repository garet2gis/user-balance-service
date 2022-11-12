package service

import (
	"context"
	"github.com/jackc/pgx/v5"
	"user_balance_service/internal/repository"
	"user_balance_service/pkg/logging"
)

type TransactionRepository interface {
	BeginTransaction(ctx context.Context) (pgx.Tx, error)
	RollbackTransaction(ctx context.Context, tx pgx.Tx)
	CommitTransaction(ctx context.Context, tx pgx.Tx)
}

type Service struct {
	BalanceService
	HistoryService
	ReservationService
}

func NewService(r *repository.Repository, l *logging.Logger) *Service {
	return &Service{
		BalanceService:     *NewBalanceService(r, l),
		HistoryService:     *NewHistoryService(r, l),
		ReservationService: *NewReservationService(r, l),
	}
}
