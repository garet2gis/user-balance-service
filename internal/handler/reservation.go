package handler

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"path"
	"user_balance_service/internal/apperror"
	"user_balance_service/internal/model"
	"user_balance_service/pkg/logging"
	"user_balance_service/pkg/utils"
)

const (
	basePathReservation = "/reservation/"
	reserve             = "/reserve/"
	confirm             = "/confirm/"
	cancel              = "/cancel/"
)

type ReservationService interface {
	ReserveMoney(ctx context.Context, rm model.Reservation) error
	CommitReservation(ctx context.Context, rm model.Reservation, status model.ReservationStatus) error
}

type reservationHandler struct {
	service  ReservationService
	logger   *logging.Logger
	validate *validator.Validate
}

func NewReservationHandler(s ReservationService, l *logging.Logger) Handler {
	return &reservationHandler{
		logger:   l,
		service:  s,
		validate: validator.New(),
	}
}

func (h *reservationHandler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, path.Join(basePathReservation, reserve), apperror.Middleware(h.Reserve, h.logger))
	router.HandlerFunc(http.MethodPost, path.Join(basePathReservation, confirm), apperror.Middleware(h.ConfirmReservation, h.logger))
	router.HandlerFunc(http.MethodPost, path.Join(basePathReservation, cancel), apperror.Middleware(h.CancelReservation, h.logger))
}

// Reserve godoc
// @Summary Резервация денег на услугу
// @ID      reservation-reserve
// @Param   reservation body model.Reservation true "Reservation"
// @Tags    Reservation
// @Success 204
// @Failure 400 {object} apperror.AppError
// @Failure 418 {object} apperror.AppError
// @Router  /reservation/reserve/ [post]
func (h *reservationHandler) Reserve(w http.ResponseWriter, r *http.Request) error {
	h.logger.Tracef("url:%s host:%s", r.URL, r.Host)
	w = utils.LogWriter{ResponseWriter: w}

	var reservation model.Reservation
	err := utils.DecodeJSON(w, r, &reservation)
	if err != nil {
		return toJSONDecodeError(err)
	}

	err = h.validate.Struct(reservation)
	err = validate(err)
	if err != nil {
		return err
	}

	err = h.service.ReserveMoney(context.Background(), reservation)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)

	return nil
}

// ConfirmReservation godoc
// @Summary Подтверждение списывания денег за услугу
// @ID      reservation-confirm
// @Param   reservation body model.Reservation true "Reservation"
// @Tags    Reservation
// @Success 204
// @Failure 400 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Failure 418 {object} apperror.AppError
// @Router  /reservation/confirm/ [post]
func (h *reservationHandler) ConfirmReservation(w http.ResponseWriter, r *http.Request) error {
	return h.commitReservation(w, r, model.Confirm)
}

// CancelReservation godoc
// @Summary Отмена резервации денег за услугу
// @ID      reservation-cancel
// @Param   reservation body model.Reservation true "Reservation"
// @Tags    Reservation
// @Success 204
// @Failure 400 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Failure 418 {object} apperror.AppError
// @Router  /reservation/cancel/ [post]
func (h *reservationHandler) CancelReservation(w http.ResponseWriter, r *http.Request) error {
	return h.commitReservation(w, r, model.Cancel)
}

func (h *reservationHandler) commitReservation(w http.ResponseWriter, r *http.Request, status model.ReservationStatus) error {
	h.logger.Tracef("url:%s host:%s", r.URL, r.Host)
	w = utils.LogWriter{ResponseWriter: w}

	var reservation model.Reservation
	err := utils.DecodeJSON(w, r, &reservation)
	if err != nil {
		return toJSONDecodeError(err)
	}

	err = h.validate.Struct(reservation)
	err = validate(err)
	if err != nil {
		return err
	}

	err = h.service.CommitReservation(context.Background(), reservation, status)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}
