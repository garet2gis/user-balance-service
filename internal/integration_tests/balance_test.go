package integration_tests

import (
	"bytes"
	"context"
	"encoding/json"
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
	"os"
	"path"
	"testing"
)

func TestMain(m *testing.M) {
	logging.Init()
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestGetBalance(t *testing.T) {
	logger := logging.GetLogger()
	client, err := initTestDB()
	require.NoError(t, err, "Failed to connect to db")
	defer client.Close()

	r := repository.NewRepository(client, logger)
	rr := httptest.NewRecorder()
	router := httprouter.New()
	c := csv.NewBuilder(logger)
	s := service.NewService(r, c, logger)
	balanceHandler := h.NewBalanceHandler(s, logger)
	balanceHandler.Register(router)

	var data = []byte(`
	{
		"user_id": "7a13445c-d6df-4111-abc0-abb12f610062"
	}`)

	b := bytes.NewBuffer(data)

	req, err := http.NewRequest(http.MethodGet, h.BasePathBalance, b)
	require.NoError(t, err, "Failed to create request")

	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "Failed to get balance")

	var balance model.Balance
	err = json.NewDecoder(rr.Body).Decode(&balance)
	require.NoError(t, err, "Failed to decode response")

	var expectedBalance = model.Balance{
		Balance: 32.32,
		UserID:  "7a13445c-d6df-4111-abc0-abb12f610062",
	}

	require.Equal(t, expectedBalance, balance, "Failed to get correct balance")
}

func TestGetNonexistentBalance(t *testing.T) {
	logger := logging.GetLogger()
	client, err := initTestDB()
	require.NoError(t, err, "Failed to connect to db")
	defer client.Close()

	r := repository.NewRepository(client, logger)
	rr := httptest.NewRecorder()
	router := httprouter.New()
	c := csv.NewBuilder(logger)
	s := service.NewService(r, c, logger)
	balanceHandler := h.NewBalanceHandler(s, logger)
	balanceHandler.Register(router)

	var data = []byte(`
	{
		"user_id": "7a13445c-d6df-4111-abc0-abb12f610000"
	}`)

	b := bytes.NewBuffer(data)

	req, err := http.NewRequest(http.MethodGet, h.BasePathBalance, b)
	require.NoError(t, err, "Failed to create request")

	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code, "Failed to not found")
}

func TestCreateReplenishBalance(t *testing.T) {
	logger := logging.GetLogger()
	client, err := initTestDB()
	require.NoError(t, err, "Failed to connect to db")
	defer client.Close()

	r := repository.NewRepository(client, logger)
	rr := httptest.NewRecorder()
	router := httprouter.New()
	c := csv.NewBuilder(logger)
	s := service.NewService(r, c, logger)
	balanceHandler := h.NewBalanceHandler(s, logger)
	balanceHandler.Register(router)

	var data = []byte(`
	{
		"amount": 20.25,
  		"comment": "+20.25",
  		"user_id": "7a13445c-d6df-4111-abc0-abb12f610060"
	}`)

	b := bytes.NewBuffer(data)
	req, err := http.NewRequest(http.MethodPost, path.Join(h.BasePathBalance, h.Replenish), b)
	require.NoError(t, err, "Failed to create request")

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "Failed to get balance")

	expectedBalance := dto.BalanceChangeRequest{
		Amount:  20.25,
		UserID:  "7a13445c-d6df-4111-abc0-abb12f610060",
		Comment: "+20.25",
	}

	var balance dto.BalanceChangeRequest
	err = json.NewDecoder(rr.Body).Decode(&balance)
	require.NoError(t, err, "Failed to decode response")

	require.Equal(t, expectedBalance, balance, "Failed to correct replenish balance")
}

func TestFailedReplenishBalance(t *testing.T) {
	logger := logging.GetLogger()
	client, err := initTestDB()
	require.NoError(t, err, "Failed to connect to db")
	defer client.Close()

	r := repository.NewRepository(client, logger)
	rr := httptest.NewRecorder()
	router := httprouter.New()
	c := csv.NewBuilder(logger)
	s := service.NewService(r, c, logger)
	balanceHandler := h.NewBalanceHandler(s, logger)
	balanceHandler.Register(router)

	var data = []byte(`
	{
		"amount": -20.25,
  		"comment": "-20.25",
  		"user_id": "7a13445c-d6df-4111-abc0-abb12f610062"
	}`)

	b := bytes.NewBuffer(data)
	req, err := http.NewRequest(http.MethodPost, path.Join(h.BasePathBalance, h.Replenish), b)
	require.NoError(t, err, "Failed to create request")

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code, "Failed to get balance")

	balance, err := r.GetBalanceByUserID(context.Background(), "7a13445c-d6df-4111-abc0-abb12f610062")
	require.NoError(t, err, "Failed to get existing balance")
	require.Equal(t, 32.32, balance, "Balance wrong replenish")
}

