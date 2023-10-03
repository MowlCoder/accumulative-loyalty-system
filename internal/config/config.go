package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v9"
)

type AppConfig struct {
	BaseHTTPAddr   string `env:"SERVER_ADDRESS"`
	AppEnvironment string `env:"APP_ENV"`
	DatabaseDSN    string `env:"DATABASE_DSN"`
}

const (
	AppProductionEnv = "production"
	AppDevEnv        = "development"
)

func (appConfig *AppConfig) Parse() {
	flag.StringVar(&appConfig.BaseHTTPAddr, "a", "localhost:8080", "Base http address that server running on")
	flag.StringVar(&appConfig.DatabaseDSN, "d", "", "Database DSN")
	flag.Parse()

	if err := env.Parse(appConfig); err != nil {
		fmt.Printf("%+v\n", err)
	}
}
