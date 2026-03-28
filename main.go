package main

import (
	"github.com/rs/zerolog/log"

	"github.com/Lagwick/catalog-service/internal/app/config"
	rhealth "github.com/Lagwick/catalog-service/internal/app/handler/http/health"
	rprocessor "github.com/Lagwick/catalog-service/internal/app/processor/http"
)

func main() {
	config.Load()

	cfg := config.Root

	// Создание handlers
	hHealth := rhealth.NewHandler()

	// Создание и запуск HTTP сервера
	httpServer := rprocessor.NewHttp(hHealth, cfg.Processor.WebServer)
	if err := httpServer.Serve(); err != nil {
		log.Fatal().Err(err).Msg("HTTP server failed")
	}
}
