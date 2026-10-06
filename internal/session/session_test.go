package session

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/Tianbo-Qiu/whoopctl/internal/auth"
	"github.com/Tianbo-Qiu/whoopctl/internal/config"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestTokenManagerReturnsStoredAccessTokenWhenValid(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)

	saveCredentialsAndToken(t, dir, config.StoredToken{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresAt:    now.Add(time.Hour),
		Scope:        "offline read:recovery",
		TokenType:    "bearer",
	})

	manager := &TokenManager{
		ConfigDir: dir,
		Now: func() time.Time {
			return now
		},
		Client: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				t.Fatal("RefreshAccessToken should not be called")
				return nil, nil
			}),
		},
	}

	accessToken, err := manager.AccessToken(context.Background())
	if err != nil {
		t.Fatalf("AccessToken returned error: %v", err)
	}

	if got, want := accessToken, "access-token"; got != want {
		t.Fatalf("accessToken = %q, want %q", got, want)
	}
}

func TestTokenManagerRefreshesExpiredToken(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)

	saveCredentialsAndToken(t, dir, config.StoredToken{
		AccessToken:  "old-access-token",
		RefreshToken: "old-refresh-token",
		ExpiresAt:    now.Add(-time.Minute),
		Scope:        "offline read:recovery",
		TokenType:    "bearer",
	})

	manager := &TokenManager{
		ConfigDir: dir,
		Now: func() time.Time {
			return now
		},
		Client: refreshClient(t, "old-refresh-token", http.StatusOK, "200 OK", map[string]any{
			"access_token":  "new-access-token",
			"refresh_token": "new-refresh-token",
			"expires_in":    3600,
			"scope":         "offline read:recovery",
			"token_type":    "Bearer",
		}),
	}

	accessToken, err := manager.AccessToken(context.Background())
	if err != nil {
		t.Fatalf("AccessToken returned error: %v", err)
	}

	if got, want := accessToken, "new-access-token"; got != want {
		t.Fatalf("accessToken = %q, want %q", got, want)
	}

	token, err := config.LoadToken(dir)
	if err != nil {
		t.Fatalf("LoadToken returned error: %v", err)
	}

	if got, want := token.AccessToken, "new-access-token"; got != want {
		t.Fatalf("AccessToken = %q, want %q", got, want)
	}
	if got, want := token.RefreshToken, "new-refresh-token"; got != want {
		t.Fatalf("RefreshToken = %q, want %q", got, want)
	}
	if got, want := token.ExpiresAt, now.Add(time.Hour); !got.Equal(want) {
		t.Fatalf("ExpiresAt = %v, want %v", got, want)
	}
}

func TestTokenManagerRefreshesTokenWithinSkew(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)

	saveCredentialsAndToken(t, dir, config.StoredToken{
		AccessToken:  "old-access-token",
		RefreshToken: "old-refresh-token",
		ExpiresAt:    now.Add(RefreshSkew / 2),
		Scope:        "offline read:recovery",
		TokenType:    "bearer",
	})

	manager := &TokenManager{
		ConfigDir: dir,
		Now: func() time.Time {
			return now
		},
		Client: refreshClient(t, "old-refresh-token", http.StatusOK, "200 OK", map[string]any{
			"access_token":  "new-access-token",
			"refresh_token": "new-refresh-token",
			"expires_in":    3600,
			"scope":         "offline read:recovery",
			"token_type":    "Bearer",
		}),
	}

	accessToken, err := manager.AccessToken(context.Background())
	if err != nil {
		t.Fatalf("AccessToken returned error: %v", err)
	}

	if got, want := accessToken, "new-access-token"; got != want {
		t.Fatalf("accessToken = %q, want %q", got, want)
	}
}

