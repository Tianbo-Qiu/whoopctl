package mcpserver

import (
	"context"
	"fmt"
	"testing"

	"github.com/Tianbo-Qiu/whoopctl/internal/whoop"
)

func TestGetProfileBasic(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient: fakeWhoopClient{
			t:               t,
			wantAccessToken: "access-token",
			profile: whoop.BasicProfile{
				UserID:    10129,
				Email:     "user@example.test",
				FirstName: "Example",
				LastName:  "User",
			},
		},
	}

	result, output, err := service.getProfileBasic(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("getProfileBasic returned error: %v", err)
	}
	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if got, want := output.Email, "user@example.test"; got != want {
		t.Fatalf("Email = %q, want %q", got, want)
	}
}

func TestGetProfileBasicReturnsAccessTokenError(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessTokenErr: fmt.Errorf("access token failed")},
		WhoopClient:  fakeWhoopClient{},
	}

	_, _, err := service.getProfileBasic(context.Background(), nil, struct{}{})
	if err == nil {
		t.Fatal("getProfileBasic returned nil error")
	}
	if got, want := err.Error(), "access token failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestGetProfileBasicReturnsWhoopError(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessToken: "access-token"},
		WhoopClient:  fakeWhoopClient{profileErr: fmt.Errorf("profile failed")},
	}

	_, _, err := service.getProfileBasic(context.Background(), nil, struct{}{})
	if err == nil {
		t.Fatal("getProfileBasic returned nil error")
	}
	if got, want := err.Error(), "profile failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}
