package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v9"
)

type AccrualConfig struct {
	RunAddress  string `env:"RUN_ADDRESS"`
	DatabaseURI string `env:"DATABASE_URI"`
}

func (appConfig *AccrualConfig) Parse() {
	flag.StringVar(&appConfig.RunAddress, "a", "localhost:8081", "Base http address that server running on")
	flag.StringVar(&appConfig.DatabaseURI, "d", "", "Database uri")
	flag.Parse()

	if err := env.Parse(appConfig); err != nil {
		fmt.Printf("%+v\n", err)
	}
}
