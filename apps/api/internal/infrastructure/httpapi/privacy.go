package httpapi

import (
	"net/http"
	"time"

	"github.com/leventkok/tale-role/apps/api/internal/shared/httperr"
)

func (s *Server) exportMe(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	subject, err := s.svc.ExportSubject(u.ID)
	if err != nil {
		s.writeAppError(w, err)
		return
	}
	rooms, chars := s.table.ExportFor(u.ID)
	licenses := s.svc.Licenses(u.ID)
	licJSON := make([]map[string]any, 0, len(licenses))
	for _, l := range licenses {
		licJSON = append(licJSON, map[string]any{
			"id":         l.ID,
			"device_id":  l.DeviceID,
			"platform":   l.Platform,
			"created_at": l.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	httperr.JSON(w, http.StatusOK, map[string]any{
		"exported_at": time.Now().UTC().Format(time.RFC3339),
		"subject":     subject,
		"licenses":    licJSON,
		"universes":   s.worlds.ExportFor(u.ID),
		"rooms":       rooms,
		"characters":  chars,
	})
}

func (s *Server) eraseMe(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	s.table.ForgetUser(u.ID)
	s.worlds.ForgetOwner(u.ID)
	if err := s.svc.Erase(u.ID); err != nil {
		s.writeAppError(w, err)
		return
	}
	httperr.JSON(w, http.StatusOK, map[string]any{"ok": true, "erased": true})
}

func (s *Server) ready(w http.ResponseWriter, _ *http.Request) {
	rt := s.llm.Runtime()
	httperr.JSON(w, http.StatusOK, map[string]any{
		"status":      "ready",
		"persistence": s.cfg.Persistence(),
		"llm":         rt.Inference,
		"images":      "stub",
	})
}
