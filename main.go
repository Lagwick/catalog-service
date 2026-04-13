package main

import (
	"context"

	"github.com/Lagwick/catalog-service/internal/app/config"
	hcategory "github.com/Lagwick/catalog-service/internal/app/handler/http/category"
	hhealth "github.com/Lagwick/catalog-service/internal/app/handler/http/health"
	hproduct "github.com/Lagwick/catalog-service/internal/app/handler/http/product"

	rprocessor "github.com/Lagwick/catalog-service/internal/app/processor/http"

	pcategory "github.com/Lagwick/catalog-service/internal/app/repository/category"
	rcpostgres "github.com/Lagwick/catalog-service/internal/app/repository/conn/postgres"
	pproduct "github.com/Lagwick/catalog-service/internal/app/repository/product"

	scategory "github.com/Lagwick/catalog-service/internal/app/service/category"
	sproduct "github.com/Lagwick/catalog-service/internal/app/service/product"

	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()
	config.Load()
	cfg := config.Root

	pgClient, err := rcpostgres.NewConn(ctx, cfg.Repository.Postgres)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}

	oldVer, newVer, err := pgClient.Migrate(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	if oldVer != newVer {
		log.Info().
			Int64("old_version", oldVer).
			Int64("new_version", newVer).
			Msg("Database migrated")
	}

	categoryRepo := pcategory.NewRepoFromPostgres(pgClient)
	productRepo := pproduct.NewRepoFromPostgres(pgClient)

	categorySvc := scategory.NewService(categoryRepo, productRepo)
	productSvc := sproduct.NewService(productRepo, categoryRepo)

	healthHandler := hhealth.NewHandler()
	categoryHandler := hcategory.NewHandler(categorySvc)
	productHandler := hproduct.NewHandler(productSvc)

	server := rprocessor.NewHttp(
		healthHandler,
		categoryHandler,
		productHandler,
		cfg.Processor.WebServer,
	)

	if err := server.Serve(); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}
