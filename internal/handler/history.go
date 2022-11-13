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
)

const (
	history = "/history/"
)

type HistoryService interface {
	GetHistory(ctx context.Context, bh dto.BalanceHistory) ([]model.HistoryRow, error)
}

type historyHandler struct {
	service  HistoryService
	logger   *logging.Logger
	validate *validator.Validate
}

func NewHistoryHandler(s HistoryService, l *logging.Logger) Handler {
	return &historyHandler{
		logger:   l,
		service:  s,
		validate: validator.New(),
	}
}

func (h *historyHandler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, history, apperror.Middleware(h.GetHistory, h.logger))
}

// GetHistory godoc
// @Summary Получение истории баланса пользователя
// @ID      get-balance-history
// @Param   user_id body dto.BalanceHistory true "User ID"
// @Tags    History
// @Success 200 {array}  model.HistoryRow
// @Failure 400 {object} apperror.AppError
// @Failure 404 {object} apperror.AppError
// @Failure 418 {object} apperror.AppError
// @Router  /history/ [get]
func (h *historyHandler) GetHistory(w http.ResponseWriter, r *http.Request) error {
	h.logger.Tracef("url:%s host:%s", r.URL, r.Host)
	w = utils.LogWriter{ResponseWriter: w}

	bh := dto.BalanceHistory{
		OrderBy:    "desc",
		OrderField: "create_date",
	}
	err := utils.DecodeJSON(w, r, &bh)
	if err != nil {
		return toJSONDecodeError(err)
	}

	err = h.validate.Struct(bh)
	err = validate(err)
	if err != nil {
		return err
	}

	b, err := h.service.GetHistory(context.Background(), bh)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %+v", b)
	}

	w.Write(response)

	return nil
}
