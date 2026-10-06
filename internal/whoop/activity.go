package whoop

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ActivityMapping struct {
	V2ActivityID string `json:"v2_activity_id"`
}

func (c *Client) ActivityMapping(ctx context.Context, accessToken string, activityV1ID int64) (ActivityMapping, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return ActivityMapping{}, fmt.Errorf("access token is required")
	}
	if activityV1ID <= 0 {
		return ActivityMapping{}, fmt.Errorf("activity v1 id is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(fmt.Sprintf("/v1/activity-mapping/%d", activityV1ID)), nil)
	if err != nil {
		return ActivityMapping{}, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.do(req)
	if err != nil {
		return ActivityMapping{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return ActivityMapping{}, apiError(resp)
	}

	var mapping ActivityMapping
	if err := json.NewDecoder(resp.Body).Decode(&mapping); err != nil {
		return ActivityMapping{}, err
	}

	return mapping, nil
}
