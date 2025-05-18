package utils

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func NewMetrics() (*sdkmetric.MeterProvider, error) {
	ctx := context.Background()
	otelCollectorEndpoint := "0.0.0.0:4317"

	exporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(otelCollectorEndpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics exporter: %w", err)
	}

	serviceName := "main-api"
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var boundaries []float64
	for i := 0; i < 11; i++ {
		boundary := 0.01 * math.Pow(2, float64(i))
		boundaries = append(boundaries, boundary)
	}
	view := metric.NewView(
		metric.Instrument{Kind: metric.InstrumentKindHistogram},
		metric.Stream{Aggregation: metric.AggregationExplicitBucketHistogram{
			Boundaries: boundaries,
			NoMinMax:   false,
		}},
	)
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(
			exporter,
			sdkmetric.WithInterval(1*time.Minute),
		)),
		sdkmetric.WithResource(res),
		metric.WithView(view),
	)
	otel.SetMeterProvider(mp)

	return mp, nil
}
