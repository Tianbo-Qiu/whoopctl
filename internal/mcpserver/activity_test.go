package mcpserver

import (
	"context"
	"fmt"
	"testing"

	"github.com/Tianbo-Qiu/whoopctl/internal/whoop"
)

func TestGetActivityMapping(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient: fakeWhoopClient{
			t:                t,
			wantAccessToken:  "access-token",
			wantActivityV1ID: 12345,
			activityMapping:  whoop.ActivityMapping{V2ActivityID: "ecfc6a15-4661-442f-a9a4-f160dd7afae8"},
		},
	}

	result, output, err := service.getActivityMapping(context.Background(), nil, GetActivityMappingInput{ActivityV1ID: 12345})
	if err != nil {
		t.Fatalf("getActivityMapping returned error: %v", err)
	}
	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if got, want := output.V2ActivityID, "ecfc6a15-4661-442f-a9a4-f160dd7afae8"; got != want {
		t.Fatalf("V2ActivityID = %q, want %q", got, want)
	}
}

func TestGetActivityMappingReturnsAccessTokenError(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessTokenErr: fmt.Errorf("access token failed")},
		WhoopClient:  fakeWhoopClient{},
	}

	_, _, err := service.getActivityMapping(context.Background(), nil, GetActivityMappingInput{ActivityV1ID: 12345})
	if err == nil {
		t.Fatal("getActivityMapping returned nil error")
	}
	if got, want := err.Error(), "access token failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestGetActivityMappingReturnsWhoopError(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient:  fakeWhoopClient{activityMappingErr: fmt.Errorf("activity mapping failed")},
	}

	_, _, err := service.getActivityMapping(context.Background(), nil, GetActivityMappingInput{ActivityV1ID: 12345})
	if err == nil {
		t.Fatal("getActivityMapping returned nil error")
	}
	if got, want := err.Error(), "activity mapping failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}
