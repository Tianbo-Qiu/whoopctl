package whoop

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const cyclePath = "/v2/cycle"

type CycleQuery struct {
	Limit     int
	Start     time.Time
	End       time.Time
	NextToken string
}

type CycleCollection struct {
	Records   []Cycle `json:"records"`
	NextToken string  `json:"next_token,omitempty"`
}

type Cycle struct {
	ID             int64       `json:"id"`
	UserID         int64       `json:"user_id"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	Start          time.Time   `json:"start"`
	End            *time.Time  `json:"end,omitempty"`
	TimezoneOffset string      `json:"timezone_offset"`
	ScoreState     string      `json:"score_state"`
	Score          *CycleScore `json:"score,omitempty"`
	StepCount      *int        `json:"step_count,omitempty"`
}

type CycleScore struct {
	Strain           float64 `json:"strain"`
	Kilojoule        float64 `json:"kilojoule"`
	AverageHeartRate int     `json:"average_heart_rate"`
	MaxHeartRate     int     `json:"max_heart_rate"`
}

func (c *Client) Cycles(ctx context.Context, accessToken string, query CycleQuery) (CycleCollection, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return CycleCollection{}, fmt.Errorf("access token is required")
	}

	endpoint, err := url.Parse(c.endpoint(cyclePath))
	if err != nil {
		return CycleCollection{}, err
	}
	values := endpoint.Query()
	if query.Limit > 0 {
		values.Set("limit", fmt.Sprintf("%d", query.Limit))
	}
	if !query.Start.IsZero() {
		values.Set("start", query.Start.Format(time.RFC3339Nano))
	}
	if !query.End.IsZero() {
		values.Set("end", query.End.Format(time.RFC3339Nano))
	}
	if strings.TrimSpace(query.NextToken) != "" {
		values.Set("nextToken", strings.TrimSpace(query.NextToken))
	}
	endpoint.RawQuery = values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return CycleCollection{}, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.do(req)
	if err != nil {
		return CycleCollection{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return CycleCollection{}, apiError(resp)
	}

	var cycles CycleCollection
	if err := json.NewDecoder(resp.Body).Decode(&cycles); err != nil {
		return CycleCollection{}, err
	}

	return cycles, nil
}

func (c *Client) Cycle(ctx context.Context, accessToken string, cycleID int64) (Cycle, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return Cycle{}, fmt.Errorf("access token is required")
	}
	if cycleID <= 0 {
		return Cycle{}, fmt.Errorf("cycle id is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(fmt.Sprintf("/v2/cycle/%d", cycleID)), nil)
	if err != nil {
		return Cycle{}, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.do(req)
	if err != nil {
		return Cycle{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Cycle{}, apiError(resp)
	}

	var cycle Cycle
	if err := json.NewDecoder(resp.Body).Decode(&cycle); err != nil {
		return Cycle{}, err
	}

	return cycle, nil
}
