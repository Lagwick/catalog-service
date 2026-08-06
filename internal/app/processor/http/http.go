package rprocessor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/Lagwick/catalog-service/internal/app/config/section"
	rhandler "github.com/Lagwick/catalog-service/internal/app/handler/http"
	"github.com/Lagwick/catalog-service/internal/app/util"
	"github.com/Lagwick/catalog-service/internal/pkg/http/httph"
	"github.com/Lagwick/catalog-service/internal/pkg/http/mzerolog"
)

type httpProc struct {
	server http.Server
	addr   string
}

func NewHTTP(
	hHealth rhandler.Health,
	hCategory rhandler.Category,
	hProduct rhandler.Product,
	cfg section.ProcessorWebServer,
	middlewares []httph.Middleware,
) *httpProc {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(handlerNotFound)

	r.Use(middlewaresToGorilla(middlewares)...)

	r.Use(
		httph.NewErrorMiddleware(),
		mzerolog.NewMiddleware(
			mzerolog.WithSkipper(util.IsFilteredHttpRoute),
		),
	)

	vGenericRegHealthCheck(r, hHealth)

	rV1 := r.PathPrefix("/v1").Subrouter()
	v1RegCategoryHandler(rV1, hCategory)
	v1RegProductHandler(rV1, hProduct)

	_ = r.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		if path == "" || len(methods) == 0 {
			return nil
		}
		log.Info().
			Str("path", path).
			Strs("methods", methods).
			Msg("registered route")
		return nil
	})

	p := httpProc{addr: fmt.Sprintf(":%d", cfg.ListenPort)}
	p.server.Addr = p.addr
	p.server.Handler = r

	return &p
}

func (p *httpProc) Serve() error {
	log.Info().Str("addr", p.addr).Msg("Starting HTTP server")
	return p.server.ListenAndServe()
}

func (p *httpProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		go func() {
			<-ctx.Done()

			if err := p.server.Close(); err != nil {
				log.Error().Err(err).Msg("HTTP server close error")
			}
		}()

		if err := p.Serve(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("HTTP server failed")
		}
	}()
}
