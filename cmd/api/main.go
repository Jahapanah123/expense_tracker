package main

import (
	"expense-tracker/internal/config"
	"expense-tracker/internal/db"
	"expense-tracker/internal/handler"
	"expense-tracker/internal/middleware"
	"expense-tracker/internal/repository"
	"expense-tracker/internal/services"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	pool := db.Connect(cfg)
	defer pool.Close()

	uow := db.NewUnitOfWork(pool)

	// user
	userRepo := repository.NewUserRepository(pool)
	userService := services.NewUserService(userRepo, uow)
	userHandler := handler.NewUserHandler(userService)

	// Expense
	expenseRepo := repository.NewExpenseRepository(pool)
	expenseService := services.NewExpenseService(expenseRepo, uow)
	expenseHandler := handler.NewExpenseHandler(expenseService)

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

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

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	slog.Info("server running", "port", cfg.Server.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server crashed", "error", err)
		os.Exit(1)
	}
}
