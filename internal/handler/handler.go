package handler

import (
	"context"
	"github.com/julienschmidt/httprouter"
)

type Handler interface {
	Register(router *httprouter.Router)
}

type BalanceRepository interface {
	GetBalanceByUserID(ctx context.Context, id string) (float64, error)
}
