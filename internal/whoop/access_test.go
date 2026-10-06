package whoop

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestRevokeAccessSendsDeleteRequest(t *testing.T) {
	called := false
	client := &Client{
		BaseURL: "https://example.test/developer/",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				called = true
				if got, want := req.Method, http.MethodDelete; got != want {
					t.Fatalf("method = %q, want %q", got, want)
				}
				if got, want := req.URL.String(), "https://example.test/developer/v2/user/access"; got != want {
					t.Fatalf("url = %q, want %q", got, want)
				}
				if got, want := req.Header.Get("Authorization"), "Bearer access-token"; got != want {
					t.Fatalf("Authorization = %q, want %q", got, want)
				}

				return &http.Response{
					StatusCode: http.StatusNoContent,
					Status:     "204 No Content",
					Body:       io.NopCloser(strings.NewReader("")),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	if err := client.RevokeAccess(context.Background(), " access-token "); err != nil {
		t.Fatalf("RevokeAccess returned error: %v", err)
	}
	if !called {
		t.Fatal("RevokeAccess did not send a request")
	}
}

func TestRevokeAccessRequiresAccessToken(t *testing.T) {
	client := &Client{}

	err := client.RevokeAccess(context.Background(), "")
	if err == nil {
		t.Fatal("RevokeAccess returned nil error")
	}
	if got, want := err.Error(), "access token is required"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestRevokeAccessReturnsAPIErrorForNon2xx(t *testing.T) {
	client := &Client{
		BaseURL: "https://example.test/developer",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return jsonResponse(t, http.StatusUnauthorized, "401 Unauthorized", map[string]any{}), nil
			}),
		},
	}

	err := client.RevokeAccess(context.Background(), "access-token")
	if err == nil {
		t.Fatal("RevokeAccess returned nil error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *APIError", err)
	}
	if got, want := apiErr.StatusCode, http.StatusUnauthorized; got != want {
		t.Fatalf("StatusCode = %d, want %d", got, want)
	}
}
