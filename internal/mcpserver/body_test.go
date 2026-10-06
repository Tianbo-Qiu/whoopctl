package mcpserver

import (
	"context"
	"fmt"
	"testing"

	"github.com/Tianbo-Qiu/whoopctl/internal/whoop"
)

func TestGetBodyMeasurement(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient: fakeWhoopClient{
			t:               t,
			wantAccessToken: "access-token",
			body: whoop.BodyMeasurement{
				HeightMeter:    1.8288,
				WeightKilogram: 90.7185,
				MaxHeartRate:   200,
			},
		},
	}

	result, output, err := service.getBodyMeasurement(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("getBodyMeasurement returned error: %v", err)
	}
	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if got, want := output.MaxHeartRate, 200; got != want {
		t.Fatalf("MaxHeartRate = %d, want %d", got, want)
	}
}

func TestGetBodyMeasurementReturnsAccessTokenError(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessTokenErr: fmt.Errorf("access token failed")},
		WhoopClient:  fakeWhoopClient{},
	}

	_, _, err := service.getBodyMeasurement(context.Background(), nil, struct{}{})
	if err == nil {
		t.Fatal("getBodyMeasurement returned nil error")
	}
	if got, want := err.Error(), "access token failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestGetBodyMeasurementReturnsWhoopError(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient:  fakeWhoopClient{bodyErr: fmt.Errorf("body failed")},
	}

	_, _, err := service.getBodyMeasurement(context.Background(), nil, struct{}{})
	if err == nil {
		t.Fatal("getBodyMeasurement returned nil error")
	}
	if got, want := err.Error(), "body failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}
