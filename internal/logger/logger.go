package logger

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

type ctxKey struct{}

var once sync.Once
var globalLogger zerolog.Logger

// Logger initializes and configures the global logger with options
func Logger(opts ...Option) {
	once.Do(func() {
		config := &Config{
			ServiceName:    "unknown-service",
			ServiceVersion: "unknown-version",
		}

		for _, opt := range opts {
			opt(config)
		}

		globalLogger = log.With().
			Str("service", config.ServiceName).
			Str("version", config.ServiceVersion).
			Str("environment", config.Environment).
			Logger()
	})
}

// Get returns the global logger instance
func Get() zerolog.Logger {
	return globalLogger
}

// FromCtx returns the Logger associated with the ctx. If no logger
// is associated, the global logger is returned. Automatically includes trace context.
func FromCtx(ctx context.Context) zerolog.Logger {
	var logger zerolog.Logger
	if l, ok := ctx.Value(ctxKey{}).(zerolog.Logger); ok {
		logger = l
	} else {
		logger = globalLogger
		// If global logger is not initialized, use a default logger
		if logger.GetLevel() == zerolog.Disabled {
			logger = log.With().Logger()
		}
	}

	// Add trace context if available
	return withTraceContext(ctx, logger)
}

// withTraceContext adds trace and span IDs to the logger if available in context
func withTraceContext(ctx context.Context, logger zerolog.Logger) zerolog.Logger {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return logger
	}

	spanContext := span.SpanContext()
	return logger.With().
		Str("trace_id", spanContext.TraceID().String()).
		Str("span_id", spanContext.SpanID().String()).
		Logger()
}

// FromContext is an alias for FromCtx for backward compatibility
func FromContext(ctx context.Context) zerolog.Logger {
	return FromCtx(ctx)
}

// WithCtx returns a copy of ctx with the Logger attached.
func WithCtx(ctx context.Context, l zerolog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// WithContext returns a copy of ctx with the Logger attached.
func WithContext(logger zerolog.Logger, ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

// Config holds logger configuration
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
}

// Option configures the logger
type Option func(*Config)

// WithServerName sets the service name
func WithServerName(name string) Option {
	return func(c *Config) {
		c.ServiceName = name
	}
}

// WithVersion sets the service version
func WithVersion(version string) Option {
	return func(c *Config) {
		c.ServiceVersion = version
	}
}

// WithEnvironment sets the environment
func WithEnvironment(env string) Option {
	return func(c *Config) {
		c.Environment = env
	}
}

// LogRequest logs an HTTP request with standard fields
func LogRequest(ctx context.Context, method, path, remoteAddr, userAgent string, duration time.Duration, statusCode int) {
	log := FromCtx(ctx)

	log.Info().
		Str("method", method).
		Str("path", path).
		Str("remote_addr", remoteAddr).
		Str("user_agent", userAgent).
		Dur("duration", duration).
		Int("status_code", statusCode).
		Msg("HTTP request completed")
}

// LogError logs an error with context
func LogError(ctx context.Context, err error, msg string) {
	log := FromCtx(ctx)
	log.Error().Err(err).Msg(msg)
}

// LogJWTValidation logs JWT validation attempts
func LogJWTValidation(ctx context.Context, success bool, duration time.Duration, userID string) {
	log := FromCtx(ctx)

	event := log.Info()
	if !success {
		event = log.Warn()
	}

	event.
		Bool("success", success).
		Dur("duration", duration).
		Str("user_id", userID).
		Msg("JWT validation completed")
}

// LogKeyFetch logs key fetching operations
func LogKeyFetch(ctx context.Context, success bool, duration time.Duration, keyCount int) {
	log := FromCtx(ctx)

	event := log.Info()
	if !success {
		event = log.Error()
	}

	event.
		Bool("success", success).
		Dur("duration", duration).
		Int("key_count", keyCount).
		Msg("Key fetch operation completed")
}