package gateway

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ProbeWeights is true when dir looks like our private adapter export.
// Inference is "local" only when this is true and a runner URL is set.
func ProbeWeights(dir string) bool {
	if strings.TrimSpace(dir) == "" {
		return false
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}
	found := false
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || found {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		if strings.Count(rel, string(os.PathSeparator)) > 2 {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		name := strings.ToLower(d.Name())
		if strings.HasSuffix(name, ".safetensors") || strings.HasSuffix(name, ".gguf") || name == "adapter_config.json" {
			found = true
			return fs.SkipAll
		}
		return nil
	})
	return found
}
