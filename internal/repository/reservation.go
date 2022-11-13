package repository

import (
	"context"
	"fmt"
	"github.com/garet2gis/user_balance_service/internal/apperror"
	"github.com/garet2gis/user_balance_service/internal/model"
	"github.com/garet2gis/user_balance_service/pkg/logging"
	"github.com/garet2gis/user_balance_service/pkg/postgresql"
	"github.com/garet2gis/user_balance_service/pkg/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReservationRepository struct {
	TransactionHelper
	BalanceChanger
	client postgresql.Client
	logger *logging.Logger
}

func NewReservationRepository(c *pgxpool.Pool, l *logging.Logger) *ReservationRepository {
	return &ReservationRepository{
		TransactionHelper: *NewTransactionHelper(c, l),
		BalanceChanger:    *NewBalanceChanger(c, l),
		client:            c,
		logger:            l,
	}
}

func (r *ReservationRepository) createReservation(ctx context.Context, tx pgx.Tx, rm model.Reservation) error {
	q := `
		INSERT INTO reservation (user_id, order_id, service_id, cost, comment)
		VALUES ($1, $2, $3, $4, $5)
		`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	_, err := tx.Exec(ctx, q, rm.UserID, rm.OrderID, rm.ServiceID, rm.Cost, rm.Comment)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return err
	}

	return nil
}

func (r *ReservationRepository) createCommitReservation(ctx context.Context, tx pgx.Tx, rm model.Reservation, status model.ReservationStatus) error {
	q := `
		INSERT INTO history_reservation (user_id, order_id, service_id, cost, status, comment)
		VALUES ($1, $2, $3, $4, $5, $6)
		`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	_, err := tx.Exec(ctx, q, rm.UserID, rm.OrderID, rm.ServiceID, rm.Cost, status, rm.Comment)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return err
	}

	return nil
}

func (r *ReservationRepository) deleteReservation(ctx context.Context, tx pgx.Tx, rm model.Reservation) error {
	q := `
		DELETE
		FROM reservation
		WHERE user_id = $1
  		AND order_id = $2
  		AND service_id = $3
  		AND cost = $4
		`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	commandTag, err := tx.Exec(ctx, q, rm.UserID, rm.OrderID, rm.ServiceID, rm.Cost)
	if err != nil {
		err = PgxErrorLog(err, r.logger)
		return err
	}

	if commandTag.RowsAffected() != 1 {
		return apperror.ErrNotFound
	}

	return nil
}

func (r *ReservationRepository) ReserveMoney(ctx context.Context, rm model.Reservation) (err error) {
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

	_, err = r.changeBalance(ctx, t, rm.UserID, -rm.Cost)
	if err != nil {
		return err
	}

	err = r.createReservation(ctx, t, rm)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReservationRepository) CommitReservation(ctx context.Context, rm model.Reservation, status model.ReservationStatus) (err error) {
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

	err = r.deleteReservation(ctx, t, rm)
	if err != nil {
		return err
	}

	if status == model.Confirm {
		rm.Cost = -rm.Cost
	}

	err = r.createCommitReservation(ctx, t, rm, status)
	if err != nil {
		return err
	}

	if status == model.Cancel {
		_, err = r.changeBalance(ctx, t, rm.UserID, rm.Cost)
		if err != nil {
			return err
		}
	}

	return nil
}
