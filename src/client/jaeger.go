package client

import (
	"context"
	"time"

	"rcabench/config"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var (
	TraceProvider *sdktrace.TracerProvider
)

func InitTraceProvider() {
	ctx := context.Background()

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpoint(config.GetString("jaeger.endpoint")),
	)
	if err != nil {
		logrus.Errorf("failed to create OTLP HTTP exporter: %v", err)
		return
	}

	resource, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.GetString("name")),
			semconv.ServiceVersion(config.GetString("version")),
		),
	)
	if err != nil {
		logrus.Errorf("failed to create OTLP sdk resource: %v", err)
		return
	}

	TraceProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)
	otel.SetTracerProvider(TraceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}

func ShutdownTraceProvider(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	if err := TraceProvider.Shutdown(ctx); err != nil {
		logrus.Errorf("failed to shutdown tracer provider: %v", err)
	}
}
