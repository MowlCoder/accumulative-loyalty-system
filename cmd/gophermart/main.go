package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/config"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/handlers"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/middlewares"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/repositories"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/services"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/storage/postgresql"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/workers"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Println("No .env provided")
	}

	appConfig := &config.GophermartConfig{}
	appConfig.Parse()

	dbPool, err := postgresql.InitPool(appConfig.DatabaseURI)

	if err != nil {
		log.Panic(err)
	}

	err = postgresql.RunMigrations(appConfig.DatabaseURI)

	if err != nil {
		log.Panic(err)
	}

	userRepository := repositories.NewUserRepository(dbPool)
	balanceActionsRepository := repositories.NewBalanceActionsRepository(dbPool)
	userOrderRepository := repositories.NewUserOrderRepository(dbPool)

	userService := services.NewUserService(userRepository, balanceActionsRepository)
	ordersService := services.NewOrdersService(userOrderRepository)
	withdrawalService := services.NewWithdrawalsService(balanceActionsRepository)

	authHandler := handlers.NewAuthHandler(&handlers.AuthHandlerOptions{
		UserService: userService,
	})
	balanceHandler := handlers.NewBalanceHandler(&handlers.BalanceHandlerOptions{
		UserService:       userService,
		WithdrawalService: withdrawalService,
	})
	ordersHandler := handlers.NewOrdersHandler(&handlers.OrdersHandlerOptions{
		OrdersService: ordersService,
	})

	ctx := context.Background()

	orderAccrualCheckingWorker := workers.NewOrderAccrualCheckingWorker(
		userOrderRepository,
		balanceActionsRepository,
		appConfig.AccrualSystemAddress,
	)
	orderAccrualCheckingWorker.Start(ctx)

	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)

	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)

		r.Get("/orders", middlewares.AuthMiddleware(http.HandlerFunc(ordersHandler.GetOrders)))
		r.Post("/orders", middlewares.AuthMiddleware(http.HandlerFunc(ordersHandler.RegisterOrder)))

		r.Get("/balance", middlewares.AuthMiddleware(http.HandlerFunc(balanceHandler.GetUserBalance)))
		r.Post("/balance/withdraw", middlewares.AuthMiddleware(http.HandlerFunc(balanceHandler.WithdrawBalance)))
		r.Get("/withdrawals", middlewares.AuthMiddleware(http.HandlerFunc(balanceHandler.GetWithdrawalHistory)))
	})

	log.Println("Gophermart server is running on", appConfig.RunAddress)

	if err := http.ListenAndServe(appConfig.RunAddress, router); err != nil {
		log.Panic(err)
	}
}
