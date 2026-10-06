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

func TestRunCyclePrintsJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	createdAt := time.Date(2022, 4, 24, 11, 25, 44, 774000000, time.UTC)
	updatedAt := time.Date(2022, 4, 24, 14, 25, 44, 774000000, time.UTC)
	start := time.Date(2022, 4, 24, 2, 25, 44, 774000000, time.UTC)
	end := time.Date(2022, 4, 24, 10, 25, 44, 774000000, time.UTC)
	stepCount := 8234

	app := NewApp(t.TempDir())
	app.TokenManager = fakeTokenManager{accessToken: "access-token"}
	app.WhoopClient = fakeWhoopClient{
		t:               t,
		wantAccessToken: "access-token",
		wantCycleQuery:  whoop.CycleQuery{},
		cycles: whoop.CycleCollection{
			Records: []whoop.Cycle{
				{
					ID:             93845,
					UserID:         10129,
					CreatedAt:      createdAt,
					UpdatedAt:      updatedAt,
					Start:          start,
					End:            &end,
					TimezoneOffset: "-05:00",
					ScoreState:     "SCORED",
					Score: &whoop.CycleScore{
						Strain:           5.2951527,
						Kilojoule:        8288.297,
						AverageHeartRate: 68,
						MaxHeartRate:     141,
					},
					StepCount: &stepCount,
				},
			},
			NextToken: "next-token",
		},
	}

	err := app.Run(context.Background(), []string{"cycle"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	want := `{
  "records": [
    {
      "id": 93845,
      "user_id": 10129,
      "created_at": "2022-04-24T11:25:44.774Z",
      "updated_at": "2022-04-24T14:25:44.774Z",
      "start": "2022-04-24T02:25:44.774Z",
      "end": "2022-04-24T10:25:44.774Z",
      "timezone_offset": "-05:00",
      "score_state": "SCORED",
      "score": {
        "strain": 5.2951527,
        "kilojoule": 8288.297,
        "average_heart_rate": 68,
        "max_heart_rate": 141
      },
      "step_count": 8234
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

func TestRunCyclePassesQuery(t *testing.T) {
	var stdout, stderr bytes.Buffer
	start := time.Date(2026, 1, 2, 3, 4, 5, 123456789, time.UTC)
	end := time.Date(2026, 1, 3, 3, 4, 5, 987654321, time.UTC)

	app := NewApp(t.TempDir())
	app.TokenManager = fakeTokenManager{accessToken: "access-token"}
	app.WhoopClient = fakeWhoopClient{
		t:               t,
		wantAccessToken: "access-token",
		wantCycleQuery: whoop.CycleQuery{
			Limit:     25,
			Start:     start,
			End:       end,
			NextToken: "next-token",
		},
		cycles: whoop.CycleCollection{Records: []whoop.Cycle{}},
	}

	err := app.Run(context.Background(), []string{
		"cycle",
		"--limit", "25",
		"--start", start.Format(time.RFC3339Nano),
		"--end", end.Format(time.RFC3339Nano),
		"--next-token", "next-token",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
}

func TestRunCycleByIDPrintsJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	createdAt := time.Date(2022, 4, 24, 11, 25, 44, 774000000, time.UTC)
	updatedAt := time.Date(2022, 4, 24, 14, 25, 44, 774000000, time.UTC)
	start := time.Date(2022, 4, 24, 2, 25, 44, 774000000, time.UTC)

	app := NewApp(t.TempDir())
	app.TokenManager = fakeTokenManager{accessToken: "access-token"}
	app.WhoopClient = fakeWhoopClient{
		t:               t,
		wantAccessToken: "access-token",
		wantCycleID:     93845,
		cycle: whoop.Cycle{
			ID:             93845,
			UserID:         10129,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			Start:          start,
			TimezoneOffset: "-05:00",
			ScoreState:     "PENDING_SCORE",
		},
	}

	err := app.Run(context.Background(), []string{"cycle", "get", "93845"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	want := `{
  "id": 93845,
  "user_id": 10129,
  "created_at": "2022-04-24T11:25:44.774Z",
  "updated_at": "2022-04-24T14:25:44.774Z",
  "start": "2022-04-24T02:25:44.774Z",
  "timezone_offset": "-05:00",
  "score_state": "PENDING_SCORE"
}
`
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestRunCycleByIDRejectsInvalidArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing cycle id",
			args:    []string{"cycle", "get"},
			wantErr: "usage: whoopctl cycle get <cycle-id>",
		},
		{
			name:    "extra arg",
			args:    []string{"cycle", "get", "93845", "extra"},
			wantErr: "usage: whoopctl cycle get <cycle-id>",
		},
		{
			name:    "not integer",
			args:    []string{"cycle", "get", "nope"},
			wantErr: "cycle id must be a positive integer",
		},
		{
			name:    "zero",
			args:    []string{"cycle", "get", "0"},
			wantErr: "cycle id must be a positive integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			accessTokenCalled := false

			app := NewApp(t.TempDir())
			app.TokenManager = fakeTokenManager{
				onAccessToken: func() {
					accessTokenCalled = true
				},
			}

			err := app.Run(context.Background(), tt.args, &stdout, &stderr)
			if err == nil {
				t.Fatal("Run returned nil error")
			}
			if got := err.Error(); got != tt.wantErr {
				t.Fatalf("error = %q, want %q", got, tt.wantErr)
			}
			if accessTokenCalled {
				t.Fatal("AccessToken was called")
			}
			if got := stdout.String(); got != "" {
				t.Fatalf("stdout = %q, want empty", got)
			}
		})
	}
}

func TestRunCycleByIDReturnsCycleError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	app := NewApp(t.TempDir())
	app.TokenManager = fakeTokenManager{accessToken: "access-token"}
	app.WhoopClient = fakeWhoopClient{cycleErr: fmt.Errorf("cycle failed")}

	err := app.Run(context.Background(), []string{"cycle", "get", "93845"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("Run returned nil error")
	}
	if got, want := err.Error(), "cycle failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
}

func TestRunCycleRejectsInvalidLimit(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing limit",
			args:    []string{"cycle", "--limit"},
			wantErr: "--limit requires a value",
		},
		{
			name:    "invalid limit",
			args:    []string{"cycle", "--limit", "nope"},
			wantErr: "--limit must be an integer",
		},
		{
			name:    "missing start",
			args:    []string{"cycle", "--start"},
			wantErr: "--start requires a value",
		},
		{
			name:    "invalid start",
			args:    []string{"cycle", "--start", "2026-01-02"},
			wantErr: "--start must be RFC3339",
		},
		{
			name:    "missing end",
			args:    []string{"cycle", "--end"},
			wantErr: "--end requires a value",
		},
		{
			name:    "invalid end",
			args:    []string{"cycle", "--end", "2026-01-03"},
			wantErr: "--end must be RFC3339",
		},
		{
			name:    "missing next token",
			args:    []string{"cycle", "--next-token"},
			wantErr: "--next-token requires a value",
		},
		{
			name:    "unknown option",
			args:    []string{"cycle", "--wat"},
			wantErr: "unknown option: --wat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			accessTokenCalled := false

			app := NewApp(t.TempDir())
			app.TokenManager = fakeTokenManager{
				onAccessToken: func() {
					accessTokenCalled = true
				},
			}

			err := app.Run(context.Background(), tt.args, &stdout, &stderr)
			if err == nil {
				t.Fatal("Run returned nil error")
			}
			if got := err.Error(); got != tt.wantErr {
				t.Fatalf("error = %q, want %q", got, tt.wantErr)
			}
			if accessTokenCalled {
				t.Fatal("AccessToken was called")
			}
			if got := stdout.String(); got != "" {
				t.Fatalf("stdout = %q, want empty", got)
			}
		})
	}
}

func TestRunCycleReturnsCycleError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	app := NewApp(t.TempDir())
	app.TokenManager = fakeTokenManager{accessToken: "access-token"}
	app.WhoopClient = fakeWhoopClient{cyclesErr: fmt.Errorf("cycles failed")}

	err := app.Run(context.Background(), []string{"cycle"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("Run returned nil error")
	}
	if got, want := err.Error(), "cycles failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
}

func TestRunSleepPrintsJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer

	app := NewApp(t.TempDir())
	app.TokenManager = fakeTokenManager{accessToken: "access-token"}
	app.WhoopClient = fakeWhoopClient{
		t:               t,
		wantAccessToken: "access-token",
		sleeps: whoop.SleepCollection{
			Records:   []whoop.Sleep{testSleep()},
			NextToken: "next-token",
		},
	}

	err := app.Run(context.Background(), []string{"sleep", "--limit", "1"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !strings.Contains(stdout.String(), `"id": "123e4567-e89b-12d3-a456-426614174000"`) {
		t.Fatalf("stdout = %q, missing sleep id", stdout.String())
	}
	if !strings.Contains(stdout.String(), `"next_token": "next-token"`) {
		t.Fatalf("stdout = %q, missing next_token", stdout.String())
	}
}

func TestRunSleepByIDPrintsJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer

	app := NewApp(t.TempDir())
	app.TokenManager = fakeTokenManager{accessToken: "access-token"}
	app.WhoopClient = fakeWhoopClient{
		t:               t,
		wantAccessToken: "access-token",
		wantSleepID:     "123e4567-e89b-12d3-a456-426614174000",
		sleep:           testSleep(),
	}

	err := app.Run(context.Background(), []string{"sleep", "get", "123e4567-e89b-12d3-a456-426614174000"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !strings.Contains(stdout.String(), `"score_state": "SCORED"`) {
		t.Fatalf("stdout = %q, missing score_state", stdout.String())
	}
}

func TestRunSleepForCyclePrintsJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer

	app := NewApp(t.TempDir())
	app.TokenManager = fakeTokenManager{accessToken: "access-token"}
	app.WhoopClient = fakeWhoopClient{
		t:               t,
		wantAccessToken: "access-token",
		wantCycleID:     93845,
		sleepForCycle:   testSleep(),
	}

	err := app.Run(context.Background(), []string{"sleep", "cycle", "93845"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !strings.Contains(stdout.String(), `"cycle_id": 93845`) {
		t.Fatalf("stdout = %q, missing cycle_id", stdout.String())
	}
}

func TestRunSleepRejectsInvalidArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "invalid limit", args: []string{"sleep", "--limit", "nope"}, wantErr: "--limit must be an integer"},
		{name: "missing sleep id", args: []string{"sleep", "get"}, wantErr: "usage: whoopctl sleep get <sleep-id>"},
		{name: "missing cycle id", args: []string{"sleep", "cycle"}, wantErr: "usage: whoopctl sleep cycle <cycle-id>"},
		{name: "invalid cycle id", args: []string{"sleep", "cycle", "nope"}, wantErr: "cycle id must be a positive integer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			app := NewApp(t.TempDir())
			err := app.Run(context.Background(), tt.args, &stdout, &stderr)
			if err == nil {
				t.Fatal("Run returned nil error")
			}
			if got := err.Error(); got != tt.wantErr {
				t.Fatalf("error = %q, want %q", got, tt.wantErr)
			}
			if got := stdout.String(); got != "" {
				t.Fatalf("stdout = %q, want empty", got)
			}
		})
	}
}

func TestRunWorkoutPrintsJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer

	app := NewApp(t.TempDir())
	app.TokenManager = fakeTokenManager{accessToken: "access-token"}
	app.WhoopClient = fakeWhoopClient{
		t:               t,
		wantAccessToken: "access-token",
		workouts: whoop.WorkoutCollection{
			Records:   []whoop.Workout{testWorkout()},
			NextToken: "next-token",
		},
	}

	err := app.Run(context.Background(), []string{"workout", "--limit", "1"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !strings.Contains(stdout.String(), `"sport_name": "running"`) {
		t.Fatalf("stdout = %q, missing sport_name", stdout.String())
	}
	if !strings.Contains(stdout.String(), `"next_token": "next-token"`) {
		t.Fatalf("stdout = %q, missing next_token", stdout.String())
	}
}

func TestRunWorkoutByIDPrintsJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer

	app := NewApp(t.TempDir())
	app.TokenManager = fakeTokenManager{accessToken: "access-token"}
	app.WhoopClient = fakeWhoopClient{
		t:               t,
		wantAccessToken: "access-token",
		wantWorkoutID:   "123e4567-e89b-12d3-a456-426614174000",
		workout:         testWorkout(),
	}

	err := app.Run(context.Background(), []string{"workout", "get", "123e4567-e89b-12d3-a456-426614174000"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !strings.Contains(stdout.String(), `"score_state": "SCORED"`) {
		t.Fatalf("stdout = %q, missing score_state", stdout.String())
	}
}

func TestRunWorkoutRejectsInvalidArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "invalid limit", args: []string{"workout", "--limit", "nope"}, wantErr: "--limit must be an integer"},
		{name: "missing workout id", args: []string{"workout", "get"}, wantErr: "usage: whoopctl workout get <workout-id>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			app := NewApp(t.TempDir())
			err := app.Run(context.Background(), tt.args, &stdout, &stderr)
			if err == nil {
				t.Fatal("Run returned nil error")
			}
			if got := err.Error(); got != tt.wantErr {
				t.Fatalf("error = %q, want %q", got, tt.wantErr)
			}
			if got := stdout.String(); got != "" {
				t.Fatalf("stdout = %q, want empty", got)
			}
		})
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
		wantQuery:       whoop.RecoveryQuery{},
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

func TestRunRecoveryPassesQuery(t *testing.T) {
	var stdout, stderr bytes.Buffer
	start := time.Date(2026, 1, 2, 3, 4, 5, 123456789, time.UTC)
	end := time.Date(2026, 1, 3, 3, 4, 5, 987654321, time.UTC)

	app := NewApp(t.TempDir())
	app.TokenManager = fakeTokenManager{accessToken: "access-token"}
	app.WhoopClient = fakeWhoopClient{
		t:               t,
		wantAccessToken: "access-token",
		wantQuery: whoop.RecoveryQuery{
			Limit:     25,
			Start:     start,
			End:       end,
			NextToken: "next-token",
		},
		recovery: whoop.RecoveryCollection{Records: []whoop.Recovery{}},
	}

	err := app.Run(context.Background(), []string{
		"recovery",
		"--limit", "25",
		"--start", start.Format(time.RFC3339Nano),
		"--end", end.Format(time.RFC3339Nano),
		"--next-token", "next-token",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
}

func TestRunRecoveryForCyclePrintsJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	createdAt := time.Date(2022, 4, 24, 11, 25, 44, 774000000, time.UTC)
	updatedAt := time.Date(2022, 4, 24, 14, 25, 44, 774000000, time.UTC)

	app := NewApp(t.TempDir())
	app.TokenManager = fakeTokenManager{accessToken: "access-token"}
	app.WhoopClient = fakeWhoopClient{
		t:               t,
		wantAccessToken: "access-token",
		wantCycleID:     93845,
		recoveryForCycle: whoop.Recovery{
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
			},
		},
	}

	err := app.Run(context.Background(), []string{"recovery", "cycle", "93845"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	want := `{
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
    "hrv_rmssd_milli": 31.813562
  }
}
`
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestRunRecoveryForCycleRejectsInvalidArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing cycle id",
			args:    []string{"recovery", "cycle"},
			wantErr: "usage: whoopctl recovery cycle <cycle-id>",
		},
		{
			name:    "extra arg",
			args:    []string{"recovery", "cycle", "93845", "extra"},
			wantErr: "usage: whoopctl recovery cycle <cycle-id>",
		},
		{
			name:    "not integer",
			args:    []string{"recovery", "cycle", "nope"},
			wantErr: "cycle id must be a positive integer",
		},
		{
			name:    "zero",
			args:    []string{"recovery", "cycle", "0"},
			wantErr: "cycle id must be a positive integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			accessTokenCalled := false

			app := NewApp(t.TempDir())
			app.TokenManager = fakeTokenManager{
				onAccessToken: func() {
					accessTokenCalled = true
				},
			}

			err := app.Run(context.Background(), tt.args, &stdout, &stderr)
			if err == nil {
				t.Fatal("Run returned nil error")
			}
			if got := err.Error(); got != tt.wantErr {
				t.Fatalf("error = %q, want %q", got, tt.wantErr)
			}
			if accessTokenCalled {
				t.Fatal("AccessToken was called")
			}
			if got := stdout.String(); got != "" {
				t.Fatalf("stdout = %q, want empty", got)
			}
		})
	}
}

func TestRunRecoveryForCycleReturnsRecoveryError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	app := NewApp(t.TempDir())
	app.TokenManager = fakeTokenManager{accessToken: "access-token"}
	app.WhoopClient = fakeWhoopClient{recoveryForCycleErr: fmt.Errorf("recovery for cycle failed")}

	err := app.Run(context.Background(), []string{"recovery", "cycle", "93845"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("Run returned nil error")
	}
	if got, want := err.Error(), "recovery for cycle failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
}

func TestRunRecoveryRejectsInvalidLimit(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing limit",
			args:    []string{"recovery", "--limit"},
			wantErr: "--limit requires a value",
		},
		{
			name:    "invalid limit",
			args:    []string{"recovery", "--limit", "nope"},
			wantErr: "--limit must be an integer",
		},
		{
			name:    "missing start",
			args:    []string{"recovery", "--start"},
			wantErr: "--start requires a value",
		},
		{
			name:    "invalid start",
			args:    []string{"recovery", "--start", "2026-01-02"},
			wantErr: "--start must be RFC3339",
		},
		{
			name:    "missing end",
			args:    []string{"recovery", "--end"},
			wantErr: "--end requires a value",
		},
		{
			name:    "invalid end",
			args:    []string{"recovery", "--end", "2026-01-03"},
			wantErr: "--end must be RFC3339",
		},
		{
			name:    "missing next token",
			args:    []string{"recovery", "--next-token"},
			wantErr: "--next-token requires a value",
		},
		{
			name:    "unknown option",
			args:    []string{"recovery", "--wat"},
			wantErr: "unknown option: --wat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			accessTokenCalled := false

			app := NewApp(t.TempDir())
			app.TokenManager = fakeTokenManager{
				onAccessToken: func() {
					accessTokenCalled = true
				},
			}

			err := app.Run(context.Background(), tt.args, &stdout, &stderr)
			if err == nil {
				t.Fatal("Run returned nil error")
			}
			if got := err.Error(); got != tt.wantErr {
				t.Fatalf("error = %q, want %q", got, tt.wantErr)
			}
			if accessTokenCalled {
				t.Fatal("AccessToken was called")
			}
			if got := stdout.String(); got != "" {
				t.Fatalf("stdout = %q, want empty", got)
			}
		})
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
	onAccessToken  func()
	onRefresh      func()
}

func (f fakeTokenManager) AccessToken(ctx context.Context) (string, error) {
	if f.onAccessToken != nil {
		f.onAccessToken()
	}
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
	t                   *testing.T
	wantAccessToken     string
	wantCycleQuery      whoop.CycleQuery
	wantQuery           whoop.RecoveryQuery
	wantCycleID         int64
	cycle               whoop.Cycle
	cycleErr            error
	cycles              whoop.CycleCollection
	cyclesErr           error
	recovery            whoop.RecoveryCollection
	recoveryErr         error
	recoveryForCycle    whoop.Recovery
	recoveryForCycleErr error
	sleep               whoop.Sleep
	sleepErr            error
	sleepForCycle       whoop.Sleep
	sleepForCycleErr    error
	sleeps              whoop.SleepCollection
	sleepsErr           error
	wantSleepID         string
	workout             whoop.Workout
	workoutErr          error
	workouts            whoop.WorkoutCollection
	workoutsErr         error
	wantWorkoutID       string
}

func (f fakeWhoopClient) Cycle(ctx context.Context, accessToken string, cycleID int64) (whoop.Cycle, error) {
	if f.t != nil {
		f.t.Helper()
	}
	if f.wantAccessToken != "" && accessToken != f.wantAccessToken {
		f.t.Fatalf("accessToken = %q, want %q", accessToken, f.wantAccessToken)
	}
	if f.wantCycleID != 0 && cycleID != f.wantCycleID {
		f.t.Fatalf("cycleID = %d, want %d", cycleID, f.wantCycleID)
	}
	if f.cycleErr != nil {
		return whoop.Cycle{}, f.cycleErr
	}
	return f.cycle, nil
}

func (f fakeWhoopClient) Cycles(ctx context.Context, accessToken string, query whoop.CycleQuery) (whoop.CycleCollection, error) {
	if f.t != nil {
		f.t.Helper()
	}
	if f.wantAccessToken != "" && accessToken != f.wantAccessToken {
		f.t.Fatalf("accessToken = %q, want %q", accessToken, f.wantAccessToken)
	}
	if query != f.wantCycleQuery {
		f.t.Fatalf("query = %+v, want %+v", query, f.wantCycleQuery)
	}
	if f.cyclesErr != nil {
		return whoop.CycleCollection{}, f.cyclesErr
	}
	return f.cycles, nil
}

func (f fakeWhoopClient) Recovery(ctx context.Context, accessToken string, query whoop.RecoveryQuery) (whoop.RecoveryCollection, error) {
	if f.t != nil {
		f.t.Helper()
	}
	if f.wantAccessToken != "" && accessToken != f.wantAccessToken {
		f.t.Fatalf("accessToken = %q, want %q", accessToken, f.wantAccessToken)
	}
	if query != f.wantQuery {
		f.t.Fatalf("query = %+v, want %+v", query, f.wantQuery)
	}
	if f.recoveryErr != nil {
		return whoop.RecoveryCollection{}, f.recoveryErr
	}
	return f.recovery, nil
}

func (f fakeWhoopClient) RecoveryForCycle(ctx context.Context, accessToken string, cycleID int64) (whoop.Recovery, error) {
	if f.t != nil {
		f.t.Helper()
	}
	if f.wantAccessToken != "" && accessToken != f.wantAccessToken {
		f.t.Fatalf("accessToken = %q, want %q", accessToken, f.wantAccessToken)
	}
	if f.wantCycleID != 0 && cycleID != f.wantCycleID {
		f.t.Fatalf("cycleID = %d, want %d", cycleID, f.wantCycleID)
	}
	if f.recoveryForCycleErr != nil {
		return whoop.Recovery{}, f.recoveryForCycleErr
	}
	return f.recoveryForCycle, nil
}

func (f fakeWhoopClient) Sleep(ctx context.Context, accessToken string, sleepID string) (whoop.Sleep, error) {
	if f.t != nil {
		f.t.Helper()
	}
	if f.wantAccessToken != "" && accessToken != f.wantAccessToken {
		f.t.Fatalf("accessToken = %q, want %q", accessToken, f.wantAccessToken)
	}
	if f.wantSleepID != "" && sleepID != f.wantSleepID {
		f.t.Fatalf("sleepID = %q, want %q", sleepID, f.wantSleepID)
	}
	if f.sleepErr != nil {
		return whoop.Sleep{}, f.sleepErr
	}
	return f.sleep, nil
}

func (f fakeWhoopClient) SleepForCycle(ctx context.Context, accessToken string, cycleID int64) (whoop.Sleep, error) {
	if f.t != nil {
		f.t.Helper()
	}
	if f.wantAccessToken != "" && accessToken != f.wantAccessToken {
		f.t.Fatalf("accessToken = %q, want %q", accessToken, f.wantAccessToken)
	}
	if f.wantCycleID != 0 && cycleID != f.wantCycleID {
		f.t.Fatalf("cycleID = %d, want %d", cycleID, f.wantCycleID)
	}
	if f.sleepForCycleErr != nil {
		return whoop.Sleep{}, f.sleepForCycleErr
	}
	return f.sleepForCycle, nil
}

func (f fakeWhoopClient) Sleeps(ctx context.Context, accessToken string, query whoop.SleepQuery) (whoop.SleepCollection, error) {
	if f.t != nil {
		f.t.Helper()
	}
	if f.wantAccessToken != "" && accessToken != f.wantAccessToken {
		f.t.Fatalf("accessToken = %q, want %q", accessToken, f.wantAccessToken)
	}
	if f.sleepsErr != nil {
		return whoop.SleepCollection{}, f.sleepsErr
	}
	return f.sleeps, nil
}

func testSleep() whoop.Sleep {
	timestamp := time.Date(2022, 4, 24, 2, 25, 44, 774000000, time.UTC)
	return whoop.Sleep{
		ID:             "123e4567-e89b-12d3-a456-426614174000",
		CycleID:        93845,
		UserID:         10129,
		CreatedAt:      timestamp,
		UpdatedAt:      timestamp,
		Start:          timestamp,
		End:            timestamp.Add(8 * time.Hour),
		TimezoneOffset: "-05:00",
		Nap:            false,
		ScoreState:     "SCORED",
	}
}

func (f fakeWhoopClient) Workout(ctx context.Context, accessToken string, workoutID string) (whoop.Workout, error) {
	if f.t != nil {
		f.t.Helper()
	}
	if f.wantAccessToken != "" && accessToken != f.wantAccessToken {
		f.t.Fatalf("accessToken = %q, want %q", accessToken, f.wantAccessToken)
	}
	if f.wantWorkoutID != "" && workoutID != f.wantWorkoutID {
		f.t.Fatalf("workoutID = %q, want %q", workoutID, f.wantWorkoutID)
	}
	if f.workoutErr != nil {
		return whoop.Workout{}, f.workoutErr
	}
	return f.workout, nil
}

func (f fakeWhoopClient) Workouts(ctx context.Context, accessToken string, query whoop.WorkoutQuery) (whoop.WorkoutCollection, error) {
	if f.t != nil {
		f.t.Helper()
	}
	if f.wantAccessToken != "" && accessToken != f.wantAccessToken {
		f.t.Fatalf("accessToken = %q, want %q", accessToken, f.wantAccessToken)
	}
	if f.workoutsErr != nil {
		return whoop.WorkoutCollection{}, f.workoutsErr
	}
	return f.workouts, nil
}

func testWorkout() whoop.Workout {
	timestamp := time.Date(2022, 4, 24, 2, 25, 44, 774000000, time.UTC)
	return whoop.Workout{
		ID:             "123e4567-e89b-12d3-a456-426614174000",
		UserID:         10129,
		CreatedAt:      timestamp,
		UpdatedAt:      timestamp,
		Start:          timestamp,
		End:            timestamp.Add(time.Hour),
		TimezoneOffset: "-05:00",
		SportName:      "running",
		ScoreState:     "SCORED",
		Score: &whoop.WorkoutScore{
			Strain:           8.2463,
			AverageHeartRate: 123,
			MaxHeartRate:     146,
			Kilojoule:        1569.34033203125,
			PercentRecorded:  100,
			ZoneDurations: whoop.ZoneDurations{
				ZoneZeroMilli:  300000,
				ZoneOneMilli:   600000,
				ZoneTwoMilli:   900000,
				ZoneThreeMilli: 900000,
				ZoneFourMilli:  600000,
				ZoneFiveMilli:  300000,
			},
		},
	}
}
