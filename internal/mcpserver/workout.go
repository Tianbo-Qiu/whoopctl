package mcpserver

import (
	"context"
	"time"

	"github.com/Tianbo-Qiu/whoopctl/internal/whoop"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetWorkoutInput struct {
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum number of workout records to return, up to 25."`
	Start     string `json:"start,omitempty" jsonschema:"Start timestamp in RFC3339 format."`
	End       string `json:"end,omitempty" jsonschema:"End timestamp in RFC3339 format."`
	NextToken string `json:"next_token,omitempty" jsonschema:"Pagination token from a previous response."`
}

type GetWorkoutOutput struct {
	Records   []WorkoutRecord `json:"records" jsonschema:"Workout records returned by WHOOP."`
	NextToken string          `json:"next_token,omitempty" jsonschema:"Pagination token for fetching the next page, if available."`
}

type GetWorkoutByIDInput struct {
	WorkoutID string `json:"workout_id" jsonschema:"WHOOP workout activity ID."`
}

type WorkoutRecord struct {
	ID             string        `json:"id" jsonschema:"WHOOP workout activity ID."`
	V1ID           *int64        `json:"v1_id,omitempty" jsonschema:"Legacy WHOOP v1 activity ID, present when available."`
	UserID         int64         `json:"user_id" jsonschema:"WHOOP user ID."`
	CreatedAt      string        `json:"created_at" jsonschema:"Record creation timestamp in RFC3339 format."`
	UpdatedAt      string        `json:"updated_at" jsonschema:"Record update timestamp in RFC3339 format."`
	Start          string        `json:"start" jsonschema:"Workout start timestamp in RFC3339 format."`
	End            string        `json:"end" jsonschema:"Workout end timestamp in RFC3339 format."`
	TimezoneOffset string        `json:"timezone_offset" jsonschema:"User timezone offset for the workout, such as -05:00."`
	SportName      string        `json:"sport_name" jsonschema:"WHOOP sport name for this workout."`
	ScoreState     string        `json:"score_state" jsonschema:"Scoring state for the workout, such as SCORED."`
	Score          *WorkoutScore `json:"score,omitempty" jsonschema:"Workout score measurements, present only when scored."`
	SportID        *int          `json:"sport_id,omitempty" jsonschema:"Legacy WHOOP sport ID, present when available."`
}

type WorkoutScore struct {
	Strain              float64       `json:"strain" jsonschema:"WHOOP strain score from 0 to 21."`
	AverageHeartRate    int           `json:"average_heart_rate" jsonschema:"Average heart rate during the workout in beats per minute."`
	MaxHeartRate        int           `json:"max_heart_rate" jsonschema:"Max heart rate during the workout in beats per minute."`
	Kilojoule           float64       `json:"kilojoule" jsonschema:"Energy expended during the workout in kilojoules."`
	PercentRecorded     float64       `json:"percent_recorded" jsonschema:"Percentage of heart-rate data recorded during the workout."`
	DistanceMeter       *float64      `json:"distance_meter,omitempty" jsonschema:"Distance traveled during the workout, if available."`
	AltitudeGainMeter   *float64      `json:"altitude_gain_meter,omitempty" jsonschema:"Altitude gained during the workout, if available."`
	AltitudeChangeMeter *float64      `json:"altitude_change_meter,omitempty" jsonschema:"Altitude difference from start to end, if available."`
	ZoneDurations       ZoneDurations `json:"zone_durations" jsonschema:"Heart-rate zone durations during the workout."`
}

type ZoneDurations struct {
	ZoneZeroMilli  int64 `json:"zone_zero_milli" jsonschema:"Duration in heart-rate zone 0, in milliseconds."`
	ZoneOneMilli   int64 `json:"zone_one_milli" jsonschema:"Duration in heart-rate zone 1, in milliseconds."`
	ZoneTwoMilli   int64 `json:"zone_two_milli" jsonschema:"Duration in heart-rate zone 2, in milliseconds."`
	ZoneThreeMilli int64 `json:"zone_three_milli" jsonschema:"Duration in heart-rate zone 3, in milliseconds."`
	ZoneFourMilli  int64 `json:"zone_four_milli" jsonschema:"Duration in heart-rate zone 4, in milliseconds."`
	ZoneFiveMilli  int64 `json:"zone_five_milli" jsonschema:"Duration in heart-rate zone 5, in milliseconds."`
}

func registerWorkoutTool(server *mcp.Server, service *WhoopService) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "get_workout",
		Description: "Fetch WHOOP workout records for the authenticated user. " +
			"Supports optional limit, start, end, and next_token pagination inputs. " +
			"Returns workout timing, sport, score state, and scored strain/heart-rate metrics when available.",
	}, service.getWorkout)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_workout_by_id",
		Description: "Fetch a single WHOOP workout activity by workout ID.",
	}, service.getWorkoutByID)
}

