package telemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// InitTracer initializes the OpenTelemetry tracer provider for distributed tracing
func InitTracer() (trace.TracerProvider, error) {
	// Get Jaeger endpoint from environment or use default
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "http://localhost:14268/api/traces"
	}

	// Create Jaeger exporter
	exp, err := jaeger.New(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)),
	)
	if err != nil {
		logrus.Warnf("Failed to initialize Jaeger exporter: %v", err)
		return nil, err
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(
			NewResource("rrctl"),
		),
	)

	otel.SetTracerProvider(tp)
	logrus.Info("OpenTelemetry tracer initialized")

	return tp, nil
}

// NewMetricsRecorder creates a structured metrics recorder
type MetricsRecorder struct {
	tracer trace.Tracer
	logger *logrus.Logger
	ctx    context.Context
}

// NewRecorder creates a new metrics recorder
func NewRecorder(ctx context.Context, logger *logrus.Logger) *MetricsRecorder {
	return &MetricsRecorder{
		tracer: otel.GetTracerProvider().Tracer("github.com/elevatediq/rrctl"),
		logger: logger,
		ctx:    ctx,
	}
}

// RecordCommand records a CLI command execution
func (m *MetricsRecorder) RecordCommand(command string, args []string) func(error) {
	_, span := m.tracer.Start(
		m.ctx,
		fmt.Sprintf("command.%s", command),
		trace.WithAttributes(
			attribute.String("command.name", command),
			attribute.StringSlice("command.args", args),
		),
	)

	startTime := time.Now()
	m.logger.WithFields(logrus.Fields{
		"command": command,
		"args":    args,
		"trace":   span.SpanContext().TraceID(),
		"span":    span.SpanContext().SpanID(),
	}).Info("Command started")

	return func(err error) {
		duration := time.Since(startTime)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			m.logger.WithFields(logrus.Fields{
				"command":  command,
				"duration": duration,
				"error":    err,
				"trace":    span.SpanContext().TraceID(),
			}).Error("Command failed")
		} else {
			span.SetStatus(codes.Ok, "success")
			m.logger.WithFields(logrus.Fields{
				"command":  command,
				"duration": duration,
				"status":   "success",
				"trace":    span.SpanContext().TraceID(),
			}).Info("Command completed")
		}
		span.End()
	}
}

// RecordRCAAnalysis records RCA analysis execution
func (m *MetricsRecorder) RecordRCAAnalysis(targetFile string, useOllama bool) func(error) {
	_, span := m.tracer.Start(
		m.ctx,
		"rca.analyze",
		trace.WithAttributes(
			attribute.String("rca.target", targetFile),
			attribute.Bool("rca.use_ollama", useOllama),
		),
	)

	startTime := time.Now()
	m.logger.WithFields(logrus.Fields{
		"target":      targetFile,
		"use_ollama":  useOllama,
		"trace":       span.SpanContext().TraceID(),
		"timestamp":   startTime.Unix(),
	}).Info("RCA analysis started")

	return func(err error) {
		duration := time.Since(startTime)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			m.logger.WithFields(logrus.Fields{
				"target":    targetFile,
				"duration":  duration.Milliseconds(),
				"error":     err,
				"trace":     span.SpanContext().TraceID(),
			}).Error("RCA analysis failed")
		} else {
			span.SetStatus(codes.Ok, "success")
			m.logger.WithFields(logrus.Fields{
				"target":    targetFile,
				"duration":  duration.Milliseconds(),
				"trace":     span.SpanContext().TraceID(),
			}).Info("RCA analysis completed")
		}
		span.End()
	}
}

// RecordSecurityScan records security scan execution
func (m *MetricsRecorder) RecordSecurityScan(scanType, target string) func(int, error) {
	_, span := m.tracer.Start(
		m.ctx,
		fmt.Sprintf("scan.%s", scanType),
		trace.WithAttributes(
			attribute.String("scan.type", scanType),
			attribute.String("scan.target", target),
		),
	)

	startTime := time.Now()

	return func(findingsCount int, err error) {
		duration := time.Since(startTime)
		span.SetAttributes(attribute.Int("scan.findings", findingsCount))

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "success")
		}

		m.logger.WithFields(logrus.Fields{
			"scan_type": scanType,
			"target":    target,
			"findings":  findingsCount,
			"duration":  duration.Milliseconds(),
			"trace":     span.SpanContext().TraceID(),
		}).Info("Security scan completed")

		span.End()
	}
}

// RecordGitAnalysis records Git repository analysis
func (m *MetricsRecorder) RecordGitAnalysis(repoPath string, commitCount int) func(error) {
	_, span := m.tracer.Start(
		m.ctx,
		"git.analyze",
		trace.WithAttributes(
			attribute.String("git.repo", repoPath),
			attribute.Int("git.commit_count", commitCount),
		),
	)

	startTime := time.Now()

	return func(err error) {
		duration := time.Since(startTime)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "success")
		}

		m.logger.WithFields(logrus.Fields{
			"repo":      repoPath,
			"commits":   commitCount,
			"duration":  duration.Milliseconds(),
			"trace":     span.SpanContext().TraceID(),
		}).Info("Git analysis completed")

		span.End()
	}
}