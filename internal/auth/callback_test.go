package auth

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestParseAuthorizationCallbackReturnsCodeAndState(t *testing.T) {
	req := httptest.NewRequest("GET", "/callback?code=auth-code&state=state-value", nil)

	res, err := ParseAuthorizationCallback(req, "state-value")
	if err != nil {
		t.Fatalf("ParseAuthorizationCallback returned error: %v", err)
	}

	if got, want := res.Code, "auth-code"; got != want {
		t.Fatalf("Code = %q, want %q", got, want)
	}

	if got, want := res.State, "state-value"; got != want {
		t.Fatalf("State = %q, want %q", got, want)
	}
}

func TestParseAuthorizationCallbackRequiresMatchingState(t *testing.T) {
	req := httptest.NewRequest("GET", "/callback?code=auth-code&state=wrong-state", nil)

	_, err := ParseAuthorizationCallback(req, "state-value")
	if err == nil {
		t.Fatalf("ParseAuthorizationCallback returned nil error")
	}
	if got, want := err.Error(), "state mismatch"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestWaitForAuthorizationCallbackReceivesCallback(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen returned error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resCh := make(chan AuthorizationCallback, 1)
	errCh := make(chan error, 1)

	go func() {
		res, err := waitForAuthorizationCallback(ctx, listener, "state-value")
		if err != nil {
			errCh <- err
			return
		}
		resCh <- res
	}()

	url := "http://" + listener.Addr().String() + "/callback?code=auth-code&state=state-value"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext returned error: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}

	if got, want := resp.StatusCode, http.StatusOK; got != want {
		t.Fatalf("status = %d, want = %d", got, want)
	}

	if !strings.Contains(string(body), "Authorization complete") {
		t.Fatalf("body = %q, want completion message", string(body))
	}

	select {
	case res := <-resCh:
		if got, want := res.Code, "auth-code"; got != want {
			t.Fatalf("Code = %q, want %q", got, want)
		}
		if got, want := res.State, "state-value"; got != want {
			t.Fatalf("State = %q, want %q", got, want)
		}
	case err := <-errCh:
		t.Fatalf("waitForAuthorizationCallback returned error: %v", err)
	case <-ctx.Done():
		t.Fatalf("context ended before callback result: %v", ctx.Err())
	}
}
