package cli

import (
	"bytes"
	"context"
	"testing"
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
