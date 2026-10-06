package mcpserver

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Tianbo-Qiu/whoopctl/internal/whoop"
)

func TestGetRecovery(t *testing.T) {
	start := time.Date(2026, 1, 2, 3, 4, 5, 123456789, time.UTC)
	end := time.Date(2026, 1, 3, 3, 4, 5, 987654321, time.UTC)
	spo2Percentage := 95.6875
	skinTempCelsius := 33.7

	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient: fakeWhoopClient{
			t:               t,
			wantAccessToken: "access-token",
			wantQuery: whoop.RecoveryQuery{
				Limit:     25,
				Start:     start,
				End:       end,
				NextToken: "next-token",
			},
			recovery: whoop.RecoveryCollection{
				Records: []whoop.Recovery{
					{
						CycleID:    93845,
						SleepID:    "123e4567-e89b-12d3-a456-426614174000",
						UserID:     10129,
						CreatedAt:  start,
						UpdatedAt:  end,
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
				NextToken: "next-page",
			},
		},
	}

	result, output, err := service.getRecovery(context.Background(), nil, GetRecoveryInput{
		Limit:     25,
		Start:     start.Format(time.RFC3339Nano),
		End:       end.Format(time.RFC3339Nano),
		NextToken: "next-token",
	})
	if err != nil {
		t.Fatalf("getRecovery returned error: %v", err)
	}
	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if got, want := output.NextToken, "next-page"; got != want {
		t.Fatalf("NextToken = %q, want %q", got, want)
	}
	if got, want := len(output.Records), 1; got != want {
		t.Fatalf("len(Records) = %d, want %d", got, want)
	}

	record := output.Records[0]
	if got, want := record.CycleID, int64(93845); got != want {
		t.Fatalf("CycleID = %d, want %d", got, want)
	}
	if got, want := record.CreatedAt, start.Format(time.RFC3339Nano); got != want {
		t.Fatalf("CreatedAt = %q, want %q", got, want)
	}
	if got, want := record.UpdatedAt, end.Format(time.RFC3339Nano); got != want {
		t.Fatalf("UpdatedAt = %q, want %q", got, want)
	}
	if record.Score == nil {
		t.Fatal("Score is nil")
	}
	if got, want := record.Score.HrvRmssdMilli, 31.813562; got != want {
		t.Fatalf("HrvRmssdMilli = %f, want %f", got, want)
	}
	if record.Score.Spo2Percentage == nil {
		t.Fatal("Spo2Percentage is nil")
	}
	if got, want := *record.Score.Spo2Percentage, spo2Percentage; got != want {
		t.Fatalf("Spo2Percentage = %f, want %f", got, want)
	}
}

func TestGetRecoveryForCycle(t *testing.T) {
	start := time.Date(2026, 1, 2, 3, 4, 5, 123456789, time.UTC)
	end := time.Date(2026, 1, 3, 3, 4, 5, 987654321, time.UTC)
	spo2Percentage := 95.6875

	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient: fakeWhoopClient{
			t:               t,
			wantAccessToken: "access-token",
			wantCycleID:     93845,
			recoveryForCycle: whoop.Recovery{
				CycleID:    93845,
				SleepID:    "123e4567-e89b-12d3-a456-426614174000",
				UserID:     10129,
				CreatedAt:  start,
				UpdatedAt:  end,
				ScoreState: "SCORED",
				Score: &whoop.RecoveryScore{
					UserCalibrating:  false,
					RecoveryScore:    44,
					RestingHeartRate: 64,
					HrvRmssdMilli:    31.813562,
					Spo2Percentage:   &spo2Percentage,
				},
			},
		},
	}

	result, output, err := service.getRecoveryForCycle(context.Background(), nil, GetRecoveryForCycleInput{
		CycleID: 93845,
	})
	if err != nil {
		t.Fatalf("getRecoveryForCycle returned error: %v", err)
	}
	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if got, want := output.CycleID, int64(93845); got != want {
		t.Fatalf("CycleID = %d, want %d", got, want)
	}
	if got, want := output.CreatedAt, start.Format(time.RFC3339Nano); got != want {
		t.Fatalf("CreatedAt = %q, want %q", got, want)
	}
	if output.Score == nil {
		t.Fatal("Score is nil")
	}
	if output.Score.Spo2Percentage == nil {
		t.Fatal("Spo2Percentage is nil")
	}
	if got, want := *output.Score.Spo2Percentage, spo2Percentage; got != want {
		t.Fatalf("Spo2Percentage = %f, want %f", got, want)
	}
}

