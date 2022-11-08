package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
	"user_balance_service/internal/model"
	"user_balance_service/pkg/logging"
	"user_balance_service/pkg/postgresql"
	"user_balance_service/pkg/utils"
)

type HistoryRepository struct {
	client postgresql.Client
	logger *logging.Logger
}

func NewHistoryRepository(c *pgxpool.Pool, l *logging.Logger) *HistoryRepository {
	return &HistoryRepository{
		client: c,
		logger: l,
	}
}

func (r *HistoryRepository) GetUserBalanceHistory(ctx context.Context, userID string) ([]model.HistoryRow, error) {
	q := `
		SELECT balance_history.order_id, 
		        balance_history.service_name,
		       	balance_history.create_date,
		       	balance_history.amount,
		       	balance_history.transaction_type
		FROM balance_history
		WHERE balance_history.user_id = $1
	`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	rows, err := r.client.Query(ctx, q, userID)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return nil, err
	}

	var historyRows []model.HistoryRow

	for rows.Next() {
		var row model.HistoryRow

		var createAt pgtype.Timestamp
		var orderID pgtype.UUID

		err = rows.Scan(&orderID, &row.ServiceName, &createAt, &row.Amount, &row.TransactionType)
		if err != nil {
			return nil, err
		}

		// перевод к московскому времени
		row.CreateAt = createAt.Time.Add(time.Hour * 3).String()
		if orderID.Valid {
			row.OrderID = utils.EncodeUUID(orderID)
		}

		historyRows = append(historyRows, row)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return historyRows, nil
}
