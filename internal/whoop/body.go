package whoop

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const bodyMeasurementPath = "/v2/user/measurement/body"

type BodyMeasurement struct {
	HeightMeter    float64 `json:"height_meter"`
	WeightKilogram float64 `json:"weight_kilogram"`
	MaxHeartRate   int     `json:"max_heart_rate"`
}

func (c *Client) BodyMeasurement(ctx context.Context, accessToken string) (BodyMeasurement, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return BodyMeasurement{}, fmt.Errorf("access token is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(bodyMeasurementPath), nil)
	if err != nil {
		return BodyMeasurement{}, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.do(req)
	if err != nil {
		return BodyMeasurement{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return BodyMeasurement{}, apiError(resp)
	}

	var measurement BodyMeasurement
	if err := json.NewDecoder(resp.Body).Decode(&measurement); err != nil {
		return BodyMeasurement{}, err
	}

	return measurement, nil
}
