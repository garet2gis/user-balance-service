package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"user_balance_service/internal/apperror"
	"user_balance_service/internal/dto"
	"user_balance_service/pkg/logging"
	"user_balance_service/pkg/postgresql"
	"user_balance_service/pkg/utils"
)

var (
	NotEnoughMoney = errors.New("not enough money on balance")
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

func (r *BalanceRepository) createHistoryDeposit(ctx context.Context, b dto.BalanceRequest) error {
	q := `
		INSERT INTO history_deposit (user_id, amount, comment) 
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

func (r *BalanceRepository) ChangeUserBalance(ctx context.Context, b dto.BalanceRequest) (bm *dto.BalanceRequest, err error) {
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
		if errors.Is(err, apperror.ErrNotFound) {
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
	err = r.createHistoryDeposit(ctx, b)
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
			return 0, apperror.ErrNotFound
		}

		err = PgxErrorLog(err, r.logger)
		return 0, err
	}

	return balance, nil
}
