package utils

import (
	"github.com/garet2gis/user_balance_service/pkg/logging"
	"net/http"
)

type LogWriter struct {
	http.ResponseWriter
}

func (w LogWriter) Write(p []byte) (n int, err error) {
	logger := logging.GetLogger()
	n, err = w.ResponseWriter.Write(p)
	if err != nil {
		logger.Errorf("Write failed: %v", err)
	}
	return
}
