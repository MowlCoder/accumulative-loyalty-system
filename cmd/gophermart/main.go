package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/config"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/handlers"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/middlewares"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/repositories"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/services"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/storage/postgresql"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/workers"

	_ "github.com/MowlCoder/accumulative-loyalty-system/docs/gophermart"
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
	defer dbPool.Close()

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

	workersCtx, workersStopCtx := context.WithCancel(context.Background())

	orderAccrualCheckingWorker := workers.NewOrderAccrualCheckingWorker(
		userOrderRepository,
		appConfig.AccrualSystemAddress,
	)
	go orderAccrualCheckingWorker.Start(workersCtx)

	server := &http.Server{
		Addr:    appConfig.RunAddress,
		Handler: makeRouter(appConfig, authHandler, balanceHandler, ordersHandler),
	}

	log.Println("Gophermart server is running on", appConfig.RunAddress)

	go func() {
		err = server.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sig

	log.Println("start graceful shutdown...")

	shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCtxCancel()

	go func() {
		<-shutdownCtx.Done()
		if shutdownCtx.Err() == context.DeadlineExceeded {
			log.Fatal("graceful shutdown timed out... forcing exit")
		}
	}()

	err = server.Shutdown(shutdownCtx)
	if err != nil {
		log.Fatal(err)
	}

	workersStopCtx()

	log.Println("graceful shutdown server successfully")
}

// @title Gophermart Loyalty Service
// @version 1.0
// @description Gophermart Loyalty Service responsible for saving user orders, saving user balance and withdraw balance
// @BasePath /api/user
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func makeRouter(
	appConfig *config.GophermartConfig,
	authHandler *handlers.AuthHandler,
	balanceHandler *handlers.BalanceHandler,
	ordersHandler *handlers.OrdersHandler,
) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.Recoverer)

	router.Group(func(publicRouter chi.Router) {
		publicRouter.Use(middleware.Logger)
		publicRouter.Use(middleware.Compress(5, "gzip"))

		publicRouter.Post("/register", authHandler.Register)
		publicRouter.Post("/login", authHandler.Login)
	})

	router.Group(func(authRouter chi.Router) {
		authRouter.Use(middlewares.AuthMiddleware)
		authRouter.Use(middleware.Logger)
		authRouter.Use(middleware.Compress(5, "gzip"))

		authRouter.Get("/orders", ordersHandler.GetOrders)
		authRouter.Post("/orders", ordersHandler.RegisterOrder)

		authRouter.Get("/balance", balanceHandler.GetUserBalance)
		authRouter.Post("/balance/withdraw", balanceHandler.WithdrawBalance)
		authRouter.Get("/withdrawals", balanceHandler.GetWithdrawalHistory)
	})

	router.Mount("/api/user", router)

	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://%s/swagger/doc.json", appConfig.RunAddress)),
	))

	return router
}
