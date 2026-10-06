package whoop

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestWorkoutsSendsRequestAndDecodesResponse(t *testing.T) {
	client := &Client{
		BaseURL: "https://example.test/developer/",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if got, want := req.Method, http.MethodGet; got != want {
					t.Fatalf("method = %q, want %q", got, want)
				}
				if got, want := req.URL.String(), "https://example.test/developer/v2/activity/workout"; got != want {
					t.Fatalf("url = %q, want %q", got, want)
				}
				if got, want := req.Header.Get("Authorization"), "Bearer access-token"; got != want {
					t.Fatalf("Authorization = %q, want %q", got, want)
				}

				return jsonResponse(t, http.StatusOK, "200 OK", map[string]any{
					"records":    []any{workoutBody()},
					"next_token": "next-token",
				}), nil
			}),
		},
	}

	workouts, err := client.Workouts(context.Background(), " access-token ", WorkoutQuery{})
	if err != nil {
		t.Fatalf("Workouts returned error: %v", err)
	}
	if got, want := workouts.NextToken, "next-token"; got != want {
		t.Fatalf("NextToken = %q, want %q", got, want)
	}
	if got, want := len(workouts.Records), 1; got != want {
		t.Fatalf("len(Records) = %d, want %d", got, want)
	}
	workout := workouts.Records[0]
	if got, want := workout.ID, "123e4567-e89b-12d3-a456-426614174000"; got != want {
		t.Fatalf("ID = %q, want %q", got, want)
	}
	if workout.Score == nil {
		t.Fatal("Score is nil")
	}
	if got, want := workout.Score.ZoneDurations.ZoneFiveMilli, int64(300000); got != want {
		t.Fatalf("ZoneFiveMilli = %d, want %d", got, want)
	}
}

func TestWorkoutsSendsQueryParams(t *testing.T) {
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

	_, err := client.Workouts(context.Background(), "access-token", WorkoutQuery{
		Limit:     25,
		Start:     start,
		End:       end,
		NextToken: " next-token ",
	})
	if err != nil {
		t.Fatalf("Workouts returned error: %v", err)
	}
}

func TestWorkoutSendsRequestAndDecodesResponse(t *testing.T) {
	client := &Client{
		BaseURL: "https://example.test/developer/",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if got, want := req.URL.String(), "https://example.test/developer/v2/activity/workout/123e4567-e89b-12d3-a456-426614174000"; got != want {
					t.Fatalf("url = %q, want %q", got, want)
				}
				return jsonResponse(t, http.StatusOK, "200 OK", workoutBody()), nil
			}),
		},
	}

	workout, err := client.Workout(context.Background(), "access-token", "123e4567-e89b-12d3-a456-426614174000")
	if err != nil {
		t.Fatalf("Workout returned error: %v", err)
	}
	if got, want := workout.SportName, "running"; got != want {
		t.Fatalf("SportName = %q, want %q", got, want)
	}
}

func TestWorkoutRequiresInputs(t *testing.T) {
	client := &Client{}

	if _, err := client.Workouts(context.Background(), "", WorkoutQuery{}); err == nil || err.Error() != "access token is required" {
		t.Fatalf("Workouts error = %v, want access token is required", err)
	}
	if _, err := client.Workout(context.Background(), "access-token", ""); err == nil || err.Error() != "workout id is required" {
		t.Fatalf("Workout error = %v, want workout id is required", err)
	}
}

func TestWorkoutReturnsAPIErrorForNon2xx(t *testing.T) {
	client := &Client{
		BaseURL: "https://example.test/developer",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return jsonResponse(t, http.StatusNotFound, "404 Not Found", map[string]any{}), nil
			}),
		},
	}

	_, err := client.Workout(context.Background(), "access-token", "123e4567-e89b-12d3-a456-426614174000")
	if err == nil {
		t.Fatal("Workout returned nil error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *APIError", err)
	}
	if got, want := apiErr.StatusCode, http.StatusNotFound; got != want {
		t.Fatalf("StatusCode = %d, want %d", got, want)
	}
}

func workoutBody() map[string]any {
	return map[string]any{
		"id":              "123e4567-e89b-12d3-a456-426614174000",
		"user_id":         10129,
		"created_at":      "2022-04-24T11:25:44.774Z",
		"updated_at":      "2022-04-24T14:25:44.774Z",
		"start":           "2022-04-24T02:25:44.774Z",
		"end":             "2022-04-24T10:25:44.774Z",
		"timezone_offset": "-05:00",
		"sport_name":      "running",
		"score_state":     "SCORED",
		"score": map[string]any{
			"strain":             8.2463,
			"average_heart_rate": 123,
			"max_heart_rate":     146,
			"kilojoule":          1569.34033203125,
			"percent_recorded":   100.0,
			"zone_durations": map[string]any{
				"zone_zero_milli":  300000,
				"zone_one_milli":   600000,
				"zone_two_milli":   900000,
				"zone_three_milli": 900000,
				"zone_four_milli":  600000,
				"zone_five_milli":  300000,
			},
		},
	}
}
