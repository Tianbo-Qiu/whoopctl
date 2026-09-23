package auth

import (
	"strings"
	"testing"
)

func TestNewStateReturnsURLSafeRandomState(t *testing.T) {
	state, err := NewState()
	if err != nil {
		t.Fatalf("NewState returned error: %v", err)
	}
	if state == "" {
		t.Fatal("state is empty")
	}
	if strings.ContainsAny(state, "+/=") {
		t.Fatalf("state = %q, contains non-URL-safe chars", state)
	}
}
