package config_test

import (
	"errors"
	"testing"

	"github.com/Brunotlps/codda/internal/config"
)

const testDatabaseURL = "postgres://codda:codda@localhost:5433/codda?sslmode=disable"

func TestLoad(t *testing.T) {
	t.Run("uses default HTTP port", func(t *testing.T) {
		t.Setenv("DATABASE_URL", testDatabaseURL)
		t.Setenv("HTTP_PORT", "")

		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("Load() returned unexpected error: %v", err)
		}

		if cfg.HTTPAddr != ":8080" {
			t.Errorf("HTTPAddr = %q, want %q", cfg.HTTPAddr, ":8080")
		}
		if cfg.DatabaseURL != testDatabaseURL {
			t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, testDatabaseURL)
		}
	})

	t.Run("accepts boundary ports", func(t *testing.T) {
		tests := []struct {
			name     string
			port     string
			wantAddr string
		}{
			{name: "lowest valid port", port: "1", wantAddr: ":1"},
			{name: "highest valid port", port: "65535", wantAddr: ":65535"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Setenv("DATABASE_URL", testDatabaseURL)
				t.Setenv("HTTP_PORT", tt.port)

				cfg, err := config.Load()
				if err != nil {
					t.Fatalf("Load() returned unexpected error: %v", err)
				}

				if cfg.HTTPAddr != tt.wantAddr {
					t.Errorf("HTTPAddr = %q, want %q", cfg.HTTPAddr, tt.wantAddr)
				}
			})
		}
	})

	t.Run("rejects invalid ports", func(t *testing.T) {
		tests := []struct {
			name    string
			port    string
			wantErr error
		}{
			{name: "non numeric", port: "abc", wantErr: nil},
			{name: "zero", port: "0", wantErr: config.ErrInvalidPort},
			{name: "above max", port: "65536", wantErr: config.ErrInvalidPort},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Setenv("DATABASE_URL", testDatabaseURL)
				t.Setenv("HTTP_PORT", tt.port)

				cfg, err := config.Load()
				if tt.wantErr != nil {
					if !errors.Is(err, tt.wantErr) {
						t.Fatalf("Load() error = %v, want %v", err, tt.wantErr)
					}
				} else if err == nil {
					t.Fatal("Load() error = nil, want an error")
				}
				if cfg != nil {
					t.Errorf("Load() cfg = %v, want nil", cfg)
				}
			})
		}
	})

	t.Run("requires database URL", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "")
		t.Setenv("HTTP_PORT", "8080")

		cfg, err := config.Load()
		if !errors.Is(err, config.ErrMissingDatabaseURL) {
			t.Fatalf("Load() error = %v, want %v", err, config.ErrMissingDatabaseURL)
		}
		if cfg != nil {
			t.Errorf("Load() cfg = %v, want nil", cfg)
		}
	})
}
