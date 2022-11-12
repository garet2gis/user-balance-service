package service

import (
	"context"
	"user_balance_service/internal/apperror"
	"user_balance_service/internal/dto"
	"user_balance_service/internal/model"
	"user_balance_service/pkg/logging"
)

type HistoryRepository interface {
	GetUserBalanceHistory(ctx context.Context, bh dto.BalanceHistory) ([]model.HistoryRow, error)
}

type HistoryService struct {
	repo   HistoryRepository
	logger *logging.Logger
}

func NewHistoryService(r HistoryRepository, l *logging.Logger) *HistoryService {
	return &HistoryService{
		repo:   r,
		logger: l,
	}
}

func (hs *HistoryService) GetHistory(ctx context.Context, bh dto.BalanceHistory) ([]model.HistoryRow, error) {
	history, err := hs.repo.GetUserBalanceHistory(ctx, bh)
	if err != nil {
		return nil, err
	}
	if len(history) == 0 {
		return nil, apperror.ErrNotFound
	}

	return history, nil
}
