package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"user_balance_service/pkg/logging"
	"user_balance_service/pkg/postgresql"
	"user_balance_service/pkg/utils"
)

type BalanceChanger struct {
	client postgresql.Client
	logger *logging.Logger
}

func NewBalanceChanger(c *pgxpool.Pool, l *logging.Logger) *BalanceChanger {
	return &BalanceChanger{
		client: c,
		logger: l,
	}
}

func (r *BalanceChanger) ChangeBalance(ctx context.Context, userID string, diff float64) (float64, error) {
	q := `
		UPDATE balance
    	SET balance= balance + $1
   		WHERE user_id = $2
    	RETURNING balance
		`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	var newBalance float64

	if err := r.client.QueryRow(ctx, q, diff, userID).Scan(&newBalance); err != nil {
		err = PgxErrorLog(err, r.logger)

		return 0, err
	}

	return newBalance, nil
}
