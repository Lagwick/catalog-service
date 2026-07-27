package builder

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/Lagwick/catalog-service/internal/app/config"
	ghcatalogv1 "github.com/Lagwick/catalog-service/internal/app/handler/grpc/catalog/v1"
	rhandler "github.com/Lagwick/catalog-service/internal/app/handler/http"
	hcategory "github.com/Lagwick/catalog-service/internal/app/handler/http/category"
	rhealth "github.com/Lagwick/catalog-service/internal/app/handler/http/health"
	hproduct "github.com/Lagwick/catalog-service/internal/app/handler/http/product"
	"github.com/Lagwick/catalog-service/internal/app/processor"
	pgateway "github.com/Lagwick/catalog-service/internal/app/processor/gateway"
	pgrpc "github.com/Lagwick/catalog-service/internal/app/processor/grpc"
	rprocessor "github.com/Lagwick/catalog-service/internal/app/processor/http"
	pprocessor "github.com/Lagwick/catalog-service/internal/app/processor/other"
	"github.com/Lagwick/catalog-service/internal/app/repository"
	pcategory "github.com/Lagwick/catalog-service/internal/app/repository/category"
	rcpostgres "github.com/Lagwick/catalog-service/internal/app/repository/conn/postgres"
	pproduct "github.com/Lagwick/catalog-service/internal/app/repository/product"
	"github.com/Lagwick/catalog-service/internal/app/service"
	scategory "github.com/Lagwick/catalog-service/internal/app/service/category"
	sproduct "github.com/Lagwick/catalog-service/internal/app/service/product"
	catalogv1 "github.com/Lagwick/catalog-service/internal/pkg/grpc/gen/catalog/v1"
)

type Builder struct {
	cCtx       *cli.Context
	ctx        context.Context
	wg         sync.WaitGroup
	err        error
	cfg        config.Config
	cancelFunc context.CancelFunc

	chErrors chan error

	connPostgres *rcpostgres.Client

	categoryRepository repository.Category
	productRepository  repository.Product
	categoryService    service.Category
	productService     service.Product
	healthHandler      rhandler.Health
	categoryHandler    rhandler.Category
	productHandler     rhandler.Product

	catalogV1Handler catalogv1.CatalogServiceServer

	processors []processor.Processor
}

func NewBuilder(cCtx *cli.Context) *Builder {
	b := Builder{
		cCtx:     cCtx,
		chErrors: make(chan error, 4096),
	}

	b.healthHandler = rhealth.NewHandler()

	ctx, cancelFunc := context.WithCancel(context.Background())
	b.ctx = ctx
	b.cancelFunc = cancelFunc
	sig := make(chan os.Signal, 1)
	signal.Notify(
		sig, syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
	)
	go b.waitForSignal(sig, cancelFunc)
	go b.printErrors()
	return &b
}

func (b *Builder) BuildConfig() {
	b.exec(true, func(b *Builder) {
		b.buildConfig()
	})
}

func (b *Builder) Run() {
	if b.ctx.Err() != nil {
		log.Info().Msg("Shutdown during initialization")
		return
	}
	if b.err != nil {
		log.Fatal().Err(b.err).Msg("Failed to initialize application")
	}
	log.Info().Msg("Application initialized")
	defer log.Info().Msg("Application completed")
	for _, proc := range b.processors {
		proc.StartAsync(b.ctx, &b.wg)
	}

	b.wg.Wait()
}

////////////////////////////////////////////////////////////////////////////////
///// REPOSITORY CONNECTIONS ///////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildRepoConnPostgres() {
	b.exec(true, func(b *Builder) {
		client, err := rcpostgres.NewClient(b.ctx, b.cfg.Repository.Postgres)
		if err != nil {
			b.err = err
			return
		}
		b.connPostgres = client
	})
}

func (b *Builder) BuildRepoConnMigrator() {
	b.exec(b.connPostgres != nil, func(b *Builder) {
		b.processors = append(b.processors, pprocessor.NewMigrator(b.connPostgres))
	})
}

