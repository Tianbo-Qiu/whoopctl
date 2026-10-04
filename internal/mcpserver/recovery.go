package mcpserver

import (
	"context"
	"time"

	"github.com/Tianbo-Qiu/whoopctl/internal/whoop"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetRecoveryInput struct {
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum number of recovery records to return, up to 25."`
	Start     string `json:"start,omitempty" jsonschema:"Start timestamp in RFC3339 format."`
	End       string `json:"end,omitempty" jsonschema:"End timestamp in RFC3339 format."`
	NextToken string `json:"next_token,omitempty" jsonschema:"Pagination token from a previous response."`
}

type GetRecoveryOutput struct {
	Records   []RecoveryRecord `json:"records" jsonschema:"Recovery records returned by WHOOP."`
	NextToken string           `json:"next_token,omitempty" jsonschema:"Pagination token for fetching the next page, if available."`
}

type RecoveryRecord struct {
	CycleID    int64          `json:"cycle_id" jsonschema:"WHOOP physiological cycle ID associated with this recovery record."`
	SleepID    string         `json:"sleep_id" jsonschema:"WHOOP sleep ID associated with this recovery record."`
	UserID     int64          `json:"user_id" jsonschema:"WHOOP user ID."`
	CreatedAt  string         `json:"created_at" jsonschema:"Record creation timestamp in RFC3339 format."`
	UpdatedAt  string         `json:"updated_at" jsonschema:"Record update timestamp in RFC3339 format."`
	ScoreState string         `json:"score_state" jsonschema:"Scoring state for the recovery record, such as SCORED."`
	Score      *RecoveryScore `json:"score,omitempty" jsonschema:"Recovery score measurements, present only when scored."`
}

type RecoveryScore struct {
	UserCalibrating  bool     `json:"user_calibrating" jsonschema:"Whether WHOOP is still calibrating recovery for this user."`
	RecoveryScore    float64  `json:"recovery_score" jsonschema:"Recovery score from 0 to 100."`
	RestingHeartRate float64  `json:"resting_heart_rate" jsonschema:"Resting heart rate in beats per minute."`
	HrvRmssdMilli    float64  `json:"hrv_rmssd_milli" jsonschema:"Heart rate variability using RMSSD, in milliseconds."`
	Spo2Percentage   *float64 `json:"spo2_percentage,omitempty" jsonschema:"Blood oxygen saturation percentage, present only for supported devices/data."`
	SkinTempCelsius  *float64 `json:"skin_temp_celsius,omitempty" jsonschema:"Skin temperature in degrees Celsius, present only for supported devices/data."`
}

func registerRecoveryTool(server *mcp.Server, service *WhoopService) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "get_recovery",
		Description: "Fetch WHOOP recovery records for the authenticated user. " +
			"Supports optional limit, start, end, and next_token pagination inputs. " +
			"Returns recovery score measurements including resting heart rate in beats per minute, HRV RMSSD in milliseconds, optional SpO2 percentage, and optional skin temperature in Celsius.",
	}, service.getRecovery)
}

func (s *WhoopService) getRecovery(ctx context.Context, req *mcp.CallToolRequest, input GetRecoveryInput) (*mcp.CallToolResult, GetRecoveryOutput, error) {
	accessToken, err := s.TokenManager.AccessToken(ctx)
	if err != nil {
		return nil, GetRecoveryOutput{}, err
	}

	query, err := recoveryQuery(input)
	if err != nil {
		return nil, GetRecoveryOutput{}, err
	}

	recovery, err := s.WhoopClient.Recovery(ctx, accessToken, query)
	if err != nil {
		return nil, GetRecoveryOutput{}, err
	}

	return nil, recoveryOutput(recovery), nil
}

func recoveryQuery(input GetRecoveryInput) (whoop.RecoveryQuery, error) {
	query := whoop.RecoveryQuery{
		Limit:     input.Limit,
		NextToken: input.NextToken,
	}

	if input.Start != "" {
		start, err := time.Parse(time.RFC3339Nano, input.Start)
		if err != nil {
			return whoop.RecoveryQuery{}, err
		}
		query.Start = start
	}

	if input.End != "" {
		end, err := time.Parse(time.RFC3339Nano, input.End)
		if err != nil {
			return whoop.RecoveryQuery{}, err
		}
		query.End = end
	}

	return query, nil
}

func recoveryOutput(collection whoop.RecoveryCollection) GetRecoveryOutput {
	out := GetRecoveryOutput{
		Records:   make([]RecoveryRecord, 0, len(collection.Records)),
		NextToken: collection.NextToken,
	}

	for _, record := range collection.Records {
		item := RecoveryRecord{
			CycleID:    record.CycleID,
			SleepID:    record.SleepID,
			UserID:     record.UserID,
			CreatedAt:  record.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:  record.UpdatedAt.Format(time.RFC3339Nano),
			ScoreState: record.ScoreState,
		}

		if record.Score != nil {
			item.Score = &RecoveryScore{
				UserCalibrating:  record.Score.UserCalibrating,
				RecoveryScore:    record.Score.RecoveryScore,
				RestingHeartRate: record.Score.RestingHeartRate,
				HrvRmssdMilli:    record.Score.HrvRmssdMilli,
				Spo2Percentage:   record.Score.Spo2Percentage,
				SkinTempCelsius:  record.Score.SkinTempCelsius,
			}
		}

		out.Records = append(out.Records, item)
	}

	return out
}
