package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Credentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func ConfigPath(configDir string) (string, error) {
	if configDir == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(dir, "whoopctl")
	}
	return filepath.Join(configDir, "config.json"), nil
}

func SaveCredentials(configDir string, creds Credentials) error {
	ClientID := strings.TrimSpace(creds.ClientID)
	ClientSecret := strings.TrimSpace(creds.ClientSecret)
	if ClientID == "" {
		return fmt.Errorf("client id is required")
	}
	if ClientSecret == "" {
		return fmt.Errorf("client secret is required")
	}

	path, err := ConfigPath(configDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(Credentials{ClientID, ClientSecret}, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return os.WriteFile(path, data, 0o600)
}
