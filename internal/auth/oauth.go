package auth

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	AuthorizeURL = "https://api.prod.whoop.com/oauth/oauth2/auth"
	RedirectURI  = "http://127.0.0.1:1061/callback"
)

type Scope string

const (
	ScopeOffline         Scope = "offline"
	ScopeReadRecovery    Scope = "read:recovery"
	ScopeReadCycles      Scope = "read:cycles"
	ScopeReadSleep       Scope = "read:sleep"
	ScopeReadWorkout     Scope = "read:workout"
	ScopeReadProfile     Scope = "read:profile"
	ScopeReadBodyMeasure Scope = "read:body_measurement"
)

var DefaultScopes = []Scope{
	ScopeOffline,
	ScopeReadRecovery,
	ScopeReadCycles,
	ScopeReadSleep,
	ScopeReadWorkout,
	ScopeReadProfile,
	ScopeReadBodyMeasure,
}

func AuthCodeURL(clientID, state string, scopes []Scope) (string, error) {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return "", fmt.Errorf("client id is required")
	}
	if state == "" {
		return "", fmt.Errorf("state is required")
	}
	if len(scopes) == 0 {
		return "", fmt.Errorf("at least one scope is required")
	}

	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("redirect_uri", RedirectURI)
	values.Set("response_type", "code")
	values.Set("scope", joinScopes(scopes))
	values.Set("state", state)

	return AuthorizeURL + "?" + values.Encode(), nil
}

func joinScopes(scopes []Scope) string {
	values := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		values = append(values, string(scope))
	}
	return strings.Join(values, " ")
}
