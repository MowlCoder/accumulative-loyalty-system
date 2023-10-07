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

	appConfig := &config.AccrualConfig{}
	appConfig.Parse()

	dbPool, err := postgresql.InitPool(appConfig.DatabaseURI)

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

	ctx := context.Background()

	calculateOrderAccrualWorker := workers.NewCalculateOrderAccrualWorker(
		registeredOrdersRepository,
		goodRewardRepository,
	)
	calculateOrderAccrualWorker.Start(ctx)

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

	log.Println("Accrual server is running on", appConfig.RunAddress)

	if err := http.ListenAndServe(appConfig.RunAddress, router); err != nil {
		log.Panic(err)
	}
}
