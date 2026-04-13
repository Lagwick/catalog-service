package rprocessor

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/Lagwick/catalog-service/internal/app/config/section"
	rhandler "github.com/Lagwick/catalog-service/internal/app/handler/http"
)

type httpProc struct {
	server http.Server
	addr   string
}

func NewHttp(
	hHealth rhandler.Health,
	hCategory rhandler.Category,
	hProduct rhandler.Product,
	cfg section.ProcessorWebServer,
) *httpProc {

	r := mux.NewRouter()

	r.NotFoundHandler = http.HandlerFunc(handlerNotFound)

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

	p := httpProc{
		addr: fmt.Sprintf(":%d", cfg.ListenPort),
	}

	p.server.Addr = p.addr
	p.server.Handler = r

	return &p
}

func (p *httpProc) Serve() error {
	log.Info().Str("addr", p.addr).Msg("Starting HTTP server")
	return p.server.ListenAndServe()
}
