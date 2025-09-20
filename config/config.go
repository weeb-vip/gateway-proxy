package config

import (
	"fmt"
	"github.com/jinzhu/configor"
	"net/url"
	"path"
	"runtime"
)

type Config struct {
	Version                    string   `env:"APP__VERSION" default:"local"`
	Port                       int      `env:"CONFIG__PORT" default:"8080"`
	GraphQLEndpoint            string   `env:"INTERNAL_GRAPHQL_URL" default:"http://key-management:5001/graphql"`
	ProxyAddress               string   `env:"CONFIG__PROXY_URL" default:"http://apollo-router:4000" json:"proxy_address"`
	OverrideOrigin             *string  `env:"CONFIG__OVERRIDE_ORIGIN" json:"override_origin"`
	KeysPollingDurationMinutes uint     `env:"CONFIG__KEYS_POLLING_DURATION_MINUTES" default:"15"`
	CORSAllowedOrigins         []string `env:"CONFIG__CORS_ALLOWED_ORIGINS" json:"cors_allowed_origins"`
	CORSAllowCredentials       bool     `env:"CONFIG__CORS_ALLOW_CREDENTIALS" default:"true" json:"cors_allow_credentials"`
	CORSMaxAge                 int      `env:"CONFIG__CORS_MAX_AGE" default:"86400" json:"cors_max_age"`
	ProxyURL                   *url.URL
}

func LoadConfig() (*Config, error) {
	var config Config
	err := configor.
		New(&configor.Config{AutoReload: false}).
		Load(&config, fmt.Sprintf("%s/config.json", getConfigLocation()))

	if err != nil {
		return nil, err
	}
	u, err := url.Parse(config.ProxyAddress)
	if err != nil {
		return nil, err
	}
	config.ProxyURL = u

	return &config, nil
}

func getConfigLocation() string {
	_, filename, _, _ := runtime.Caller(0)

	return path.Join(path.Dir(filename), "../config")
}
