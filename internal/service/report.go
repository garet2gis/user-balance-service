package service

import (
	"context"
	"github.com/garet2gis/user_balance_service/internal/apperror"
	"github.com/garet2gis/user_balance_service/internal/dto"
	"github.com/garet2gis/user_balance_service/internal/model"
	"github.com/garet2gis/user_balance_service/pkg/logging"
	"time"
)

type ReportRepository interface {
	GetReport(ctx context.Context, year int, month int) ([]model.ReportRow, error)
}

type CSVBuilder interface {
	CreateReport(rows []model.ReportRow, year, month int) (string, error)
	IsCreated(year, month int) (bool, string, error)
}

type ReportService struct {
	repo       ReportRepository
	csvBuilder CSVBuilder
	logger     *logging.Logger
}

func NewReportService(r ReportRepository, csv CSVBuilder, l *logging.Logger) *ReportService {
	return &ReportService{
		repo:       r,
		logger:     l,
		csvBuilder: csv,
	}
}

func (rs *ReportService) GetReport(ctx context.Context, ro dto.ReportRequest) (*dto.ReportResponse, error) {
	year, month, _ := time.Now().Date()

	isRecreate := false
	if year == ro.Year && int(month) == ro.Month {
		isRecreate = true
	}

	isCreated, createdFilePath, err := rs.csvBuilder.IsCreated(ro.Year, ro.Month)
	if err != nil {
		return nil, err
	}

	if isCreated && !isRecreate {
		return &dto.ReportResponse{
			FileURL: createdFilePath,
		}, nil
	}

	reportRows, err := rs.repo.GetReport(ctx, ro.Year, ro.Month)
	if err != nil {
		return nil, err
	}
	if len(reportRows) == 0 {
		return nil, apperror.ErrNotFound
	}

	reportPath, err := rs.csvBuilder.CreateReport(reportRows, ro.Year, ro.Month)
	if err != nil {
		return nil, err
	}
	return &dto.ReportResponse{
		FileURL: reportPath,
	}, nil
}
