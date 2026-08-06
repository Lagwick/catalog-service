package monitor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Lagwick/catalog-service/internal/app/config/section"
	"github.com/Lagwick/catalog-service/internal/app/constant"
	"github.com/Lagwick/catalog-service/internal/app/processor"
)

const (
	initTimeout     = 5 * time.Second
	shutdownTimeout = 5 * time.Second
)

type (
	openTelemetryProc struct {
		traceProvider *sdktrace.TracerProvider
		conn          *grpc.ClientConn
	}

	openTelemetryErrorHandler struct{}
)

func NewOpenTelemetryController(
	ctx context.Context,
	env string,
	cfg section.MonitorOpenTelemetry,
) (processor.Processor, error) {
	ctx, cancel := context.WithTimeout(ctx, initTimeout)
	defer cancel()

	var p openTelemetryProc

	attrs := []attribute.KeyValue{
		semconv.ServiceName(constant.AppName),
	}

	if env != "" {
		attrs = append(
			attrs,
			semconv.DeploymentEnvironment(strings.ToLower(env)),
		)
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(attrs...),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	conn, err := grpc.NewClient(
		cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("create gRPC client for Jaeger: %w", err)
	}

	if err = waitForReady(ctx, conn); err != nil {
		_ = conn.Close()

		return nil, fmt.Errorf(
			"connect to Jaeger at %s: %w",
			cfg.Address,
			err,
		)
	}

	p.conn = conn

	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithGRPCConn(conn),
	)
	if err != nil {
		_ = conn.Close()

		return nil, fmt.Errorf("create OTLP exporter: %w", err)
	}

	cfg.SampleRatio = min(1, max(0, cfg.SampleRatio))

	maxBatchSize := int(cfg.MaxBatchSize) //nolint:gosec // G115: значение берётся из конфига
	maxQueueSize := int(cfg.MaxQueueSize) //nolint:gosec // G115: значение берётся из конфига

	batchProcessor := sdktrace.NewBatchSpanProcessor(
		exporter,
		sdktrace.WithBatchTimeout(cfg.SendBatchTimeout),
		sdktrace.WithExportTimeout(cfg.ExportTimeout),
		sdktrace.WithMaxExportBatchSize(maxBatchSize),
		sdktrace.WithMaxQueueSize(maxQueueSize),
	)

	p.traceProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(
			sdktrace.TraceIDRatioBased(cfg.SampleRatio),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(batchProcessor),
	)

	otel.SetTracerProvider(p.traceProvider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)
	otel.SetErrorHandler(openTelemetryErrorHandler{})

	log.Info().
		Str("service", constant.AppName).
		Str("environment", strings.ToLower(env)).
		Msg("OpenTelemetry has been initialized")

	return &p, nil
}

func waitForReady(ctx context.Context, conn *grpc.ClientConn) error {
	conn.Connect()

	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			return nil
		}

		if !conn.WaitForStateChange(ctx, state) {
			return ctx.Err()
		}
	}
}

func (p *openTelemetryProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	processor.WatchForShutdown(ctx, wg, processor.CloserFunc(p.shutdown))
}

func (p *openTelemetryProc) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := p.traceProvider.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to shutdown trace provider")
	}
	if err := p.conn.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close Jaeger gRPC connection")
	}
	return nil
}

func (openTelemetryErrorHandler) Handle(err error) {
	log.Error().Err(err).Msg("OpenTelemetry error")
}
