package mtracelog

import (
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

const fieldTraceID = "trace_id"

type Hook struct{}

func (Hook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	ctx := e.GetCtx()
	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() {
		return
	}
	e.Str(fieldTraceID, spanContext.TraceID().String())
}
