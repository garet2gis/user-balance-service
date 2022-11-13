package service

import (
	"user_balance_service/internal/csv"
	"user_balance_service/internal/repository"
	"user_balance_service/pkg/logging"
)

type Service struct {
	BalanceService
	HistoryService
	ReservationService
	ReportService
}

func NewService(r *repository.Repository, csv *csv.Builder, l *logging.Logger) *Service {
	return &Service{
		BalanceService:     *NewBalanceService(r, l),
		HistoryService:     *NewHistoryService(r, l),
		ReservationService: *NewReservationService(r, l),
		ReportService:      *NewReportService(r, csv, l),
	}
}
