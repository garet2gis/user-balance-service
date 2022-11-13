package postgresql

import (
	"context"
	"errors"
	"fmt"
	"github.com/garet2gis/user_balance_service/internal/config"
	"github.com/garet2gis/user_balance_service/pkg/logging"
	"github.com/golang-migrate/migrate/v4"
	pgxMigrate "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

func NewClient(ctx context.Context, maxAttempts int, sc config.DBConfig, logger *logging.Logger) (pool *pgxpool.Pool, err error) {
	var dsn string
	if sc.DBPassword == "" {
		dsn = fmt.Sprintf("postgresql://%s@%s:%s/%s", sc.DBUsername, sc.DBHost, sc.DBPort, sc.DBName)
	} else {
		dsn = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", sc.DBUsername, sc.DBPassword, sc.DBHost, sc.DBPort, sc.DBName)
	}
	err = DoWithTries(func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		pool, err = pgxpool.New(ctx, dsn)
		if err != nil {
			return err
		}

		return nil
	}, maxAttempts, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed: do with tries connect to postgresql, error: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed: ping postgresql, error: %w", err)
	}

	if sc.AutoMigrate {
		err = migrateUp(dsn, sc.DBName, logger)
		if err != nil {
			return
		}
	}

	return pool, nil
}

func migrateUp(dsn, dbName string, logger *logging.Logger) error {
	p := pgxMigrate.Postgres{}

	d, err := p.Open(dsn)
	if err != nil {
		return fmt.Errorf("failed: connect to database and migrate due to error: %w", err)
	}
	defer func() {
		if err = d.Close(); err != nil {
			logger.Fatalf("failed: close connection(migrations) due to error: %v", err)
		}
	}()

	m, err := migrate.NewWithDatabaseInstance("file://./migrations", dbName, d)
	if err != nil {
		return fmt.Errorf("failed: connect to database and migrate due to error: %w", err)
	}

	logger.Tracef("migrations: %+v", *m)
	err = m.Up()

	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("no changes in migrations")
			return nil
		}

		return fmt.Errorf("failed: do migrate up due to error: %w", err)
	}

	return nil
}

func DoWithTries(fn func() error, attempts int, delay time.Duration) (err error) {
	for attempts > 0 {
		if err = fn(); err != nil {
			time.Sleep(delay)
			attempts--
			continue
		}
		return nil
	}
	return
}
