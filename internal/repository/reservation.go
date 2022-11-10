package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
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
	ReservationNotFound = errors.New("reservation not found")
)

type ReservationRepository struct {
	BalanceChanger
	client postgresql.Client
	logger *logging.Logger
}

func NewReservationRepository(c *pgxpool.Pool, l *logging.Logger) *ReservationRepository {
	return &ReservationRepository{
		BalanceChanger: *NewBalanceChanger(c, l),
		client:         c,
		logger:         l,
	}
}

func (r *ReservationRepository) getReservation(ctx context.Context, rm model.Reserve) (*model.ReserveDBModel, error) {
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

	var m model.ReserveDBModel

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

func (r *ReservationRepository) createReservation(ctx context.Context, rm model.Reserve) error {
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

func (r *ReservationRepository) createPreviousReservation(ctx context.Context, rm model.ReserveDBModel) error {
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

func (r *ReservationRepository) createCommitReservation(ctx context.Context, rm model.Reserve, status Status) error {
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

func (r *ReservationRepository) deleteReservation(ctx context.Context, rm model.Reserve) error {
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

func (r *ReservationRepository) ReserveMoney(ctx context.Context, rm model.Reserve) error {
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

func (r *ReservationRepository) CommitReservation(ctx context.Context, rm model.Reserve, status Status) error {
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
