package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"user_balance_service/internal/model"
	"user_balance_service/pkg/logging"
	"user_balance_service/pkg/postgresql"
	"user_balance_service/pkg/utils"
)

type BalanceRepository struct {
	client postgresql.Client
	logger *logging.Logger
}

func NewBalanceRepository(c *pgxpool.Pool, l *logging.Logger) *BalanceRepository {
	return &BalanceRepository{
		client: c,
		logger: l,
	}
}

func PgxErrorLog(err error, l *logging.Logger) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		pgErr = err.(*pgconn.PgError)
		newErr := fmt.Errorf("Code: %s, Message: %s, Where: %s, Detail: %s, SQLState: %s", pgErr.Code, pgErr.Message, pgErr.Where, pgErr.Detail, pgErr.SQLState())
		l.Error(newErr)
		return newErr
	}
	return err
}

func (r *BalanceRepository) GetBalanceByUserID(ctx context.Context, id string) (float64, error) {
	q := `
		SELECT 
		       balance.balance
		FROM balance
		WHERE user_id = $1
	`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	var balance float64

	if err := r.client.QueryRow(ctx, q, id).Scan(&balance); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errors.New("user's balance not found")
		}

		err = PgxErrorLog(err, r.logger)
		return 0, err
	}

	return balance, nil
}

func (r *BalanceRepository) ReplenishUserBalance(ctx context.Context, b model.BalanceModel) (bm *model.BalanceModel, err error) {
	conn, err := r.client.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})

	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			errTx := tx.Rollback(ctx)
			if errTx != nil {
				r.logger.Errorf("transaction rollback failed")
			}
		} else {
			errTx := tx.Commit(ctx)
			if errTx != nil {
				r.logger.Errorf("transaction commit failed")
			}
		}
	}()

	prevBalance, err := r.GetBalanceByUserID(ctx, b.UserID)
	if err != nil {
		return nil, err
	}

	if prevBalance-b.Amount < 0 {
		return nil, errors.New("failed to update balance, balance must be positive")
	}

	q := `
		UPDATE balance
    	SET balance= $1 + balance
   		WHERE user_id = $2
    	RETURNING balance;
	`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	var newBalance float64

	if err = r.client.QueryRow(ctx, q, b.Amount, b.UserID).Scan(&newBalance); err != nil {
		err = PgxErrorLog(err, r.logger)
		return nil, err
	}

	// записываем пополнение баланса для отображения истории
	q = `
		INSERT INTO replenishment (user_id, amount) 
		VALUES ($1, $2)
		`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	_, err = r.client.Exec(ctx, q, b.UserID, b.Amount)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return nil, err
	}

	b.Amount = newBalance
	return &b, nil
}
