package auth

import (
	"strings"
	"testing"
)

func TestStateGeneratorReturnsURLSafeRandomState(t *testing.T) {
	state, err := StateGenerator()
	if err != nil {
		t.Fatalf("StateGenerator returned error: %v", err)
	}
	if state == "" {
		t.Fatal("state is empty")
	}
	if strings.ContainsAny(state, "+/=") {
		t.Fatalf("state = %q, contains non-URL-safe chars", state)
	}
}
