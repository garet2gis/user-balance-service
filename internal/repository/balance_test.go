package repository

import (
	"context"
	"github.com/garet2gis/user_balance_service/internal/config"
	"github.com/garet2gis/user_balance_service/pkg/logging"
	"github.com/garet2gis/user_balance_service/pkg/postgresql"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"testing"
)

func initTestDB() (pool *pgxpool.Pool, err error) {
	logger := logging.GetLogger()

	cfg := config.DBConfig{
		DBPort:      "5432",
		DBHost:      "localhost",
		DBName:      "testdb",
		DBPassword:  "",
		DBUsername:  "postgres",
		AutoMigrate: false,
	}

	return postgresql.NewClient(context.Background(), 3, cfg, logger)
}

func TestMain(m *testing.M) {
	logging.Init()
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestGetBalance(t *testing.T) {
	logger := logging.GetLogger()
	client, err := initTestDB()
	if err != nil {
		return
	}
	defer client.Close()
	r := NewBalanceRepository(client, logger)

	balance, err := r.GetBalanceByUserID(context.Background(), "7a13445c-d6df-4111-abc0-abb12f610069")
	if err != nil {
		t.Error("Get a User Balance Failed")
	}

	if balance != 500.34 {
		t.Error("Balance did not return correct values.")
	}
}
