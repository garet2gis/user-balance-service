package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/garet2gis/user_balance_service/internal/apperror"
	"github.com/garet2gis/user_balance_service/internal/dto"
	"github.com/garet2gis/user_balance_service/internal/model"
	"github.com/garet2gis/user_balance_service/pkg/logging"
	"github.com/garet2gis/user_balance_service/pkg/postgresql"
	"github.com/garet2gis/user_balance_service/pkg/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	NotEnoughMoney = errors.New("not enough money on balance")
)

type BalanceRepository struct {
	TransactionHelper
	BalanceChanger
	client postgresql.Client
	logger *logging.Logger
}

func NewBalanceRepository(c *pgxpool.Pool, l *logging.Logger) *BalanceRepository {
	return &BalanceRepository{
		TransactionHelper: *NewTransactionHelper(c, l),
		BalanceChanger:    *NewBalanceChanger(c, l),
		client:            c,
		logger:            l,
	}
}

func (r *BalanceRepository) createBalance(ctx context.Context, tx pgx.Tx, id string) error {
	q := `
		INSERT INTO balance (user_id)
		VALUES ($1)
	`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	_, err := tx.Exec(ctx, q, id)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return err
	}
	return nil
}

func (r *BalanceRepository) createHistoryDeposit(ctx context.Context, tx pgx.Tx, b dto.BalanceChangeRequest) error {
	q := `
		INSERT INTO history_deposit (user_id, amount, comment) 
		VALUES ($1, $2, $3)
		`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	_, err := tx.Exec(ctx, q, b.UserID, b.Amount, b.Comment)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return err
	}

	return nil
}

func (r *BalanceRepository) createHistoryTransfer(ctx context.Context, tx pgx.Tx, b dto.TransferRequest) error {

	q := `
		INSERT INTO history_deposit (user_id, to_user_id, amount, comment) 
		VALUES ($1, $2 ,$3, $4)
		`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	_, err := tx.Exec(ctx, q, b.UserIDFrom, b.UserIDTo, -b.Amount, b.Comment)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return err
	}

	q = `
		INSERT INTO history_deposit (user_id, from_user_id, amount, comment) 
		VALUES ($1, $2 ,$3, $4)
		`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	_, err = tx.Exec(ctx, q, b.UserIDTo, b.UserIDFrom, b.Amount, b.Comment)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return err
	}

	return nil
}

func (r *BalanceRepository) GetBalanceByUserID(ctx context.Context, id string) (float64, error) {
	var fakeTx pgx.Tx
	return r.getBalanceByUserID(ctx, fakeTx, id)
}

func (r *BalanceRepository) getBalanceByUserID(ctx context.Context, tx pgx.Tx, id string) (float64, error) {
	q := `
		SELECT 
		       balance.balance
		FROM balance
		WHERE user_id = $1
	`
	r.logger.Infof("is not transaction: %t", tx == nil)

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	var err error
	var balance float64

	if tx == nil {
		err = r.client.QueryRow(ctx, q, id).Scan(&balance)
	} else {
		err = tx.QueryRow(ctx, q, id).Scan(&balance)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, apperror.ErrNotFound
		}

		err = PgxErrorLog(err, r.logger)
		return 0, err
	}

	return balance, nil
}

func (r *BalanceRepository) ChangeUserBalance(ctx context.Context, b dto.BalanceChangeRequest, depositType model.DepositType) (bm *dto.BalanceChangeRequest, err error) {
	t, err := r.beginTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			r.rollbackTransaction(ctx, t)
		} else {
			r.commitTransaction(ctx, t)
		}
	}()

	_, err = r.getBalanceByUserID(ctx, t, b.UserID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) && depositType == model.Replenish {
			err = r.createBalance(ctx, t, b.UserID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if depositType == model.Reduce {
		b.Amount = -b.Amount
	}

	newBalance, err := r.changeBalance(ctx, t, b.UserID, b.Amount)
	if err != nil {
		return nil, err
	}

	// записываем пополнение баланса в таблицу для отображения истории
	err = r.createHistoryDeposit(ctx, t, b)
	if err != nil {
		return nil, err
	}

	b.Amount = newBalance
	return &b, nil
}

func (r *BalanceRepository) TransferMoney(ctx context.Context, transfer dto.TransferRequest) (err error) {
	t, err := r.beginTransaction(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			r.rollbackTransaction(ctx, t)
		} else {
			r.commitTransaction(ctx, t)
		}
	}()

	_, err = r.changeBalance(ctx, t, transfer.UserIDFrom, -transfer.Amount)
	if err != nil {
		return err
	}

	_, err = r.changeBalance(ctx, t, transfer.UserIDTo, transfer.Amount)
	if err != nil {
		return err
	}

	err = r.createHistoryTransfer(ctx, t, transfer)
	if err != nil {
		return err
	}

	return nil
}
