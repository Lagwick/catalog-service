package main

import (
	"context"

	rcpostgres "github.com/Lagwick/catalog-service/internal/app/repository/conn/postgres"
	"github.com/rs/zerolog/log"

	"github.com/Lagwick/catalog-service/internal/app/config"
	rhealth "github.com/Lagwick/catalog-service/internal/app/handler/http/health"
	rprocessor "github.com/Lagwick/catalog-service/internal/app/processor/http"
)

func main() {
	ctx := context.Background()
	config.Load()
	cfg := config.Root

	// Подключение к PostgreSQL
	pgClient, err := rcpostgres.NewConn(ctx, cfg.Repository.Postgres)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}

	// Применение миграций
	oldVer, newVer, err := pgClient.Migrate(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	if oldVer != newVer {
		log.Info().
			Int64("old_version", oldVer).
			Int64("new_version", newVer).
			Msg("Database migrated")
	} else {
		log.Info().
			Int64("version", newVer).
			Msg("Database is up to date")
	}

	cfg = config.Root

	// Создание handlers
	hHealth := rhealth.NewHandler()

	// Создание и запуск HTTP сервера
	httpServer := rprocessor.NewHttp(hHealth, cfg.Processor.WebServer)
	if err := httpServer.Serve(); err != nil {
		log.Fatal().Err(err).Msg("HTTP server failed")
	}
}
