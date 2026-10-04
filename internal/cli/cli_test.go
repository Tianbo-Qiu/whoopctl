package cli

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Tianbo-Qiu/whoopctl/internal/auth"
	"github.com/Tianbo-Qiu/whoopctl/internal/config"
	"github.com/Tianbo-Qiu/whoopctl/internal/whoop"
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

func TestRunRecoveryPrintsJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	createdAt := time.Date(2022, 4, 24, 11, 25, 44, 774000000, time.UTC)
	updatedAt := time.Date(2022, 4, 24, 14, 25, 44, 774000000, time.UTC)
	spo2Percentage := 95.6875
	skinTempCelsius := 33.7

	app := NewApp(t.TempDir())
	app.TokenManager = fakeTokenManager{accessToken: "access-token"}
	app.WhoopClient = fakeWhoopClient{
		t:               t,
		wantAccessToken: "access-token",
		recovery: whoop.RecoveryCollection{
			Records: []whoop.Recovery{
				{
					CycleID:    93845,
					SleepID:    "123e4567-e89b-12d3-a456-426614174000",
					UserID:     10129,
					CreatedAt:  createdAt,
					UpdatedAt:  updatedAt,
					ScoreState: "SCORED",
					Score: &whoop.RecoveryScore{
						UserCalibrating:  false,
						RecoveryScore:    44,
						RestingHeartRate: 64,
						HrvRmssdMilli:    31.813562,
						Spo2Percentage:   &spo2Percentage,
						SkinTempCelsius:  &skinTempCelsius,
					},
				},
			},
			NextToken: "next-token",
		},
	}

	err := app.Run(context.Background(), []string{"recovery"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	want := `{
  "records": [
    {
      "cycle_id": 93845,
      "sleep_id": "123e4567-e89b-12d3-a456-426614174000",
      "user_id": 10129,
      "created_at": "2022-04-24T11:25:44.774Z",
      "updated_at": "2022-04-24T14:25:44.774Z",
      "score_state": "SCORED",
      "score": {
        "user_calibrating": false,
        "recovery_score": 44,
        "resting_heart_rate": 64,
        "hrv_rmssd_milli": 31.813562,
        "spo2_percentage": 95.6875,
        "skin_temp_celsius": 33.7
      }
    }
  ],
  "next_token": "next-token"
}
`
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestRunRecoveryReturnsAccessTokenError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	app := NewApp(t.TempDir())
	app.TokenManager = fakeTokenManager{accessTokenErr: fmt.Errorf("access token failed")}

	err := app.Run(context.Background(), []string{"recovery"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("Run returned nil error")
	}
	if got, want := err.Error(), "access token failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
}

func TestRunRecoveryReturnsRecoveryError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	app := NewApp(t.TempDir())
	app.TokenManager = fakeTokenManager{accessToken: "access-token"}
	app.WhoopClient = fakeWhoopClient{recoveryErr: fmt.Errorf("recovery failed")}

	err := app.Run(context.Background(), []string{"recovery"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("Run returned nil error")
	}
	if got, want := err.Error(), "recovery failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
}

func TestRunAuthSetupSaveCredentials(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := NewApp(t.TempDir())

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
	app := NewApp(t.TempDir())

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
	app := NewApp(t.TempDir())

	err := app.Run(context.Background(), []string{"auth", "nope"}, &stdout, &stderr)
	if err == nil {
		t.Fatalf("Run returned nil error")
	}
	if got, want := err.Error(), "unknown auth command: nope"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestRunAuthStatusShowsMissingCredentialsAndToken(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := NewApp(t.TempDir())

	err := app.Run(context.Background(), []string{"auth", "status"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if got, want := stdout.String(), "Credentials: missing\nToken: missing\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRunAuthStatusShowsValidToken(t *testing.T) {
	var stdout, stderr bytes.Buffer
	dir := t.TempDir()
	expiresAt := time.Now().Add(time.Hour).UTC().Truncate(time.Second)

	err := config.SaveCredentials(dir, config.Credentials{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	})
	if err != nil {
		t.Fatalf("SaveCredentials returned error: %v", err)
	}

	err = config.SaveToken(dir, config.StoredToken{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresAt:    expiresAt,
		Scope:        "offline read:recovery",
		TokenType:    "bearer",
	})
	if err != nil {
		t.Fatalf("SaveToken returned error: %v", err)
	}

	app := NewApp(dir)

	err = app.Run(context.Background(), []string{"auth", "status"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	want := fmt.Sprintf("Credentials: configured\nToken: valid until %s\n", expiresAt.Format(time.RFC3339))
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRunAuthStatusShowsExpiredToken(t *testing.T) {
	var stdout, stderr bytes.Buffer
	dir := t.TempDir()
	expiresAt := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)

	err := config.SaveCredentials(dir, config.Credentials{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	})
	if err != nil {
		t.Fatalf("SaveCredentials returned error: %v", err)
	}

	err = config.SaveToken(dir, config.StoredToken{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresAt:    expiresAt,
		Scope:        "offline read:recovery",
		TokenType:    "bearer",
	})
	if err != nil {
		t.Fatalf("SaveToken returned error: %v", err)
	}

	app := NewApp(dir)

	err = app.Run(context.Background(), []string{"auth", "status"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	want := fmt.Sprintf("Credentials: configured\nToken: expired at %s\n", expiresAt.Format(time.RFC3339))
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRunAuthRefresh(t *testing.T) {
	var stdout, stderr bytes.Buffer
	called := false

	app := NewApp(t.TempDir())
	app.TokenManager = fakeTokenManager{
		refreshToken: "access-token",
		onRefresh: func() {
			called = true
		},
	}

	err := app.Run(context.Background(), []string{"auth", "refresh"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if !called {
		t.Fatal("RefreshSession was not called")
	}
	if got, want := stdout.String(), "Token refreshed\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestRunAuthRefreshReturnsError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	app := NewApp(t.TempDir())
	app.TokenManager = fakeTokenManager{refreshErr: fmt.Errorf("refresh failed")}

	err := app.Run(context.Background(), []string{"auth", "refresh"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("Run returned nil error")
	}
	if got, want := err.Error(), "refresh failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
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

	app := NewApp(dir)
	app.StateGenerator = func() (string, error) {
		return "state-value", nil
	}
	app.WaitForCallback = func(ctx context.Context, state string) (auth.AuthorizationCallback, error) {
		return auth.AuthorizationCallback{
			Code:  "auth-code",
			State: state,
		}, nil
	}
	app.ExchangeAuthorizationCode = func(ctx context.Context, creds config.Credentials, code string) (auth.TokenResponse, error) {
		if got, want := creds.ClientID, "client-id"; got != want {
			t.Fatalf("ClientID = %q, want %q", got, want)
		}
		if got, want := creds.ClientSecret, "client-secret"; got != want {
			t.Fatalf("ClientSecret = %q, want %q", got, want)
		}
		if got, want := code, "auth-code"; got != want {
			t.Fatalf("code = %q, want %q", got, want)
		}
		return auth.TokenResponse{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresIn:    3600,
			Scope:        "offline read:recovery",
			TokenType:    "bearer",
		}, nil
	}

	beforeExpiresAt := time.Now().Add(time.Hour)
	err = app.Run(context.Background(), []string{"auth", "login"}, &stdout, &stderr)
	afterExpiresAt := time.Now().Add(time.Hour)
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

	if !strings.Contains(out, "Authorization complete.\n") {
		t.Fatalf("stdout = %q, missing callback confirmation", out)
	}

	token, err := config.LoadToken(dir)
	if err != nil {
		t.Fatalf("LoadToken returned error: %v", err)
	}
	if got, want := token.AccessToken, "access-token"; got != want {
		t.Fatalf("AccessToken = %q, want %q", got, want)
	}
	if got, want := token.RefreshToken, "refresh-token"; got != want {
		t.Fatalf("RefreshToken = %q, want %q", got, want)
	}
	if token.ExpiresAt.Before(beforeExpiresAt) || token.ExpiresAt.After(afterExpiresAt) {
		t.Fatalf("ExpiresAt = %v, want between %v and %v", token.ExpiresAt, beforeExpiresAt, afterExpiresAt)
	}
	if got, want := token.Scope, "offline read:recovery"; got != want {
		t.Fatalf("Scope = %q, want %q", got, want)
	}
	if got, want := token.TokenType, "bearer"; got != want {
		t.Fatalf("TokenType = %q, want %q", got, want)
	}
}

func TestRunAuthLoginReturnsCallbackContextError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	dir := t.TempDir()

	err := config.SaveCredentials(dir, config.Credentials{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	})
	if err != nil {
		t.Fatalf("SaveCredentials returned error: %v", err)
	}

	app := NewApp(dir)
	app.StateGenerator = func() (string, error) {
		return "state-value", nil
	}
	app.WaitForCallback = func(ctx context.Context, state string) (auth.AuthorizationCallback, error) {
		<-ctx.Done()
		return auth.AuthorizationCallback{}, ctx.Err()
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = app.Run(ctx, []string{"auth", "login"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("Run returned nil error")
	}
	if got, want := err, context.Canceled; got != want {
		t.Fatalf("error = %v, want %v", got, want)
	}
}

type fakeTokenManager struct {
	accessToken    string
	accessTokenErr error
	refreshToken   string
	refreshErr     error
	onRefresh      func()
}

func (f fakeTokenManager) AccessToken(ctx context.Context) (string, error) {
	if f.accessTokenErr != nil {
		return "", f.accessTokenErr
	}
	return f.accessToken, nil
}

func (f fakeTokenManager) Refresh(ctx context.Context) (string, error) {
	if f.onRefresh != nil {
		f.onRefresh()
	}
	if f.refreshErr != nil {
		return "", f.refreshErr
	}
	return f.refreshToken, nil
}

type fakeWhoopClient struct {
	t               *testing.T
	wantAccessToken string
	recovery        whoop.RecoveryCollection
	recoveryErr     error
}

func (f fakeWhoopClient) Recovery(ctx context.Context, accessToken string) (whoop.RecoveryCollection, error) {
	if f.t != nil {
		f.t.Helper()
	}
	if f.wantAccessToken != "" && accessToken != f.wantAccessToken {
		f.t.Fatalf("accessToken = %q, want %q", accessToken, f.wantAccessToken)
	}
	if f.recoveryErr != nil {
		return whoop.RecoveryCollection{}, f.recoveryErr
	}
	return f.recovery, nil
}
