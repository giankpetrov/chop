package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DiscoveryInfo contains information for AI agents to discover chop.
type DiscoveryInfo struct {
	Version string `json:"version"`
	Path    string `json:"path"`
}

// WriteDiscoveryInfo writes discovery metadata to ~/.chop/path.json.
func WriteDiscoveryInfo(version string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return err
	}

	info := DiscoveryInfo{
		Version: version,
		Path:    exe,
	}

	dir := filepath.Join(home, ".chop")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	path := filepath.Join(dir, "path.json")
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// DiscoveryPath returns the path to the discovery file.
func DiscoveryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".chop", "path.json"), nil
}
