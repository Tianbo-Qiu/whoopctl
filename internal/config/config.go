package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Credentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type StoredToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope"`
	TokenType    string    `json:"token_type"`
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

func TokenPath(configDir string) (string, error) {
	if configDir == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(dir, "whoopctl")
	}
	return filepath.Join(configDir, "token.json"), nil
}

func SaveCredentials(configDir string, creds Credentials) error {
	creds, err := NormalizeCredentials(creds)
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

	creds, err = NormalizeCredentials(creds)
	if err != nil {
		return Credentials{}, err
	}

	return creds, nil
}

func NormalizeCredentials(creds Credentials) (Credentials, error) {
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

func SaveToken(configDir string, token StoredToken) error {
	token, err := NormalizeToken(token)
	if err != nil {
		return err
	}

	path, err := TokenPath(configDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return os.WriteFile(path, data, 0o600)
}

func LoadToken(configDir string) (StoredToken, error) {
	path, err := TokenPath(configDir)
	if err != nil {
		return StoredToken{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return StoredToken{}, err
	}

	var token StoredToken
	if err := json.Unmarshal(data, &token); err != nil {
		return StoredToken{}, err
	}

	return NormalizeToken(token)
}

func DeleteToken(configDir string) error {
	path, err := TokenPath(configDir)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func NormalizeToken(token StoredToken) (StoredToken, error) {
	token.AccessToken = strings.TrimSpace(token.AccessToken)
	token.RefreshToken = strings.TrimSpace(token.RefreshToken)
	token.Scope = strings.TrimSpace(token.Scope)
	token.TokenType = strings.TrimSpace(token.TokenType)

	if token.AccessToken == "" {
		return StoredToken{}, fmt.Errorf("access token is required")
	}
	if token.RefreshToken == "" {
		return StoredToken{}, fmt.Errorf("refresh token is required")
	}
	if token.ExpiresAt.IsZero() {
		return StoredToken{}, fmt.Errorf("expires at is required")
	}
	if token.TokenType == "" {
		return StoredToken{}, fmt.Errorf("token type is required")
	}
	if !strings.EqualFold(token.TokenType, "bearer") {
		return StoredToken{}, fmt.Errorf("unsupported token type: %s", token.TokenType)
	}
	token.TokenType = "bearer"

	return token, nil
}
