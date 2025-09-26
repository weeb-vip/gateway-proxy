package config

import (
	"fmt"
	"os"

	"github.com/jinzhu/configor"
	"net/url"
	"path"
	"runtime"
)

type APPConfig struct {
	APPName string
	Version string
	Env     string
}

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
	AuthMode                   string   `env:"CONFIG__AUTH_MODE" default:"both" json:"auth_mode"` // "cookie", "header", or "both"
	CacheEnabled               bool     `env:"CONFIG__CACHE_ENABLED" default:"true" json:"cache_enabled"`
	CacheTTLMinutes            int      `env:"CONFIG__CACHE_TTL_MINUTES" default:"5" json:"cache_ttl_minutes"`
	RedisURL                   string   `env:"CONFIG__REDIS_URL" default:"redis://localhost:6379" json:"redis_url"`
	RedisPassword              string   `env:"CONFIG__REDIS_PASSWORD" default:"" json:"redis_password"`
	RedisDB                    int      `env:"CONFIG__REDIS_DB" default:"0" json:"redis_db"`
	ProxyURL                   *url.URL
	APPConfig                  APPConfig
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

	// Populate APPConfig from main config fields
	config.APPConfig = APPConfig{
		APPName: getEnvOrDefault("APP__NAME", "gateway-proxy"),
		Version: config.Version,
		Env:     getEnvOrDefault("APP__ENV", "development"),
	}

	return &config, nil
}

// getEnvOrDefault gets an environment variable with a fallback default
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// LoadConfigOrPanic loads the config and panics on error
// This is useful for initialization where we can't proceed without config
func LoadConfigOrPanic() *Config {
	cfg, err := LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

func getConfigLocation() string {
	_, filename, _, _ := runtime.Caller(0)

	return path.Join(path.Dir(filename), "../config")
}
