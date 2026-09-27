package session

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Tianbo-Qiu/whoopctl/internal/auth"
	"github.com/Tianbo-Qiu/whoopctl/internal/config"
)

const RefreshSkew = 2 * time.Minute

var (
	ErrSetupRequired = errors.New("setup required")
	ErrLoginRequired = errors.New("login required")
)

type TokenManager struct {
	ConfigDir string
	Client    *http.Client
	Now       func() time.Time
}

func (m *TokenManager) AccessToken(ctx context.Context) (string, error) {
	creds, err := config.LoadCredentials(m.ConfigDir)
	if err != nil {
		return "", fmt.Errorf("%w; load credentials failed: %v", ErrSetupRequired, err)
	}

	token, err := config.LoadToken(m.ConfigDir)
	if err != nil {
		return "", fmt.Errorf("%w; load token failed: %v", ErrLoginRequired, err)
	}

	now := time.Now
	if m.Now != nil {
		now = m.Now
	}

	if now().Add(RefreshSkew).Before(token.ExpiresAt) {
		return token.AccessToken, nil
	}

	client := m.Client
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := auth.RefreshAccessToken(ctx, client, creds, token.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("%w; %v", ErrLoginRequired, err)
	}

	refreshed := config.StoredToken{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    now().Add(time.Duration(resp.ExpiresIn) * time.Second),
		Scope:        resp.Scope,
		TokenType:    resp.TokenType,
	}

	if err := config.SaveToken(m.ConfigDir, refreshed); err != nil {
		return "", err
	}

	return refreshed.AccessToken, nil
}
