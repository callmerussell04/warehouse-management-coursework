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
	"warehouse-management-system/internal/middleware"
	"warehouse-management-system/internal/repository"
	"warehouse-management-system/internal/service"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func Run() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := loadConfig()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/warehouse_db?sslmode=disable"
	}

	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	db, err := newDB(dsn)
	if err != nil {
		logger.Error("failed to init db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}

	userRepository := repository.NewUserRepository(db, logger)
	otpRepository := repository.NewRedisOTPRepository(redisClient, logger)
	refreshSessionRepository := repository.NewRedisRefreshSessionRepository(redisClient)
	rateLimiter := repository.NewRedisRateLimiter(redisClient)
	productRepository := repository.NewProductRepository(db, logger)
	counterpartyRepository := repository.NewCounterpartyRepository(db, logger)
	orderRepository := repository.NewOrderRepository(db, logger)
	reportRepository := repository.NewReportRepository(db, logger)

	emailService := service.NewSMTPNotificationService(logger)

	userService := service.NewUserService(userRepository, otpRepository, emailService, refreshSessionRepository, logger, cfg.jwtSecret)

	productService := service.NewProductService(productRepository, logger)
	counterpartyService := service.NewCounterpartyService(counterpartyRepository, logger)
	orderService := service.NewOrderService(orderRepository, counterpartyService, productService, logger)
	reportService := service.NewReportService(reportRepository, logger)

	ctxInit, cancelInit := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelInit()

	if err := userService.EnsureAdminExists(ctxInit, cfg.adminUsername, cfg.adminEmail, cfg.adminPassword); err != nil {
		logger.Error("Failed to ensure admin user exists", "error", err)
		os.Exit(1)
	}

	userHandler := handler.NewUserHandler(userService, rateLimiter, cfg.cookieSecure, logger)
	productHandler := handler.NewProductHandler(productService, logger)
	counterpartyHandler := handler.NewCounterpartyHandler(counterpartyService, logger)
	orderHandler := handler.NewOrderHandler(orderService, logger)
	reportHandler := handler.NewReportHandler(reportService, logger)

	router := gin.Default()
	if err := router.SetTrustedProxies(cfg.trustedProxies); err != nil {
		logger.Error("failed to configure trusted proxies", "error", err)
		return
	}

	router.Use(middleware.LimitRequestBody())
	router.Use(middleware.CORSMiddleware(cfg.allowedOrigins))

	handler.InitRoutes(router, logger, cfg.jwtSecret, productHandler, counterpartyHandler, orderHandler, userHandler, reportHandler)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
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
