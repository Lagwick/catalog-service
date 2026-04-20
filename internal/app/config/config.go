package config

import (
	"log"

	"github.com/Lagwick/catalog-service/internal/app/config/section"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Repository section.Repository `split_words:"true"`
	Processor  section.Processor  `split_words:"true"`
	Monitor    section.Monitor    `split_words:"true"`
}

var Root Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found")
	}
	err := envconfig.Process("APP", &Root)
	if err != nil {
		log.Fatalf("config parse error: %v", err)
	}
}
