package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"user_balance_service/internal/apperror"
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

func (r *ReservationRepository) CreateReservation(ctx context.Context, rm model.Reserve) error {
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

func (r *ReservationRepository) CreateCommitReservation(ctx context.Context, rm model.Reserve, status Status) error {
	q := `
		INSERT INTO history_reservation (user_id, order_id, service_id, cost, status, comment)
		VALUES ($1, $2, $3, $4, $5, $6)
		`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	_, err := r.client.Exec(ctx, q, rm.UserID, rm.OrderID, rm.ServiceID, rm.Cost, status, rm.Comment)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return err
	}

	return nil
}

func (r *ReservationRepository) DeleteReservation(ctx context.Context, rm model.Reserve) error {
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
		return apperror.ErrNotFound
	}

	return nil
}

// service

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

	_, err = r.ChangeBalance(ctx, rm.UserID, -rm.Cost)
	if err != nil {
		return err
	}

	err = r.CreateReservation(ctx, rm)
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

	err = r.DeleteReservation(ctx, rm)
	if err != nil {
		return err
	}

	if status == Confirm {
		rm.Cost = -rm.Cost
	}

	err = r.CreateCommitReservation(ctx, rm, status)
	if err != nil {
		return err
	}

	if status == Cancel {
		_, err = r.ChangeBalance(ctx, rm.UserID, rm.Cost)
		if err != nil {
			return err
		}
	}

	return nil
}
