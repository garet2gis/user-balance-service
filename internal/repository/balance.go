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

var (
	BalanceNotFound = errors.New("user's balance not found")
	NotEnoughMoney  = errors.New("not enough money on balance")
)

type BalanceRepository struct {
	BalanceChanger
	client postgresql.Client
	logger *logging.Logger
}

func NewBalanceRepository(c *pgxpool.Pool, l *logging.Logger) *BalanceRepository {
	return &BalanceRepository{
		BalanceChanger: *NewBalanceChanger(c, l),
		client:         c,
		logger:         l,
	}
}

func PgxErrorLog(err error, l *logging.Logger) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		pgErr = err.(*pgconn.PgError)
		if pgErr.Code == "23514" && pgErr.ConstraintName == "balance_balance_check" {
			return NotEnoughMoney
		}
		newErr := fmt.Errorf("Code: %s, Message: %s, Where: %s, Detail: %s, SQLState: %s", pgErr.Code, pgErr.Message, pgErr.Where, pgErr.Detail, pgErr.SQLState())
		l.Error(newErr)
		return newErr
	}
	return err
}

func (r *BalanceRepository) createBalance(ctx context.Context, id string) error {
	q := `
		INSERT INTO balance (user_id)
		VALUES ($1)
	`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	_, err := r.client.Exec(ctx, q, id)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return err
	}
	return nil
}

func (r *BalanceRepository) createReplenishment(ctx context.Context, b model.Balance) error {
	q := `
		INSERT INTO replenishment (user_id, amount, comment) 
		VALUES ($1, $2, $3)
		`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	_, err := r.client.Exec(ctx, q, b.UserID, b.Amount, b.Comment)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return err
	}

	return nil
}

func (r *BalanceRepository) ReplenishUserBalance(ctx context.Context, b model.Balance) (bm *model.Balance, err error) {
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

	_, err = r.GetBalanceByUserID(ctx, b.UserID)
	if err != nil {
		if errors.Is(err, BalanceNotFound) {
			err = r.createBalance(ctx, b.UserID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	newBalance, err := r.changeBalance(ctx, b.UserID, b.Amount)
	if err != nil {
		return nil, err
	}

	// записываем пополнение баланса в таблицу replenishment для отображения истории
	err = r.createReplenishment(ctx, b)
	if err != nil {
		return nil, err
	}

	b.Amount = newBalance
	return &b, nil
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
			return 0, BalanceNotFound
		}

		err = PgxErrorLog(err, r.logger)
		return 0, err
	}

	return balance, nil
}
