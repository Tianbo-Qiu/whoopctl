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
	creds, err := normalizeCredentials(creds)
	if err != nil {
		return err
	}

	path, err := ConfigPath(configDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return os.WriteFile(path, data, 0o600)
}

func LoadCredentials(configDir string) (Credentials, error) {
	path, err := ConfigPath(configDir)
	if err != nil {
		return Credentials{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Credentials{}, err
	}

	var creds Credentials
	err = json.Unmarshal(data, &creds)
	if err != nil {
		return Credentials{}, err
	}

	creds, err = normalizeCredentials(creds)
	if err != nil {
		return Credentials{}, err
	}

	return creds, nil
}

func normalizeCredentials(creds Credentials) (Credentials, error) {
	creds.ClientID = strings.TrimSpace(creds.ClientID)
	creds.ClientSecret = strings.TrimSpace(creds.ClientSecret)
	if creds.ClientID == "" {
		return Credentials{}, fmt.Errorf("client id is required")
	}
	if creds.ClientSecret == "" {
		return Credentials{}, fmt.Errorf("client secret is required")
	}

	return creds, nil
}
