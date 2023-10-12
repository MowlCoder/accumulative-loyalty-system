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

	_ "github.com/MowlCoder/accumulative-loyalty-system/cmd/accrual/docs"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Println("No .env provided")
	}

	appConfig := &config.AccrualConfig{}
	appConfig.Parse()

	dbPool, err := postgresql.InitPool(appConfig.DatabaseURI)
	defer dbPool.Close()

	if err != nil {
		log.Panic(err)
	}

	err = postgresql.RunMigrations(appConfig.DatabaseURI)

	if err != nil {
		log.Panic(err)
	}

	goodRewardRepository := repositories.NewGoodRewardRepository(dbPool)
	registeredOrdersRepository := repositories.NewRegisteredOrdersRepository(dbPool)

	goodRewardsService := services.NewGoodRewardsService(goodRewardRepository)
	accrualOrdersService := services.NewAccrualOrdersService(registeredOrdersRepository)

	goodsHandler := handlers.NewGoodsHandler(goodRewardsService)
	accrualOrdersHandler := handlers.NewAccrualOrdersHandler(accrualOrdersService)

	workersCtx, workersStopCtx := context.WithCancel(context.Background())

	calculateOrderAccrualWorker := workers.NewCalculateOrderAccrualWorker(
		registeredOrdersRepository,
		goodRewardRepository,
	)
	go calculateOrderAccrualWorker.Start(workersCtx)

	server := &http.Server{
		Addr:    appConfig.RunAddress,
		Handler: makeRouter(appConfig, goodsHandler, accrualOrdersHandler),
	}

	log.Println("Accrual server is running on", appConfig.RunAddress)

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

// @title Gophermart Accrual Service
// @version 1.0
// @description Accrual service responsible for calculating accrual for registered orders
// @BasePath /api
func makeRouter(
	appConfig *config.AccrualConfig,
	goodsHandler *handlers.GoodsHandler,
	accrualOrdersHandler *handlers.AccrualOrdersHandler,
) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)

	router.Route("/api/goods", func(r chi.Router) {
		r.Post("/", goodsHandler.SaveNewGoodReward)
	})

	router.Route("/api/orders", func(r chi.Router) {
		r.Get("/{orderID}", middlewares.RateLimit(http.HandlerFunc(accrualOrdersHandler.GetRegisteredOrderInfo)))
		r.Post("/", accrualOrdersHandler.RegisterOrderForAccrual)
	})

	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://%s/swagger/doc.json", appConfig.RunAddress)),
	))

	return router
}
