package auth

import (
	"net/url"
	"testing"
)

func TestAuthCodeURLBuildsAuthorizeURL(t *testing.T) {
	res, err := AuthCodeURL("client-id", "state", []Scope{
		ScopeOffline,
		ScopeReadRecovery,
	})
	if err != nil {
		t.Fatalf("AuthCodeURL returned error: %v", err)
	}

	parsed, err := url.Parse(res)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if got, want := parsed.Scheme, "https"; got != want {
		t.Fatalf("schema = %q, want = %q", got, want)
	}
	if got, want := parsed.Host, "api.prod.whoop.com"; got != want {
		t.Fatalf("host = %q, want = %q", got, want)
	}
	if got, want := parsed.Path, "/oauth/oauth2/auth"; got != want {
		t.Fatalf("path = %q, want = %q", got, want)
	}

	query := parsed.Query()

	if got, want := query.Get("client_id"), "client-id"; got != want {
		t.Fatalf("client_id = %q, want = %q", got, want)
	}
	if got, want := query.Get("redirect_uri"), RedirectURI; got != want {
		t.Fatalf("redirect_uri = %q, want = %q", got, want)
	}
	if got, want := query.Get("response_type"), "code"; got != want {
		t.Fatalf("response_type = %q, want = %q", got, want)
	}
	if got, want := query.Get("scope"), "offline read:recovery"; got != want {
		t.Fatalf("scope = %q, want = %q", got, want)
	}
	if got, want := query.Get("state"), "state"; got != want {
		t.Fatalf("state = %q, want = %q", got, want)
	}
}

func TestAuthCodeURLRequiresClientID(t *testing.T) {
	_, err := AuthCodeURL("", "state", []Scope{ScopeOffline})
	if err == nil {
		t.Fatal("AuthCodeURL returned nil error")
	}
	if got, want := err.Error(), "client id is required"; got != want {
		t.Fatalf("error = %q, want = %q", got, want)
	}
}

func TestAuthCodeURLRequiresState(t *testing.T) {
	_, err := AuthCodeURL("client-id", "", []Scope{ScopeOffline})
	if err == nil {
		t.Fatal("AuthCodeURL returned nil error")
	}
	if got, want := err.Error(), "state is required"; got != want {
		t.Fatalf("error = %q, want = %q", got, want)
	}
}

func TestAuthCodeURLRequiresScopes(t *testing.T) {
	_, err := AuthCodeURL("client-id", "state", []Scope{})
	if err == nil {
		t.Fatal("AuthCodeURL returned nil error")
	}
	if got, want := err.Error(), "at least one scope is required"; got != want {
		t.Fatalf("error = %q, want = %q", got, want)
	}
}
