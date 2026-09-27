package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Tianbo-Qiu/whoopctl/internal/config"
)

const TokenEndpoint = "https://api.prod.whoop.com/oauth/oauth2/token"

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

func ExchangeAuthorizationCode(ctx context.Context, client *http.Client, creds config.Credentials, code string) (TokenResponse, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return TokenResponse{}, fmt.Errorf("authorization code is required")
	}

	creds, err := config.NormalizeCredentials(creds)
	if err != nil {
		return TokenResponse{}, err
	}

	if client == nil {
		client = http.DefaultClient
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", RedirectURI)
	form.Set("client_id", creds.ClientID)
	form.Set("client_secret", creds.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return TokenResponse{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return TokenResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return TokenResponse{}, fmt.Errorf("token exchange failed: %s", resp.Status)
	}

	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return TokenResponse{}, err
	}

	return normalizeTokenResponse(tokenResponse)
}

func normalizeTokenResponse(tokenResponse TokenResponse) (TokenResponse, error) {
	tokenResponse.AccessToken = strings.TrimSpace(tokenResponse.AccessToken)
	tokenResponse.RefreshToken = strings.TrimSpace(tokenResponse.RefreshToken)
	tokenResponse.Scope = strings.TrimSpace(tokenResponse.Scope)
	tokenResponse.TokenType = strings.TrimSpace(tokenResponse.TokenType)

	if tokenResponse.AccessToken == "" {
		return TokenResponse{}, fmt.Errorf("access token is required")
	}

	if tokenResponse.RefreshToken == "" {
		return TokenResponse{}, fmt.Errorf("refresh token is required")
	}

	if tokenResponse.ExpiresIn <= 0 {
		return TokenResponse{}, fmt.Errorf("expires in must be positive")
	}

	if tokenResponse.TokenType == "" {
		return TokenResponse{}, fmt.Errorf("token type is required")
	}
	if !strings.EqualFold(tokenResponse.TokenType, "bearer") {
		return TokenResponse{}, fmt.Errorf("unsupported token type: %s", tokenResponse.TokenType)
	}
	tokenResponse.TokenType = "bearer"

	return tokenResponse, nil
}
