package whoop

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const profileBasicPath = "/v2/user/profile/basic"

type BasicProfile struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (c *Client) BasicProfile(ctx context.Context, accessToken string) (BasicProfile, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return BasicProfile{}, fmt.Errorf("access token is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(profileBasicPath), nil)
	if err != nil {
		return BasicProfile{}, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.do(req)
	if err != nil {
		return BasicProfile{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return BasicProfile{}, apiError(resp)
	}

	var profile BasicProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return BasicProfile{}, err
	}

	return profile, nil
}
