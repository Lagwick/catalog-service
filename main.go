package main

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/Lagwick/catalog-service/internal/app/config"
	"github.com/Lagwick/catalog-service/internal/app/entity"
	rhealth "github.com/Lagwick/catalog-service/internal/app/handler/http/health"
	rprocessor "github.com/Lagwick/catalog-service/internal/app/processor/http"
	pcategory "github.com/Lagwick/catalog-service/internal/app/repository/category"
	rcpostgres "github.com/Lagwick/catalog-service/internal/app/repository/conn/postgres"
	pproduct "github.com/Lagwick/catalog-service/internal/app/repository/product"
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
	} else {
		log.Info().
			Int64("version", newVer).
			Msg("Database is up to date")
	}

	// === Проверка репозиториев ===

	categoryRepo := pcategory.NewRepoFromPostgres(pgClient)
	productRepo := pproduct.NewRepoFromPostgres(pgClient)

	cat := entity.Category{
		GUID:      uuid.New(),
		Name:      "Электроника",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := categoryRepo.Create(ctx, cat); err != nil {
		log.Fatal().Err(err).Msg("Category create failed")
	}

	log.Info().
		Str("guid", cat.GUID.String()).
		Msg("Category created")

	foundCategory, err := categoryRepo.GetByGUID(ctx, cat.GUID)
	if err != nil {
		log.Fatal().Err(err).Msg("Category GetByGUID failed")
	}

	log.Info().
		Str("name", foundCategory.Name).
		Msg("Category found")

	foundCategory.Name = "Бытовая техника"
	foundCategory.UpdatedAt = time.Now()

	if err := categoryRepo.Update(ctx, foundCategory); err != nil {
		log.Fatal().Err(err).Msg("Category update failed")
	}

	log.Info().Msg("Category updated")

	categories, err := categoryRepo.List(ctx, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Category list failed")
	}

	log.Info().
		Int("count", len(categories)).
		Msg("All categories")

	desc := "Мощный пылесос"

	product := entity.Product{
		GUID:         uuid.New(),
		Name:         "Пылесос Dyson V15",
		Description:  &desc,
		Price:        49999.99,
		CategoryGUID: foundCategory.GUID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := productRepo.Create(ctx, product); err != nil {
		log.Fatal().Err(err).Msg("Product create failed")
	}

	log.Info().
		Str("guid", product.GUID.String()).
		Msg("Product created")

	products, err := productRepo.List(ctx, nil, &foundCategory.GUID)
	if err != nil {
		log.Fatal().Err(err).Msg("Product list failed")
	}

	log.Info().
		Int("count", len(products)).
		Msg("Products found")

	if err := productRepo.Delete(ctx, product.GUID); err != nil {
		log.Fatal().Err(err).Msg("Product delete failed")
	}

	if err := categoryRepo.Delete(ctx, foundCategory.GUID); err != nil {
		log.Fatal().Err(err).Msg("Category delete failed")
	}

	log.Info().Msg("Repository test completed")

	// === HTTP сервер запускаем после тестов ===

	hHealth := rhealth.NewHandler()

	httpServer := rprocessor.NewHttp(hHealth, cfg.Processor.WebServer)

	if err := httpServer.Serve(); err != nil {
		log.Fatal().Err(err).Msg("HTTP server failed")
	}
}
