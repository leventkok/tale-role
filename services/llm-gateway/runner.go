package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/leventkok/tale-role/services/llm-gateway/internal/packs"
)

// RunnerURLsFromEnv reads local runner addresses. LLM_RUNNER_URL is the shared fallback.
// These are our processes — never a paid API host.
func RunnerURLsFromEnv() (storyteller, mechanics string) {
	storyteller = os.Getenv("LLM_STORYTELLER_URL")
	mechanics = os.Getenv("LLM_MECHANICS_URL")
	if storyteller == "" {
		storyteller = os.Getenv("LLM_RUNNER_URL")
	}
	return storyteller, mechanics
}

// SetRunners points Storyteller and mechanics at local HTTP processes.
// Empty mechanics URL reuses the storyteller URL. Neither is a paid API.
func (s *Service) SetRunners(storyteller, mechanics string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.storytellerURL = strings.TrimRight(strings.TrimSpace(storyteller), "/")
	s.mechanicsURL = strings.TrimRight(strings.TrimSpace(mechanics), "/")
	if s.mechanicsURL == "" {
		s.mechanicsURL = s.storytellerURL
	}
	if s.client == nil {
		s.client = &http.Client{Timeout: 8 * time.Second}
	}
}

func (s *Service) inferenceLocked() string {
	if s.adapter == packs.Local && s.weightsReady && s.storytellerURL != "" {
		return packs.Local
	}
	return packs.Stub
}

func (s *Service) useLocal(role string) (base string, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.adapter != packs.Local || !s.weightsReady {
		return "", false
	}
	if role == "mechanics" {
		return s.mechanicsURL, s.mechanicsURL != ""
	}
	return s.storytellerURL, s.storytellerURL != ""
}

func (s *Service) callJSON(url string, payload any, dest any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	s.mu.Lock()
	cli := s.client
	s.mu.Unlock()
	if cli == nil {
		cli = &http.Client{Timeout: 8 * time.Second}
	}
	res, err := cli.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("runner status %d", res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(dest)
}
