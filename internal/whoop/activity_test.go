package whoop

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestActivityMappingSendsRequestAndDecodesResponse(t *testing.T) {
	client := &Client{
		BaseURL: "https://example.test/developer/",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if got, want := req.Method, http.MethodGet; got != want {
					t.Fatalf("method = %q, want %q", got, want)
				}
				if got, want := req.URL.String(), "https://example.test/developer/v1/activity-mapping/12345"; got != want {
					t.Fatalf("url = %q, want %q", got, want)
				}
				if got, want := req.Header.Get("Authorization"), "Bearer access-token"; got != want {
					t.Fatalf("Authorization = %q, want %q", got, want)
				}

				return jsonResponse(t, http.StatusOK, "200 OK", map[string]any{
					"v2_activity_id": "ecfc6a15-4661-442f-a9a4-f160dd7afae8",
				}), nil
			}),
		},
	}

	mapping, err := client.ActivityMapping(context.Background(), " access-token ", 12345)
	if err != nil {
		t.Fatalf("ActivityMapping returned error: %v", err)
	}
	if got, want := mapping.V2ActivityID, "ecfc6a15-4661-442f-a9a4-f160dd7afae8"; got != want {
		t.Fatalf("V2ActivityID = %q, want %q", got, want)
	}
}

func TestActivityMappingRequiresAccessToken(t *testing.T) {
	client := &Client{}

	_, err := client.ActivityMapping(context.Background(), "", 12345)
	if err == nil {
		t.Fatal("ActivityMapping returned nil error")
	}
	if got, want := err.Error(), "access token is required"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestActivityMappingRequiresActivityV1ID(t *testing.T) {
	client := &Client{}

	_, err := client.ActivityMapping(context.Background(), "access-token", 0)
	if err == nil {
		t.Fatal("ActivityMapping returned nil error")
	}
	if got, want := err.Error(), "activity v1 id is required"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestActivityMappingReturnsAPIErrorForNon2xx(t *testing.T) {
	client := &Client{
		BaseURL: "https://example.test/developer",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return jsonResponse(t, http.StatusNotFound, "404 Not Found", map[string]any{}), nil
			}),
		},
	}

	_, err := client.ActivityMapping(context.Background(), "access-token", 12345)
	if err == nil {
		t.Fatal("ActivityMapping returned nil error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *APIError", err)
	}
	if got, want := apiErr.StatusCode, http.StatusNotFound; got != want {
		t.Fatalf("StatusCode = %d, want %d", got, want)
	}
}
