package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Tianbo-Qiu/whoopctl/internal/auth"
	"github.com/Tianbo-Qiu/whoopctl/internal/config"
	"github.com/Tianbo-Qiu/whoopctl/internal/session"
	"github.com/Tianbo-Qiu/whoopctl/internal/whoop"
)

type App struct {
	ConfigDir                 string
	StateGenerator            func() (string, error)
	WaitForCallback           func(ctx context.Context, state string) (auth.AuthorizationCallback, error)
	ExchangeAuthorizationCode func(ctx context.Context, creds config.Credentials, code string) (auth.TokenResponse, error)
	TokenManager              tokenManager
	WhoopClient               whoopClient
}

type tokenManager interface {
	AccessToken(ctx context.Context) (string, error)
	Refresh(ctx context.Context) (string, error)
}

type whoopClient interface {
	BasicProfile(ctx context.Context, accessToken string) (whoop.BasicProfile, error)
	BodyMeasurement(ctx context.Context, accessToken string) (whoop.BodyMeasurement, error)
	Cycle(ctx context.Context, accessToken string, cycleID int64) (whoop.Cycle, error)
	Cycles(ctx context.Context, accessToken string, query whoop.CycleQuery) (whoop.CycleCollection, error)
	Recovery(ctx context.Context, accessToken string, query whoop.RecoveryQuery) (whoop.RecoveryCollection, error)
	RecoveryForCycle(ctx context.Context, accessToken string, cycleID int64) (whoop.Recovery, error)
	Sleep(ctx context.Context, accessToken string, sleepID string) (whoop.Sleep, error)
	SleepForCycle(ctx context.Context, accessToken string, cycleID int64) (whoop.Sleep, error)
	Sleeps(ctx context.Context, accessToken string, query whoop.SleepQuery) (whoop.SleepCollection, error)
	Workout(ctx context.Context, accessToken string, workoutID string) (whoop.Workout, error)
	Workouts(ctx context.Context, accessToken string, query whoop.WorkoutQuery) (whoop.WorkoutCollection, error)
}

func NewApp(configDir string) *App {
	return &App{
		ConfigDir:    configDir,
		TokenManager: &session.TokenManager{ConfigDir: configDir},
		WhoopClient:  &whoop.Client{},
	}
}

func Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	app := NewApp("")
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
	case "body":
		return app.runBody(ctx, args[1:], stdout)
	case "cycle":
		return app.runCycle(ctx, args[1:], stdout)
	case "profile":
		return app.runProfile(ctx, args[1:], stdout)
	case "recovery":
		return app.runRecovery(ctx, args[1:], stdout)
	case "sleep":
		return app.runSleep(ctx, args[1:], stdout)
	case "workout":
		return app.runWorkout(ctx, args[1:], stdout)
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func (app *App) runBody(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: whoopctl body")
	}

	token, err := app.tokenManager().AccessToken(ctx)
	if err != nil {
		return err
	}

	measurement, err := app.whoopClient().BodyMeasurement(ctx, token)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(measurement)
}

func (app *App) runProfile(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return app.runProfileBasic(ctx, stdout)
	}
	if len(args) == 1 && args[0] == "basic" {
		return app.runProfileBasic(ctx, stdout)
	}

	return fmt.Errorf("usage: whoopctl profile")
}

func (app *App) runProfileBasic(ctx context.Context, stdout io.Writer) error {
	token, err := app.tokenManager().AccessToken(ctx)
	if err != nil {
		return err
	}

	profile, err := app.whoopClient().BasicProfile(ctx, token)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(profile)
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
	if _, err := app.tokenManager().Refresh(ctx); err != nil {
		return err
	}

	fmt.Fprintln(stdout, "Token refreshed")
	return nil
}

func (app *App) runCycle(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) > 0 && args[0] == "get" {
		return app.runCycleByID(ctx, args[1:], stdout)
	}

	query, err := cycleQuery(args)
	if err != nil {
		return err
	}

	token, err := app.tokenManager().AccessToken(ctx)
	if err != nil {
		return err
	}

	cycles, err := app.whoopClient().Cycles(ctx, token, query)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cycles)
}

func (app *App) runCycleByID(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: whoopctl cycle get <cycle-id>")
	}

	cycleID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || cycleID <= 0 {
		return fmt.Errorf("cycle id must be a positive integer")
	}

	token, err := app.tokenManager().AccessToken(ctx)
	if err != nil {
		return err
	}

	cycle, err := app.whoopClient().Cycle(ctx, token, cycleID)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cycle)
}

