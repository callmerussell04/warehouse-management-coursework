package app

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"warehouse-management-system/internal/handler"
	"warehouse-management-system/internal/repository"
	"warehouse-management-system/internal/service"

	"github.com/gin-gonic/gin"
)

func Run() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/warehouse_db?sslmode=disable"
	}

	db, err := newDB(dsn)
	if err != nil {
		logger.Error("failed to init db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	productRepository := repository.NewProductRepository(db, logger)
	counterpartyRepository := repository.NewCounterpartyRepository(db, logger)
	orderRepository := repository.NewOrderRepository(db, logger)

	productService := service.NewProductService(productRepository, logger)
	counterpartyService := service.NewCounterpartyService(counterpartyRepository, logger)
	orderService := service.NewOrderService(orderRepository, counterpartyService, productService, logger)

	productHandler := handler.NewProductHandler(productService, logger)
	counterpartyHandler := handler.NewCounterpartyHandler(counterpartyService, logger)
	orderHandler := handler.NewOrderHandler(orderService, logger)

	router := gin.Default()

	handler.InitRoutes(router, productHandler, counterpartyHandler, orderHandler)

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		logger.Info("Starting server", "port", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}
	<-ctx.Done()
	logger.Info("Server exiting")
}

func newDB(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
