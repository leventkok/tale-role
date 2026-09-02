package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gateway "github.com/leventkok/tale-role/services/llm-gateway"
	"github.com/leventkok/tale-role/services/llm-gateway/internal/httpapi"
)

func TestAdminTokenRequiredForSwap(t *testing.T) {
	h := httpapi.New(gateway.New(), "change-me")
	req := httptest.NewRequest(http.MethodPut, "/v1/runtime", strings.NewReader(`{"prompt_pack":"v1-terse"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthed swap: %d", rec.Code)
	}
	req = httptest.NewRequest(http.MethodPut, "/v1/runtime", strings.NewReader(`{"prompt_pack":"v1-terse"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer change-me")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("authed swap: %d %s", rec.Code, rec.Body.String())
	}
}