func (app *App) runRecovery(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) > 0 && args[0] == "cycle" {
		return app.runRecoveryForCycle(ctx, args[1:], stdout)
	}

	query, err := recoveryQuery(args)
	if err != nil {
		return err
	}

	token, err := app.tokenManager().AccessToken(ctx)
	if err != nil {
		return err
	}

	recovery, err := app.whoopClient().Recovery(ctx, token, query)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(recovery)
}

func (app *App) runRecoveryForCycle(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: whoopctl recovery cycle <cycle-id>")
	}

	cycleID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || cycleID <= 0 {
		return fmt.Errorf("cycle id must be a positive integer")
	}

	token, err := app.tokenManager().AccessToken(ctx)
	if err != nil {
		return err
	}

	recovery, err := app.whoopClient().RecoveryForCycle(ctx, token, cycleID)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(recovery)
}

func (app *App) runSleep(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) > 0 && args[0] == "get" {
		return app.runSleepByID(ctx, args[1:], stdout)
	}
	if len(args) > 0 && args[0] == "cycle" {
		return app.runSleepForCycle(ctx, args[1:], stdout)
	}

	query, err := sleepQuery(args)
	if err != nil {
		return err
	}

	token, err := app.tokenManager().AccessToken(ctx)
	if err != nil {
		return err
	}

	sleeps, err := app.whoopClient().Sleeps(ctx, token, query)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(sleeps)
}

func (app *App) runSleepByID(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: whoopctl sleep get <sleep-id>")
	}

	token, err := app.tokenManager().AccessToken(ctx)
	if err != nil {
		return err
	}

	sleep, err := app.whoopClient().Sleep(ctx, token, args[0])
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(sleep)
}

func (app *App) runSleepForCycle(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: whoopctl sleep cycle <cycle-id>")
	}

	cycleID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || cycleID <= 0 {
		return fmt.Errorf("cycle id must be a positive integer")
	}

	token, err := app.tokenManager().AccessToken(ctx)
	if err != nil {
		return err
	}

	sleep, err := app.whoopClient().SleepForCycle(ctx, token, cycleID)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(sleep)
}

func (app *App) runWorkout(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) > 0 && args[0] == "get" {
		return app.runWorkoutByID(ctx, args[1:], stdout)
	}

	query, err := workoutQuery(args)
	if err != nil {
		return err
	}

	token, err := app.tokenManager().AccessToken(ctx)
	if err != nil {
		return err
	}

	workouts, err := app.whoopClient().Workouts(ctx, token, query)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(workouts)
}

func (app *App) runWorkoutByID(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: whoopctl workout get <workout-id>")
	}

	token, err := app.tokenManager().AccessToken(ctx)
	if err != nil {
		return err
	}

	workout, err := app.whoopClient().Workout(ctx, token, args[0])
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(workout)
}

func cycleQuery(args []string) (whoop.CycleQuery, error) {
	var query whoop.CycleQuery

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--limit":
			i++
			if i >= len(args) {
				return whoop.CycleQuery{}, fmt.Errorf("--limit requires a value")
			}
			limit, err := strconv.Atoi(args[i])
			if err != nil {
				return whoop.CycleQuery{}, fmt.Errorf("--limit must be an integer")
			}
			query.Limit = limit
		case "--start":
			i++
			if i >= len(args) {
				return whoop.CycleQuery{}, fmt.Errorf("--start requires a value")
			}
			start, err := time.Parse(time.RFC3339Nano, args[i])
			if err != nil {
				return whoop.CycleQuery{}, fmt.Errorf("--start must be RFC3339")
			}
			query.Start = start
		case "--end":
			i++
			if i >= len(args) {
				return whoop.CycleQuery{}, fmt.Errorf("--end requires a value")
			}
			end, err := time.Parse(time.RFC3339Nano, args[i])
			if err != nil {
				return whoop.CycleQuery{}, fmt.Errorf("--end must be RFC3339")
			}
			query.End = end
		case "--next-token":
			i++
			if i >= len(args) {
				return whoop.CycleQuery{}, fmt.Errorf("--next-token requires a value")
			}
			query.NextToken = args[i]
		default:
			return whoop.CycleQuery{}, fmt.Errorf("unknown option: %s", args[i])
		}
	}

	return query, nil
}

