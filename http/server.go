package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/weeb-vip/gateway-proxy/config"
	"github.com/weeb-vip/gateway-proxy/http/middlewares"
	"github.com/weeb-vip/gateway-proxy/internal/handlers"
	"github.com/weeb-vip/gateway-proxy/internal/jwt"
	"github.com/weeb-vip/gateway-proxy/internal/keys"
	"github.com/weeb-vip/gateway-proxy/internal/logger"
	"github.com/weeb-vip/gateway-proxy/internal/poller"
	"github.com/weeb-vip/gateway-proxy/tracing"
)

func Start(cfg *config.Config, formatter logrus.Formatter) error {
	// Initialize context
	ctx := context.Background()

	// Initialize structured logging
	logger.Logger(
		logger.WithServerName(cfg.APPConfig.APPName),
		logger.WithVersion(cfg.APPConfig.Version),
		logger.WithEnvironment(cfg.APPConfig.Env),
	)

	log := logger.Get()
	log.Info().Msg("Initializing telemetry...")

	// Initialize tracing
	tracedCtx, err := tracing.InitTracing(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize tracing")
		return fmt.Errorf("failed to initialize tracing: %w", err)
	}

	// Setup graceful shutdown for tracing
	defer func() {
		log.Info().Msg("Shutting down tracing...")
		if shutdownErr := tracing.Shutdown(context.Background()); shutdownErr != nil {
			log.Error().Err(shutdownErr).Msg("Error shutting down tracing")
		}
	}()

	// Keep logrus for compatibility
	logrus.SetFormatter(formatter)

	jwtParser, err := getJWTParser(cfg)
	if err != nil {
		return err
	}

	log.Info().
		Str("proxy_host", cfg.ProxyURL.Host).
		Int("port", cfg.Port).
		Str("service", cfg.APPConfig.APPName).
		Str("version", cfg.APPConfig.Version).
		Str("environment", cfg.APPConfig.Env).
		Msg("Starting gateway proxy server")

	fmt.Printf("proxy requests to: %s\n", cfg.ProxyURL.Host)
	fmt.Println(fmt.Sprintf("listening on http://localhost:%d", cfg.Port))

	mux := http.NewServeMux()
	// Updated middleware chain with new telemetry
	mux.Handle("/",
		middlewares.CORS(cfg)(
			middlewares.Tracer()(
				middlewares.MetricsMiddleware()(
					middlewares.Logger()(
						handlers.GetProxy(cfg, jwtParser),
					),
				),
			),
		),
	)

	// Use traced context for the server (although http.ListenAndServe doesn't directly use it)
	_ = tracedCtx

	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), mux)
}

func getMinimumDuration(askedDuration time.Duration, minimumDuration time.Duration) time.Duration {
	if askedDuration < minimumDuration {
		return minimumDuration
	}

	return askedDuration
}

func getJWTParser(cfg *config.Config) (jwt.Parser, error) {
	fetcher := keys.NewFetcher(cfg.GraphQLEndpoint)

	p, err := poller.Keys(fetcher)
	if err != nil {
		return nil, err
	}

	requestedDuration := time.Duration(cfg.KeysPollingDurationMinutes) * time.Minute
	p.SetupBackgroundPolling(getMinimumDuration(requestedDuration, time.Minute))

	return jwt.NewParser(p), nil
}
