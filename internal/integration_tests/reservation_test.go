package integration_tests

import (
	"bytes"
	"context"
	"github.com/garet2gis/user_balance_service/internal/csv"
	h "github.com/garet2gis/user_balance_service/internal/handler"
	"github.com/garet2gis/user_balance_service/internal/repository"
	"github.com/garet2gis/user_balance_service/internal/service"
	"github.com/garet2gis/user_balance_service/pkg/logging"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
)

func TestCreateReservation(t *testing.T) {
	logger := logging.GetLogger()
	client, err := initTestDB()
	require.NoError(t, err, "Failed to connect to db")
	defer client.Close()

	r := repository.NewRepository(client, logger)
	rr := httptest.NewRecorder()
	router := httprouter.New()
	c := csv.NewBuilder(logger)
	s := service.NewService(r, c, logger)
	reservationHandler := h.NewReservationHandler(s, logger)
	reservationHandler.Register(router)

	var data = []byte(`
	{
		"cost": 21,
  		"order_id": "983e8792-6736-41bd-9f1a-7c67f8501645",
  		"service_id": "34e16535-480c-43f8-95a9-b7a503499af0",
  		"user_id": "7a13445c-d6df-4111-abc0-abb12f610068",
		"comment": "reserve 21"
	}`)

	b := bytes.NewBuffer(data)

	req, err := http.NewRequest(http.MethodPost, path.Join(h.BasePathReservation, h.Reserve), b)
	require.NoError(t, err, "Failed to create request")

	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code, "Wrong status")

	balance, err := r.GetBalanceByUserID(context.Background(), "7a13445c-d6df-4111-abc0-abb12f610068")
	require.NoError(t, err, "Failed to get existing balance")

	require.Equal(t, 100.0, balance, "Balance wrong reserve")
}

func TestFailedCreateReservation(t *testing.T) {
	logger := logging.GetLogger()
	client, err := initTestDB()
	require.NoError(t, err, "Failed to connect to db")
	defer client.Close()

	r := repository.NewRepository(client, logger)
	rr := httptest.NewRecorder()
	router := httprouter.New()
	c := csv.NewBuilder(logger)
	s := service.NewService(r, c, logger)
	reservationHandler := h.NewReservationHandler(s, logger)
	reservationHandler.Register(router)

	var data = []byte(`
	{
		"cost": 101,
  		"order_id": "983e8792-6736-41bd-9f1a-7c67f8501645",
  		"service_id": "34e16535-480c-43f8-95a9-b7a503499af0",
  		"user_id": "7a13445c-d6df-4111-abc0-abb12f610068",
		"comment": "reserve 101"
	}`)

	b := bytes.NewBuffer(data)

	req, err := http.NewRequest(http.MethodPost, path.Join(h.BasePathReservation, h.Reserve), b)
	require.NoError(t, err, "Failed to create request")

	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code, "Wrong status")

	balance, err := r.GetBalanceByUserID(context.Background(), "7a13445c-d6df-4111-abc0-abb12f610068")
	require.NoError(t, err, "Failed to get existing balance")

	require.Equal(t, 100.0, balance, "Balance wrong reserve")
}

func TestConfirmReservation(t *testing.T) {
	logger := logging.GetLogger()
	client, err := initTestDB()
	require.NoError(t, err, "Failed to connect to db")
	defer client.Close()

	r := repository.NewRepository(client, logger)
	rr := httptest.NewRecorder()
	router := httprouter.New()
	c := csv.NewBuilder(logger)
	s := service.NewService(r, c, logger)
	reservationHandler := h.NewReservationHandler(s, logger)
	reservationHandler.Register(router)

	var data = []byte(`
	{
		"cost": 21,
  		"order_id": "983e8792-6736-41bd-9f1a-7c67f8501645",
  		"service_id": "34e16535-480c-43f8-95a9-b7a503499af0",
  		"user_id": "7a13445c-d6df-4111-abc0-abb12f610068",
		"comment": "confirm 21"
	}`)

	b := bytes.NewBuffer(data)

	req, err := http.NewRequest(http.MethodPost, path.Join(h.BasePathReservation, h.Confirm), b)
	require.NoError(t, err, "Failed to create request")

	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code, "Wrong status")

	balance, err := r.GetBalanceByUserID(context.Background(), "7a13445c-d6df-4111-abc0-abb12f610068")
	require.NoError(t, err, "Failed to get existing balance")

	require.Equal(t, 100.0, balance, "Balance wrong confirm")
}

func TestCancelReservation(t *testing.T) {
	logger := logging.GetLogger()
	client, err := initTestDB()
	require.NoError(t, err, "Failed to connect to db")
	defer client.Close()

	r := repository.NewRepository(client, logger)
	rr := httptest.NewRecorder()
	router := httprouter.New()
	c := csv.NewBuilder(logger)
	s := service.NewService(r, c, logger)
	reservationHandler := h.NewReservationHandler(s, logger)
	reservationHandler.Register(router)

	var data = []byte(`
	{
		"cost": 50,
  		"order_id": "983e8792-6736-41bd-9f1a-7c67f8501645",
  		"service_id": "34e16535-480c-43f8-95a9-b7a503499af2",
  		"user_id": "7a13445c-d6df-4111-abc0-abb12f610068",
		"comment": "cancel 50"
	}`)

	b := bytes.NewBuffer(data)

	req, err := http.NewRequest(http.MethodPost, path.Join(h.BasePathReservation, h.Cancel), b)
	require.NoError(t, err, "Failed to create request")

	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code, "Wrong status")

	balance, err := r.GetBalanceByUserID(context.Background(), "7a13445c-d6df-4111-abc0-abb12f610068")
	require.NoError(t, err, "Failed to get existing balance")

	require.Equal(t, 150.0, balance, "Balance wrong cancel")
}
