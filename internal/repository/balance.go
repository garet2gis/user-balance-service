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

type Status string

const (
	Confirm Status = "confirm"
	Cancel         = "cancel"
)

var (
	BalanceNotFound     = errors.New("user's balance not found")
	NotEnoughMoney      = errors.New("not enough money on balance")
	ReservationNotFound = errors.New("reservation not found")
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
		if pgErr.Code == "23514" && pgErr.ConstraintName == "balance_balance_check" {
			return NotEnoughMoney
		}
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
			return 0, BalanceNotFound
		}

		err = PgxErrorLog(err, r.logger)
		return 0, err
	}

	return balance, nil
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

func (r *BalanceRepository) createReplenishment(ctx context.Context, b model.BalanceModel) error {
	q := `
		INSERT INTO replenishment (user_id, amount) 
		VALUES ($1, $2)
		`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	_, err := r.client.Exec(ctx, q, b.UserID, b.Amount)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return err
	}

	return nil
}

func (r *BalanceRepository) createReservation(ctx context.Context, rm model.ReserveModel) error {
	q := `
		INSERT INTO reservation (user_id, order_id, service_id, cost)
		VALUES ($1, $2, $3, $4)
		`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	_, err := r.client.Exec(ctx, q, rm.UserID, rm.OrderID, rm.ServiceID, rm.Cost)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return err
	}

	return nil
}

func (r *BalanceRepository) reduceBalance(ctx context.Context, rm model.ReserveModel) (float64, error) {
	q := `
		UPDATE balance
    	SET balance= balance - $1
   		WHERE user_id = $2
    	RETURNING balance
		`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	var newBalance float64

	if err := r.client.QueryRow(ctx, q, rm.Cost, rm.UserID).Scan(&newBalance); err != nil {
		err = PgxErrorLog(err, r.logger)

		return 0, err
	}

	return newBalance, nil
}

func (r *BalanceRepository) deleteReservation(ctx context.Context, rm model.ReserveModel) error {
	q := `
		DELETE
		FROM reservation
		WHERE user_id = $1
  		AND order_id = $2
  		AND service_id = $3
  		AND cost = $4
		`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	commandTag, err := r.client.Exec(ctx, q, rm.UserID, rm.OrderID, rm.ServiceID, rm.Cost)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return err
	}

	if commandTag.RowsAffected() != 1 {
		return ReservationNotFound
	}

	return nil
}

func (r *BalanceRepository) createCommitReservation(ctx context.Context, rm model.ReserveModel, status Status) error {
	q := `
		INSERT INTO commit_reservation (user_id, order_id, service_id, cost, status)
		VALUES ($1, $2, $3, $4, $5)
		`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	_, err := r.client.Exec(ctx, q, rm.UserID, rm.OrderID, rm.ServiceID, rm.Cost, status)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return err
	}

	return nil
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

	q := `
		UPDATE balance
    	SET balance= $1 + balance
   		WHERE user_id = $2
    	RETURNING balance
	`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	var newBalance float64

	if err = r.client.QueryRow(ctx, q, b.Amount, b.UserID).Scan(&newBalance); err != nil {
		err = PgxErrorLog(err, r.logger)
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

func (r *BalanceRepository) ReserveMoney(ctx context.Context, rm model.ReserveModel) error {
	conn, err := r.client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})

	if err != nil {
		return err
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

	_, err = r.reduceBalance(ctx, rm)
	if err != nil {
		return err
	}

	err = r.createReservation(ctx, rm)
	if err != nil {
		return err
	}

	return nil
}

func (r *BalanceRepository) CommitReservation(ctx context.Context, rm model.ReserveModel, status Status) error {
	conn, err := r.client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})

	if err != nil {
		return err
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

	err = r.deleteReservation(ctx, rm)
	if err != nil {
		return err
	}

	err = r.createCommitReservation(ctx, rm, status)
	if err != nil {
		return err
	}

	return nil
}

func (r *BalanceRepository) GetReport(ctx context.Context, year int, month int) ([]model.ReportRow, error) {
	q := `
		SELECT service.name, SUM(commit_reservation.cost) as "sum"
		FROM commit_reservation
        JOIN service USING (service_id)
		WHERE commit_reservation.status = 'confirm'
  			AND EXTRACT(YEAR FROM commit_reservation.created_at) = $1
  			AND EXTRACT(MONTH FROM commit_reservation.created_at) = $2
		GROUP BY service.name
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
