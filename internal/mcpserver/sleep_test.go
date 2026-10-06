package mcpserver

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Tianbo-Qiu/whoopctl/internal/whoop"
)

func TestGetSleep(t *testing.T) {
	start := time.Date(2026, 1, 2, 3, 4, 5, 123456789, time.UTC)
	end := time.Date(2026, 1, 3, 3, 4, 5, 987654321, time.UTC)

	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient: fakeWhoopClient{
			t:               t,
			wantAccessToken: "access-token",
			sleeps: whoop.SleepCollection{
				Records:   []whoop.Sleep{mcpTestSleep(start, end)},
				NextToken: "next-page",
			},
		},
	}

	result, output, err := service.getSleep(context.Background(), nil, GetSleepInput{
		Limit:     25,
		Start:     start.Format(time.RFC3339Nano),
		End:       end.Format(time.RFC3339Nano),
		NextToken: "next-token",
	})
	if err != nil {
		t.Fatalf("getSleep returned error: %v", err)
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
	if got, want := output.Records[0].ID, "123e4567-e89b-12d3-a456-426614174000"; got != want {
		t.Fatalf("ID = %q, want %q", got, want)
	}
}

func TestGetSleepByID(t *testing.T) {
	start := time.Date(2026, 1, 2, 3, 4, 5, 123456789, time.UTC)
	end := time.Date(2026, 1, 3, 3, 4, 5, 987654321, time.UTC)

	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient: fakeWhoopClient{
			t:               t,
			wantAccessToken: "access-token",
			wantSleepID:     "123e4567-e89b-12d3-a456-426614174000",
			sleep:           mcpTestSleep(start, end),
		},
	}

	_, output, err := service.getSleepByID(context.Background(), nil, GetSleepByIDInput{
		SleepID: "123e4567-e89b-12d3-a456-426614174000",
	})
	if err != nil {
		t.Fatalf("getSleepByID returned error: %v", err)
	}
	if got, want := output.ScoreState, "SCORED"; got != want {
		t.Fatalf("ScoreState = %q, want %q", got, want)
	}
}

func TestGetSleepForCycle(t *testing.T) {
	start := time.Date(2026, 1, 2, 3, 4, 5, 123456789, time.UTC)
	end := time.Date(2026, 1, 3, 3, 4, 5, 987654321, time.UTC)

	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient: fakeWhoopClient{
			t:               t,
			wantAccessToken: "access-token",
			wantCycleID:     93845,
			sleepForCycle:   mcpTestSleep(start, end),
		},
	}

	_, output, err := service.getSleepForCycle(context.Background(), nil, GetSleepForCycleInput{
		CycleID: 93845,
	})
	if err != nil {
		t.Fatalf("getSleepForCycle returned error: %v", err)
	}
	if got, want := output.CycleID, int64(93845); got != want {
		t.Fatalf("CycleID = %d, want %d", got, want)
	}
}

func TestGetSleepReturnsInvalidStartError(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient:  fakeWhoopClient{},
	}

	_, _, err := service.getSleep(context.Background(), nil, GetSleepInput{Start: "2026-01-02"})
	if err == nil {
		t.Fatal("getSleep returned nil error")
	}
}

func TestGetSleepReturnsWhoopError(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient:  fakeWhoopClient{sleepsErr: fmt.Errorf("sleeps failed")},
	}

	_, _, err := service.getSleep(context.Background(), nil, GetSleepInput{})
	if err == nil {
		t.Fatal("getSleep returned nil error")
	}
	if got, want := err.Error(), "sleeps failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func mcpTestSleep(start time.Time, end time.Time) whoop.Sleep {
	respRate := 16.1
	performance := 98.0
	return whoop.Sleep{
		ID:             "123e4567-e89b-12d3-a456-426614174000",
		CycleID:        93845,
		UserID:         10129,
		CreatedAt:      start,
		UpdatedAt:      end,
		Start:          start,
		End:            end,
		TimezoneOffset: "-05:00",
		Nap:            false,
		ScoreState:     "SCORED",
		Score: &whoop.SleepScore{
			StageSummary: whoop.SleepStageSummary{
				TotalInBedTimeMilli:         30272735,
				TotalAwakeTimeMilli:         1403507,
				TotalNoDataTimeMilli:        0,
				TotalLightSleepTimeMilli:    14905851,
				TotalSlowWaveSleepTimeMilli: 6630370,
				TotalRemSleepTimeMilli:      5879573,
				SleepCycleCount:             3,
				DisturbanceCount:            12,
			},
			SleepNeeded: whoop.SleepNeeded{
				BaselineMilli:             27395716,
				NeedFromSleepDebtMilli:    352230,
				NeedFromRecentStrainMilli: 208595,
				NeedFromRecentNapMilli:    -12312,
			},
			RespiratoryRate:            &respRate,
			SleepPerformancePercentage: &performance,
		},
	}
}
