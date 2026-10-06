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

const workoutPath = "/v2/activity/workout"

type WorkoutQuery struct {
	Limit     int
	Start     time.Time
	End       time.Time
	NextToken string
}

type WorkoutCollection struct {
	Records   []Workout `json:"records"`
	NextToken string    `json:"next_token,omitempty"`
}

type Workout struct {
	ID             string        `json:"id"`
	V1ID           *int64        `json:"v1_id,omitempty"`
	UserID         int64         `json:"user_id"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	Start          time.Time     `json:"start"`
	End            time.Time     `json:"end"`
	TimezoneOffset string        `json:"timezone_offset"`
	SportName      string        `json:"sport_name"`
	ScoreState     string        `json:"score_state"`
	Score          *WorkoutScore `json:"score,omitempty"`
	SportID        *int          `json:"sport_id,omitempty"`
}

type WorkoutScore struct {
	Strain              float64       `json:"strain"`
	AverageHeartRate    int           `json:"average_heart_rate"`
	MaxHeartRate        int           `json:"max_heart_rate"`
	Kilojoule           float64       `json:"kilojoule"`
	PercentRecorded     float64       `json:"percent_recorded"`
	DistanceMeter       *float64      `json:"distance_meter,omitempty"`
	AltitudeGainMeter   *float64      `json:"altitude_gain_meter,omitempty"`
	AltitudeChangeMeter *float64      `json:"altitude_change_meter,omitempty"`
	ZoneDurations       ZoneDurations `json:"zone_durations"`
}

type ZoneDurations struct {
	ZoneZeroMilli  int64 `json:"zone_zero_milli"`
	ZoneOneMilli   int64 `json:"zone_one_milli"`
	ZoneTwoMilli   int64 `json:"zone_two_milli"`
	ZoneThreeMilli int64 `json:"zone_three_milli"`
	ZoneFourMilli  int64 `json:"zone_four_milli"`
	ZoneFiveMilli  int64 `json:"zone_five_milli"`
}

func (c *Client) Workouts(ctx context.Context, accessToken string, query WorkoutQuery) (WorkoutCollection, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return WorkoutCollection{}, fmt.Errorf("access token is required")
	}

	endpoint, err := url.Parse(c.endpoint(workoutPath))
	if err != nil {
		return WorkoutCollection{}, err
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
		return WorkoutCollection{}, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.do(req)
	if err != nil {
		return WorkoutCollection{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return WorkoutCollection{}, apiError(resp)
	}

	var workouts WorkoutCollection
	if err := json.NewDecoder(resp.Body).Decode(&workouts); err != nil {
		return WorkoutCollection{}, err
	}

	return workouts, nil
}

func (c *Client) Workout(ctx context.Context, accessToken string, workoutID string) (Workout, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return Workout{}, fmt.Errorf("access token is required")
	}
	workoutID = strings.TrimSpace(workoutID)
	if workoutID == "" {
		return Workout{}, fmt.Errorf("workout id is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(fmt.Sprintf("%s/%s", workoutPath, url.PathEscape(workoutID))), nil)
	if err != nil {
		return Workout{}, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.do(req)
	if err != nil {
		return Workout{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Workout{}, apiError(resp)
	}

	var workout Workout
	if err := json.NewDecoder(resp.Body).Decode(&workout); err != nil {
		return Workout{}, err
	}

	return workout, nil
}
