package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/garet2gis/user_balance_service/internal/apperror"
	"github.com/garet2gis/user_balance_service/internal/dto"
	"github.com/garet2gis/user_balance_service/internal/model"
	"github.com/garet2gis/user_balance_service/pkg/logging"
	"github.com/garet2gis/user_balance_service/pkg/utils"
	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"path"
)

const (
	BasePathBalance = "/balance/"
	Replenish       = "/replenish/"
	Reduce          = "/reduce/"
	Transfer        = "/transfer/"
)

type BalanceService interface {
	ChangeUserBalance(ctx context.Context, b dto.BalanceChangeRequest, depositType model.DepositType) (bm *dto.BalanceChangeRequest, err error)
	GetBalanceByUserID(ctx context.Context, id string) (*model.Balance, error)
	TransferMoney(ctx context.Context, transfer dto.TransferRequest) (err error)
}

type balanceHandler struct {
	logger   *logging.Logger
	service  BalanceService
	validate *validator.Validate
}

func NewBalanceHandler(s BalanceService, l *logging.Logger) Handler {
	return &balanceHandler{
		logger:   l,
		service:  s,
		validate: validator.New(),
	}
}

func (h *balanceHandler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, BasePathBalance, apperror.Middleware(h.GetBalance, h.logger))
	router.HandlerFunc(http.MethodPost, path.Join(BasePathBalance, Replenish), apperror.Middleware(h.ReplenishBalance, h.logger))
	router.HandlerFunc(http.MethodPost, path.Join(BasePathBalance, Reduce), apperror.Middleware(h.ReduceBalance, h.logger))
	router.HandlerFunc(http.MethodPost, path.Join(BasePathBalance, Transfer), apperror.Middleware(h.TransferBalance, h.logger))
}

// GetBalance godoc
// @Summary Получение баланса пользователя
// @ID      get-balance
// @Param   user_id body dto.BalanceGetRequest true "User ID"
// @Tags    Balance
// @Success 200 {object} model.Balance
// @Failure 400 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Failure 418 {object} apperror.AppError
// @Router  /balance/ [get]
func (h *balanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) error {
	h.logger.Tracef("url:%s host:%s", r.URL, r.Host)
	w = utils.LogWriter{ResponseWriter: w}

	var uID dto.BalanceGetRequest
	err := utils.DecodeJSON(w, r, &uID)
	if err != nil {
		return toJSONDecodeError(err)
	}

	err = h.validate.Struct(uID)
	err = validate(err)
	if err != nil {
		return err
	}

	b, err := h.service.GetBalanceByUserID(context.Background(), uID.UserID)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %+v", b)
	}

	w.Write(response)

	return nil
}

// ReplenishBalance godoc
// @Summary     Пополняет баланс пользователя
// @Description В случае пополнения баланса ранее не упомянутого пользователя, он создается в БД
// @ID          replenish-balance
// @Param       balance body dto.BalanceChangeRequest true "User balance"
// @Tags        Balance
// @Success     200 {object} dto.BalanceChangeRequest
// @Failure     400 {object} apperror.AppError
// @Failure     418 {object} apperror.AppError
// @Router      /balance/replenish/ [post]
func (h *balanceHandler) ReplenishBalance(w http.ResponseWriter, r *http.Request) error {
	return h.changeBalance(w, r, model.Replenish)
}

// ReduceBalance godoc
// @Summary     Уменьшает баланс пользователя
// @Description В случае уменьшения баланса ранее не упомянутого пользователя, он НЕ создается в БД (возвращается 404)
// @ID          reduce-balance
// @Param       balance body dto.BalanceChangeRequest true "User balance"
// @Tags        Balance
// @Success     200 {object} dto.BalanceChangeRequest
// @Failure     400 {object} apperror.AppError
// @Failure     418 {object} apperror.AppError
// @Router      /balance/reduce/ [post]
func (h *balanceHandler) ReduceBalance(w http.ResponseWriter, r *http.Request) error {
	return h.changeBalance(w, r, model.Reduce)
}

func (h *balanceHandler) changeBalance(w http.ResponseWriter, r *http.Request, depositType model.DepositType) error {
	h.logger.Tracef("url:%s host:%s", r.URL, r.Host)
	w = utils.LogWriter{ResponseWriter: w}

	var b dto.BalanceChangeRequest
	err := utils.DecodeJSON(w, r, &b)
	if err != nil {
		return toJSONDecodeError(err)
	}

	err = h.validate.Struct(b)
	err = validate(err)
	if err != nil {
		return err
	}

	newBalance, err := h.service.ChangeUserBalance(context.Background(), b, depositType)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response, err := json.Marshal(newBalance)
	if err != nil {
		return toJSONDecodeError(fmt.Errorf("failed to marshal balance: %+v", newBalance))
	}

	w.Write(response)

	return nil
}

// TransferBalance godoc
// @Summary Переводит деньги с одного счета на другой
// @ID      transfer-balance
// @Param   balance body dto.TransferRequest true "Transfer money"
// @Tags    Balance
// @Success 204
// @Failure 400 {object} apperror.AppError
// @Failure 418 {object} apperror.AppError
// @Router  /balance/transfer/ [post]
func (h *balanceHandler) TransferBalance(w http.ResponseWriter, r *http.Request) error {
	h.logger.Tracef("url:%s host:%s", r.URL, r.Host)
	w = utils.LogWriter{ResponseWriter: w}

	var b dto.TransferRequest
	err := utils.DecodeJSON(w, r, &b)
	if err != nil {
		return toJSONDecodeError(err)
	}

	err = h.validate.Struct(b)
	err = validate(err)
	if err != nil {
		return err
	}

	err = h.service.TransferMoney(context.Background(), b)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}
