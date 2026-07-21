package config

import (
	"errors"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	Addr          string
	FrontendURL   string
	AuthServerURL string
	DatabaseURL   string
}

func Load() (Config, error) {
	// 環境変数から API の起動設定と sandbox_auth 連携設定を読み込む。
	cfg := Config{
		Addr:          GetEnv("ADDR", ":8080"),
		FrontendURL:   strings.TrimRight(GetEnv("FRONTEND_URL", "http://localhost:3000"), "/"),
		AuthServerURL: strings.TrimRight(GetEnv("AUTH_SERVER_URL", "http://localhost:8080"), "/"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
	}

	if !isAbsoluteHTTPURL(cfg.FrontendURL) {
		return cfg, errors.New("FRONTEND_URL must be an absolute http or https URL")
	}
	if !isAbsoluteHTTPURL(cfg.AuthServerURL) {
		return cfg, errors.New("AUTH_SERVER_URL must be an absolute http or https URL")
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
