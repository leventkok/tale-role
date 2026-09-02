package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/leventkok/tale-role/apps/api/internal/application/app"
	"github.com/leventkok/tale-role/apps/api/internal/application/game"
	"github.com/leventkok/tale-role/apps/api/internal/application/world"
	"github.com/leventkok/tale-role/apps/api/internal/shared/config"
	"github.com/leventkok/tale-role/apps/api/internal/shared/httperr"
	worker "github.com/leventkok/tale-role/services/image-worker"
	gateway "github.com/leventkok/tale-role/services/llm-gateway"
)

type Server struct {
	svc        *app.Service
	table      *game.Table
	worlds     *world.Catalog
	llm        *gateway.Service
	images     *worker.Service
	log        *slog.Logger
	cfg        config.Config
	adminEmail string
}

func New(svc *app.Service, table *game.Table, worlds *world.Catalog, llm *gateway.Service, log *slog.Logger, cfg config.Config, adminEmail string) http.Handler {
	if llm == nil {
		llm = gateway.New()
	}
	if worlds == nil {
		worlds = world.NewCatalog()
	}
	s := &Server{svc: svc, table: table, worlds: worlds, llm: llm, images: worker.New(), log: log, cfg: cfg, adminEmail: adminEmail}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(maxBody(cfg.MaxBodyBytes))
	r.Use(securityHeaders)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: allowCredentials(cfg.CORSAllowedOrigins),
		MaxAge:           300,
	}))

	r.Get("/health/live", func(w http.ResponseWriter, _ *http.Request) {
		httperr.JSON(w, http.StatusOK, map[string]string{"status": "alive"})
	})
	r.Get("/health/ready", s.ready)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", s.register)
		r.Post("/auth/login", s.login)
		r.Post("/auth/otp/verify", s.verifyOTP)
		r.Group(func(r chi.Router) {
			r.Use(s.auth)
			r.Get("/me", s.me)
			r.Get("/me/export", s.exportMe)
			r.Delete("/me", s.eraseMe)
			r.Post("/licenses/register", s.registerLicense)
			r.Get("/licenses/me", s.myLicenses)
			r.Post("/rooms", s.createRoom)
			r.Get("/universes", s.listUniverses)
			r.Post("/universes", s.createUniverse)
			r.Get("/universes/{universeID}", s.getUniverse)
			r.Get("/rooms/{roomID}", s.getRoom)
			r.Post("/rooms/{roomID}/join", s.joinRoom)
			r.Post("/rooms/{roomID}/characters", s.setCharacter)
			r.Post("/rooms/{roomID}/start", s.startRoom)
			r.Post("/rooms/{roomID}/turns", s.actRoom)
			r.Group(func(r chi.Router) {
				r.Use(s.adminOnly)
				r.Get("/admin/runtime", s.adminRuntime)
				r.Put("/admin/runtime", s.adminSwap)
				r.Get("/admin/traces", s.adminTraces)
			})
		})
	})
	return r
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httperr.Write(w, s.log, http.StatusBadRequest, "invalid request", err)
		return
	}
	if err := s.svc.Register(body.Email, body.Password); err != nil {
		s.writeAppError(w, err)
		return
	}
	httperr.JSON(w, http.StatusAccepted, map[string]any{"otp_required": true})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httperr.Write(w, s.log, http.StatusBadRequest, "invalid request", err)
		return
	}
	token, err := s.svc.Login(body.Email, body.Password)
	if errors.Is(err, app.ErrOTPRequired) {
		httperr.JSON(w, http.StatusUnauthorized, map[string]any{"otp_required": true, "error": "otp required"})
		return
	}
	if err != nil {
		s.writeAppError(w, err)
		return
	}
	httperr.JSON(w, http.StatusOK, map[string]any{"token": token})
}

func (s *Server) verifyOTP(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httperr.Write(w, s.log, http.StatusBadRequest, "invalid request", err)
		return
	}
	token, err := s.svc.VerifyOTP(body.Email, body.Code)
	if err != nil {
		s.writeAppError(w, err)
		return
	}
	httperr.JSON(w, http.StatusOK, map[string]any{"token": token})
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	httperr.JSON(w, http.StatusOK, map[string]any{
		"id":       u.ID,
		"email":    u.Email,
		"verified": u.Verified,
	})
}

func (s *Server) registerLicense(w http.ResponseWriter, r *http.Request) {
	var body struct {
		DeviceID string `json:"device_id"`
		Platform string `json:"platform"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httperr.Write(w, s.log, http.StatusBadRequest, "invalid request", err)
		return
	}
	u := userFrom(r)
	lic, err := s.svc.RegisterLicense(u.ID, body.DeviceID, body.Platform)
	if err != nil {
		s.writeAppError(w, err)
		return
	}
	httperr.JSON(w, http.StatusCreated, map[string]any{
		"id":        lic.ID,
		"device_id": lic.DeviceID,
		"platform":  lic.Platform,
	})
}

func (s *Server) myLicenses(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	list := s.svc.Licenses(u.ID)
	rows := make([]map[string]any, 0, len(list))
	for _, l := range list {
		rows = append(rows, map[string]any{
			"id":         l.ID,
			"device_id":  l.DeviceID,
			"platform":   l.Platform,
			"created_at": l.CreatedAt.Format(time.RFC3339),
		})
	}
	httperr.JSON(w, http.StatusOK, map[string]any{"licenses": rows})
}

func (s *Server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			httperr.Write(w, s.log, http.StatusUnauthorized, "unauthorized", nil)
			return
		}
		u, err := s.svc.UserFromToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			httperr.Write(w, s.log, http.StatusUnauthorized, "unauthorized", nil)
			return
		}
		ctx := withUser(r.Context(), u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) adminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := userFrom(r)
		if u == nil || s.adminEmail == "" || u.Email != s.adminEmail {
			httperr.Write(w, s.log, http.StatusForbidden, "forbidden", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) writeAppError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, app.ErrEmailTaken):
		httperr.Write(w, s.log, http.StatusConflict, "email taken", err)
	case errors.Is(err, app.ErrOTPInvalid), errors.Is(err, app.ErrInvalidCredentials):
		httperr.Write(w, s.log, http.StatusUnauthorized, "unauthorized", err)
	case errors.Is(err, app.ErrUnauthorized):
		httperr.Write(w, s.log, http.StatusUnauthorized, "unauthorized", err)
	case errors.Is(err, game.ErrNotFound), errors.Is(err, world.ErrNotFound):
		httperr.Write(w, s.log, http.StatusNotFound, "not found", err)
	case errors.Is(err, game.ErrBadPassword), errors.Is(err, game.ErrForbidden), errors.Is(err, world.ErrForbidden):
		httperr.Write(w, s.log, http.StatusForbidden, "forbidden", err)
	case errors.Is(err, game.ErrBadStats), errors.Is(err, game.ErrUnknownDice), errors.Is(err, game.ErrUnknownSkill), errors.Is(err, game.ErrHasCharacter), errors.Is(err, game.ErrNoCharacter), errors.Is(err, world.ErrInvalid):
		httperr.Write(w, s.log, http.StatusBadRequest, "invalid request", err)
	default:
		httperr.Write(w, s.log, http.StatusInternalServerError, "an internal error occurred", err)
	}
}

func maxBody(n int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, n)
			next.ServeHTTP(w, r)
		})
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func allowCredentials(origins []string) bool {
	if len(origins) == 0 {
		return false
	}
	for _, o := range origins {
		if o == "*" {
			return false
		}
	}
	return true
}
