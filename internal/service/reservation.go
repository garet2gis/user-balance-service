package service

import (
	"context"
	"user_balance_service/internal/model"
	"user_balance_service/pkg/logging"
)

type ReservationRepository interface {
	TransactionRepository
	CreateReservation(ctx context.Context, rm model.Reservation) error
	CreateCommitReservation(ctx context.Context, rm model.Reservation, status model.ReservationStatus) error
	DeleteReservation(ctx context.Context, rm model.Reservation) error
	ChangeBalance(ctx context.Context, userID string, diff float64) (float64, error)
}

type ReservationService struct {
	repo   ReservationRepository
	logger *logging.Logger
}

func NewReservationService(r ReservationRepository, l *logging.Logger) *ReservationService {
	return &ReservationService{
		repo:   r,
		logger: l,
	}
}

func (rs *ReservationService) ReserveMoney(ctx context.Context, rm model.Reservation) error {
	t, err := rs.repo.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			rs.repo.RollbackTransaction(ctx, t)
		} else {
			rs.repo.CommitTransaction(ctx, t)
		}
	}()

	_, err = rs.repo.ChangeBalance(ctx, rm.UserID, -rm.Cost)
	if err != nil {
		return err
	}

	err = rs.repo.CreateReservation(ctx, rm)
	if err != nil {
		return err
	}

	return nil
}

func (rs *ReservationService) CommitReservation(ctx context.Context, rm model.Reservation, status model.ReservationStatus) error {
	t, err := rs.repo.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			rs.repo.RollbackTransaction(ctx, t)
		} else {
			rs.repo.CommitTransaction(ctx, t)
		}
	}()

	err = rs.repo.DeleteReservation(ctx, rm)
	if err != nil {
		return err
	}

	if status == model.Confirm {
		rm.Cost = -rm.Cost
	}

	err = rs.repo.CreateCommitReservation(ctx, rm, status)
	if err != nil {
		return err
	}

	if status == model.Cancel {
		_, err = rs.repo.ChangeBalance(ctx, rm.UserID, rm.Cost)
		if err != nil {
			return err
		}
	}

	return nil
}
