package httperr

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type body struct {
	Error string `json:"error"`
}

func Write(w http.ResponseWriter, log *slog.Logger, status int, public string, err error) {
	if status >= 500 && log != nil && err != nil {
		log.Error("internal error", "error", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body{Error: public})
}

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
