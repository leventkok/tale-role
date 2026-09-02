package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/leventkok/tale-role/apps/api/internal/application/app"
	"github.com/leventkok/tale-role/apps/api/internal/application/game"
	"github.com/leventkok/tale-role/apps/api/internal/application/world"
	"github.com/leventkok/tale-role/apps/api/internal/infrastructure/httpapi"
	"github.com/leventkok/tale-role/apps/api/internal/infrastructure/mail"
	"github.com/leventkok/tale-role/apps/api/internal/infrastructure/memory"
	mongostore "github.com/leventkok/tale-role/apps/api/internal/infrastructure/mongo"
	"github.com/leventkok/tale-role/apps/api/internal/shared/config"
	gateway "github.com/leventkok/tale-role/services/llm-gateway"
)

func main() {
	cfg := config.Load()
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if cfg.JWTSecretIsDefault() {
		log.Warn("JWT_SECRET is unset; authentication uses a known default value")
	}

	var ident app.Identity = memory.NewStore()
	table := game.NewTable()
	worlds := world.NewCatalog()
	if cfg.MongoURI != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		store, err := mongostore.Connect(ctx, cfg.MongoURI, cfg.MongoDB)
		cancel()
		if err != nil {
			log.Error("mongo connect failed", "error", err)
			os.Exit(1)
		}
		ident = store
		loadCtx, loadCancel := context.WithTimeout(context.Background(), 15*time.Second)
		rooms, err := store.LoadRooms(loadCtx)
		if err != nil {
			log.Error("mongo load rooms", "error", err)
			os.Exit(1)
		}
		unis, err := store.LoadUniverses(loadCtx)
		loadCancel()
		if err != nil {
			log.Error("mongo load universes", "error", err)
			os.Exit(1)
		}
		table.Load(rooms)
		table.SetSink(store)
		worlds.Load(unis)
		worlds.SetSink(store)
		log.Info("persistence", "engine", "mongo", "rooms", len(rooms), "universes", len(unis))
	} else {
		log.Warn("MONGO_URI unset; using in-memory store (restart wipes data)")
	}

	svc := app.NewService(ident, cfg.JWTSecret, cfg.JWTExpiry, cfg.OTPTTL)
	mailer := mail.SMTP{
		Host: cfg.SMTPHost, Port: cfg.SMTPPort, From: cfg.SMTPFrom,
		User: cfg.SMTPUser, Pass: cfg.SMTPPass,
	}
	svc.Mailer = mailer
	if mailer.Enabled() {
		log.Info("mail", "transport", "smtp", "addr", mailer.Addr())
	} else {
		log.Warn("SMTP_HOST unset; OTP email is not sent")
	}
	if devOTP := os.Getenv("TALEROLE_DEV_OTP"); devOTP != "" {
		log.Warn("TALEROLE_DEV_OTP is set; using a fixed OTP issuer (local only)")
		svc.IssueOTP = func() (string, error) { return devOTP, nil }
	}
	llm := gateway.New()
	llm.ConfigureLocal(os.Getenv("TALEROLE_ADAPTER_DIR"))
	if os.Getenv("TALEROLE_ADAPTER_DIR") != "" {
		rt := llm.Runtime()
		log.Info("llm adapters", "dir_configured", rt.AdapterDirConfigured, "weights_ready", rt.WeightsReady, "inference", rt.Inference)
	}
	handler := httpapi.New(svc, table, worlds, llm, log, cfg, os.Getenv("TALEROLE_ADMIN_EMAIL"))

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("listening", "addr", addr, "persistence", cfg.Persistence())
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
