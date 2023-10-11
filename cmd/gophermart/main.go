package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

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

	orderAccrualCheckingWorker := workers.NewOrderAccrualCheckingWorker(
		userOrderRepository,
		orderAccrualFacade,
		appConfig.AccrualSystemAddress,
	)
	orderAccrualCheckingWorker.Start(serverCtx)

	server := &http.Server{
		Addr:    appConfig.RunAddress,
		Handler: makeRouter(authHandler, balanceHandler, ordersHandler),
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("gracefully shutdown server")
		serverStopCtx()
	}()

	log.Println("Gophermart server is running on", appConfig.RunAddress)
	err = server.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-serverCtx.Done()
}

func makeRouter(
	authHandler *handlers.AuthHandler,
	balanceHandler *handlers.BalanceHandler,
	ordersHandler *handlers.OrdersHandler,
) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.Recoverer)

	router.Group(func(publicRouter chi.Router) {
		publicRouter.Use(middleware.Logger)

		publicRouter.Post("/register", authHandler.Register)
		publicRouter.Post("/login", authHandler.Login)
	})

	router.Group(func(authRouter chi.Router) {
		authRouter.Use(middlewares.AuthMiddleware)
		authRouter.Use(middleware.Logger)

		authRouter.Get("/orders", ordersHandler.GetOrders)
		authRouter.Post("/orders", ordersHandler.RegisterOrder)

		authRouter.Get("/balance", balanceHandler.GetUserBalance)
		authRouter.Post("/balance/withdraw", balanceHandler.WithdrawBalance)
		authRouter.Get("/withdrawals", balanceHandler.GetWithdrawalHistory)
	})

	router.Mount("/api/user", router)

	return router
}
