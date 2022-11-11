package service

import (
	"context"
	"errors"
	"user_balance_service/internal/apperror"
	"user_balance_service/internal/dto"
	"user_balance_service/internal/model"
	"user_balance_service/pkg/logging"
)

type BalanceRepository interface {
	TransactionRepository
	GetBalanceByUserID(ctx context.Context, id string) (float64, error)
	CreateHistoryDeposit(ctx context.Context, b dto.BalanceChangeRequest) error
	CreateHistoryTransfer(ctx context.Context, b dto.TransferRequest) error
	CreateBalance(ctx context.Context, id string) error
	ChangeBalance(ctx context.Context, userID string, diff float64) (float64, error)
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

func (bs *BalanceService) ChangeUserBalance(ctx context.Context, b dto.BalanceChangeRequest, depositType model.DepositType) (bm *dto.BalanceChangeRequest, err error) {
	t, err := bs.repo.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			bs.repo.RollbackTransaction(ctx, t)
		} else {
			bs.repo.CommitTransaction(ctx, t)
		}
	}()

	_, err = bs.repo.GetBalanceByUserID(ctx, b.UserID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) && depositType == model.Replenish {
			err = bs.repo.CreateBalance(ctx, b.UserID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if depositType == model.Reduce {
		b.Amount = -b.Amount
	}

	newBalance, err := bs.repo.ChangeBalance(ctx, b.UserID, b.Amount)
	if err != nil {
		return nil, err
	}

	// записываем пополнение баланса в таблицу для отображения истории
	err = bs.repo.CreateHistoryDeposit(ctx, b)
	if err != nil {
		return nil, err
	}

	b.Amount = newBalance
	return &b, nil
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

func (bs *BalanceService) TransferMoney(ctx context.Context, transfer dto.TransferRequest) (err error) {
	t, err := bs.repo.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			bs.repo.RollbackTransaction(ctx, t)
		} else {
			bs.repo.CommitTransaction(ctx, t)
		}
	}()

	_, err = bs.repo.ChangeBalance(ctx, transfer.UserIDFrom, -transfer.Amount)
	if err != nil {
		return err
	}

	_, err = bs.repo.ChangeBalance(ctx, transfer.UserIDTo, transfer.Amount)
	if err != nil {
		return err
	}

	err = bs.repo.CreateHistoryTransfer(ctx, transfer)
	if err != nil {
		return err
	}

	return nil
}
