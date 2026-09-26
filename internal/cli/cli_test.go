package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/Tianbo-Qiu/whoopctl/internal/config"
)

func TestRunVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Run(context.Background(), []string{"version"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if got, want := stdout.String(), "whoopctl dev\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestRunNoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Run(context.Background(), []string{}, &stdout, &stderr)
	if err == nil {
		t.Fatal("Run returned nil error")
	}

	if got, want := err.Error(), "usage: whoopctl <command>"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Run(context.Background(), []string{"nope"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("Run returned nil error")
	}
	if got, want := err.Error(), "unknown command: nope"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestRunAuthSetupSaveCredentials(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := &App{ConfigDir: t.TempDir()}

	err := app.Run(context.Background(), []string{
		"auth", "setup",
		"--client-id", "client-id",
		"--client-secret", "client-secret",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if got, want := stdout.String(), "WHOOP credentials saved\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}

	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestRunAuthNoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := &App{ConfigDir: t.TempDir()}

	err := app.Run(context.Background(), []string{"auth"}, &stdout, &stderr)
	if err == nil {
		t.Fatalf("Run returned nil error")
	}
	if got, want := err.Error(), "usage: whoopctl auth <command>"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestRunAuthUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := &App{ConfigDir: t.TempDir()}

	err := app.Run(context.Background(), []string{"auth", "nope"}, &stdout, &stderr)
	if err == nil {
		t.Fatalf("Run returned nil error")
	}
	if got, want := err.Error(), "unknown auth command: nope"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestRunAuthLoginPrintsAuthorizeURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	dir := t.TempDir()

	err := config.SaveCredentials(dir, config.Credentials{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	})
	if err != nil {
		t.Fatalf("SaveCredentials returned error: %v", err)
	}

	app := &App{
		ConfigDir: dir,
		StateGenerator: func() (string, error) {
			return "state-value", nil
		},
	}

	err = app.Run(context.Background(), []string{"auth", "login"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run auth login returned error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Open to authorize whoopctl:\n") {
		t.Fatalf("stdout = %q, missing authorization prompt", out)
	}

	if !strings.Contains(out, "client_id=client-id") {
		t.Fatalf("stdout = %q, missing client_id param", out)
	}

	if strings.Contains(out, "client_secret=client-secret") {
		t.Fatalf("stdout = %q, should not include client_secret param", out)
	}

	if !strings.Contains(out, "state=state-value") {
		t.Fatalf("stdout = %q, missing state param", out)
	}

	if !strings.Contains(out, "redirect_uri=http%3A%2F%2F127.0.0.1%3A1061%2Fcallback") {
		t.Fatalf("stdout = %q, missing redirect_uri param", out)
	}
}
