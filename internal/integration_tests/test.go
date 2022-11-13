package integration_tests

import (
	"context"
	"github.com/garet2gis/user_balance_service/internal/config"
	"github.com/garet2gis/user_balance_service/pkg/logging"
	"github.com/garet2gis/user_balance_service/pkg/postgresql"
	"github.com/jackc/pgx/v5/pgxpool"
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
