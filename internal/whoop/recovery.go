package whoop

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const recoveryPath = "/v2/recovery"

type RecoveryCollection struct {
	Records   []Recovery `json:"records"`
	NextToken string     `json:"next_token,omitempty"`
}

type Recovery struct {
	CycleID    int64          `json:"cycle_id"`
	SleepID    string         `json:"sleep_id"`
	UserID     int64          `json:"user_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	ScoreState string         `json:"score_state"`
	Score      *RecoveryScore `json:"score,omitempty"`
}

type RecoveryScore struct {
	UserCalibrating  bool     `json:"user_calibrating"`
	RecoveryScore    float64  `json:"recovery_score"`
	RestingHeartRate float64  `json:"resting_heart_rate"`
	HrvRmssdMilli    float64  `json:"hrv_rmssd_milli"`
	Spo2Percentage   *float64 `json:"spo2_percentage,omitempty"`
	SkinTempCelsius  *float64 `json:"skin_temp_celsius,omitempty"`
}

func (c *Client) Recovery(ctx context.Context, accessToken string) (RecoveryCollection, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return RecoveryCollection{}, fmt.Errorf("access token is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(recoveryPath), nil)
	if err != nil {
		return RecoveryCollection{}, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.do(req)
	if err != nil {
		return RecoveryCollection{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return RecoveryCollection{}, apiError(resp)
	}

	var recovery RecoveryCollection
	if err := json.NewDecoder(resp.Body).Decode(&recovery); err != nil {
		return RecoveryCollection{}, err
	}

	return recovery, nil
}
