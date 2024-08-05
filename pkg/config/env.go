package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port    string `env:"PORT,required" envDefault:"8080"`
	DBHost  string `env:"DB_HOST,required" envDefault:"localhost"`
	DBPort  string `env:"DB_PORT,required" envDefault:"5432"`
	DBUser  string `env:"DB_USER,required" envDefault:"postgres"`
	DBPass  string `env:"DB_PASS,required" envDefault:"qwerty"`
	DBName  string `env:"DB_NAME,required" envDefault:"banking"`
	SSLMode bool   `env:"DB_SSL_MODE,required" envDefault:"false"`
}

func Parse() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	return &cfg, nil
}
