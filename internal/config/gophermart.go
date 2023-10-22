package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v9"
)

type GophermartConfig struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func (appConfig *GophermartConfig) Parse() {
	flag.StringVar(&appConfig.RunAddress, "a", "localhost:8080", "Base http address that server running on")
	flag.StringVar(&appConfig.DatabaseURI, "d", "", "Database uri")
	flag.StringVar(&appConfig.AccrualSystemAddress, "r", "http://localhost:8081", "Address of accrual system")
	flag.Parse()

	if err := env.Parse(appConfig); err != nil {
		fmt.Printf("%+v\n", err)
	}
}
