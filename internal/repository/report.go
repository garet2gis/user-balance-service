package repository

import (
	"context"
	"fmt"
	"github.com/garet2gis/user_balance_service/internal/model"
	"github.com/garet2gis/user_balance_service/pkg/logging"
	"github.com/garet2gis/user_balance_service/pkg/postgresql"
	"github.com/garet2gis/user_balance_service/pkg/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportRepository struct {
	client postgresql.Client
	logger *logging.Logger
}

func NewReportRepository(c *pgxpool.Pool, l *logging.Logger) *ReportRepository {
	return &ReportRepository{
		client: c,
		logger: l,
	}
}

func (r *ReportRepository) GetReport(ctx context.Context, year int, month int) ([]model.ReportRow, error) {
	q := `
		SELECT service.name, SUM(-history_reservation.cost) as "sum"
		FROM history_reservation
		JOIN service USING (service_id)
		WHERE history_reservation.status = 'confirm'
  			AND EXTRACT(YEAR FROM history_reservation.created_at) = $1
  			AND EXTRACT(MONTH FROM history_reservation.created_at) = $2
		GROUP BY service.name;
	`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	rows, err := r.client.Query(ctx, q, year, month)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return nil, err
	}

	var reportRows []model.ReportRow

	for rows.Next() {
		var row model.ReportRow

		err = rows.Scan(&row.ServiceName, &row.Cost)

		if err != nil {
			return nil, err
		}

		reportRows = append(reportRows, row)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reportRows, nil
}
