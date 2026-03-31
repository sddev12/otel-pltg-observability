package observability

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	otelglobal "go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

var Logger *slog.Logger

type ShutdownOtelFunc func(context.Context) error

func SetupOpenTelemetry(ctx context.Context, resource *resource.Resource) (ShutdownOtelFunc, error) {
	loggerShutdown, logger, err := SetupLogging(ctx, resource)
	if err != nil {
		return nil, err
	}

	metricProvider, err := SetupMetrics(ctx, resource)
	if err != nil {
		return nil, err
	}

	var shutdownFuncs []ShutdownOtelFunc = []ShutdownOtelFunc{
		loggerShutdown,
		metricProvider.Shutdown,
	}

	shutdownOtel := func(ctx context.Context) error {
		var errs []error
		for _, fn := range shutdownFuncs {
			if err := fn(ctx); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	}

	Logger = logger

	return shutdownOtel, nil

}

func SetupLogging(ctx context.Context, resource *resource.Resource) (func(context.Context) error, *slog.Logger, error) {

	// Set up stdout log exporter
	stdoutExporter, err := stdoutlog.New(
		stdoutlog.WithWriter(os.Stdout),
	)
	if err != nil {
		return nil, nil, err
	}

	// Set up Otel Collector Exporter
	otelCollectorExporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint("localhost:4317"),
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		return nil, nil, err
	}

	// Set up provider with batch processor
	logProvider := sdklog.NewLoggerProvider(
		sdklog.WithResource(resource),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(stdoutExporter)),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(otelCollectorExporter)),
	)

	// Make the provider global
	otelglobal.SetLoggerProvider(logProvider)

	// Set up the log handler for integrating otel with slog
	logHandler := otelslog.NewHandler(
		"go-gin-api.logger",
		otelslog.WithLoggerProvider(logProvider),
		otelslog.WithSource(true),
	)

	// Set up slog with otel
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	return logProvider.Shutdown, logger, nil
}

func SetupMetrics(ctx context.Context, resource *resource.Resource) (*metric.MeterProvider, error) {
	// Set up stdout metric exporter
	stdoutExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	// Set up Otel Collector exporter
	otelCollectorExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint("localhost:4317"),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(resource),
		metric.WithReader(metric.NewPeriodicReader(
			stdoutExporter,
			metric.WithInterval(3*time.Second))),
		metric.WithReader(metric.NewPeriodicReader(
			otelCollectorExporter,
			metric.WithInterval(3*time.Second))),
	)

	otel.SetMeterProvider(meterProvider)

	return meterProvider, nil
}
