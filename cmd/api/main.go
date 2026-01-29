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
	"time"

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

	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	slog.Info("Connected to Database")

	defer pool.Close()

	// user
	userRepo := repository.NewUserRepository(pool)
	userService := services.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// Expense
	expenseRepo := repository.NewExpenseRepository(pool)
	expenseService := services.NewExpenseService(expenseRepo)
	expenseHandler := handler.NewExpenseHandler(expenseService)

	router := gin.Default()
	router.SetTrustedProxies(nil)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		slog.Info("health check endpoint called")
		c.JSON(200, gin.H{
			"Message": "I love you",
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
		userRoute.POST("/users/expenses", expenseHandler.AddExpenseHandler)
		userRoute.GET("/users/expenses", expenseHandler.GetAllExpenseHandler)
		userRoute.GET("/expenses/:id", expenseHandler.GetExpenseByIDHandler)
		userRoute.PUT("/expenses/:id", expenseHandler.UpdateExpenseHandler)
		userRoute.DELETE("/expenses/:id", expenseHandler.DeleteExpenseHandler)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		slog.Info("server running", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server crashed", "error", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info(" Shutting down")

	// Give active requests 5 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Forced shutdown", "error", err)
	}
	slog.Info("Server stopped cleanly")
}
