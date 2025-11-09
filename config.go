package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type appConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func loadOrCreateConfig(path string) (appConfig, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg := appConfig{
				Username: "admin",
				Password: "admin123",
			}
			if err := writeConfig(path, cfg); err != nil {
				return appConfig{}, false, err
			}
			return cfg, true, nil
		}
		return appConfig{}, false, err
	}

	// Allow UTF-8 BOM that some editors add.
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

	var cfg appConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return appConfig{}, false, fmt.Errorf("parse config: %w", err)
	}

	cfg.Username = strings.TrimSpace(cfg.Username)
	cfg.Password = strings.TrimSpace(cfg.Password)
	if cfg.Username == "" || cfg.Password == "" {
		return appConfig{}, false, errors.New("config requires non-empty username and password")
	}

	return cfg, false, nil
}

func writeConfig(path string, cfg appConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
