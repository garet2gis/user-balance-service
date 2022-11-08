package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"user_balance_service/pkg/logging"
	"user_balance_service/pkg/utils"
)

func insertTestDataInServicesTable(pool *pgxpool.Pool, logger *logging.Logger) {
	q := `
		INSERT INTO service (service_id, name)
		VALUES ('34e16535-480c-43f8-95a9-b7a503499afd', 'Услуга 1'),
				('bf13b3f8-503d-4e41-8f71-a541a20583e6', 'Услуга 2'),
				('b55e4e01-5152-4cb0-95f2-ee27d5d2e9cd', 'Услуга 3');
		`
	logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))
	_, err := pool.Exec(context.Background(), q)
	if err != nil {
		logger.Errorf(err.Error())
	}
}
