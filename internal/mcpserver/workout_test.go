package mcpserver

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Tianbo-Qiu/whoopctl/internal/whoop"
)

func TestGetWorkout(t *testing.T) {
	start := time.Date(2026, 1, 2, 3, 4, 5, 123456789, time.UTC)
	end := time.Date(2026, 1, 3, 3, 4, 5, 987654321, time.UTC)

	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient: fakeWhoopClient{
			t:               t,
			wantAccessToken: "access-token",
			workouts: whoop.WorkoutCollection{
				Records:   []whoop.Workout{mcpTestWorkout(start, end)},
				NextToken: "next-page",
			},
		},
	}

	result, output, err := service.getWorkout(context.Background(), nil, GetWorkoutInput{
		Limit:     25,
		Start:     start.Format(time.RFC3339Nano),
		End:       end.Format(time.RFC3339Nano),
		NextToken: "next-token",
	})
	if err != nil {
		t.Fatalf("getWorkout returned error: %v", err)
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
	if got, want := output.Records[0].SportName, "running"; got != want {
		t.Fatalf("SportName = %q, want %q", got, want)
	}
}

func TestGetWorkoutByID(t *testing.T) {
	start := time.Date(2026, 1, 2, 3, 4, 5, 123456789, time.UTC)
	end := time.Date(2026, 1, 3, 3, 4, 5, 987654321, time.UTC)

	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient: fakeWhoopClient{
			t:               t,
			wantAccessToken: "access-token",
			wantWorkoutID:   "123e4567-e89b-12d3-a456-426614174000",
			workout:         mcpTestWorkout(start, end),
		},
	}

	_, output, err := service.getWorkoutByID(context.Background(), nil, GetWorkoutByIDInput{
		WorkoutID: "123e4567-e89b-12d3-a456-426614174000",
	})
	if err != nil {
		t.Fatalf("getWorkoutByID returned error: %v", err)
	}
	if got, want := output.ScoreState, "SCORED"; got != want {
		t.Fatalf("ScoreState = %q, want %q", got, want)
	}
}

func TestGetWorkoutReturnsInvalidStartError(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient:  fakeWhoopClient{},
	}

	_, _, err := service.getWorkout(context.Background(), nil, GetWorkoutInput{Start: "2026-01-02"})
	if err == nil {
		t.Fatal("getWorkout returned nil error")
	}
}

func TestGetWorkoutReturnsWhoopError(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient:  fakeWhoopClient{workoutsErr: fmt.Errorf("workouts failed")},
	}

	_, _, err := service.getWorkout(context.Background(), nil, GetWorkoutInput{})
	if err == nil {
		t.Fatal("getWorkout returned nil error")
	}
	if got, want := err.Error(), "workouts failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func mcpTestWorkout(start time.Time, end time.Time) whoop.Workout {
	return whoop.Workout{
		ID:             "123e4567-e89b-12d3-a456-426614174000",
		UserID:         10129,
		CreatedAt:      start,
		UpdatedAt:      end,
		Start:          start,
		End:            end,
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
