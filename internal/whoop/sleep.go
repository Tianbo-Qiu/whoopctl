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

const sleepPath = "/v2/activity/sleep"

type SleepQuery struct {
	Limit     int
	Start     time.Time
	End       time.Time
	NextToken string
}

type SleepCollection struct {
	Records   []Sleep `json:"records"`
	NextToken string  `json:"next_token,omitempty"`
}

type Sleep struct {
	ID             string      `json:"id"`
	CycleID        int64       `json:"cycle_id"`
	V1ID           *int64      `json:"v1_id,omitempty"`
	UserID         int64       `json:"user_id"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	Start          time.Time   `json:"start"`
	End            time.Time   `json:"end"`
	TimezoneOffset string      `json:"timezone_offset"`
	Nap            bool        `json:"nap"`
	ScoreState     string      `json:"score_state"`
	Score          *SleepScore `json:"score,omitempty"`
}

type SleepScore struct {
	StageSummary               SleepStageSummary `json:"stage_summary"`
	SleepNeeded                SleepNeeded       `json:"sleep_needed"`
	RespiratoryRate            *float64          `json:"respiratory_rate,omitempty"`
	SleepPerformancePercentage *float64          `json:"sleep_performance_percentage,omitempty"`
	SleepConsistencyPercentage *float64          `json:"sleep_consistency_percentage,omitempty"`
	SleepEfficiencyPercentage  *float64          `json:"sleep_efficiency_percentage,omitempty"`
}

type SleepStageSummary struct {
	TotalInBedTimeMilli         int `json:"total_in_bed_time_milli"`
	TotalAwakeTimeMilli         int `json:"total_awake_time_milli"`
	TotalNoDataTimeMilli        int `json:"total_no_data_time_milli"`
	TotalLightSleepTimeMilli    int `json:"total_light_sleep_time_milli"`
	TotalSlowWaveSleepTimeMilli int `json:"total_slow_wave_sleep_time_milli"`
	TotalRemSleepTimeMilli      int `json:"total_rem_sleep_time_milli"`
	SleepCycleCount             int `json:"sleep_cycle_count"`
	DisturbanceCount            int `json:"disturbance_count"`
}

type SleepNeeded struct {
	BaselineMilli             int64 `json:"baseline_milli"`
	NeedFromSleepDebtMilli    int64 `json:"need_from_sleep_debt_milli"`
	NeedFromRecentStrainMilli int64 `json:"need_from_recent_strain_milli"`
	NeedFromRecentNapMilli    int64 `json:"need_from_recent_nap_milli"`
}

func (c *Client) Sleeps(ctx context.Context, accessToken string, query SleepQuery) (SleepCollection, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return SleepCollection{}, fmt.Errorf("access token is required")
	}

	endpoint, err := url.Parse(c.endpoint(sleepPath))
	if err != nil {
		return SleepCollection{}, err
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
		return SleepCollection{}, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.do(req)
	if err != nil {
		return SleepCollection{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return SleepCollection{}, apiError(resp)
	}

	var sleeps SleepCollection
	if err := json.NewDecoder(resp.Body).Decode(&sleeps); err != nil {
		return SleepCollection{}, err
	}

	return sleeps, nil
}

func (c *Client) Sleep(ctx context.Context, accessToken string, sleepID string) (Sleep, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return Sleep{}, fmt.Errorf("access token is required")
	}
	sleepID = strings.TrimSpace(sleepID)
	if sleepID == "" {
		return Sleep{}, fmt.Errorf("sleep id is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(fmt.Sprintf("%s/%s", sleepPath, url.PathEscape(sleepID))), nil)
	if err != nil {
		return Sleep{}, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.do(req)
	if err != nil {
		return Sleep{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Sleep{}, apiError(resp)
	}

	var sleep Sleep
	if err := json.NewDecoder(resp.Body).Decode(&sleep); err != nil {
		return Sleep{}, err
	}

	return sleep, nil
}

func (c *Client) SleepForCycle(ctx context.Context, accessToken string, cycleID int64) (Sleep, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return Sleep{}, fmt.Errorf("access token is required")
	}
	if cycleID <= 0 {
		return Sleep{}, fmt.Errorf("cycle id is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(fmt.Sprintf("/v2/cycle/%d/sleep", cycleID)), nil)
	if err != nil {
		return Sleep{}, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.do(req)
	if err != nil {
		return Sleep{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Sleep{}, apiError(resp)
	}

	var sleep Sleep
	if err := json.NewDecoder(resp.Body).Decode(&sleep); err != nil {
		return Sleep{}, err
	}

	return sleep, nil
}
