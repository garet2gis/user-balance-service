package main

import (
	"context"
	"fmt"
	"github.com/julienschmidt/httprouter"
	httpSwagger "github.com/swaggo/http-swagger"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"
	"user_balance_service/internal/config"
	"user_balance_service/internal/handler"
	"user_balance_service/internal/repository"
	"user_balance_service/internal/service"
	"user_balance_service/pkg/logging"
	"user_balance_service/pkg/postgresql"

	"user_balance_service/cmd/main/docs"
)

// @title   API User Balance Service
// @version 1.0.0

// @BasePath /
// @produce  json

func main() {
	logging.Init()
	logger := logging.GetLogger()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		logger.Fatal(err)
	}
}

func run(ctx context.Context) error {
	logger := logging.GetLogger()
	cfg := config.GetConfig()

	client, err := postgresql.NewClient(context.Background(), 3, cfg.DBConfig, logger)
	if err != nil {
		return err
	}

	defer func() {
		client.Close()
		logger.Info("db shutdown gracefully")
	}()

	// Для тестирования нужна заполненная таблица услуг
	insertTestDataInServicesTable(client, logger)

	r := repository.NewRepository(client, logger)
	s := service.NewService(r, logger)

	router := httprouter.New()

	balanceHandler := handler.NewBalanceHandler(s, logger)
	balanceHandler.Register(router)

	historyHandler := handler.NewHistoryHandler(s, logger)
	historyHandler.Register(router)

	host := fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.Port)
	swaggerInit(router, host)
	startServer(ctx, router, host)

	return nil
}

func swaggerInit(router *httprouter.Router, host string) {
	docs.SwaggerInfo.Host = host
	router.Handler(http.MethodGet, "/swagger/*filename", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://%s/swagger/doc.json", host)),
	))
}

func startServer(ctx context.Context, router *httprouter.Router, host string) {
	logger := logging.GetLogger()

	listener, listenErr := net.Listen("tcp", host)
	logger.Infof("server is listening %s", host)
	if listenErr != nil {
		logger.Fatal(listenErr)
	}
	server := &http.Server{
		Handler:      router,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen and serve: %v", err)
		}
	}()

	// graceful shutdown
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Infof("error shutting down server %s", err)
	} else {
		logger.Info("server shutdown gracefully")
	}
}
