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
	Confirm    Status = "confirm"
	Cancel            = "cancel"
	WasReserve        = "was_reserve"
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

func (r *BalanceRepository) getReservation(ctx context.Context, rm model.ReserveDTO) (*model.ReserveModel, error) {
	q := `
		SELECT reservation_id,
			user_id,
			order_id,
			service_id,
			cost,
			created_at,
			comment
		FROM reservation
		WHERE user_id = $1
  			AND order_id = $2
  			AND service_id = $3
  			AND cost = $4
	`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	var m model.ReserveModel

	s := r.client.QueryRow(ctx, q, rm.UserID, rm.OrderID, rm.ServiceID, rm.Cost)
	err := s.Scan(&m.ReservationID, &m.UserID, &m.OrderID, &m.ServiceID, &m.Cost, &m.CreatedAt, &m.Comment)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ReservationNotFound
		}

		err = PgxErrorLog(err, r.logger)
		return nil, err
	}

	return &m, nil
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

func (r *BalanceRepository) createReplenishment(ctx context.Context, b model.BalanceDTO) error {
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

func (r *BalanceRepository) createReservation(ctx context.Context, rm model.ReserveDTO) error {
	q := `
		INSERT INTO reservation (user_id, order_id, service_id, cost, comment)
		VALUES ($1, $2, $3, $4, $5)
		`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	_, err := r.client.Exec(ctx, q, rm.UserID, rm.OrderID, rm.ServiceID, rm.Cost, rm.Comment)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return err
	}

	return nil
}

func (r *BalanceRepository) changeBalance(ctx context.Context, userID string, diff float64) (float64, error) {
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

func (r *BalanceRepository) deleteReservation(ctx context.Context, rm model.ReserveDTO) error {
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

func (r *BalanceRepository) createPreviousReservation(ctx context.Context, rm model.ReserveModel) error {
	q := `
		INSERT INTO history_reservation (user_id, order_id, service_id, cost, status, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	_, err := r.client.Exec(ctx, q, rm.UserID, rm.OrderID, rm.ServiceID, -rm.Cost, WasReserve, rm.Comment, rm.CreatedAt)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return err
	}

	return nil
}

func (r *BalanceRepository) createCommitReservation(ctx context.Context, rm model.ReserveDTO, status Status) error {
	q := `
		INSERT INTO history_reservation (user_id, order_id, service_id, cost, status, comment)
		VALUES ($1, $2, $3, $4, $5, $6)
		`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	if status == Confirm {
		rm.Cost = 0
	}

	_, err := r.client.Exec(ctx, q, rm.UserID, rm.OrderID, rm.ServiceID, rm.Cost, status, rm.Comment)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return err
	}

	return nil
}

func (r *BalanceRepository) ReplenishUserBalance(ctx context.Context, b model.BalanceDTO) (bm *model.BalanceDTO, err error) {
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

func (r *BalanceRepository) ReserveMoney(ctx context.Context, rm model.ReserveDTO) error {
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

	_, err = r.changeBalance(ctx, rm.UserID, -rm.Cost)
	if err != nil {
		return err
	}

	err = r.createReservation(ctx, rm)
	if err != nil {
		return err
	}

	return nil
}

func (r *BalanceRepository) CommitReservation(ctx context.Context, rm model.ReserveDTO, status Status) error {
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

	reservation, err := r.getReservation(ctx, rm)
	if err != nil {
		return err
	}

	err = r.createPreviousReservation(ctx, *reservation)
	if err != nil {
		return err
	}

	err = r.deleteReservation(ctx, rm)
	if err != nil {
		return err
	}

	err = r.createCommitReservation(ctx, rm, status)
	if err != nil {
		return err
	}

	if status == Cancel {
		_, err = r.changeBalance(ctx, rm.UserID, rm.Cost)
		if err != nil {
			return err
		}
	}

	return nil
}
