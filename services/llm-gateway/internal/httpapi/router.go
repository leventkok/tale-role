package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	gateway "github.com/leventkok/tale-role/services/llm-gateway"
)

func New(svc *gateway.Service, adminToken string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Get("/health/live", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
	})
	r.Post("/v1/narrate", func(w http.ResponseWriter, r *http.Request) {
		var req gateway.NarrateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, svc.Narrate(req))
	})
	r.Post("/v1/intent", func(w http.ResponseWriter, r *http.Request) {
		var req gateway.IntentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, svc.ProposeIntent(req))
	})
	r.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if adminToken == "" || !strings.HasPrefix(req.Header.Get("Authorization"), "Bearer ") {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				got := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
				if got != adminToken {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				next.ServeHTTP(w, req)
			})
		})
		r.Get("/v1/runtime", func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, http.StatusOK, svc.Runtime())
		})
		r.Put("/v1/runtime", func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				PromptPack string `json:"prompt_pack"`
				AdapterID  string `json:"adapter_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid request", http.StatusBadRequest)
				return
			}
			if err := svc.Swap(body.PromptPack, body.AdapterID); err != nil {
				http.Error(w, "invalid request", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, svc.Runtime())
		})
		r.Get("/v1/traces", func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{"traces": svc.Traces()})
		})
	})
	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
