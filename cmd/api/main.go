package main

import (
	"context"
	"expense-tracker/internal/config"
	"expense-tracker/internal/db"
	"expense-tracker/internal/handler"
	"expense-tracker/internal/middleware"
	"expense-tracker/internal/repository"
	"expense-tracker/internal/services"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	slog.Info("Config loaded successfully")

	// ✅ Fix 1: Pass context and cfg.DB (not full cfg)
	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DB)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	slog.Info("Connected to Database")
	defer pool.Close()

	// user
	userRepo := repository.NewUserRepository()
	userService := services.NewUserService(userRepo, pool)
	userHandler := handler.NewUserHandler(userService)

	// Expense
	expenseRepo := repository.NewExpenseRepository()
	expenseService := services.NewExpenseService(expenseRepo, pool)
	expenseHandler := handler.NewExpenseHandler(expenseService)

	router := gin.Default()
	router.SetTrustedProxies(nil)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		slog.Info("health check endpoint called")
		c.JSON(200, gin.H{
			"status": "API is working",
		})
	})

	publicRoute := router.Group("/")
	{
		publicRoute.POST("/users/register", userHandler.CreateUserHandler)
		publicRoute.POST("/users/login", userHandler.LogInUserHandler)
	}

	userRoute := router.Group("/")
	userRoute.Use(middleware.AuthMiddleware())
	{
		userRoute.GET("/users/me", userHandler.GetUserHandler)
		userRoute.POST("/users/expenses", expenseHandler.CreateExpenseHandler)
		userRoute.GET("/users/expenses", expenseHandler.GetAllExpenseHandler)
		userRoute.GET("/expenses/:id", expenseHandler.GetExpenseByIDHandler)
		userRoute.PUT("/expenses/:id", expenseHandler.UpdateExpenseHandler)
		userRoute.DELETE("/expenses/:id", expenseHandler.DeleteExpenseHandler)
	}

	// ✅ Fix 2: Use cfg.HTTP.Port (not os.Getenv + fallback)
	port := cfg.HTTP.Port

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		slog.Info("server running", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server crashed", "error", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down")

	// Give active requests time to finish (use configured timeout)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Forced shutdown", "error", err)
	}
	slog.Info("Server stopped cleanly")
}
