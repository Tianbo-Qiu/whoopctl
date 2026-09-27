package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/Tianbo-Qiu/whoopctl/internal/config"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestExchangeAuthorizationCodeSendsTokenRequest(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if got, want := req.Method, http.MethodPost; got != want {
				t.Fatalf("method = %q, want %q", got, want)
			}

			if got, want := req.URL.String(), TokenEndpoint; got != want {
				t.Fatalf("url = %q, want = %q", got, want)
			}

			if got, want := req.Header.Get("Content-Type"), "application/x-www-form-urlencoded"; got != want {
				t.Fatalf("Content-Type = %q, want %q", got, want)
			}

			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("ReadAll returned error: %v", err)
			}

			form, err := url.ParseQuery(string(body))
			if err != nil {
				t.Fatalf("ParseQuery returned error: %v", err)
			}

			assertFormValue(t, form, "grant_type", "authorization_code")
			assertFormValue(t, form, "code", "auth-code")
			assertFormValue(t, form, "redirect_uri", RedirectURI)
			assertFormValue(t, form, "client_id", "client-id")
			assertFormValue(t, form, "client_secret", "client-secret")

			return jsonResponse(t, http.StatusOK, "200 OK", map[string]any{
				"access_token":  "access-token",
				"refresh_token": "refresh-token",
				"expires_in":    3600,
				"scope":         "offline read:recovery",
				"token_type":    "Bearer",
			}), nil
		}),
	}

	resp, err := ExchangeAuthorizationCode(context.Background(), client, config.Credentials{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}, "auth-code")
	if err != nil {
		t.Fatalf("ExchangeAuthorizationCode returned error: %v", err)
	}

	if got, want := resp.AccessToken, "access-token"; got != want {
		t.Fatalf("AccessToken = %q, want %q", got, want)
	}

	if got, want := resp.RefreshToken, "refresh-token"; got != want {
		t.Fatalf("RefreshToken = %q, want %q", got, want)
	}

	if got, want := resp.ExpiresIn, 3600; got != want {
		t.Fatalf("ExpiresIn = %d, want %d", got, want)
	}

	if got, want := resp.Scope, "offline read:recovery"; got != want {
		t.Fatalf("Scope = %q, want %q", got, want)
	}

	if got, want := resp.TokenType, "bearer"; got != want {
		t.Fatalf("TokenType = %q, want %q", got, want)
	}
}

func assertFormValue(t *testing.T, form url.Values, key, want string) {
	t.Helper()

	if got := form.Get(key); got != want {
		t.Fatalf("%s = %q, want %q", key, got, want)
	}
}

func jsonResponse(t *testing.T, statusCode int, status string, body map[string]any) *http.Response {
	t.Helper()

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	return &http.Response{
		StatusCode: statusCode,
		Status:     status,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(data)),
	}
}

func TestExchangeAuthorizationCodeReturnsErrorForNon2xx(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return jsonResponse(t, http.StatusBadRequest, "400 Bad Request", map[string]any{}), nil
		}),
	}

	_, err := ExchangeAuthorizationCode(context.Background(), client, config.Credentials{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}, "auth-code")
	if err == nil {
		t.Fatalf("ExchangeAuthorizationCode returned nil error")
	}

	if got, want := err.Error(), "token exchange failed: 400 Bad Request"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestExchangeAuthorizationCodeRequiresAuthorizationCode(t *testing.T) {
	_, err := ExchangeAuthorizationCode(context.Background(), nil, config.Credentials{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}, "")
	if err == nil {
		t.Fatalf("ExchangeAuthorizationCode returned nil error")
	}

	if got, want := err.Error(), "authorization code is required"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestExchangeAuthorizationCodeValidatesAccessToken(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return jsonResponse(t, http.StatusOK, "200 OK", map[string]any{
				"refresh_token": "refresh-token",
				"expires_in":    3600,
				"scope":         "offline read:recovery",
				"token_type":    "Bearer",
			}), nil
		}),
	}

	_, err := ExchangeAuthorizationCode(context.Background(), client, config.Credentials{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}, "auth-code")
	if err == nil {
		t.Fatal("ExchangeAuthorizationCode returned nil error")
	}

	if got, want := err.Error(), "access token is required"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestExchangeAuthorizationCodeValidatesTokenExpiration(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return jsonResponse(t, http.StatusOK, "200 OK", map[string]any{
				"access_token":  "access-token",
				"refresh_token": "refresh-token",
				"expires_in":    0,
				"scope":         "offline read:recovery",
				"token_type":    "Bearer",
			}), nil
		}),
	}

	_, err := ExchangeAuthorizationCode(context.Background(), client, config.Credentials{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}, "auth-code")
	if err == nil {
		t.Fatal("ExchangeAuthorizationCode returned nil error")
	}

	if got, want := err.Error(), "expires in must be positive"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}
