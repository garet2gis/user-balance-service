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

func (r *BalanceRepository) CreateBalance(ctx context.Context, id string) error {
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

func (r *BalanceRepository) CreateHistoryDeposit(ctx context.Context, b dto.BalanceChangeRequest) error {
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

func (r *BalanceRepository) CreateHistoryTransfer(ctx context.Context, b dto.TransferRequest) error {

	q := `
		INSERT INTO history_deposit (user_id, to_user_id, amount, comment) 
		VALUES ($1, $2 ,$3, $4)
		`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	_, err := r.client.Exec(ctx, q, b.UserIDFrom, b.UserIDTo, -b.Amount, b.Comment)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return err
	}

	q = `
		INSERT INTO history_deposit (user_id, from_user_id, amount, comment) 
		VALUES ($1, $2 ,$3, $4)
		`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	_, err = r.client.Exec(ctx, q, b.UserIDTo, b.UserIDFrom, b.Amount, b.Comment)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return err
	}

	return nil
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
