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

	ddHttp "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

func Start(cfg *config.Config, formatter logrus.Formatter) error {
	_ = profiler.Start(profiler.WithProfileTypes(profiler.CPUProfile, profiler.HeapProfile))
	logrus.SetFormatter(formatter)
	http.DefaultTransport = ddHttp.WrapRoundTripper(http.DefaultTransport)
	tracer.Start(tracer.WithServiceName("proxy"))
	defer tracer.Stop()
	jwtParser, err := getJWTParser(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("proxy requests to: %s\n", cfg.ProxyURL.Host)
	fmt.Println(fmt.Sprintf("listening on http://localhost:%d", cfg.Port))

	mux := ddHttp.NewServeMux(ddHttp.WithServiceName("proxy"))
	mux.Handle("/", middlewares.Measurement(measurements.NewClient())(middlewares.Tracer()(middlewares.Logger()(handlers.GetProxy(cfg, jwtParser)))))

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
