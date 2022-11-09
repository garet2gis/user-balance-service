package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"path"
	"strings"
	"user_balance_service/internal/apperror"
	"user_balance_service/pkg/logging"
	"user_balance_service/pkg/utils"
)

const (
	basePath  = "/balance/"
	balanceID = "/:id"
)

type handler struct {
	logger *logging.Logger
	repo   BalanceRepository
}

func NewHandler(r BalanceRepository, l *logging.Logger) Handler {
	return &handler{
		logger: l,
		repo:   r,
	}
}

func (h *handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, path.Join(basePath, balanceID), apperror.Middleware(h.GetBalance, h.logger))
}

type GetBalanceResponse struct {
	// Баланс пользователя
	Balance float64 `json:"balance"`
	// UUID баланса пользователя
	UserID string `json:"user_id"`
}

// GetBalance godoc
// @Summary Получение баланса пользователя
// @ID      get-balance
// @Param   user_id path string true "User ID" default(7a13445c-d6df-4111-abc0-abb12f610069)
// @Tags    Balance
// @Success 200 {object} GetBalanceResponse
// @Router  /{user_id} [get]
func (h *handler) GetBalance(w http.ResponseWriter, r *http.Request) error {
	h.logger.Tracef("url:%s host:%s", r.URL, r.Host)
	w = utils.LogWriter{ResponseWriter: w}

	splitPath := strings.Split(r.URL.Path, "/")
	id := splitPath[len(splitPath)-1]

	newBalance, err := h.repo.GetBalanceByUserID(context.Background(), id)
	if err != nil {
		return err
	}

	b := &GetBalanceResponse{
		Balance: newBalance,
		UserID:  id,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("failed to marshal user, user: %+v", b)
	}

	w.Write(response)

	return nil
}
