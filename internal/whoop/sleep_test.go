package whoop

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestSleepsSendsRequestAndDecodesResponse(t *testing.T) {
	start := "2022-04-24T02:25:44.774Z"
	end := "2022-04-24T10:25:44.774Z"
	respiratoryRate := 16.11328125
	sleepPerformance := 98.0
	sleepConsistency := 90.0
	sleepEfficiency := 91.69533848

	client := &Client{
		BaseURL: "https://example.test/developer/",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if got, want := req.Method, http.MethodGet; got != want {
					t.Fatalf("method = %q, want %q", got, want)
				}
				if got, want := req.URL.String(), "https://example.test/developer/v2/activity/sleep"; got != want {
					t.Fatalf("url = %q, want %q", got, want)
				}
				if got, want := req.Header.Get("Authorization"), "Bearer access-token"; got != want {
					t.Fatalf("Authorization = %q, want %q", got, want)
				}

				return jsonResponse(t, http.StatusOK, "200 OK", map[string]any{
					"records": []any{
						map[string]any{
							"id":              "123e4567-e89b-12d3-a456-426614174000",
							"cycle_id":        93845,
							"user_id":         10129,
							"created_at":      "2022-04-24T11:25:44.774Z",
							"updated_at":      "2022-04-24T14:25:44.774Z",
							"start":           start,
							"end":             end,
							"timezone_offset": "-05:00",
							"nap":             false,
							"score_state":     "SCORED",
							"score": map[string]any{
								"stage_summary": map[string]any{
									"total_in_bed_time_milli":          30272735,
									"total_awake_time_milli":           1403507,
									"total_no_data_time_milli":         0,
									"total_light_sleep_time_milli":     14905851,
									"total_slow_wave_sleep_time_milli": 6630370,
									"total_rem_sleep_time_milli":       5879573,
									"sleep_cycle_count":                3,
									"disturbance_count":                12,
								},
								"sleep_needed": map[string]any{
									"baseline_milli":                27395716,
									"need_from_sleep_debt_milli":    352230,
									"need_from_recent_strain_milli": 208595,
									"need_from_recent_nap_milli":    -12312,
								},
								"respiratory_rate":             respiratoryRate,
								"sleep_performance_percentage": sleepPerformance,
								"sleep_consistency_percentage": sleepConsistency,
								"sleep_efficiency_percentage":  sleepEfficiency,
							},
						},
					},
					"next_token": "next-token",
				}), nil
			}),
		},
	}

	sleeps, err := client.Sleeps(context.Background(), " access-token ", SleepQuery{})
	if err != nil {
		t.Fatalf("Sleeps returned error: %v", err)
	}
	if got, want := sleeps.NextToken, "next-token"; got != want {
		t.Fatalf("NextToken = %q, want %q", got, want)
	}
	if got, want := len(sleeps.Records), 1; got != want {
		t.Fatalf("len(Records) = %d, want %d", got, want)
	}

	sleep := sleeps.Records[0]
	if got, want := sleep.ID, "123e4567-e89b-12d3-a456-426614174000"; got != want {
		t.Fatalf("ID = %q, want %q", got, want)
	}
	if got, want := sleep.CycleID, int64(93845); got != want {
		t.Fatalf("CycleID = %d, want %d", got, want)
	}
	if sleep.Score == nil {
		t.Fatal("Score is nil")
	}
	if got, want := sleep.Score.StageSummary.SleepCycleCount, 3; got != want {
		t.Fatalf("SleepCycleCount = %d, want %d", got, want)
	}
	if sleep.Score.RespiratoryRate == nil {
		t.Fatal("RespiratoryRate is nil")
	}
	if got, want := *sleep.Score.RespiratoryRate, respiratoryRate; got != want {
		t.Fatalf("RespiratoryRate = %f, want %f", got, want)
	}
}

