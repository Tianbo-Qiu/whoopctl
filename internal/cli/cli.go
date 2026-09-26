package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Tianbo-Qiu/whoopctl/internal/auth"
	"github.com/Tianbo-Qiu/whoopctl/internal/config"
)

type App struct {
	ConfigDir       string
	StateGenerator  func() (string, error)
	WaitForCallback func(ctx context.Context, state string) (auth.AuthorizationCallback, error)
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

	if _, err := waitForCallback(loginCtx, state); err != nil {
		return err
	}

	fmt.Fprintln(stdout, "Authorization code received.")
	return nil
}
