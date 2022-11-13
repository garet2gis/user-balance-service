package handler

import (
	"errors"
	"github.com/garet2gis/user_balance_service/internal/apperror"
	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
)

type Handler interface {
	Register(router *httprouter.Router)
}

func toJSONDecodeError(err error) error {
	return apperror.NewAppError(err, "JSON Decode Error", err.Error())
}

func toValidateError(err error) error {
	return apperror.NewAppError(err, "Validate error", err.Error())
}

func validate(err error) error {
	if err != nil {
		var invalid *validator.InvalidValidationError
		if errors.As(err, &invalid) {
			return err
		}
		return toValidateError(err)
	}
	return nil
}
