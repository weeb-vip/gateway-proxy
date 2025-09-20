package http

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/weeb-vip/gateway-proxy/http/middlewares"
	"github.com/weeb-vip/gateway-proxy/internal/measurements"
	"net/http"
	"time"

	"github.com/weeb-vip/gateway-proxy/config"
	"github.com/weeb-vip/gateway-proxy/internal/handlers"
	"github.com/weeb-vip/gateway-proxy/internal/jwt"
	"github.com/weeb-vip/gateway-proxy/internal/keys"
	"github.com/weeb-vip/gateway-proxy/internal/poller"

)

func Start(cfg *config.Config, formatter logrus.Formatter) error {
	logrus.SetFormatter(formatter)
	jwtParser, err := getJWTParser(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("proxy requests to: %s\n", cfg.ProxyURL.Host)
	fmt.Println(fmt.Sprintf("listening on http://localhost:%d", cfg.Port))

	mux := http.NewServeMux()
	mux.Handle("/", middlewares.CORS(cfg)(middlewares.Measurement(measurements.NewClient())(middlewares.Logger()(handlers.GetProxy(cfg, jwtParser)))))

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
