package main

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/sddev12/go-gin-api/handlers"
	"github.com/sddev12/go-gin-api/observability"
	"github.com/sddev12/go-gin-api/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Metrics

func main() {
	ctx := context.Background()

	// Set up resource
	resource, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("go-gin-api"),
			semconv.DeploymentEnvironment("dev"),
		),
	)
	if err != nil {
		panic(err)
	}

	// Set up Open Telemetry to regiser the global metric provider and the logger
	shutdownOtel, err := observability.SetupOpenTelemetry(ctx, resource)
	if err != nil {
		panic(err)
	}
	defer shutdownOtel(ctx)

	// Initialise handler metrics
	if err := handlers.InitMetrics(otel.Meter("go-gin-api")); err != nil {
		panic(err)
	}

	// Load env vars for the application
	if err := utils.LoadEnvVars(); err != nil {
		slog.Error("Failed to load environment variables", "error", err)
		return
	}

	// Set up the api server
	router := gin.Default()

	// Routes
	router.GET("/healthz", handlers.HandleHealthz)
	router.GET("/slow", handlers.HandleSlow)
	router.GET("/errorgen", handlers.HandleErrorGen)

	// Start the server
	slog.Info("starting server...")
	router.Run() // listens on 0.0.0.0:8080 by default
}
