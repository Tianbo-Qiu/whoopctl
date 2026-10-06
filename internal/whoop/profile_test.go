package whoop

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestBasicProfileSendsRequestAndDecodesResponse(t *testing.T) {
	client := &Client{
		BaseURL: "https://example.test/developer/",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if got, want := req.Method, http.MethodGet; got != want {
					t.Fatalf("method = %q, want %q", got, want)
				}
				if got, want := req.URL.String(), "https://example.test/developer/v2/user/profile/basic"; got != want {
					t.Fatalf("url = %q, want %q", got, want)
				}
				if got, want := req.Header.Get("Authorization"), "Bearer access-token"; got != want {
					t.Fatalf("Authorization = %q, want %q", got, want)
				}

				return jsonResponse(t, http.StatusOK, "200 OK", map[string]any{
					"user_id":    10129,
					"email":      "user@example.test",
					"first_name": "Example",
					"last_name":  "User",
				}), nil
			}),
		},
	}

	profile, err := client.BasicProfile(context.Background(), " access-token ")
	if err != nil {
		t.Fatalf("BasicProfile returned error: %v", err)
	}
	if got, want := profile.UserID, int64(10129); got != want {
		t.Fatalf("UserID = %d, want %d", got, want)
	}
	if got, want := profile.Email, "user@example.test"; got != want {
		t.Fatalf("Email = %q, want %q", got, want)
	}
}

func TestBasicProfileRequiresAccessToken(t *testing.T) {
	client := &Client{}

	_, err := client.BasicProfile(context.Background(), "")
	if err == nil {
		t.Fatal("BasicProfile returned nil error")
	}
	if got, want := err.Error(), "access token is required"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestBasicProfileReturnsAPIErrorForNon2xx(t *testing.T) {
	client := &Client{
		BaseURL: "https://example.test/developer",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return jsonResponse(t, http.StatusUnauthorized, "401 Unauthorized", map[string]any{}), nil
			}),
		},
	}

	_, err := client.BasicProfile(context.Background(), "access-token")
	if err == nil {
		t.Fatal("BasicProfile returned nil error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *APIError", err)
	}
	if got, want := apiErr.StatusCode, http.StatusUnauthorized; got != want {
		t.Fatalf("StatusCode = %d, want %d", got, want)
	}
}
