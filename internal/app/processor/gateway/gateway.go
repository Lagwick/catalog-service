package pgateway

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Lagwick/catalog-service/internal/app/config/section"
	"github.com/Lagwick/catalog-service/internal/app/processor"
	catalogv1 "github.com/Lagwick/catalog-service/internal/pkg/grpc/gen/catalog/v1"
)

const (
	gatewayReadHeaderTimeout = 5 * time.Second
	gatewayReadTimeout       = 30 * time.Second
	gatewayWriteTimeout      = 30 * time.Second
	gatewayIdleTimeout       = 120 * time.Second
	gatewayShutdownTimeout   = 5 * time.Second
)

type gatewayProc struct {
	server   *http.Server
	addr     string
	grpcAddr string
}

func NewGateway(
	cfgGateway section.ProcessorGateway,
	cfgGrpc section.ProcessorGrpc,
) processor.Processor {
	addr := fmt.Sprintf(":%d", cfgGateway.ListenPort)

	grpcAddr := net.JoinHostPort(
		"localhost",
		strconv.Itoa(int(cfgGrpc.ListenPort)),
	)

	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: gatewayReadHeaderTimeout,
		ReadTimeout:       gatewayReadTimeout,
		WriteTimeout:      gatewayWriteTimeout,
		IdleTimeout:       gatewayIdleTimeout,
	}

	return &gatewayProc{
		server:   server,
		addr:     addr,
		grpcAddr: grpcAddr,
	}
}

func (p *gatewayProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := catalogv1.RegisterCatalogServiceHandlerFromEndpoint(ctx, mux, p.grpcAddr, opts); err != nil {
		log.Fatal().Err(err).Str("grpc_addr", p.grpcAddr).Msg("failed to register grpc gateway")
		return
	}

	p.server.Handler = mux

	lc := net.ListenConfig{}

	l, err := lc.Listen(ctx, "tcp", p.addr)
	if err != nil {
		log.Fatal().Err(err).Str("addr", p.addr).Msg("failed to listen gateway")
		return
	}

	log.Info().Str("addr", p.addr).Str("grpc_addr", p.grpcAddr).Msg("Starting gRPC Gateway")

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := p.server.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("grpc gateway stopped")
		}
	}()

	processor.WatchForShutdown(ctx, wg, l)

	processor.WatchForShutdown(ctx, wg, processor.NewCloserContextFunc(
		p.server.Shutdown,
		context.Background(),
		gatewayShutdownTimeout,
	))
}
