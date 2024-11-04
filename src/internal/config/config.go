package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"log"
	"time"
)

type Config struct {
	HttpServer
	Postgres
}

type HttpServer struct {
	ServerAddress string        `envconfig:"SERVER_ADDRESS" default:"0.0.0.0:8080"`
	Timeout       time.Duration `envconfig:"SERVER_TIMEOUT" default:"4s"`
	IdleTimeout   time.Duration `envconfig:"SERVER_IDLE_TIMEOUT" default:"60s"`
}

type Postgres struct {
	ConnURL  string `envconfig:"POSTGRES_CONN"`
	JDBCURL  string `envconfig:"POSTGRES_JDBC_URL"`
	Username string `envconfig:"POSTGRES_USERNAME"`
	Password string `envconfig:"POSTGRES_PASSWORD"`
	Host     string `envconfig:"POSTGRES_HOST"`
	Port     int    `envconfig:"POSTGRES_PORT"`
	Database string `envconfig:"POSTGRES_DATABASE"`
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading env variables", err)
	}

	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal("Error processing env variables:", err)
	}

	return &cfg
}
