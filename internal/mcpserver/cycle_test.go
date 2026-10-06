package mcpserver

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Tianbo-Qiu/whoopctl/internal/whoop"
)

func TestGetCycle(t *testing.T) {
	start := time.Date(2026, 1, 2, 3, 4, 5, 123456789, time.UTC)
	end := time.Date(2026, 1, 3, 3, 4, 5, 987654321, time.UTC)
	stepCount := 8234

	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient: fakeWhoopClient{
			t:               t,
			wantAccessToken: "access-token",
			wantCycleQuery: whoop.CycleQuery{
				Limit:     25,
				Start:     start,
				End:       end,
				NextToken: "next-token",
			},
			cycles: whoop.CycleCollection{
				Records: []whoop.Cycle{
					{
						ID:             93845,
						UserID:         10129,
						CreatedAt:      start,
						UpdatedAt:      end,
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
				NextToken: "next-page",
			},
		},
	}

	result, output, err := service.getCycle(context.Background(), nil, GetCycleInput{
		Limit:     25,
		Start:     start.Format(time.RFC3339Nano),
		End:       end.Format(time.RFC3339Nano),
		NextToken: "next-token",
	})
	if err != nil {
		t.Fatalf("getCycle returned error: %v", err)
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
	if got, want := record.ID, int64(93845); got != want {
		t.Fatalf("ID = %d, want %d", got, want)
	}
	if got, want := record.End, end.Format(time.RFC3339Nano); got != want {
		t.Fatalf("End = %q, want %q", got, want)
	}
	if record.Score == nil {
		t.Fatal("Score is nil")
	}
	if got, want := record.Score.AverageHeartRate, 68; got != want {
		t.Fatalf("AverageHeartRate = %d, want %d", got, want)
	}
	if record.StepCount == nil {
		t.Fatal("StepCount is nil")
	}
	if got, want := *record.StepCount, stepCount; got != want {
		t.Fatalf("StepCount = %d, want %d", got, want)
	}
}

func TestGetCycleReturnsAccessTokenError(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessTokenErr: fmt.Errorf("access token failed")},
		WhoopClient:  fakeWhoopClient{},
	}

	_, _, err := service.getCycle(context.Background(), nil, GetCycleInput{})
	if err == nil {
		t.Fatal("getCycle returned nil error")
	}
	if got, want := err.Error(), "access token failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestGetCycleReturnsInvalidStartError(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient:  fakeWhoopClient{},
	}

	_, _, err := service.getCycle(context.Background(), nil, GetCycleInput{
		Start: "2026-01-02",
	})
	if err == nil {
		t.Fatal("getCycle returned nil error")
	}
}

func TestGetCycleReturnsWhoopError(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient:  fakeWhoopClient{cyclesErr: fmt.Errorf("cycles failed")},
	}

	_, _, err := service.getCycle(context.Background(), nil, GetCycleInput{})
	if err == nil {
		t.Fatal("getCycle returned nil error")
	}
	if got, want := err.Error(), "cycles failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}
