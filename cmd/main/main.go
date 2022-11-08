package main

import (
	"context"
	"user_balance_service/internal/config"
	"user_balance_service/pkg/logging"
	"user_balance_service/pkg/postgresql"
)

func main() {
	logging.Init()
	logger := logging.GetLogger()
	cfg := config.GetConfig()

	client, err := postgresql.NewClient(context.Background(), 3, cfg.DBConfig, logger)
	if err != nil {
		logger.Fatalf("%v", err)
	}

	// Для тестирования нужна заполненная таблица услуг
	insertTestDataInServicesTable(client, logger)
}
