package mcpserver

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestRevokeUserAccess(t *testing.T) {
	var calls []string
	service := &WhoopService{
		TokenManager: fakeTokenManager{
			accessToken: "access-token",
			onClearToken: func() {
				calls = append(calls, "clear")
			},
		},
		WhoopClient: fakeWhoopClient{
			t:               t,
			wantAccessToken: "access-token",
			onRevoke: func() {
				calls = append(calls, "revoke")
			},
		},
	}

	result, output, err := service.revokeUserAccess(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("revokeUserAccess returned error: %v", err)
	}
	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !output.Revoked {
		t.Fatal("Revoked = false, want true")
	}
	if got, want := strings.Join(calls, ","), "revoke,clear"; got != want {
		t.Fatalf("calls = %q, want %q", got, want)
	}
}

func TestRevokeUserAccessReturnsAccessTokenError(t *testing.T) {
	service := &WhoopService{
		TokenManager: fakeTokenManager{accessTokenErr: fmt.Errorf("access token failed")},
		WhoopClient:  fakeWhoopClient{},
	}

	_, _, err := service.revokeUserAccess(context.Background(), nil, struct{}{})
	if err == nil {
		t.Fatal("revokeUserAccess returned nil error")
	}
	if got, want := err.Error(), "access token failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestRevokeUserAccessKeepsTokenWhenRevokeFails(t *testing.T) {
	cleared := false
	service := &WhoopService{
		TokenManager: fakeTokenManager{
			accessToken: "access-token",
			onClearToken: func() {
				cleared = true
			},
		},
		WhoopClient: fakeWhoopClient{revokeErr: fmt.Errorf("revoke failed")},
	}

	_, _, err := service.revokeUserAccess(context.Background(), nil, struct{}{})
	if err == nil {
		t.Fatal("revokeUserAccess returned nil error")
	}
	if got, want := err.Error(), "revoke failed"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
	if cleared {
		t.Fatal("ClearToken was called after revoke failed")
	}
}
