package main

import (
	"context"
	"fmt"
	"github.com/garet2gis/user_balance_service/pkg/logging"
	"github.com/garet2gis/user_balance_service/pkg/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

func insertTestDataInServicesTable(pool *pgxpool.Pool, logger *logging.Logger) {
	q := `
		INSERT INTO service (service_id, name)
		VALUES ('34e16535-480c-43f8-95a9-b7a503499af0', 'Курьерская доставка'),
				('34e16535-480c-43f8-95a9-b7a503499af1', 'Бронирование'),
				('34e16535-480c-43f8-95a9-b7a503499af2', 'Дополнительная гарантия для товара');
		`
	logger.Trace(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))
	_, err := pool.Exec(context.Background(), q)
	if err != nil {
		logger.Errorf(err.Error())
	}
}