func (s *WhoopService) getWorkout(ctx context.Context, req *mcp.CallToolRequest, input GetWorkoutInput) (*mcp.CallToolResult, GetWorkoutOutput, error) {
	accessToken, err := s.TokenManager.AccessToken(ctx)
	if err != nil {
		return nil, GetWorkoutOutput{}, err
	}

	query, err := workoutQuery(input)
	if err != nil {
		return nil, GetWorkoutOutput{}, err
	}

	workouts, err := s.WhoopClient.Workouts(ctx, accessToken, query)
	if err != nil {
		return nil, GetWorkoutOutput{}, err
	}

	return nil, workoutOutput(workouts), nil
}

func (s *WhoopService) getWorkoutByID(ctx context.Context, req *mcp.CallToolRequest, input GetWorkoutByIDInput) (*mcp.CallToolResult, WorkoutRecord, error) {
	accessToken, err := s.TokenManager.AccessToken(ctx)
	if err != nil {
		return nil, WorkoutRecord{}, err
	}

	workout, err := s.WhoopClient.Workout(ctx, accessToken, input.WorkoutID)
	if err != nil {
		return nil, WorkoutRecord{}, err
	}

	return nil, workoutRecord(workout), nil
}

func workoutQuery(input GetWorkoutInput) (whoop.WorkoutQuery, error) {
	query := whoop.WorkoutQuery{
		Limit:     input.Limit,
		NextToken: input.NextToken,
	}

	if input.Start != "" {
		start, err := time.Parse(time.RFC3339Nano, input.Start)
		if err != nil {
			return whoop.WorkoutQuery{}, err
		}
		query.Start = start
	}

	if input.End != "" {
		end, err := time.Parse(time.RFC3339Nano, input.End)
		if err != nil {
			return whoop.WorkoutQuery{}, err
		}
		query.End = end
	}

	return query, nil
}

func workoutOutput(collection whoop.WorkoutCollection) GetWorkoutOutput {
	out := GetWorkoutOutput{
		Records:   make([]WorkoutRecord, 0, len(collection.Records)),
		NextToken: collection.NextToken,
	}

	for _, record := range collection.Records {
		out.Records = append(out.Records, workoutRecord(record))
	}

	return out
}

func workoutRecord(record whoop.Workout) WorkoutRecord {
	out := WorkoutRecord{
		ID:             record.ID,
		V1ID:           record.V1ID,
		UserID:         record.UserID,
		CreatedAt:      record.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:      record.UpdatedAt.Format(time.RFC3339Nano),
		Start:          record.Start.Format(time.RFC3339Nano),
		End:            record.End.Format(time.RFC3339Nano),
		TimezoneOffset: record.TimezoneOffset,
		SportName:      record.SportName,
		ScoreState:     record.ScoreState,
		SportID:        record.SportID,
	}

	if record.Score != nil {
		out.Score = &WorkoutScore{
			Strain:              record.Score.Strain,
			AverageHeartRate:    record.Score.AverageHeartRate,
			MaxHeartRate:        record.Score.MaxHeartRate,
			Kilojoule:           record.Score.Kilojoule,
			PercentRecorded:     record.Score.PercentRecorded,
			DistanceMeter:       record.Score.DistanceMeter,
			AltitudeGainMeter:   record.Score.AltitudeGainMeter,
			AltitudeChangeMeter: record.Score.AltitudeChangeMeter,
			ZoneDurations: ZoneDurations{
				ZoneZeroMilli:  record.Score.ZoneDurations.ZoneZeroMilli,
				ZoneOneMilli:   record.Score.ZoneDurations.ZoneOneMilli,
				ZoneTwoMilli:   record.Score.ZoneDurations.ZoneTwoMilli,
				ZoneThreeMilli: record.Score.ZoneDurations.ZoneThreeMilli,
				ZoneFourMilli:  record.Score.ZoneDurations.ZoneFourMilli,
				ZoneFiveMilli:  record.Score.ZoneDurations.ZoneFiveMilli,
			},
		}
	}

	return out
}
