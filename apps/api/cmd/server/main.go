package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/leventkok/tale-role/apps/api/internal/application/app"
	"github.com/leventkok/tale-role/apps/api/internal/infrastructure/httpapi"
	"github.com/leventkok/tale-role/apps/api/internal/infrastructure/memory"
	"github.com/leventkok/tale-role/apps/api/internal/shared/config"
)

func main() {
	cfg := config.Load()
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if cfg.JWTSecretIsDefault() {
		log.Warn("JWT_SECRET is unset; authentication uses a known default value")
	}

	svc := app.NewService(memory.NewStore(), cfg.JWTSecret, cfg.JWTExpiry, cfg.OTPTTL)
	handler := httpapi.New(svc, log, cfg)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	_ = srv.Close()
}
