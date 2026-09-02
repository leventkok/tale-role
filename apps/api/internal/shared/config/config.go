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
	SMTPHost           string
	SMTPPort           string
	SMTPFrom           string
	SMTPUser           string
	SMTPPass           string
	ResendAPIKey       string
	ResendFrom         string
}

func Load() Config {
	origins := splitCSV(env("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001"))
	host, port := listenAddr()
	return Config{
		Host:               host,
		Port:               port,
		JWTSecret:          env("JWT_SECRET", "change-me-in-production"),
		JWTExpiry:          8 * time.Hour,
		CORSAllowedOrigins: origins,
		MaxBodyBytes:       envInt64("MAX_BODY_BYTES", 1<<20),
		OTPTTL:             10 * time.Minute,
		MongoURI:           env("MONGO_URI", ""),
		MongoDB:            env("MONGO_DB", "talerole"),
		SMTPHost:           env("SMTP_HOST", ""),
		SMTPPort:           env("SMTP_PORT", "1025"),
		SMTPFrom:           env("SMTP_FROM", "Tale Role <noreply@talerole.local>"),
		SMTPUser:           env("SMTP_USER", ""),
		SMTPPass:           env("SMTP_PASS", ""),
		ResendAPIKey:       env("RESEND_API_KEY", ""),
		ResendFrom:         env("RESEND_FROM", "Tale Role <onboarding@resend.dev>"),
	}
}

func (c Config) Persistence() string {
	if c.MongoURI != "" {
		return "mongo"
	}
	return "memory"
}

func (c Config) Mail() string {
	if c.ResendAPIKey != "" {
		return "resend"
	}
	if c.SMTPHost != "" {
		return "smtp"
	}
	return "none"
}

func (c Config) JWTSecretIsDefault() bool {
	return c.JWTSecret == "change-me-in-production"
}

// listenAddr binds loopback for laptop/tunnel sitting. PaaS injects PORT
// (Render, Fly, Cloud Run); then we listen on all interfaces unless SERVER_HOST is set.
func listenAddr() (host, port string) {
	port = os.Getenv("PORT")
	if port == "" {
		port = env("SERVER_PORT", "8080")
	}
	host = os.Getenv("SERVER_HOST")
	if host == "" {
		if os.Getenv("PORT") != "" {
			host = "0.0.0.0"
		} else {
			host = "127.0.0.1"
		}
	}
	return host, port
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
