package pprocessor

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/Lagwick/catalog-service/internal/app/processor"
	"github.com/Lagwick/catalog-service/internal/app/repository"
)

type procMigrate struct {
	migrator repository.Migrate
}

func NewMigrator(migrator repository.Migrate) processor.Processor {
	return &procMigrate{
		migrator: migrator,
	}
}

func (p *procMigrate) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	processor.Wrap(ctx, wg, func(ctx context.Context) {
		p.job(ctx)
	})
}

func (p *procMigrate) job(ctx context.Context) {

	oldVer, newVer, err := p.migrator.Migrate(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Migration error")
		return
	}
	if oldVer != newVer {
		log.Info().
			Int64("old_version", oldVer).
			Int64("new_version", newVer).
			Msg("Migration complete")
	} else {
		log.Info().Msg("Schema is up to date")
	}
}
