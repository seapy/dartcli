package cache

import (
	"os"
	"path/filepath"
)

// Dir returns the cache directory path (~/.dartcli/cache).
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".dartcli", "cache")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// CorpCodePath returns the path to the cached corp code JSON.
func CorpCodePath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "corpcode.json"), nil
}
