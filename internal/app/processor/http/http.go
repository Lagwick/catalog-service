package rprocessor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/Lagwick/catalog-service/internal/app/config/section"
	rhandler "github.com/Lagwick/catalog-service/internal/app/handler/http"
	"github.com/Lagwick/catalog-service/internal/app/processor"
	"github.com/Lagwick/catalog-service/internal/app/util"
	"github.com/Lagwick/catalog-service/internal/pkg/http/httph"
	"github.com/Lagwick/catalog-service/internal/pkg/http/mzerolog"
)

type HTTPProc struct {
	server http.Server
	addr   string
}

func NewHTTP(
	hHealth rhandler.Health,
	hCategory rhandler.Category,
	hProduct rhandler.Product,
	cfg section.ProcessorWebServer,
) processor.Processor {
	r := mux.NewRouter()

	r.NotFoundHandler = http.HandlerFunc(handlerNotFound)

	r.Use(
		httph.NewErrorMiddleware(),
		mzerolog.NewMiddleware(
			mzerolog.WithSkipper(util.IsFilteredHttpRoute),
		),
	)

	// health
	vGenericRegHealthCheck(r, hHealth)

	// API v1
	rV1 := r.PathPrefix("/v1").Subrouter()

	if hCategory != nil {
		v1RegCategoryHandler(rV1, hCategory)
	}

	if hProduct != nil {
		v1RegProductHandler(rV1, hProduct)
	}

	// лог всех роутов
	_ = r.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		if path == "" {
			return nil
		}
		methods, _ := route.GetMethods()
		if len(methods) == 0 {
			return nil
		}

		log.Debug().
			Str("path", path).
			Strs("methods", methods).
			Msg("registered route")

		return nil
	})

	p := HTTPProc{
		addr: fmt.Sprintf(":%d", cfg.ListenPort),
	}

	p.server.Handler = r

	return &p
}

func (p *HTTPProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	var lc net.ListenConfig

	l, err := lc.Listen(ctx, "tcp", p.addr)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start listener")
	}

	log.Info().
		Str("addr", p.addr).
		Msg("HTTP server started")

	go p.serve(l)

	processor.WatchForShutdown(
		ctx,
		wg,
		processor.CloserFunc(func() error {
			return l.Close()
		}),
	)

	processor.WatchForShutdown(
		ctx,
		wg,
		processor.NewCloserContextFunc(
			func(ctx context.Context) error {
				return p.server.Shutdown(ctx)
			},
			context.Background(),
			5*time.Second,
		),
	)
}

func (p *HTTPProc) serve(l net.Listener) {
	_ = p.server.Serve(l)
}
