package whoop

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

const userAccessPath = "/v2/user/access"

func (c *Client) RevokeAccess(ctx context.Context, accessToken string) error {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return fmt.Errorf("access token is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.endpoint(userAccessPath), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return apiError(resp)
	}

	return nil
}
