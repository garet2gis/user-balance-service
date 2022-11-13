package repository

import (
	"context"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"user_balance_service/internal/dto"
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

func (r *HistoryRepository) GetUserBalanceHistory(ctx context.Context, bh dto.BalanceHistory) ([]model.HistoryRow, error) {
	qb := sq.Select("order_id, service_name, from_user_id, to_user_id, create_date, amount, transaction_type, comment").
		From("balance_history").
		Where(sq.Eq{"user_id": bh.UserID}).PlaceholderFormat(sq.Dollar).
		OrderBy(fmt.Sprintf("%s %s", bh.OrderField, bh.OrderBy))

	if bh.Limit > 0 {
		qb = qb.Limit(uint64(bh.Limit))
	}

	if bh.Offset > 0 {
		qb = qb.Offset(uint64(bh.Offset))
	}

	q, i, err := qb.ToSql()
	if err != nil {
		return nil, err
	}

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	rows, err := r.client.Query(ctx, q, i...)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return nil, err
	}

	var historyRows []model.HistoryRow

	for rows.Next() {
		var row model.HistoryRow

		var createAt pgtype.Timestamp
		var orderID pgtype.UUID
		var UserIDFrom pgtype.UUID
		var UserIDTo pgtype.UUID

		err = rows.Scan(&orderID, &row.ServiceName, &UserIDFrom, &UserIDTo, &createAt, &row.Amount, &row.TransactionType, &row.Comment)
		if err != nil {
			return nil, err
		}

		row.CreateAt = createAt.Time.String()
		if orderID.Valid {
			row.OrderID = utils.EncodeUUID(orderID)
		}
		if UserIDFrom.Valid {
			row.UserIDFrom = utils.EncodeUUID(UserIDFrom)
		}
		if UserIDTo.Valid {
			row.UserIDTo = utils.EncodeUUID(UserIDTo)
		}

		historyRows = append(historyRows, row)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return historyRows, nil
}
