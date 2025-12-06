package telemetry

import (
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// NewResource creates an OTEL resource for rrctl
func NewResource(serviceName string) *resource.Resource {
	r, _ := resource.New(nil,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("1.0.0"),
			semconv.ServiceInstanceIDKey.String("rrctl-" + serviceName),
		),
	)
	return r
}