package main

import (
	"context"
	"fmt"
	"github.com/julienschmidt/httprouter"
	httpSwagger "github.com/swaggo/http-swagger"
	"net"
	"net/http"
	"time"
	"user_balance_service/internal/config"
	"user_balance_service/internal/handler"
	"user_balance_service/internal/repository"
	"user_balance_service/pkg/logging"
	"user_balance_service/pkg/postgresql"

	"user_balance_service/cmd/main/docs"
)

// @title   API User Balance Service
// @version 1.0.0

// @BasePath /

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

	r := repository.NewRepository(client, logger)

	router := httprouter.New()

	balanceHandler := handler.NewHandler(r, logger)
	balanceHandler.Register(router)

	host := fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.Port)

	swaggerInit(router, host)
	start(router, host)
}

func swaggerInit(router *httprouter.Router, host string) {
	docs.SwaggerInfo.Host = host
	router.Handler(http.MethodGet, "/swagger/*filename", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://%s/swagger/doc.json", host)),
	))
}

func start(router *httprouter.Router, host string) {
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

	logger.Fatal(server.Serve(listener))
}
