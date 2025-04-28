package tracing

import (
	"context"
	"path"
	"runtime"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// WithSpan wraps a function with OpenTelemetry tracing.
// It automatically detects the function name from the call stack.
func WithSpan(ctx context.Context, f func(context.Context) error) error {
	pc, _, _, ok := runtime.Caller(1)
	funcName := "unknown"
	if ok {
		funcName = path.Base(runtime.FuncForPC(pc).Name())
	}

	childCtx, span := otel.Tracer("rcabench/task").Start(ctx, funcName)
	defer span.End()

	return f(childCtx)
}

// WithSpanNamed wraps a function with OpenTelemetry tracing using a custom span name.
func WithSpanNamed(ctx context.Context, name string, f func(context.Context) error) error {
	childCtx, span := otel.Tracer("rcabench/task").Start(ctx, name)
	defer span.End()

	return f(childCtx)
}

func WithSpanReturnValue[T any](ctx context.Context, f func(context.Context) (T, error)) (T, error) {
	pc, _, _, ok := runtime.Caller(1)
	funcName := "unknown"
	if ok {
		funcName = path.Base(runtime.FuncForPC(pc).Name())
	}

	childCtx, span := otel.Tracer("rcabench/task").Start(ctx, funcName)
	defer span.End()

	return f(childCtx)
}

func SetSpanAttribute(ctx context.Context, key string, value string) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.String(key, value),
		)
	}
}
