package mcpserver

import (
	"context"
	"time"

	"github.com/Tianbo-Qiu/whoopctl/internal/whoop"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetSleepInput struct {
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum number of sleep records to return, up to 25."`
	Start     string `json:"start,omitempty" jsonschema:"Start timestamp in RFC3339 format."`
	End       string `json:"end,omitempty" jsonschema:"End timestamp in RFC3339 format."`
	NextToken string `json:"next_token,omitempty" jsonschema:"Pagination token from a previous response."`
}

type GetSleepOutput struct {
	Records   []SleepRecord `json:"records" jsonschema:"Sleep records returned by WHOOP."`
	NextToken string        `json:"next_token,omitempty" jsonschema:"Pagination token for fetching the next page, if available."`
}

type GetSleepByIDInput struct {
	SleepID string `json:"sleep_id" jsonschema:"WHOOP sleep activity ID."`
}

type GetSleepForCycleInput struct {
	CycleID int64 `json:"cycle_id" jsonschema:"WHOOP physiological cycle ID."`
}

type SleepRecord struct {
	ID             string      `json:"id" jsonschema:"WHOOP sleep activity ID."`
	CycleID        int64       `json:"cycle_id" jsonschema:"WHOOP physiological cycle ID associated with this sleep."`
	V1ID           *int64      `json:"v1_id,omitempty" jsonschema:"Legacy WHOOP v1 activity ID, present when available."`
	UserID         int64       `json:"user_id" jsonschema:"WHOOP user ID."`
	CreatedAt      string      `json:"created_at" jsonschema:"Record creation timestamp in RFC3339 format."`
	UpdatedAt      string      `json:"updated_at" jsonschema:"Record update timestamp in RFC3339 format."`
	Start          string      `json:"start" jsonschema:"Sleep start timestamp in RFC3339 format."`
	End            string      `json:"end" jsonschema:"Sleep end timestamp in RFC3339 format."`
	TimezoneOffset string      `json:"timezone_offset" jsonschema:"User timezone offset for the sleep, such as -05:00."`
	Nap            bool        `json:"nap" jsonschema:"Whether this sleep activity was a nap."`
	ScoreState     string      `json:"score_state" jsonschema:"Scoring state for the sleep, such as SCORED."`
	Score          *SleepScore `json:"score,omitempty" jsonschema:"Sleep score measurements, present only when scored."`
}

type SleepScore struct {
	StageSummary               SleepStageSummary `json:"stage_summary" jsonschema:"Summary of sleep stages."`
	SleepNeeded                SleepNeeded       `json:"sleep_needed" jsonschema:"Breakdown of sleep need before this sleep."`
	RespiratoryRate            *float64          `json:"respiratory_rate,omitempty" jsonschema:"Respiratory rate during sleep."`
	SleepPerformancePercentage *float64          `json:"sleep_performance_percentage,omitempty" jsonschema:"Sleep performance percentage."`
	SleepConsistencyPercentage *float64          `json:"sleep_consistency_percentage,omitempty" jsonschema:"Sleep consistency percentage."`
	SleepEfficiencyPercentage  *float64          `json:"sleep_efficiency_percentage,omitempty" jsonschema:"Sleep efficiency percentage."`
}

type SleepStageSummary struct {
	TotalInBedTimeMilli         int `json:"total_in_bed_time_milli" jsonschema:"Total time in bed, in milliseconds."`
	TotalAwakeTimeMilli         int `json:"total_awake_time_milli" jsonschema:"Total awake time, in milliseconds."`
	TotalNoDataTimeMilli        int `json:"total_no_data_time_milli" jsonschema:"Total no-data time, in milliseconds."`
	TotalLightSleepTimeMilli    int `json:"total_light_sleep_time_milli" jsonschema:"Total light sleep time, in milliseconds."`
	TotalSlowWaveSleepTimeMilli int `json:"total_slow_wave_sleep_time_milli" jsonschema:"Total slow-wave sleep time, in milliseconds."`
	TotalRemSleepTimeMilli      int `json:"total_rem_sleep_time_milli" jsonschema:"Total REM sleep time, in milliseconds."`
	SleepCycleCount             int `json:"sleep_cycle_count" jsonschema:"Number of sleep cycles."`
	DisturbanceCount            int `json:"disturbance_count" jsonschema:"Number of sleep disturbances."`
}

type SleepNeeded struct {
	BaselineMilli             int64 `json:"baseline_milli" jsonschema:"Baseline sleep need, in milliseconds."`
	NeedFromSleepDebtMilli    int64 `json:"need_from_sleep_debt_milli" jsonschema:"Sleep need from sleep debt, in milliseconds."`
	NeedFromRecentStrainMilli int64 `json:"need_from_recent_strain_milli" jsonschema:"Sleep need from recent strain, in milliseconds."`
	NeedFromRecentNapMilli    int64 `json:"need_from_recent_nap_milli" jsonschema:"Sleep need reduction from recent nap activity, in milliseconds."`
}

func registerSleepTool(server *mcp.Server, service *WhoopService) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "get_sleep",
		Description: "Fetch WHOOP sleep records for the authenticated user. " +
			"Supports optional limit, start, end, and next_token pagination inputs. " +
			"Returns sleep timing, nap flag, score state, and scored sleep stage/need metrics when available.",
	}, service.getSleep)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_sleep_by_id",
		Description: "Fetch a single WHOOP sleep activity by sleep ID.",
	}, service.getSleepByID)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_sleep_for_cycle",
		Description: "Fetch the WHOOP sleep activity for a specific physiological cycle ID.",
	}, service.getSleepForCycle)
}