func TestSleepsSendsQueryParams(t *testing.T) {
	start := time.Date(2026, 1, 2, 3, 4, 5, 123456789, time.UTC)
	end := time.Date(2026, 1, 3, 3, 4, 5, 987654321, time.UTC)

	client := &Client{
		BaseURL: "https://example.test/developer",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				query := req.URL.Query()
				if got, want := query.Get("limit"), "25"; got != want {
					t.Fatalf("limit = %q, want %q", got, want)
				}
				if got, want := query.Get("start"), start.Format(time.RFC3339Nano); got != want {
					t.Fatalf("start = %q, want %q", got, want)
				}
				if got, want := query.Get("end"), end.Format(time.RFC3339Nano); got != want {
					t.Fatalf("end = %q, want %q", got, want)
				}
				if got, want := query.Get("nextToken"), "next-token"; got != want {
					t.Fatalf("nextToken = %q, want %q", got, want)
				}

				return jsonResponse(t, http.StatusOK, "200 OK", map[string]any{"records": []any{}}), nil
			}),
		},
	}

	_, err := client.Sleeps(context.Background(), "access-token", SleepQuery{
		Limit:     25,
		Start:     start,
		End:       end,
		NextToken: " next-token ",
	})
	if err != nil {
		t.Fatalf("Sleeps returned error: %v", err)
	}
}

func TestSleepSendsRequestAndDecodesResponse(t *testing.T) {
	client := &Client{
		BaseURL: "https://example.test/developer/",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if got, want := req.URL.String(), "https://example.test/developer/v2/activity/sleep/123e4567-e89b-12d3-a456-426614174000"; got != want {
					t.Fatalf("url = %q, want %q", got, want)
				}
				return sleepResponse(t), nil
			}),
		},
	}

	sleep, err := client.Sleep(context.Background(), "access-token", "123e4567-e89b-12d3-a456-426614174000")
	if err != nil {
		t.Fatalf("Sleep returned error: %v", err)
	}
	if got, want := sleep.ID, "123e4567-e89b-12d3-a456-426614174000"; got != want {
		t.Fatalf("ID = %q, want %q", got, want)
	}
}

func TestSleepForCycleSendsRequestAndDecodesResponse(t *testing.T) {
	client := &Client{
		BaseURL: "https://example.test/developer/",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if got, want := req.URL.String(), "https://example.test/developer/v2/cycle/93845/sleep"; got != want {
					t.Fatalf("url = %q, want %q", got, want)
				}
				return sleepResponse(t), nil
			}),
		},
	}

	sleep, err := client.SleepForCycle(context.Background(), "access-token", 93845)
	if err != nil {
		t.Fatalf("SleepForCycle returned error: %v", err)
	}
	if got, want := sleep.CycleID, int64(93845); got != want {
		t.Fatalf("CycleID = %d, want %d", got, want)
	}
}

func TestSleepRequiresInputs(t *testing.T) {
	client := &Client{}

	if _, err := client.Sleeps(context.Background(), "", SleepQuery{}); err == nil || err.Error() != "access token is required" {
		t.Fatalf("Sleeps error = %v, want access token is required", err)
	}
	if _, err := client.Sleep(context.Background(), "access-token", ""); err == nil || err.Error() != "sleep id is required" {
		t.Fatalf("Sleep error = %v, want sleep id is required", err)
	}
	if _, err := client.SleepForCycle(context.Background(), "access-token", 0); err == nil || err.Error() != "cycle id is required" {
		t.Fatalf("SleepForCycle error = %v, want cycle id is required", err)
	}
}

func TestSleepReturnsAPIErrorForNon2xx(t *testing.T) {
	client := &Client{
		BaseURL: "https://example.test/developer",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return jsonResponse(t, http.StatusUnauthorized, "401 Unauthorized", map[string]any{}), nil
			}),
		},
	}

	_, err := client.Sleep(context.Background(), "access-token", "123e4567-e89b-12d3-a456-426614174000")
	if err == nil {
		t.Fatal("Sleep returned nil error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *APIError", err)
	}
	if got, want := apiErr.StatusCode, http.StatusUnauthorized; got != want {
		t.Fatalf("StatusCode = %d, want %d", got, want)
	}
}

func sleepResponse(t *testing.T) *http.Response {
	t.Helper()

	return jsonResponse(t, http.StatusOK, "200 OK", map[string]any{
		"id":              "123e4567-e89b-12d3-a456-426614174000",
		"cycle_id":        93845,
		"user_id":         10129,
		"created_at":      "2022-04-24T11:25:44.774Z",
		"updated_at":      "2022-04-24T14:25:44.774Z",
		"start":           "2022-04-24T02:25:44.774Z",
		"end":             "2022-04-24T10:25:44.774Z",
		"timezone_offset": "-05:00",
		"nap":             false,
		"score_state":     "PENDING_SCORE",
	})
}
