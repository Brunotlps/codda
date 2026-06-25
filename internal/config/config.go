package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

// defaultHTTPPort is used when the HTTP_PORT environment variable is unset.
const defaultHTTPPort = "8080"

// ErrInvalidPort is returned by Load when HTTP_PORT is set to a value
// outside the valid TCP port range.
var ErrInvalidPort = errors.New("HTTP_PORT must be between 1 and 65535")

// Config holds the application's runtime configuration, populated from
// environment variables.
type Config struct {
	// HTTPAddr is the address the HTTP server listens on (e.g. ":8080").
	HTTPAddr string
}

// Load reads configuration from environment variables, applying defaults
// where they are unset, and returns an error if any value is invalid.
func Load() (*Config, error) {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = defaultHTTPPort
	}

	n, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("HTTP_PORT: %w", err)
	}
	if n < 1 || n > 65535 {
		return nil, ErrInvalidPort
	}

	return &Config{
		HTTPAddr: ":" + port,
	}, nil
}
