package whoop

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestBodyMeasurementSendsRequestAndDecodesResponse(t *testing.T) {
	client := &Client{
		BaseURL: "https://example.test/developer/",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if got, want := req.Method, http.MethodGet; got != want {
					t.Fatalf("method = %q, want %q", got, want)
				}
				if got, want := req.URL.String(), "https://example.test/developer/v2/user/measurement/body"; got != want {
					t.Fatalf("url = %q, want %q", got, want)
				}
				if got, want := req.Header.Get("Authorization"), "Bearer access-token"; got != want {
					t.Fatalf("Authorization = %q, want %q", got, want)
				}

				return jsonResponse(t, http.StatusOK, "200 OK", map[string]any{
					"height_meter":    1.8288,
					"weight_kilogram": 90.7185,
					"max_heart_rate":  200,
				}), nil
			}),
		},
	}

	measurement, err := client.BodyMeasurement(context.Background(), " access-token ")
	if err != nil {
		t.Fatalf("BodyMeasurement returned error: %v", err)
	}
	if got, want := measurement.HeightMeter, 1.8288; got != want {
		t.Fatalf("HeightMeter = %f, want %f", got, want)
	}
	if got, want := measurement.MaxHeartRate, 200; got != want {
		t.Fatalf("MaxHeartRate = %d, want %d", got, want)
	}
}

func TestBodyMeasurementRequiresAccessToken(t *testing.T) {
	client := &Client{}

	_, err := client.BodyMeasurement(context.Background(), "")
	if err == nil {
		t.Fatal("BodyMeasurement returned nil error")
	}
	if got, want := err.Error(), "access token is required"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestBodyMeasurementReturnsAPIErrorForNon2xx(t *testing.T) {
	client := &Client{
		BaseURL: "https://example.test/developer",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return jsonResponse(t, http.StatusUnauthorized, "401 Unauthorized", map[string]any{}), nil
			}),
		},
	}

	_, err := client.BodyMeasurement(context.Background(), "access-token")
	if err == nil {
		t.Fatal("BodyMeasurement returned nil error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *APIError", err)
	}
	if got, want := apiErr.StatusCode, http.StatusUnauthorized; got != want {
		t.Fatalf("StatusCode = %d, want %d", got, want)
	}
}
