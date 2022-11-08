package main

import (
	"context"
	"user_balance_service/internal/config"
	"user_balance_service/internal/model"
	"user_balance_service/internal/repository"
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

	br := repository.NewBalanceRepository(client, logger)

	id, err := br.GetBalanceByUserID(context.TODO(), "7a13445c-d6df-4111-abc0-abb12f610069")

	if err != nil {
		logger.Errorf("%v", err)
	} else {
		logger.Infof("balance: %f", id)
	}

	balance, err := br.ReplenishUserBalance(context.TODO(), model.BalanceModel{
		UserID: "7a13445c-d6df-4111-abc0-abb12f610069",
		Amount: 12,
	})

	if err != nil {
		logger.Errorf("%v", err)
	} else {
		logger.Infof("new balance: %f", balance.Amount)
	}
}
