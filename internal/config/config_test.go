package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveCredentialsWritesConfigFile(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "whoopctl")

	err := SaveCredentials(dir, Credentials{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	})
	if err != nil {
		t.Fatalf("SaveCredentials returned error: %v", err)
	}

	dirInfo, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat dir returned error: %v", err)
	}

	if got, want := dirInfo.Mode().Perm(), os.FileMode(0o700); got != want {
		t.Fatalf("dir mode = %v, want %v", got, want)
	}

	path := filepath.Join(dir, "config.json")

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat returned error: %v", err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o600); got != want {
		t.Fatalf("mode = %v, want %v", got, want)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	if creds.ClientID != "client-id" {
		t.Fatalf("ClientID = %q, want %q", creds.ClientID, "client-id")
	}
	if creds.ClientSecret != "client-secret" {
		t.Fatalf("ClientSecret = %q, want %q", creds.ClientSecret, "client-secret")
	}
}

func TestSaveCredentialsRequiresClientID(t *testing.T) {
	err := SaveCredentials(t.TempDir(), Credentials{
		ClientSecret: "client-secret",
	})
	if err == nil {
		t.Fatal("SaveCredentials returned nil error")
	}
	if got, want := err.Error(), "client id is required"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestSaveCredentialsRequiresClientSecret(t *testing.T) {
	err := SaveCredentials(t.TempDir(), Credentials{
		ClientID: "client-id",
	})
	if err == nil {
		t.Fatal("SaveCredentials returned nil error")
	}
	if got, want := err.Error(), "client secret is required"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}
