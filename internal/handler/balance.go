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
	"user_balance_service/internal/model"
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
	router.HandlerFunc(http.MethodPost, basePath, apperror.Middleware(h.UpdateBalance, h.logger))
}

type BalanceDTO struct {
	// Баланс пользователя
	Balance float64 `json:"balance"`
	// UUID баланса пользователя
	UserID string `json:"user_id"`
} // @name BalanceDTO

// GetBalance godoc
// @Summary Получение баланса пользователя
// @ID      get-balance
// @Param   user_id path string true "User ID" example(7a13445c-d6df-4111-abc0-abb12f610069)
// @Tags    Balance
// @Success 200 {object} BalanceDTO
// @Router  /balance/{user_id} [get]
func (h *handler) GetBalance(w http.ResponseWriter, r *http.Request) error {
	h.logger.Tracef("url:%s host:%s", r.URL, r.Host)
	w = utils.LogWriter{ResponseWriter: w}

	splitPath := strings.Split(r.URL.Path, "/")
	id := splitPath[len(splitPath)-1]

	newBalance, err := h.repo.GetBalanceByUserID(context.Background(), id)
	if err != nil {
		return err
	}

	b := &BalanceDTO{
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

type BalanceRequest struct {
	// Баланс пользователя
	Balance float64 `json:"balance"`
	// UUID баланса пользователя
	UserID string `json:"user_id"  example:"7a13445c-d6df-4111-abc0-abb12f610069"`
	// UUID баланса пользователя
	Comment string `json:"comment,omitempty"`
} // @name BalanceRequest

func NewBalanceRequest(b *model.Balance) *BalanceRequest {
	return &BalanceRequest{
		Balance: b.Amount,
		UserID:  b.UserID,
		Comment: b.Comment,
	}
}

func (br *BalanceRequest) ToModel() *model.Balance {
	return &model.Balance{
		Amount:  br.Balance,
		UserID:  br.UserID,
		Comment: br.Comment,
	}
}

// UpdateBalance godoc
// @Summary     Пополнение баланса пользователя
// @Description В случае обновления баланса ранее не упомянутого пользователя, он создается в БД
// @ID          post-balance
// @Param       balance body BalanceRequest true "User balance"
// @Tags        Balance
// @Success     200 {object} BalanceRequest
// @Router      /balance/ [post]
func (h *handler) UpdateBalance(w http.ResponseWriter, r *http.Request) error {
	h.logger.Tracef("url:%s host:%s", r.URL, r.Host)
	w = utils.LogWriter{ResponseWriter: w}

	var b BalanceRequest
	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		return fmt.Errorf("failed to decode new balance")
	}

	bm := b.ToModel()
	newBalance, err := h.repo.ReplenishUserBalance(context.Background(), *bm)
	if err != nil {
		return err
	}
	b = *NewBalanceRequest(newBalance)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("failed to marshal balance: %+v", b)
	}

	w.Write(response)

	return nil
}
