package service

import (
	"context"
	"user_balance_service/internal/model"
	"user_balance_service/pkg/logging"
)

type ReservationRepository interface {
	ReserveMoney(ctx context.Context, rm model.Reservation) (err error)
	CommitReservation(ctx context.Context, rm model.Reservation, status model.ReservationStatus) (err error)
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

func (rs *ReservationService) ReserveMoney(ctx context.Context, rm model.Reservation) (err error) {
	err = rs.repo.ReserveMoney(ctx, rm)
	if err != nil {
		return err
	}
	return nil
}

func (rs *ReservationService) CommitReservation(ctx context.Context, rm model.Reservation, status model.ReservationStatus) (err error) {
	err = rs.repo.CommitReservation(ctx, rm, status)
	if err != nil {
		return err
	}
	return nil
}
