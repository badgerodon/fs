package yandex

import (
	"net/http"
	"os"
)

// Default values for the file system.
var (
	DefaultAccessToken = os.Getenv("YANDEX_ACCESS_TOKEN")
	DefaultBaseURL     = "https://cloud-api.yandex.net/"
)

type config struct {
	accessToken string
	baseURL     string
	httpClient  *http.Client
}

// An Option customizes the config.
type Option func(*config)

// WithAccessToken sets the access token in the config.
func WithAccessToken(accessToken string) Option {
	return func(cfg *config) {
		cfg.accessToken = accessToken
	}
}

// WithBaseURL sets the base URL in the config.
func WithBaseURL(baseURL string) Option {
	return func(cfg *config) {
		cfg.baseURL = baseURL
	}
}

// WithHTTPClient sets the default HTTP client in the config.
func WithHTTPClient(client *http.Client) Option {
	return func(cfg *config) {
		cfg.httpClient = client
	}
}

func getConfig(opts ...Option) *config {
	cfg := new(config)
	WithAccessToken(DefaultAccessToken)(cfg)
	WithBaseURL(DefaultBaseURL)(cfg)
	WithHTTPClient(http.DefaultClient)(cfg)
	return cfg
}
