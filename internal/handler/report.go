package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"time"
	"user_balance_service/internal/apperror"
	"user_balance_service/internal/dto"
	"user_balance_service/pkg/logging"
	"user_balance_service/pkg/utils"
)

const (
	report = "/report/"
)

type ReportService interface {
	GetReport(ctx context.Context, ro dto.ReportRequest) (*dto.ReportResponse, error)
}

type reportHandler struct {
	service  ReportService
	logger   *logging.Logger
	validate *validator.Validate
}

func NewReportHandler(s ReportService, l *logging.Logger) Handler {
	return &reportHandler{
		logger:   l,
		service:  s,
		validate: validator.New(),
	}
}

func (h *reportHandler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, report, apperror.Middleware(h.GetReport, h.logger))
}

// GetReport godoc
// @Summary     Получение ссылки на сформированный отчет
// @Description Отчет пересоздается только за текущий месяц
// @ID          get-report
// @Param       report body dto.ReportRequest true "Report options"
// @Tags        Report
// @Success     200 {array}  dto.ReportResponse
// @Failure     400 {object} apperror.AppError
// @Failure     418 {object} apperror.AppError
// @Router      /report/ [post]
func (h *reportHandler) GetReport(w http.ResponseWriter, r *http.Request) error {
	h.logger.Tracef("url:%s host:%s", r.URL, r.Host)
	w = utils.LogWriter{ResponseWriter: w}
	var ro dto.ReportRequest
	err := utils.DecodeJSON(w, r, &ro)
	if err != nil {
		return toJSONDecodeError(err)
	}

	err = h.validate.Struct(ro)
	err = validate(err)
	if err != nil {
		return err
	}

	// validate date
	year, month, _ := time.Now().Date()
	if ro.Year > year {
		return toValidateError(fmt.Errorf("year bigger than current"))
	}
	if ro.Year == year && ro.Month > int(month) {
		return toValidateError(fmt.Errorf("month bigger than current"))
	}

	reportPath, err := h.service.GetReport(context.Background(), ro)
	if err != nil {
		return err
	}

	reportPath.FileURL = fmt.Sprintf("%s/%s", r.Host, reportPath.FileURL)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response, err := json.Marshal(reportPath)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %+v", reportPath)
	}

	w.Write(response)

	return nil
}
