package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host               string
	Port               string
	JWTSecret          string
	JWTExpiry          time.Duration
	CORSAllowedOrigins []string
	MaxBodyBytes       int64
	OTPTTL             time.Duration
	MongoURI           string
	MongoDB            string
}

func Load() Config {
	origins := splitCSV(env("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001"))
	return Config{
		Host:               env("SERVER_HOST", "127.0.0.1"),
		Port:               env("SERVER_PORT", "8080"),
		JWTSecret:          env("JWT_SECRET", "change-me-in-production"),
		JWTExpiry:          8 * time.Hour,
		CORSAllowedOrigins: origins,
		MaxBodyBytes:       envInt64("MAX_BODY_BYTES", 1<<20),
		OTPTTL:             10 * time.Minute,
		MongoURI:           env("MONGO_URI", ""),
		MongoDB:            env("MONGO_DB", "talerole"),
	}
}

func (c Config) Persistence() string {
	if c.MongoURI != "" {
		return "mongo"
	}
	return "memory"
}

func (c Config) JWTSecretIsDefault() bool {
	return c.JWTSecret == "change-me-in-production"
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt64(key string, fallback int64) int64 {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
