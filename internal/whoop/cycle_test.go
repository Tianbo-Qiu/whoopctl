package whoop

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestCyclesSendsRequestAndDecodesResponse(t *testing.T) {
	createdAt := "2022-04-24T11:25:44.774Z"
	updatedAt := "2022-04-24T14:25:44.774Z"
	start := "2022-04-24T02:25:44.774Z"
	end := "2022-04-24T10:25:44.774Z"
	stepCount := 8234

	client := &Client{
		BaseURL: "https://example.test/developer/",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if got, want := req.Method, http.MethodGet; got != want {
					t.Fatalf("method = %q, want %q", got, want)
				}
				if got, want := req.URL.String(), "https://example.test/developer/v2/cycle"; got != want {
					t.Fatalf("url = %q, want %q", got, want)
				}
				if got, want := req.Header.Get("Authorization"), "Bearer access-token"; got != want {
					t.Fatalf("Authorization = %q, want %q", got, want)
				}

				return jsonResponse(t, http.StatusOK, "200 OK", map[string]any{
					"records": []any{
						map[string]any{
							"id":              93845,
							"user_id":         10129,
							"created_at":      createdAt,
							"updated_at":      updatedAt,
							"start":           start,
							"end":             end,
							"timezone_offset": "-05:00",
							"score_state":     "SCORED",
							"score": map[string]any{
								"strain":             5.2951527,
								"kilojoule":          8288.297,
								"average_heart_rate": 68,
								"max_heart_rate":     141,
							},
							"step_count": stepCount,
						},
					},
					"next_token": "next-token",
				}), nil
			}),
		},
	}

	cycles, err := client.Cycles(context.Background(), " access-token ", CycleQuery{})
	if err != nil {
		t.Fatalf("Cycles returned error: %v", err)
	}

	if got, want := cycles.NextToken, "next-token"; got != want {
		t.Fatalf("NextToken = %q, want %q", got, want)
	}
	if got, want := len(cycles.Records), 1; got != want {
		t.Fatalf("len(Records) = %d, want %d", got, want)
	}

	cycle := cycles.Records[0]
	if got, want := cycle.ID, int64(93845); got != want {
		t.Fatalf("ID = %d, want %d", got, want)
	}
	if got, want := cycle.UserID, int64(10129); got != want {
		t.Fatalf("UserID = %d, want %d", got, want)
	}
	if got, want := cycle.TimezoneOffset, "-05:00"; got != want {
		t.Fatalf("TimezoneOffset = %q, want %q", got, want)
	}
	if got, want := cycle.ScoreState, "SCORED"; got != want {
		t.Fatalf("ScoreState = %q, want %q", got, want)
	}
	if cycle.End == nil {
		t.Fatal("End is nil")
	}
	if cycle.Score == nil {
		t.Fatal("Score is nil")
	}
	if got, want := cycle.Score.Strain, 5.2951527; got != want {
		t.Fatalf("Strain = %f, want %f", got, want)
	}
	if cycle.StepCount == nil {
		t.Fatal("StepCount is nil")
	}
	if got, want := *cycle.StepCount, stepCount; got != want {
		t.Fatalf("StepCount = %d, want %d", got, want)
	}
}

func TestCyclesRequiresAccessToken(t *testing.T) {
	client := &Client{}

	_, err := client.Cycles(context.Background(), "", CycleQuery{})
	if err == nil {
		t.Fatal("Cycles returned nil error")
	}
	if got, want := err.Error(), "access token is required"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestCycleSendsRequestAndDecodesResponse(t *testing.T) {
	start := "2022-04-24T02:25:44.774Z"

	client := &Client{
		BaseURL: "https://example.test/developer/",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if got, want := req.Method, http.MethodGet; got != want {
					t.Fatalf("method = %q, want %q", got, want)
				}
				if got, want := req.URL.String(), "https://example.test/developer/v2/cycle/93845"; got != want {
					t.Fatalf("url = %q, want %q", got, want)
				}
				if got, want := req.Header.Get("Authorization"), "Bearer access-token"; got != want {
					t.Fatalf("Authorization = %q, want %q", got, want)
				}

				return jsonResponse(t, http.StatusOK, "200 OK", map[string]any{
					"id":              93845,
					"user_id":         10129,
					"created_at":      "2022-04-24T11:25:44.774Z",
					"updated_at":      "2022-04-24T14:25:44.774Z",
					"start":           start,
					"timezone_offset": "-05:00",
					"score_state":     "PENDING_SCORE",
				}), nil
			}),
		},
	}

	cycle, err := client.Cycle(context.Background(), " access-token ", 93845)
	if err != nil {
		t.Fatalf("Cycle returned error: %v", err)
	}

	if got, want := cycle.ID, int64(93845); got != want {
		t.Fatalf("ID = %d, want %d", got, want)
	}
	if got, want := cycle.ScoreState, "PENDING_SCORE"; got != want {
		t.Fatalf("ScoreState = %q, want %q", got, want)
	}
	if cycle.End != nil {
		t.Fatalf("End = %v, want nil", cycle.End)
	}
	if cycle.Score != nil {
		t.Fatalf("Score = %#v, want nil", cycle.Score)
	}
}

func TestCycleRequiresAccessToken(t *testing.T) {
	client := &Client{}

	_, err := client.Cycle(context.Background(), "", 93845)
	if err == nil {
		t.Fatal("Cycle returned nil error")
	}
	if got, want := err.Error(), "access token is required"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestCycleRequiresCycleID(t *testing.T) {
	client := &Client{}

	_, err := client.Cycle(context.Background(), "access-token", 0)
	if err == nil {
		t.Fatal("Cycle returned nil error")
	}
	if got, want := err.Error(), "cycle id is required"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestCycleReturnsAPIErrorForNon2xx(t *testing.T) {
	client := &Client{
		BaseURL: "https://example.test/developer",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return jsonResponse(t, http.StatusNotFound, "404 Not Found", map[string]any{}), nil
			}),
		},
	}

	_, err := client.Cycle(context.Background(), "access-token", 93845)
	if err == nil {
		t.Fatal("Cycle returned nil error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *APIError", err)
	}
	if got, want := apiErr.StatusCode, http.StatusNotFound; got != want {
		t.Fatalf("StatusCode = %d, want %d", got, want)
	}
}

func TestCyclesReturnsAPIErrorForNon2xx(t *testing.T) {
	client := &Client{
		BaseURL: "https://example.test/developer",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return jsonResponse(t, http.StatusUnauthorized, "401 Unauthorized", map[string]any{}), nil
			}),
		},
	}

	_, err := client.Cycles(context.Background(), "access-token", CycleQuery{})
	if err == nil {
		t.Fatal("Cycles returned nil error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *APIError", err)
	}
	if got, want := apiErr.StatusCode, http.StatusUnauthorized; got != want {
		t.Fatalf("StatusCode = %d, want %d", got, want)
	}
}

func TestCyclesSendsQueryParams(t *testing.T) {
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

				return jsonResponse(t, http.StatusOK, "200 OK", map[string]any{
					"records": []any{},
				}), nil
			}),
		},
	}

	_, err := client.Cycles(context.Background(), "access-token", CycleQuery{
		Limit:     25,
		Start:     start,
		End:       end,
		NextToken: " next-token ",
	})
	if err != nil {
		t.Fatalf("Cycles returned error: %v", err)
	}
}
