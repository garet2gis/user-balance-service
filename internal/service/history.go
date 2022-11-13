package service

import (
	"context"
	"github.com/garet2gis/user_balance_service/internal/apperror"
	"github.com/garet2gis/user_balance_service/internal/dto"
	"github.com/garet2gis/user_balance_service/internal/model"
	"github.com/garet2gis/user_balance_service/pkg/logging"
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