func TestGetRecoveryForCycleReturnsAccessTokenError(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessTokenErr: fmt.Errorf("access token failed")},
		WhoopClient:  fakeWhoopClient{},
	}

	_, _, err := service.getRecoveryForCycle(context.Background(), nil, GetRecoveryForCycleInput{
		CycleID: 93845,
	})
	if err == nil {
		t.Fatal("getRecoveryForCycle returned nil error")
	}
	if got, want := err.Error(), "access token failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestGetRecoveryForCycleReturnsWhoopError(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient:  fakeWhoopClient{recoveryForCycleErr: fmt.Errorf("recovery for cycle failed")},
	}

	_, _, err := service.getRecoveryForCycle(context.Background(), nil, GetRecoveryForCycleInput{
		CycleID: 93845,
	})
	if err == nil {
		t.Fatal("getRecoveryForCycle returned nil error")
	}
	if got, want := err.Error(), "recovery for cycle failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestGetRecoveryReturnsAccessTokenError(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessTokenErr: fmt.Errorf("access token failed")},
		WhoopClient:  fakeWhoopClient{},
	}

	_, _, err := service.getRecovery(context.Background(), nil, GetRecoveryInput{})
	if err == nil {
		t.Fatal("getRecovery returned nil error")
	}
	if got, want := err.Error(), "access token failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestGetRecoveryReturnsInvalidStartError(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient:  fakeWhoopClient{},
	}

	_, _, err := service.getRecovery(context.Background(), nil, GetRecoveryInput{
		Start: "2026-01-02",
	})
	if err == nil {
		t.Fatal("getRecovery returned nil error")
	}
}

func TestGetRecoveryReturnsWhoopError(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient:  fakeWhoopClient{recoveryErr: fmt.Errorf("recovery failed")},
	}

	_, _, err := service.getRecovery(context.Background(), nil, GetRecoveryInput{})
	if err == nil {
		t.Fatal("getRecovery returned nil error")
	}
	if got, want := err.Error(), "recovery failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

type fakeTokenManager struct {
	accessToken    string
	accessTokenErr error
	clearErr       error
	onClearToken   func()
}

func (f fakeTokenManager) ClearToken(ctx context.Context) error {
	if f.onClearToken != nil {
		f.onClearToken()
	}
	return f.clearErr
}

func (f fakeTokenManager) AccessToken(ctx context.Context) (string, error) {
	if f.accessTokenErr != nil {
		return "", f.accessTokenErr
	}
	return f.accessToken, nil
}

type fakeWhoopClient struct {
	t                   *testing.T
	wantAccessToken     string
	wantActivityV1ID    int64
	activityMapping     whoop.ActivityMapping
	activityMappingErr  error
	profile             whoop.BasicProfile
	profileErr          error
	body                whoop.BodyMeasurement
	bodyErr             error
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
	revokeErr           error
	onRevoke            func()
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

func (f fakeWhoopClient) ActivityMapping(ctx context.Context, accessToken string, activityV1ID int64) (whoop.ActivityMapping, error) {
	if f.t != nil {
		f.t.Helper()
	}
	if f.wantAccessToken != "" && accessToken != f.wantAccessToken {
		f.t.Fatalf("accessToken = %q, want %q", accessToken, f.wantAccessToken)
	}
	if f.wantActivityV1ID != 0 && activityV1ID != f.wantActivityV1ID {
		f.t.Fatalf("activityV1ID = %d, want %d", activityV1ID, f.wantActivityV1ID)
	}
	if f.activityMappingErr != nil {
		return whoop.ActivityMapping{}, f.activityMappingErr
	}
	return f.activityMapping, nil
}

func (f fakeWhoopClient) BasicProfile(ctx context.Context, accessToken string) (whoop.BasicProfile, error) {
	if f.t != nil {
		f.t.Helper()
	}
	if f.wantAccessToken != "" && accessToken != f.wantAccessToken {
		f.t.Fatalf("accessToken = %q, want %q", accessToken, f.wantAccessToken)
	}
	if f.profileErr != nil {
		return whoop.BasicProfile{}, f.profileErr
	}
	return f.profile, nil
}

func (f fakeWhoopClient) BodyMeasurement(ctx context.Context, accessToken string) (whoop.BodyMeasurement, error) {
	if f.t != nil {
		f.t.Helper()
	}
	if f.wantAccessToken != "" && accessToken != f.wantAccessToken {
		f.t.Fatalf("accessToken = %q, want %q", accessToken, f.wantAccessToken)
	}
	if f.bodyErr != nil {
		return whoop.BodyMeasurement{}, f.bodyErr
	}
	return f.body, nil
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

func (f fakeWhoopClient) RevokeAccess(ctx context.Context, accessToken string) error {
	if f.t != nil {
		f.t.Helper()
	}
	if f.wantAccessToken != "" && accessToken != f.wantAccessToken {
		f.t.Fatalf("accessToken = %q, want %q", accessToken, f.wantAccessToken)
	}
	if f.onRevoke != nil {
		f.onRevoke()
	}
	return f.revokeErr
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
