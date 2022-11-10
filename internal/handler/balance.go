package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"path"
	"strings"
	"user_balance_service/internal/apperror"
	"user_balance_service/internal/dto"
	"user_balance_service/internal/model"
	"user_balance_service/pkg/logging"
	"user_balance_service/pkg/utils"
)

const (
	basePath  = "/balance/"
	balanceID = "/:id"
)

type handler struct {
	logger   *logging.Logger
	repo     BalanceRepository
	validate *validator.Validate
}

func NewHandler(r BalanceRepository, l *logging.Logger) Handler {

	return &handler{
		logger:   l,
		repo:     r,
		validate: validator.New(),
	}
}

func (h *handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, path.Join(basePath, balanceID), apperror.Middleware(h.GetBalance, h.logger))
	router.HandlerFunc(http.MethodPost, basePath, apperror.Middleware(h.UpdateBalance, h.logger))
}

// GetBalance godoc
// @Summary Получение баланса пользователя
// @ID      get-balance
// @Param   user_id path string true "User ID" default(7a13445c-d6df-4111-abc0-abb12f610069)
// @Tags    BalanceRequest
// @Success 200 {object} model.Balance
// @Failure 400 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Failure 418 {object} apperror.AppError
// @Router  /balance/{user_id} [get]
func (h *handler) GetBalance(w http.ResponseWriter, r *http.Request) error {
	h.logger.Tracef("url:%s host:%s", r.URL, r.Host)
	w = utils.LogWriter{ResponseWriter: w}

	splitPath := strings.Split(r.URL.Path, "/")
	id := splitPath[len(splitPath)-1]

	err := h.validate.Var(id, "required,uuid")
	err = validate(err)
	if err != nil {
		return err
	}

	newBalance, err := h.repo.GetBalanceByUserID(context.Background(), id)
	if err != nil {
		return err
	}

	b := &model.Balance{
		Balance: newBalance,
		UserID:  id,
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

// UpdateBalance godoc
// @Summary     Изменяет баланс пользователя
// @Description В случае обновления баланса ранее не упомянутого пользователя, он создается в БД
// @ID          post-balance
// @Param       balance body dto.BalanceRequest true "User balance"
// @Tags        BalanceRequest
// @Success     200 {object} dto.BalanceRequest
// @Failure     400 {object} apperror.AppError
// @Failure     418 {object} apperror.AppError
// @Router      /balance/ [post]
func (h *handler) UpdateBalance(w http.ResponseWriter, r *http.Request) error {
	h.logger.Tracef("url:%s host:%s", r.URL, r.Host)
	w = utils.LogWriter{ResponseWriter: w}

	var b dto.BalanceRequest
	err := utils.DecodeJSON(w, r, &b)
	if err != nil {
		return toJSONDecodeError(err)
	}

	err = h.validate.Struct(b)
	err = validate(err)
	if err != nil {
		return err
	}

	newBalance, err := h.repo.ChangeUserBalance(context.Background(), b)
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
