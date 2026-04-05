package section

import "github.com/Lagwick/catalog-service/internal/app/util"

type (
	Repository struct {
		Postgres RepositoryPostgres `split_words:"true"`
	}

	RepositoryPostgres struct {
		Address        string        `required:"true"`
		Username       string        `required:"true"`
		Password       string        `required:"true"`
		Name           string        `required:"true"`
		ReadTimeout    util.Duration `default:"5s" split_words:"true"`
		WriteTimeout   util.Duration `default:"5s" split_words:"true"`
		MigrationTable string        `split_words:"true" default:"schema_migrations"`
	}
)
