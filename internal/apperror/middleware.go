package apperror

import (
	"errors"
	"github.com/garet2gis/user_balance_service/pkg/logging"
	"github.com/garet2gis/user_balance_service/pkg/utils"
	"net/http"
)

type appHandler func(w http.ResponseWriter, r *http.Request) error

func Middleware(h appHandler, l *logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = utils.LogWriter{
			ResponseWriter: w,
		}
		var appErr *AppError

		err := h(w, r)

		if err != nil {
			l.Errorf("%v", err)
			w.Header().Set("Content-Type", "application/json")
			if errors.As(err, &appErr) {
				if errors.Is(err, ErrNotFound) {
					w.WriteHeader(http.StatusNotFound)
					w.Write(ErrNotFound.Marshal())
					return
				}

				w.WriteHeader(http.StatusBadRequest)
				w.Write(appErr.Marshal())
				return
			}
			w.WriteHeader(http.StatusTeapot)
			w.Write(systemError(err).Marshal())
		}
	}
}
