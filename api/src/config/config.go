package config

import (
	"errors"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	Addr             string
	FrontendURL      string
	OIDCIssuerURL    string
	OIDCInternalURL  string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCRedirectURL  string
	DatabaseURL      string
	SessionSecret    []byte
}

func Load() (Config, error) {
	frontendURL := strings.TrimRight(GetEnv("FRONTEND_URL", "http://localhost:3000"), "/")
	oidcIssuerURL := strings.TrimRight(GetEnv("OIDC_ISSUER_URL", "http://localhost:8080"), "/")
	cfg := Config{
		Addr:             GetEnv("ADDR", ":8080"),
		FrontendURL:      frontendURL,
		OIDCIssuerURL:    oidcIssuerURL,
		OIDCInternalURL:  strings.TrimRight(GetEnv("OIDC_INTERNAL_URL", oidcIssuerURL), "/"),
		OIDCClientID:     os.Getenv("OIDC_CLIENT_ID"),
		OIDCClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
		OIDCRedirectURL:  GetEnv("OIDC_REDIRECT_URL", "http://localhost:8080/auth/callback"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		SessionSecret:    []byte(os.Getenv("APP_AUTH_SECRET")),
	}

	if !isAbsoluteHTTPURL(cfg.FrontendURL) {
		return cfg, errors.New("FRONTEND_URL must be an absolute http or https URL")
	}
	if !isAbsoluteHTTPURL(cfg.OIDCIssuerURL) {
		return cfg, errors.New("OIDC_ISSUER_URL must be an absolute http or https URL")
	}
	if !isAbsoluteHTTPURL(cfg.OIDCInternalURL) {
		return cfg, errors.New("OIDC_INTERNAL_URL must be an absolute http or https URL")
	}
	if !isAbsoluteHTTPURL(cfg.OIDCRedirectURL) {
		return cfg, errors.New("OIDC_REDIRECT_URL must be an absolute http or https URL")
	}
	if cfg.OIDCClientID == "" || cfg.OIDCClientSecret == "" {
		return cfg, errors.New("OIDC_CLIENT_ID and OIDC_CLIENT_SECRET are required")
	}
	if len(cfg.SessionSecret) < 32 {
		return cfg, errors.New("APP_AUTH_SECRET must be at least 32 bytes")
	}
	if cfg.DatabaseURL == "" {
		return cfg, errors.New("DATABASE_URL is required")
	}

	return cfg, nil
}

func GetEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func isAbsoluteHTTPURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return parsed.IsAbs() && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https")
}