func (s *WhoopService) getSleep(ctx context.Context, req *mcp.CallToolRequest, input GetSleepInput) (*mcp.CallToolResult, GetSleepOutput, error) {
	accessToken, err := s.TokenManager.AccessToken(ctx)
	if err != nil {
		return nil, GetSleepOutput{}, err
	}

	query, err := sleepQuery(input)
	if err != nil {
		return nil, GetSleepOutput{}, err
	}

	sleeps, err := s.WhoopClient.Sleeps(ctx, accessToken, query)
	if err != nil {
		return nil, GetSleepOutput{}, err
	}

	return nil, sleepOutput(sleeps), nil
}

func (s *WhoopService) getSleepByID(ctx context.Context, req *mcp.CallToolRequest, input GetSleepByIDInput) (*mcp.CallToolResult, SleepRecord, error) {
	accessToken, err := s.TokenManager.AccessToken(ctx)
	if err != nil {
		return nil, SleepRecord{}, err
	}

	sleep, err := s.WhoopClient.Sleep(ctx, accessToken, input.SleepID)
	if err != nil {
		return nil, SleepRecord{}, err
	}

	return nil, sleepRecord(sleep), nil
}

func (s *WhoopService) getSleepForCycle(ctx context.Context, req *mcp.CallToolRequest, input GetSleepForCycleInput) (*mcp.CallToolResult, SleepRecord, error) {
	accessToken, err := s.TokenManager.AccessToken(ctx)
	if err != nil {
		return nil, SleepRecord{}, err
	}

	sleep, err := s.WhoopClient.SleepForCycle(ctx, accessToken, input.CycleID)
	if err != nil {
		return nil, SleepRecord{}, err
	}

	return nil, sleepRecord(sleep), nil
}

func sleepQuery(input GetSleepInput) (whoop.SleepQuery, error) {
	query := whoop.SleepQuery{
		Limit:     input.Limit,
		NextToken: input.NextToken,
	}

	if input.Start != "" {
		start, err := time.Parse(time.RFC3339Nano, input.Start)
		if err != nil {
			return whoop.SleepQuery{}, err
		}
		query.Start = start
	}

	if input.End != "" {
		end, err := time.Parse(time.RFC3339Nano, input.End)
		if err != nil {
			return whoop.SleepQuery{}, err
		}
		query.End = end
	}

	return query, nil
}

func sleepOutput(collection whoop.SleepCollection) GetSleepOutput {
	out := GetSleepOutput{
		Records:   make([]SleepRecord, 0, len(collection.Records)),
		NextToken: collection.NextToken,
	}

	for _, record := range collection.Records {
		out.Records = append(out.Records, sleepRecord(record))
	}

	return out
}

func sleepRecord(record whoop.Sleep) SleepRecord {
	out := SleepRecord{
		ID:             record.ID,
		CycleID:        record.CycleID,
		V1ID:           record.V1ID,
		UserID:         record.UserID,
		CreatedAt:      record.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:      record.UpdatedAt.Format(time.RFC3339Nano),
		Start:          record.Start.Format(time.RFC3339Nano),
		End:            record.End.Format(time.RFC3339Nano),
		TimezoneOffset: record.TimezoneOffset,
		Nap:            record.Nap,
		ScoreState:     record.ScoreState,
	}

	if record.Score != nil {
		out.Score = &SleepScore{
			StageSummary: SleepStageSummary{
				TotalInBedTimeMilli:         record.Score.StageSummary.TotalInBedTimeMilli,
				TotalAwakeTimeMilli:         record.Score.StageSummary.TotalAwakeTimeMilli,
				TotalNoDataTimeMilli:        record.Score.StageSummary.TotalNoDataTimeMilli,
				TotalLightSleepTimeMilli:    record.Score.StageSummary.TotalLightSleepTimeMilli,
				TotalSlowWaveSleepTimeMilli: record.Score.StageSummary.TotalSlowWaveSleepTimeMilli,
				TotalRemSleepTimeMilli:      record.Score.StageSummary.TotalRemSleepTimeMilli,
				SleepCycleCount:             record.Score.StageSummary.SleepCycleCount,
				DisturbanceCount:            record.Score.StageSummary.DisturbanceCount,
			},
			SleepNeeded: SleepNeeded{
				BaselineMilli:             record.Score.SleepNeeded.BaselineMilli,
				NeedFromSleepDebtMilli:    record.Score.SleepNeeded.NeedFromSleepDebtMilli,
				NeedFromRecentStrainMilli: record.Score.SleepNeeded.NeedFromRecentStrainMilli,
				NeedFromRecentNapMilli:    record.Score.SleepNeeded.NeedFromRecentNapMilli,
			},
			RespiratoryRate:            record.Score.RespiratoryRate,
			SleepPerformancePercentage: record.Score.SleepPerformancePercentage,
			SleepConsistencyPercentage: record.Score.SleepConsistencyPercentage,
			SleepEfficiencyPercentage:  record.Score.SleepEfficiencyPercentage,
		}
	}

	return out
}