func sleepQuery(args []string) (whoop.SleepQuery, error) {
	var query whoop.SleepQuery

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--limit":
			i++
			if i >= len(args) {
				return whoop.SleepQuery{}, fmt.Errorf("--limit requires a value")
			}
			limit, err := strconv.Atoi(args[i])
			if err != nil {
				return whoop.SleepQuery{}, fmt.Errorf("--limit must be an integer")
			}
			query.Limit = limit
		case "--start":
			i++
			if i >= len(args) {
				return whoop.SleepQuery{}, fmt.Errorf("--start requires a value")
			}
			start, err := time.Parse(time.RFC3339Nano, args[i])
			if err != nil {
				return whoop.SleepQuery{}, fmt.Errorf("--start must be RFC3339")
			}
			query.Start = start
		case "--end":
			i++
			if i >= len(args) {
				return whoop.SleepQuery{}, fmt.Errorf("--end requires a value")
			}
			end, err := time.Parse(time.RFC3339Nano, args[i])
			if err != nil {
				return whoop.SleepQuery{}, fmt.Errorf("--end must be RFC3339")
			}
			query.End = end
		case "--next-token":
			i++
			if i >= len(args) {
				return whoop.SleepQuery{}, fmt.Errorf("--next-token requires a value")
			}
			query.NextToken = args[i]
		default:
			return whoop.SleepQuery{}, fmt.Errorf("unknown option: %s", args[i])
		}
	}

	return query, nil
}

func workoutQuery(args []string) (whoop.WorkoutQuery, error) {
	var query whoop.WorkoutQuery

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--limit":
			i++
			if i >= len(args) {
				return whoop.WorkoutQuery{}, fmt.Errorf("--limit requires a value")
			}
			limit, err := strconv.Atoi(args[i])
			if err != nil {
				return whoop.WorkoutQuery{}, fmt.Errorf("--limit must be an integer")
			}
			query.Limit = limit
		case "--start":
			i++
			if i >= len(args) {
				return whoop.WorkoutQuery{}, fmt.Errorf("--start requires a value")
			}
			start, err := time.Parse(time.RFC3339Nano, args[i])
			if err != nil {
				return whoop.WorkoutQuery{}, fmt.Errorf("--start must be RFC3339")
			}
			query.Start = start
		case "--end":
			i++
			if i >= len(args) {
				return whoop.WorkoutQuery{}, fmt.Errorf("--end requires a value")
			}
			end, err := time.Parse(time.RFC3339Nano, args[i])
			if err != nil {
				return whoop.WorkoutQuery{}, fmt.Errorf("--end must be RFC3339")
			}
			query.End = end
		case "--next-token":
			i++
			if i >= len(args) {
				return whoop.WorkoutQuery{}, fmt.Errorf("--next-token requires a value")
			}
			query.NextToken = args[i]
		default:
			return whoop.WorkoutQuery{}, fmt.Errorf("unknown option: %s", args[i])
		}
	}

	return query, nil
}

func recoveryQuery(args []string) (whoop.RecoveryQuery, error) {
	var query whoop.RecoveryQuery

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--limit":
			i++
			if i >= len(args) {
				return whoop.RecoveryQuery{}, fmt.Errorf("--limit requires a value")
			}
			limit, err := strconv.Atoi(args[i])
			if err != nil {
				return whoop.RecoveryQuery{}, fmt.Errorf("--limit must be an integer")
			}
			query.Limit = limit
		case "--start":
			i++
			if i >= len(args) {
				return whoop.RecoveryQuery{}, fmt.Errorf("--start requires a value")
			}
			start, err := time.Parse(time.RFC3339Nano, args[i])
			if err != nil {
				return whoop.RecoveryQuery{}, fmt.Errorf("--start must be RFC3339")
			}
			query.Start = start
		case "--end":
			i++
			if i >= len(args) {
				return whoop.RecoveryQuery{}, fmt.Errorf("--end requires a value")
			}
			end, err := time.Parse(time.RFC3339Nano, args[i])
			if err != nil {
				return whoop.RecoveryQuery{}, fmt.Errorf("--end must be RFC3339")
			}
			query.End = end
		case "--next-token":
			i++
			if i >= len(args) {
				return whoop.RecoveryQuery{}, fmt.Errorf("--next-token requires a value")
			}
			query.NextToken = args[i]
		default:
			return whoop.RecoveryQuery{}, fmt.Errorf("unknown option: %s", args[i])
		}
	}

	return query, nil
}

func (app *App) tokenManager() tokenManager {
	if app.TokenManager != nil {
		return app.TokenManager
	}
	return &session.TokenManager{ConfigDir: app.ConfigDir}
}

func (app *App) whoopClient() whoopClient {
	if app.WhoopClient != nil {
		return app.WhoopClient
	}
	return &whoop.Client{}
}
