package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Tianbo-Qiu/whoopctl/internal/auth"
	"github.com/Tianbo-Qiu/whoopctl/internal/config"
	"github.com/Tianbo-Qiu/whoopctl/internal/session"
)

type App struct {
	ConfigDir                 string
	StateGenerator            func() (string, error)
	WaitForCallback           func(ctx context.Context, state string) (auth.AuthorizationCallback, error)
	ExchangeAuthorizationCode func(ctx context.Context, creds config.Credentials, code string) (auth.TokenResponse, error)
	RefreshSession            func(ctx context.Context) error
}

func Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	app := &App{}
	return app.Run(ctx, args, stdout, stderr)
}

func (app *App) Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	_ = stderr

	if len(args) == 0 {
		return fmt.Errorf("usage: whoopctl <command>")
	}

	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, "whoopctl dev")
		return nil
	case "auth":
		return app.runAuth(ctx, args[1:], stdout)
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func (app *App) runAuth(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: whoopctl auth <command>")
	}

	switch args[0] {
	case "setup":
		return app.runAuthSetup(args[1:], stdout)
	case "login":
		return app.runAuthLogin(ctx, stdout)
	case "status":
		return app.runAuthStatus(stdout)
	case "refresh":
		return app.runAuthRefresh(ctx, stdout)
	default:
		return fmt.Errorf("unknown auth command: %s", args[0])
	}
}

func (app *App) runAuthSetup(args []string, stdout io.Writer) error {
	var clientID, clientSecret string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--client-id":
			i++
			if i >= len(args) {
				return fmt.Errorf("--client-id requires a value")
			}
			clientID = args[i]
		case "--client-secret":
			i++
			if i >= len(args) {
				return fmt.Errorf("--client-secret requires a value")
			}
			clientSecret = args[i]
		default:
			return fmt.Errorf("unknown option: %s", args[i])
		}
	}

	if err := config.SaveCredentials(app.ConfigDir, config.Credentials{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}); err != nil {
		return err
	}

	fmt.Fprintln(stdout, "WHOOP credentials saved")
	return nil
}

func (app *App) runAuthLogin(ctx context.Context, stdout io.Writer) error {
	creds, err := config.LoadCredentials(app.ConfigDir)
	if err != nil {
		return err
	}

	stateGenerator := app.StateGenerator
	if stateGenerator == nil {
		stateGenerator = auth.StateGenerator
	}

	state, err := stateGenerator()
	if err != nil {
		return err
	}

	authURL, err := auth.AuthCodeURL(creds.ClientID, state, auth.DefaultScopes)
	if err != nil {
		return err
	}

	fmt.Fprintln(stdout, "Open to authorize whoopctl:")
	fmt.Fprintln(stdout, authURL)

	loginCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	waitForCallback := app.WaitForCallback
	if waitForCallback == nil {
		waitForCallback = auth.WaitForAuthorizationCallback
	}

	resp, err := waitForCallback(loginCtx, state)
	if err != nil {
		return err
	}

	exchangeAuthorizationCode := app.ExchangeAuthorizationCode
	if exchangeAuthorizationCode == nil {
		exchangeAuthorizationCode = func(ctx context.Context, creds config.Credentials, code string) (auth.TokenResponse, error) {
			return auth.ExchangeAuthorizationCode(ctx, http.DefaultClient, creds, code)
		}
	}

	tokenResponse, err := exchangeAuthorizationCode(loginCtx, creds, resp.Code)
	if err != nil {
		return err
	}

	storedToken := config.StoredToken{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
		Scope:        tokenResponse.Scope,
		TokenType:    tokenResponse.TokenType,
	}

	if err := config.SaveToken(app.ConfigDir, storedToken); err != nil {
		return err
	}

	fmt.Fprintln(stdout, "Authorization complete.")
	return nil
}

func (app *App) runAuthStatus(stdout io.Writer) error {
	if _, err := config.LoadCredentials(app.ConfigDir); err != nil {
		fmt.Fprintln(stdout, "Credentials: missing")
	} else {
		fmt.Fprintln(stdout, "Credentials: configured")
	}

	token, err := config.LoadToken(app.ConfigDir)
	if err != nil {
		fmt.Fprintln(stdout, "Token: missing")
		return nil
	}

	expiresAt := token.ExpiresAt.UTC().Format(time.RFC3339)
	if time.Now().Before(token.ExpiresAt) {
		fmt.Fprintf(stdout, "Token: valid until %s\n", expiresAt)
		return nil
	}

	fmt.Fprintf(stdout, "Token: expired at %s\n", expiresAt)
	return nil
}

func (app *App) runAuthRefresh(ctx context.Context, stdout io.Writer) error {
	refreshSession := app.RefreshSession
	if refreshSession == nil {
		refreshSession = func(ctx context.Context) error {
			manager := &session.TokenManager{ConfigDir: app.ConfigDir}
			_, err := manager.Refresh(ctx)
			return err
		}
	}

	if err := refreshSession(ctx); err != nil {
		return err
	}

	fmt.Fprintln(stdout, "Token refreshed")
	return nil
}