func TestTokenManagerRefreshForcesRefreshWhenTokenStillValid(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)

	saveCredentialsAndToken(t, dir, config.StoredToken{
		AccessToken:  "old-access-token",
		RefreshToken: "old-refresh-token",
		ExpiresAt:    now.Add(time.Hour),
		Scope:        "offline read:recovery",
		TokenType:    "bearer",
	})

	manager := &TokenManager{
		ConfigDir: dir,
		Now: func() time.Time {
			return now
		},
		Client: refreshClient(t, "old-refresh-token", http.StatusOK, "200 OK", map[string]any{
			"access_token":  "new-access-token",
			"refresh_token": "new-refresh-token",
			"expires_in":    3600,
			"scope":         "offline read:recovery",
			"token_type":    "Bearer",
		}),
	}

	accessToken, err := manager.Refresh(context.Background())
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}

	if got, want := accessToken, "new-access-token"; got != want {
		t.Fatalf("accessToken = %q, want %q", got, want)
	}

	token, err := config.LoadToken(dir)
	if err != nil {
		t.Fatalf("LoadToken returned error: %v", err)
	}
	if got, want := token.RefreshToken, "new-refresh-token"; got != want {
		t.Fatalf("RefreshToken = %q, want %q", got, want)
	}
}

func TestTokenManagerReturnsRefreshError(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)

	saveCredentialsAndToken(t, dir, config.StoredToken{
		AccessToken:  "old-access-token",
		RefreshToken: "old-refresh-token",
		ExpiresAt:    now.Add(-time.Minute),
		Scope:        "offline read:recovery",
		TokenType:    "bearer",
	})

	manager := &TokenManager{
		ConfigDir: dir,
		Now: func() time.Time {
			return now
		},
		Client: refreshClient(t, "old-refresh-token", http.StatusUnauthorized, "401 Unauthorized", map[string]any{}),
	}

	_, err := manager.AccessToken(context.Background())
	if err == nil {
		t.Fatal("AccessToken returned nil error")
	}

	if !errors.Is(err, ErrLoginRequired) {
		t.Fatalf("error = %v, want ErrLoginRequired", err)
	}

	if got, want := err.Error(), "login required; token refresh failed: 401 Unauthorized"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func saveCredentialsAndToken(t *testing.T, dir string, token config.StoredToken) {
	t.Helper()

	if err := config.SaveCredentials(dir, config.Credentials{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}); err != nil {
		t.Fatalf("SaveCredentials returned error: %v", err)
	}

	if err := config.SaveToken(dir, token); err != nil {
		t.Fatalf("SaveToken returned error: %v", err)
	}
}

func refreshClient(t *testing.T, wantRefreshToken string, statusCode int, status string, body map[string]any) *http.Client {
	t.Helper()

	return &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if got, want := req.Method, http.MethodPost; got != want {
				t.Fatalf("method = %q, want %q", got, want)
			}

			if got, want := req.URL.String(), auth.TokenEndpoint; got != want {
				t.Fatalf("url = %q, want %q", got, want)
			}

			data, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("ReadAll returned error: %v", err)
			}

			form, err := url.ParseQuery(string(data))
			if err != nil {
				t.Fatalf("ParseQuery returned error: %v", err)
			}

			assertFormValue(t, form, "grant_type", "refresh_token")
			assertFormValue(t, form, "refresh_token", wantRefreshToken)
			assertFormValue(t, form, "client_id", "client-id")
			assertFormValue(t, form, "client_secret", "client-secret")

			return jsonResponse(t, statusCode, status, body), nil
		}),
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

func TestTokenManagerClearTokenRemovesStoredToken(t *testing.T) {
	dir := t.TempDir()

	saveCredentialsAndToken(t, dir, config.StoredToken{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresAt:    time.Date(2026, 9, 27, 13, 0, 0, 0, time.UTC),
		Scope:        "offline read:recovery",
		TokenType:    "bearer",
	})

	manager := &TokenManager{ConfigDir: dir}
	if err := manager.ClearToken(context.Background()); err != nil {
		t.Fatalf("ClearToken returned error: %v", err)
	}

	_, err := manager.AccessToken(context.Background())
	if !errors.Is(err, ErrLoginRequired) {
		t.Fatalf("AccessToken error = %v, want ErrLoginRequired", err)
	}
}