////////////////////////////////////////////////////////////////////////////////
///// REPOSITORIES /////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildRepoCategory() {
	b.exec(true, func(b *Builder) {
		b.categoryRepository = pcategory.NewRepoFromPostgres(b.connPostgres)
	}, b.connPostgres)
}

func (b *Builder) BuildRepoProduct() {
	b.exec(true, func(b *Builder) {
		b.productRepository = pproduct.NewRepoFromPostgres(b.connPostgres)
	}, b.connPostgres)
}

////////////////////////////////////////////////////////////////////////////////
///// PRIVATE //////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) waitForSignal(sig chan os.Signal, cancelFunc func()) {
	result := <-sig
	log.Info().Str("signal", result.String()).Msg("Shutdown is requested")
	cancelFunc()
}

func (b *Builder) printErrors() {
	for err := range b.chErrors {
		log.Error().Err(err).Msg("Got new error")
	}
}

func (b *Builder) buildConfig() {
	args := config.LoadArgs{
		Output:          b.cCtx.App.Writer,
		EnableSimpleLog: b.cCtx.Bool("no-json"),
	}

	config.Load(args)

	b.cfg = config.Root
}

func (b *Builder) exec(preCond bool, cb func(b *Builder), requiredArgs ...any) {
	if !preCond || b.err != nil || b.ctx.Err() != nil {
		return
	}

	for i, requiredArg := range requiredArgs {
		rv := reflect.ValueOf(requiredArg)
		if !rv.IsValid() {
			b.err = fmt.Errorf("BUG: required argument #%d is nil (check dependencies)", i)
			return
		}
		if rv.Type().Kind() == reflect.Struct || !rv.IsZero() {
			continue
		}
		b.err = fmt.Errorf("BUG: required %s, but empty", rv.Type().String())
		return
	}

	cb(b)
}

////////////////////////////////////////////////////////////////////////////////
///// SERVICES /////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildServiceCategory() {
	b.exec(true, func(b *Builder) {
		b.categoryService = scategory.NewService(b.categoryRepository, b.productRepository, b.connPostgres)
	}, b.categoryRepository, b.productRepository)
}

func (b *Builder) BuildServiceProduct() {
	b.exec(true, func(b *Builder) {
		b.productService = sproduct.NewService(
			b.productRepository,
			b.categoryRepository,
			b.connPostgres,
		)
	}, b.productRepository, b.categoryRepository, b.connPostgres)
}

////////////////////////////////////////////////////////////////////////////////
///// HANDLERS /////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildHandlerHttpCategory() {
	b.exec(true, func(b *Builder) {
		b.categoryHandler = hcategory.NewHandler(b.categoryService)
	}, b.categoryService)
}

func (b *Builder) BuildHandlerHttpProduct() {
	b.exec(true, func(b *Builder) {
		b.productHandler = hproduct.NewHandler(b.productService)
	}, b.productService)
}

func (b *Builder) BuildHandlerGrpcCatalogV1() {
	b.exec(true, func(b *Builder) {
		b.catalogV1Handler = ghcatalogv1.NewHandler(b.productService)
	}, b.productService)
}

////////////////////////////////////////////////////////////////////////////////
///// PROCESSORS ///////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildProcHttp() {
	b.exec(true, func(b *Builder) {
		proc := rprocessor.NewHTTP(
			b.healthHandler,
			b.categoryHandler,
			b.productHandler,
			b.cfg.Processor.WebServer)
		b.processors = append(b.processors, proc)
	}, b.healthHandler)
}

func (b *Builder) BuildProcGrpc() {
	b.exec(true, func(b *Builder) {
		proc := pgrpc.NewGRPC(
			b.catalogV1Handler,
			b.cfg.Processor.Grpc,
		)

		b.processors = append(b.processors, proc)
	}, b.catalogV1Handler)
}

func (b *Builder) BuildProcGateway() {
	b.exec(true, func(b *Builder) {
		proc := pgateway.NewGateway(
			b.cfg.Processor.Gateway,
			b.cfg.Processor.Grpc,
		)

		b.processors = append(b.processors, proc)
	})
}
