package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/config"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/facade"
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

	authHandler := handlers.NewAuthHandler(userService)
	balanceHandler := handlers.NewBalanceHandler(userService, withdrawalService)
	ordersHandler := handlers.NewOrdersHandler(ordersService)

	orderAccrualFacade := facade.NewOrderAccrualFacade(
		dbPool,
		userOrderRepository,
	)

	ctx := context.Background()

	orderAccrualCheckingWorker := workers.NewOrderAccrualCheckingWorker(
		userOrderRepository,
		orderAccrualFacade,
		appConfig.AccrualSystemAddress,
	)
	orderAccrualCheckingWorker.Start(ctx)

	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)

	router.Group(func(publicRouter chi.Router) {
		publicRouter.Post("/register", authHandler.Register)
		publicRouter.Post("/login", authHandler.Login)
	})

	router.Group(func(authRouter chi.Router) {
		authRouter.Use(middlewares.AuthMiddleware)

		authRouter.Get("/orders", ordersHandler.GetOrders)
		authRouter.Post("/orders", ordersHandler.RegisterOrder)

		authRouter.Get("/balance", balanceHandler.GetUserBalance)
		authRouter.Post("/balance/withdraw", balanceHandler.WithdrawBalance)
		authRouter.Get("/withdrawals", balanceHandler.GetWithdrawalHistory)
	})

	router.Mount("/api/user", router)

	log.Println("Gophermart server is running on", appConfig.RunAddress)

	if err := http.ListenAndServe(appConfig.RunAddress, router); err != nil {
		log.Panic(err)
	}
}
