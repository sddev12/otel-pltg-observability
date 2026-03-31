package handlers

import (
	"go.opentelemetry.io/otel/metric"
)

var meter metric.Meter
var healthzReqCounter metric.Int64Counter
var slowReqCounter metric.Int64Counter

func InitMetrics(m metric.Meter) error {
	meter = m
	var err error
	healthzReqCounter, err = m.Int64Counter("go_gin_api.healthcheck.total_requests",
		metric.WithDescription("Total number of requests on the healthz endpoint"),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		return err
	}

	slowReqCounter, err = m.Int64Counter("go_gin_api.slow.total_requests",
		metric.WithDescription("Total number of requests on the slow endpoint"),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		return err
	}

	return nil
}
