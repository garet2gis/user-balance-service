package handler

import (
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	"user_balance_service/internal/apperror"
	"user_balance_service/internal/dto"
)

type Handler interface {
	Register(router *httprouter.Router)
}

type BalanceRepository interface {
	GetBalanceByUserID(ctx context.Context, id string) (float64, error)
	ChangeUserBalance(ctx context.Context, b dto.BalanceRequest) (bm *dto.BalanceRequest, err error)
}

func toJSONDecodeError(err error) error {
	return apperror.NewAppError(err, "JSON Decode Error", err.Error())
}

func validate(err error) error {
	if err != nil {
		var invalid *validator.InvalidValidationError
		if errors.As(err, &invalid) {
			return err
		}
		return apperror.NewAppError(err, "Validate error", err.Error())
	}
	return nil
}
