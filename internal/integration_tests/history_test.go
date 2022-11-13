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
	"testing"
)

func TestHistoryBalance(t *testing.T) {
	logger := logging.GetLogger()
	client, err := initTestDB()
	require.NoError(t, err, "Failed to connect to db")
	defer client.Close()

	r := repository.NewRepository(client, logger)
	rr := httptest.NewRecorder()
	router := httprouter.New()
	c := csv.NewBuilder(logger)
	s := service.NewService(r, c, logger)
	historyHandler := h.NewHistoryHandler(s, logger)
	historyHandler.Register(router)

	var bc = dto.BalanceChangeRequest{
		Amount:  120.22,
		UserID:  "7a13445c-d6df-4111-abc0-abb12f610063",
		Comment: "+120.12",
	}

	_, err = r.ChangeUserBalance(context.Background(), bc, model.Replenish)
	require.NoError(t, err, "Failed to replenish")

	var tr = dto.TransferRequest{
		Amount:     20.11,
		UserIDFrom: "7a13445c-d6df-4111-abc0-abb12f610063",
		UserIDTo:   "7a13445c-d6df-4111-abc0-abb12f610064",
		Comment:    "transfer",
	}

	err = r.TransferMoney(context.Background(), tr)
	require.NoError(t, err, "Failed to transfer")

	var res = model.Reservation{
		UserID:    "7a13445c-d6df-4111-abc0-abb12f610063",
		ServiceID: "34e16535-480c-43f8-95a9-b7a503499af1",
		OrderID:   "34e16535-480c-43f8-95a9-b7a503499a00",
		Cost:      30.11,
		Comment:   "reserve",
	}

	err = r.ReserveMoney(context.Background(), res)
	require.NoError(t, err, "Failed to reserve")

	var data = []byte(`
	{
		"user_id": "7a13445c-d6df-4111-abc0-abb12f610063"
	}`)

	b := bytes.NewBuffer(data)
	req, err := http.NewRequest(http.MethodGet, h.History, b)
	require.NoError(t, err, "Failed to create request")

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "Wrong status code")

	var history []HistoryRowTest
	err = json.NewDecoder(rr.Body).Decode(&history)
	require.NoError(t, err, "Failed to decode response")

	require.Equal(t, expectedHistory(), history, "Failed to get correct history")
}

type HistoryRowTest struct {
	OrderID         string  `json:"order_id,omitempty"`
	ServiceName     string  `json:"service_name,omitempty"`
	UserIDFrom      string  `json:"user_id_from,omitempty"`
	UserIDTo        string  `json:"user_id_to,omitempty"`
	Amount          float64 `json:"amount"`
	TransactionType string  `json:"transaction_type"`
	Comment         string  `json:"comment"`
}

func expectedHistory() []HistoryRowTest {
	return []HistoryRowTest{{
		OrderID:         "34e16535-480c-43f8-95a9-b7a503499a00",
		ServiceName:     "Бронирование",
		UserIDFrom:      "",
		UserIDTo:        "",
		Amount:          30.11,
		TransactionType: "reserve",
		Comment:         "reserve",
	}, {
		OrderID:         "",
		ServiceName:     "",
		UserIDFrom:      "",
		UserIDTo:        "7a13445c-d6df-4111-abc0-abb12f610064",
		Amount:          -20.11,
		TransactionType: "balance_change",
		Comment:         "transfer",
	}, {
		OrderID:         "",
		ServiceName:     "",
		UserIDFrom:      "",
		UserIDTo:        "",
		Amount:          120.22,
		TransactionType: "balance_change",
		Comment:         "+120.12",
	},
	}
}
