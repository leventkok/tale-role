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

const runnerTimeout = 30 * time.Second

// RunnerURLsFromEnv reads our GPU runner addresses. Comma-separated
// LLM_*_URLS lists are round-robin replicas. Never a paid API host.
func RunnerURLsFromEnv() (storyteller, mechanics string) {
	storyteller = firstEnv("LLM_STORYTELLER_URLS", "LLM_STORYTELLER_URL", "LLM_RUNNER_URL")
	mechanics = firstEnv("LLM_MECHANICS_URLS", "LLM_MECHANICS_URL")
	return storyteller, mechanics
}

func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}

func parseURLList(raw string) []string {
	out := make([]string, 0, 2)
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimRight(strings.TrimSpace(p), "/")
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func runnerClient() *http.Client {
	return &http.Client{Timeout: runnerTimeout}
}

// SetRunners points Storyteller and mechanics at our HTTP processes.
// Comma-separated values are replicas. Empty mechanics reuses the storyteller list.
func (s *Service) SetRunners(storyteller, mechanics string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.storytellerURLs = parseURLList(storyteller)
	s.mechanicsURLs = parseURLList(mechanics)
	if len(s.mechanicsURLs) == 0 {
		s.mechanicsURLs = append([]string{}, s.storytellerURLs...)
	}
	if s.client == nil {
		s.client = runnerClient()
	}
}

func (s *Service) inferenceLocked() string {
	if (s.adapter == packs.Hub || s.adapter == packs.Local) && s.weightsReady && len(s.storytellerURLs) > 0 {
		return packs.Hub
	}
	return packs.Stub
}

func (s *Service) replicaURLs(role string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	src := s.storytellerURLs
	if role == "mechanics" {
		src = s.mechanicsURLs
	}
	out := make([]string, len(src))
	copy(out, src)
	return out
}

func (s *Service) nextReplica(role string, n int) int {
	if n <= 0 {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if role == "mechanics" {
		s.mechanicsRR++
		return int(s.mechanicsRR-1) % n
	}
	s.storytellerRR++
	return int(s.storytellerRR-1) % n
}

func (s *Service) callRole(role, path string, payload, dest any) bool {
	urls := s.replicaURLs(role)
	if len(urls) == 0 {
		return false
	}
	start := s.nextReplica(role, len(urls))
	for i := 0; i < len(urls); i++ {
		base := urls[(start+i)%len(urls)]
		if err := s.callJSON(base+path, payload, dest); err == nil {
			return true
		}
	}
	return false
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
		cli = runnerClient()
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
