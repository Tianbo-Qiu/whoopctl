package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/Tianbo-Qiu/whoopctl/internal/config"
)

type App struct {
	ConfigDir string
}

func Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	app := &App{}
	return app.Run(ctx, args, stdout, stderr)
}

func (app *App) Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	_ = ctx
	_ = stderr

	if len(args) == 0 {
		return fmt.Errorf("usage: whoopctl <command>")
	}

	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, "whoopctl dev")
		return nil
	case "auth":
		return app.runAuth(args[1:], stdout)
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func (app *App) runAuth(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: whoopctl auth <command>")
	}

	switch args[0] {
	case "setup":
		return app.runAuthSetup(args[1:], stdout)
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
