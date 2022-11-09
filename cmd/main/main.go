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

	br := repository.NewRepository(client, logger)

	id, err := br.BR.GetBalanceByUserID(context.TODO(), "7a13445c-d6df-4111-abc0-abb12f610069")

	if err != nil {
		logger.Errorf("%v", err)
	} else {
		logger.Infof("balance: %f", id)
	}

	balance, err := br.BR.ReplenishUserBalance(context.TODO(), model.BalanceDTO{
		UserID:  "7a13445c-d6df-4111-abc0-abb12f610069",
		Amount:  100,
		Comment: "Информация о зачислении",
	})

	if err != nil {
		logger.Errorf("%v", err)
	} else {
		logger.Infof("new balance: %f", balance.Amount)
	}

	err = br.BR.ReserveMoney(context.TODO(), model.ReserveDTO{
		UserID:    "7a13445c-d6df-4111-abc0-abb12f610069",
		ServiceID: "b55e4e01-5152-4cb0-95f2-ee27d5d2e9cd",
		OrderID:   "b55e4e01-5152-4cb0-95f2-ee27d5d2e9c1",
		Cost:      100,
		Comment:   "Информация о резерве денег на услугу",
	})
	if err != nil {
		logger.Errorf("%v", err)
	}

	err = br.BR.CommitReservation(context.TODO(), model.ReserveDTO{
		UserID:    "7a13445c-d6df-4111-abc0-abb12f610069",
		ServiceID: "b55e4e01-5152-4cb0-95f2-ee27d5d2e9cd",
		OrderID:   "b55e4e01-5152-4cb0-95f2-ee27d5d2e9c1",
		Comment:   "Информация о подтверждении списании денег",
		Cost:      100,
	}, repository.Confirm)
	if err != nil {
		logger.Errorf("%v", err)
	}

	report, err := br.RR.GetReport(context.TODO(), 2022, 11)
	if err != nil {
		logger.Errorf("%v", err)
	} else {
		logger.Infof("report: %v", report)
	}

	history, err := br.HR.GetUserBalanceHistory(context.TODO(), "7a13445c-d6df-4111-abc0-abb12f610069")
	if err != nil {
		logger.Errorf("%v", err)
	} else {
		logger.Infof("report: %+v", history)
	}
}
