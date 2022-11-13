package service

import (
	"context"
	"github.com/garet2gis/user_balance_service/internal/dto"
	"github.com/garet2gis/user_balance_service/internal/model"
	"github.com/garet2gis/user_balance_service/pkg/logging"
)

type BalanceRepository interface {
	GetBalanceByUserID(ctx context.Context, id string) (float64, error)
	ChangeUserBalance(ctx context.Context, b dto.BalanceChangeRequest, depositType model.DepositType) (bm *dto.BalanceChangeRequest, err error)
	TransferMoney(ctx context.Context, transfer dto.TransferRequest) (err error)
}

type BalanceService struct {
	repo   BalanceRepository
	logger *logging.Logger
}

func NewBalanceService(r BalanceRepository, l *logging.Logger) *BalanceService {
	return &BalanceService{
		repo:   r,
		logger: l,
	}
}

func (bs *BalanceService) GetBalanceByUserID(ctx context.Context, id string) (*model.Balance, error) {
	balance, err := bs.repo.GetBalanceByUserID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &model.Balance{
		Balance: balance,
		UserID:  id,
	}, nil
}

func (bs *BalanceService) ChangeUserBalance(ctx context.Context, b dto.BalanceChangeRequest, depositType model.DepositType) (bm *dto.BalanceChangeRequest, err error) {
	balance, err := bs.repo.ChangeUserBalance(ctx, b, depositType)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (bs *BalanceService) TransferMoney(ctx context.Context, transfer dto.TransferRequest) (err error) {
	err = bs.repo.TransferMoney(ctx, transfer)
	if err != nil {
		return err
	}
	return nil
}
