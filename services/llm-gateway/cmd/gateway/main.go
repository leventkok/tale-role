package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	gateway "github.com/leventkok/tale-role/services/llm-gateway"
	"github.com/leventkok/tale-role/services/llm-gateway/internal/httpapi"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	host := getenv("GATEWAY_HOST", "127.0.0.1")
	port := getenv("GATEWAY_PORT", "8090")
	token := os.Getenv("LLM_GATEWAY_ADMIN_TOKEN")
	if token == "" {
		log.Warn("LLM_GATEWAY_ADMIN_TOKEN is unset; admin swap stays locked")
	}
	svc := gateway.New()
	svc.ConfigureHub(os.Getenv("HF_STORYTELLER_MODEL"), os.Getenv("HF_MECHANICS_MODEL"))
	svc.SetRunners(gateway.RunnerURLsFromEnv())
	addr := fmt.Sprintf("%s:%s", host, port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      httpapi.New(svc, token),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	log.Info("llm-gateway listening", "addr", addr, "inference", svc.Runtime().Inference)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("gateway error", "error", err)
		os.Exit(1)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