func TestReduceBalance(t *testing.T) {
	logger := logging.GetLogger()
	client, err := initTestDB()
	require.NoError(t, err, "Failed to connect to db")
	defer client.Close()

	r := repository.NewRepository(client, logger)
	rr := httptest.NewRecorder()
	router := httprouter.New()
	c := csv.NewBuilder(logger)
	s := service.NewService(r, c, logger)
	balanceHandler := h.NewBalanceHandler(s, logger)
	balanceHandler.Register(router)

	var data = []byte(`
	{
		"amount": 100,
  		"comment": "-100",
  		"user_id": "7a13445c-d6df-4111-abc0-abb12f610069"
	}`)

	b := bytes.NewBuffer(data)
	req, err := http.NewRequest(http.MethodPost, path.Join(h.BasePathBalance, h.Reduce), b)
	require.NoError(t, err, "Failed to create request")

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "Failed to get balance")

	expectedBalance := dto.BalanceChangeRequest{
		Amount:  400.34,
		UserID:  "7a13445c-d6df-4111-abc0-abb12f610069",
		Comment: "-100",
	}

	var balance dto.BalanceChangeRequest
	err = json.NewDecoder(rr.Body).Decode(&balance)
	require.NoError(t, err, "Failed to decode response")

	require.Equal(t, expectedBalance, balance, "Failed to correct reduce balance")
}

func TestFailedReduceBalance(t *testing.T) {
	logger := logging.GetLogger()
	client, err := initTestDB()
	require.NoError(t, err, "Failed to connect to db")
	defer client.Close()

	r := repository.NewRepository(client, logger)
	rr := httptest.NewRecorder()
	router := httprouter.New()
	c := csv.NewBuilder(logger)
	s := service.NewService(r, c, logger)
	balanceHandler := h.NewBalanceHandler(s, logger)
	balanceHandler.Register(router)

	var data = []byte(`
	{
		"amount": 1000,
  		"comment": "-1000",
  		"user_id": "7a13445c-d6df-4111-abc0-abb12f610069"
	}`)

	b := bytes.NewBuffer(data)
	req, err := http.NewRequest(http.MethodPost, path.Join(h.BasePathBalance, h.Reduce), b)
	require.NoError(t, err, "Failed to create request")

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code, "Failed to get balance")

	// check what in db
	balance, err := r.GetBalanceByUserID(context.Background(), "7a13445c-d6df-4111-abc0-abb12f610069")
	require.NoError(t, err, "Failed to get existing balance")
	require.Equal(t, 400.34, balance, "Balance wrong replenish")

}

func TestTransferMoney(t *testing.T) {
	logger := logging.GetLogger()
	client, err := initTestDB()
	require.NoError(t, err, "Failed to connect to db")
	defer client.Close()

	r := repository.NewRepository(client, logger)
	rr := httptest.NewRecorder()
	router := httprouter.New()
	c := csv.NewBuilder(logger)
	s := service.NewService(r, c, logger)
	balanceHandler := h.NewBalanceHandler(s, logger)
	balanceHandler.Register(router)

	var data = []byte(`
	{
		"amount": 100,
  		"comment": "transfer 100",
  		"user_id_from": "7a13445c-d6df-4111-abc0-abb12f610069",
  		"user_id_to": "7a13445c-d6df-4111-abc0-abb12f610060"
	}`)

	b := bytes.NewBuffer(data)
	req, err := http.NewRequest(http.MethodPost, path.Join(h.BasePathBalance, h.Transfer), b)
	require.NoError(t, err, "Failed to create request")

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code, "Wrong status code")

	// check what in db
	balance, err := r.GetBalanceByUserID(context.Background(), "7a13445c-d6df-4111-abc0-abb12f610069")
	require.NoError(t, err, "Failed to get existing balance")

	require.Equal(t, 300.34, balance, "Balance wrong transfer reduce")

	balance, err = r.GetBalanceByUserID(context.Background(), "7a13445c-d6df-4111-abc0-abb12f610060")
	require.NoError(t, err, "Failed to get existing balance")

	require.Equal(t, 120.25, balance, "Balance wrong transfer replenish")
}

func TestFailedTransferMoney(t *testing.T) {
	logger := logging.GetLogger()
	client, err := initTestDB()
	require.NoError(t, err, "Failed to connect to db")
	defer client.Close()

	r := repository.NewRepository(client, logger)
	rr := httptest.NewRecorder()
	router := httprouter.New()
	c := csv.NewBuilder(logger)
	s := service.NewService(r, c, logger)
	balanceHandler := h.NewBalanceHandler(s, logger)
	balanceHandler.Register(router)

	var data = []byte(`
	{
		"amount": 10000,
  		"comment": "transfer 10000",
  		"user_id_from": "7a13445c-d6df-4111-abc0-abb12f610069",
  		"user_id_to": "7a13445c-d6df-4111-abc0-abb12f610060"
	}`)

	b := bytes.NewBuffer(data)
	req, err := http.NewRequest(http.MethodPost, path.Join(h.BasePathBalance, h.Transfer), b)
	require.NoError(t, err, "Failed to create request")

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code, "Wrong status code")

	// check what in db
	balance, err := r.GetBalanceByUserID(context.Background(), "7a13445c-d6df-4111-abc0-abb12f610069")
	require.NoError(t, err, "Failed to get existing balance")

	require.Equal(t, 300.34, balance, "Balance wrong transfer reduce")

	balance, err = r.GetBalanceByUserID(context.Background(), "7a13445c-d6df-4111-abc0-abb12f610060")
	require.NoError(t, err, "Failed to get existing balance")

	require.Equal(t, 120.25, balance, "Balance wrong transfer replenish")
}
