package handler

import (
	"context"
	"github.com/julienschmidt/httprouter"
	"user_balance_service/internal/apperror"
	"user_balance_service/internal/model"
)

type Handler interface {
	Register(router *httprouter.Router)
}

type BalanceRepository interface {
	GetBalanceByUserID(ctx context.Context, id string) (float64, error)
	ReplenishUserBalance(ctx context.Context, b model.Balance) (bm *model.Balance, err error)
}

func toJSONDecodeError(err error) error {
	return apperror.NewAppError(err, "JSON Decode Error", err.Error())
}
