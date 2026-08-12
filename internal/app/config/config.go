package config

import (
	"io"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/Lagwick/catalog-service/internal/app/config/section"
	"github.com/Lagwick/catalog-service/internal/app/constant"
	msentry "github.com/Lagwick/catalog-service/internal/app/monitor/sentry"
	mtracelog "github.com/Lagwick/catalog-service/internal/app/monitor/tracelog"
)

type LoadArgs struct {
	Output          io.Writer `json:"-"`
	EnableSimpleLog bool
}

func createLogger(level zerolog.Level, output io.Writer) zerolog.Logger {
	return zerolog.New(output).
		Level(level).
		Hook(mtracelog.Hook{}).
		With().
		Timestamp().
		Logger()
}

type Config struct {
	Repository section.Repository `split_words:"true"`
	Processor  section.Processor  `split_words:"true"`
	Monitor    section.Monitor    `split_words:"true"`
}

var Root Config

func Load(args LoadArgs) {
	zerolog.TimestampFieldName = "timestamp"
	zerolog.MessageFieldName = "msg"
	zerolog.TimeFieldFormat = time.RFC3339

	if args.EnableSimpleLog {
		args.Output = zerolog.ConsoleWriter{Out: args.Output}
	}

	log.Logger = createLogger(zerolog.DebugLevel, args.Output)

	log.Debug().Msg("Logger initialized with Debug level")

	if err := godotenv.Load(); err != nil {
		log.Warn().Err(err).Msg("No .env file found")
	}

	if err := envconfig.Process("APP", &Root); err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	level, err := zerolog.ParseLevel(Root.Monitor.LogLevel)
	if err != nil {
		log.Warn().
			Str("log_level", Root.Monitor.LogLevel).
			Msg("Unknown log level, using debug")

		level = zerolog.DebugLevel
	}

	output := args.Output

	w, ok := msentry.Init(
		Root.Monitor.Sentry,
		msentry.Options{
			ServiceName: constant.AppName,
			Environment: Root.Monitor.Environment,
			Release:     constant.Version,
		},
	)
	if ok {
		output = zerolog.MultiLevelWriter(args.Output, w)
	}

	log.Logger = createLogger(level, output)

	log.Info().
		Str("log_level", level.String()).
		Msg("Logger re-initialized with config level")
}
