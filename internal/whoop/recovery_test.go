package whoop

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestRecoverySendsRequestAndDecodesResponse(t *testing.T) {
	createdAt := "2022-04-24T11:25:44.774Z"
	updatedAt := "2022-04-24T14:25:44.774Z"
	spo2Percentage := 95.6875
	skinTempCelsius := 33.7

	client := &Client{
		BaseURL: "https://example.test/developer/",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if got, want := req.Method, http.MethodGet; got != want {
					t.Fatalf("method = %q, want %q", got, want)
				}
				if got, want := req.URL.String(), "https://example.test/developer/v2/recovery"; got != want {
					t.Fatalf("url = %q, want %q", got, want)
				}
				if got, want := req.Header.Get("Authorization"), "Bearer access-token"; got != want {
					t.Fatalf("Authorization = %q, want %q", got, want)
				}

				return jsonResponse(t, http.StatusOK, "200 OK", map[string]any{
					"records": []any{
						map[string]any{
							"cycle_id":    93845,
							"sleep_id":    "123e4567-e89b-12d3-a456-426614174000",
							"user_id":     10129,
							"created_at":  createdAt,
							"updated_at":  updatedAt,
							"score_state": "SCORED",
							"score": map[string]any{
								"user_calibrating":   false,
								"recovery_score":     44,
								"resting_heart_rate": 64,
								"hrv_rmssd_milli":    31.813562,
								"spo2_percentage":    spo2Percentage,
								"skin_temp_celsius":  skinTempCelsius,
							},
						},
					},
					"next_token": "next-token",
				}), nil
			}),
		},
	}

	recovery, err := client.Recovery(context.Background(), " access-token ", RecoveryQuery{})
	if err != nil {
		t.Fatalf("Recovery returned error: %v", err)
	}

	if got, want := recovery.NextToken, "next-token"; got != want {
		t.Fatalf("NextToken = %q, want %q", got, want)
	}
	if got, want := len(recovery.Records), 1; got != want {
		t.Fatalf("len(Records) = %d, want %d", got, want)
	}

	record := recovery.Records[0]
	if got, want := record.CycleID, int64(93845); got != want {
		t.Fatalf("CycleID = %d, want %d", got, want)
	}
	if got, want := record.SleepID, "123e4567-e89b-12d3-a456-426614174000"; got != want {
		t.Fatalf("SleepID = %q, want %q", got, want)
	}
	if got, want := record.UserID, int64(10129); got != want {
		t.Fatalf("UserID = %d, want %d", got, want)
	}

	wantCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		t.Fatalf("Parse createdAt returned error: %v", err)
	}
	if got := record.CreatedAt; !got.Equal(wantCreatedAt) {
		t.Fatalf("CreatedAt = %v, want %v", got, wantCreatedAt)
	}

	wantUpdatedAt, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		t.Fatalf("Parse updatedAt returned error: %v", err)
	}
	if got := record.UpdatedAt; !got.Equal(wantUpdatedAt) {
		t.Fatalf("UpdatedAt = %v, want %v", got, wantUpdatedAt)
	}

	if got, want := record.ScoreState, "SCORED"; got != want {
		t.Fatalf("ScoreState = %q, want %q", got, want)
	}
	if record.Score == nil {
		t.Fatal("Score is nil")
	}
	if got, want := record.Score.UserCalibrating, false; got != want {
		t.Fatalf("UserCalibrating = %t, want %t", got, want)
	}
	if got, want := record.Score.RecoveryScore, 44.0; got != want {
		t.Fatalf("RecoveryScore = %f, want %f", got, want)
	}
	if got, want := record.Score.RestingHeartRate, 64.0; got != want {
		t.Fatalf("RestingHeartRate = %f, want %f", got, want)
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
	if record.Score.SkinTempCelsius == nil {
		t.Fatal("SkinTempCelsius is nil")
	}
	if got, want := *record.Score.SkinTempCelsius, skinTempCelsius; got != want {
		t.Fatalf("SkinTempCelsius = %f, want %f", got, want)
	}
}

func TestRecoveryRequiresAccessToken(t *testing.T) {
	client := &Client{}

	_, err := client.Recovery(context.Background(), "", RecoveryQuery{})
	if err == nil {
		t.Fatal("Recovery returned nil error")
	}
	if got, want := err.Error(), "access token is required"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestRecoveryReturnsAPIErrorForNon2xx(t *testing.T) {
	client := &Client{
		BaseURL: "https://example.test/developer",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return jsonResponse(t, http.StatusUnauthorized, "401 Unauthorized", map[string]any{}), nil
			}),
		},
	}

	_, err := client.Recovery(context.Background(), "access-token", RecoveryQuery{})
	if err == nil {
		t.Fatal("Recovery returned nil error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *APIError", err)
	}
	if got, want := apiErr.StatusCode, http.StatusUnauthorized; got != want {
		t.Fatalf("StatusCode = %d, want %d", got, want)
	}
	if got, want := apiErr.Message, "invalid authorization"; got != want {
		t.Fatalf("Message = %q, want %q", got, want)
	}
	if got, want := err.Error(), "whoop api error: 401 Unauthorized - invalid authorization"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestRecoverySendsQueryParams(t *testing.T) {
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

	_, err := client.Recovery(context.Background(), "access-token", RecoveryQuery{
		Limit:     25,
		Start:     start,
		End:       end,
		NextToken: " next-token ",
	})
	if err != nil {
		t.Fatalf("Recovery returned error: %v", err)
	}
}

func jsonResponse(t *testing.T, statusCode int, status string, body map[string]any) *http.Response {
	t.Helper()

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	return &http.Response{
		StatusCode: statusCode,
		Status:     status,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(data)),
	}
}
