package integration_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/garet2gis/user_balance_service/internal/csv"
	"github.com/garet2gis/user_balance_service/internal/dto"
	h "github.com/garet2gis/user_balance_service/internal/handler"
	"github.com/garet2gis/user_balance_service/internal/model"
	"github.com/garet2gis/user_balance_service/internal/repository"
	"github.com/garet2gis/user_balance_service/internal/service"
	"github.com/garet2gis/user_balance_service/pkg/logging"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestReport(t *testing.T) {
	logger := logging.GetLogger()
	client, err := initTestDB()
	require.NoError(t, err, "Failed to connect to db")
	defer client.Close()

	r := repository.NewRepository(client, logger)
	rr := httptest.NewRecorder()
	router := httprouter.New()
	c := csv.NewBuilder(logger)
	s := service.NewService(r, c, logger)
	reportHandler := h.NewReportHandler(s, logger)
	reportHandler.Register(router)

	year, month, _ := time.Now().Date()

	var data = []byte(fmt.Sprintf(`
	{
		"year": %d,
  		"month": %d
	}`, year, int(month)))

	b := bytes.NewBuffer(data)
	req, err := http.NewRequest(http.MethodPost, h.Report, b)
	require.NoError(t, err, "Failed to create request")

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "Wrong status code")

	path := fmt.Sprintf("static/reports/%d_%d_report.csv", year, month)

	var reportPath dto.ReportResponse
	err = json.NewDecoder(rr.Body).Decode(&reportPath)
	require.NoError(t, err, "Failed to decode response")

	require.Equal(t, true, strings.HasSuffix(reportPath.FileURL, path), "Failed to get correct report")

	// check db
	report, err := r.GetReport(context.Background(), year, int(month))
	require.NoError(t, err, "Failed to get report from db")

	require.Equal(t, expectedReportRows(), report, "Failed to get report")

}

func expectedReportRows() []model.ReportRow {
	return []model.ReportRow{
		{
			ServiceName: "Бронирование",
			Cost:        "57.00",
		},
		{
			ServiceName: "Дополнительная гарантия для товара",
			Cost:        "70.74",
		},
		{
			ServiceName: "Курьерская доставка",
			Cost:        "120.78",
		},
	}
}
