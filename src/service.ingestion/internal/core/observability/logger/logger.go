package logger

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

const (
	OtelExporterOtelEndpointKey = "OTEL_EXPORTER_OTLP_ENDPOINT"
	OtelServiceNameKey          = "OTEL_SERVICE_NAME"
	OtelServiceLayerKey         = "OTEL_SERVICE_LAYER"
	OtelServiceVersionKey       = "OTEL_SERVICE_VERSION"
	OtelEnabled                 = "OTEL_ENABLED"
)

// Logger defines the logging interface with various severity levels.
type Logger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(err error, format string, args ...any)
	Fatal(err error, format string, args ...any)
}

// logSink represents a logging sink with OpenTelemetry integration.
type logSink struct {
	config   logConfig
	provider log.LoggerProvider
	logger   log.Logger
}

// logConfig holds configuration for the OpenTelemetry logger.
type logConfig struct {
	otelEnabled    bool // otelEnabled indicates whether OpenTelemetry logging is enabled (used for local testing).
	endpoint       string
	serviceName    string
	serviceLayer   string
	serviceVersion string
	minLevel       log.Severity
}

// NewOtelLogger initializes an OpenTelemetry logger for logs only, using environment variables.
func NewOtelLogger(ctx context.Context, args ...attribute.KeyValue) (Logger, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			append(args,
				attribute.String("service.name", cfg.serviceName),
				attribute.String("service.layer", cfg.serviceLayer),
				attribute.String("service.version", cfg.serviceVersion),
			)...,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	exporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint(cfg.endpoint),
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create otpl log exporter: %w", err)
	}

	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)

	logger := provider.Logger(cfg.serviceName)
	cfg.minLevel = log.SeverityInfo

	return &logSink{
		config:   cfg,
		provider: provider,
		logger:   logger,
	}, nil
}

// loadConfig loads otel configuration from environment variables.
func loadConfig() (logConfig, error) {
	endpoint := strings.TrimSpace(os.Getenv(OtelExporterOtelEndpointKey))
	serviceName := strings.TrimSpace(os.Getenv(OtelServiceNameKey))
	serviceLayer := strings.TrimSpace(os.Getenv(OtelServiceLayerKey))
	serviceVersion := strings.TrimSpace(os.Getenv(OtelServiceVersionKey))
	otelEnabled := strings.TrimSpace(os.Getenv(OtelEnabled))

	if len(endpoint) == 0 {
		return logConfig{}, fmt.Errorf("environment variable %s is required but not set", OtelExporterOtelEndpointKey)
	}

	if len(serviceName) == 0 {
		return logConfig{}, fmt.Errorf("environment variable %s is required but not set", OtelServiceNameKey)
	}

	return logConfig{
		endpoint:       endpoint,
		serviceName:    serviceName,
		serviceLayer:   serviceLayer,
		serviceVersion: serviceVersion,
		otelEnabled:    otelEnabled == "true",
	}, nil
}

// emit logs a message with the specified severity level.
func (ls *logSink) emit(ctx context.Context, severity log.Severity, err error, format string, args ...any) {
	var record log.Record

	message := fmt.Sprintf(format, args...)
	timedMessage := fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), message)

	record.SetBody(log.StringValue(timedMessage))
	record.SetSeverity(severity)

	if err != nil {
		record.AddAttributes(log.String(string(semconv.ExceptionMessageKey), err.Error()))
	}

	if !ls.config.otelEnabled {
		fmt.Printf("%s -> %s\n", severity.String(), timedMessage)
		return
	}

	if ls.config.minLevel > severity {
		return
	}

	ls.logger.Emit(ctx, record)
}

// Debug logs a debug-level message.
func (ls *logSink) Debug(format string, args ...any) {
	toCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	ls.emit(toCtx, log.SeverityDebug, nil, format, args...)
}

// Info logs an info-level message.
func (ls *logSink) Info(format string, args ...any) {
	toCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	ls.emit(toCtx, log.SeverityInfo, nil, format, args...)
}

// Warn logs a warning-level message.
func (ls *logSink) Warn(format string, args ...any) {
	toCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	ls.emit(toCtx, log.SeverityWarn, nil, format, args...)
}

// Error logs an error-level message along with an error.
func (ls *logSink) Error(err error, format string, args ...any) {
	toCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ls.emit(toCtx, log.SeverityError, err, format, args...)
}

// Fatal logs a fatal-level message along with an error and exits the application.
func (ls *logSink) Fatal(err error, format string, args ...any) {
	toCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ls.emit(toCtx, log.SeverityFatal, err, format, args...)
	os.Exit(1)
}
