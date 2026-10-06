package mcpserver

import (
	"context"
	"time"

	"github.com/Tianbo-Qiu/whoopctl/internal/whoop"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetCycleInput struct {
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum number of cycle records to return, up to 25."`
	Start     string `json:"start,omitempty" jsonschema:"Start timestamp in RFC3339 format."`
	End       string `json:"end,omitempty" jsonschema:"End timestamp in RFC3339 format."`
	NextToken string `json:"next_token,omitempty" jsonschema:"Pagination token from a previous response."`
}

type GetCycleOutput struct {
	Records   []CycleRecord `json:"records" jsonschema:"Cycle records returned by WHOOP."`
	NextToken string        `json:"next_token,omitempty" jsonschema:"Pagination token for fetching the next page, if available."`
}

type GetCycleByIDInput struct {
	CycleID int64 `json:"cycle_id" jsonschema:"WHOOP physiological cycle ID."`
}

type CycleRecord struct {
	ID             int64       `json:"id" jsonschema:"WHOOP physiological cycle ID."`
	UserID         int64       `json:"user_id" jsonschema:"WHOOP user ID."`
	CreatedAt      string      `json:"created_at" jsonschema:"Record creation timestamp in RFC3339 format."`
	UpdatedAt      string      `json:"updated_at" jsonschema:"Record update timestamp in RFC3339 format."`
	Start          string      `json:"start" jsonschema:"Cycle start timestamp in RFC3339 format."`
	End            string      `json:"end,omitempty" jsonschema:"Cycle end timestamp in RFC3339 format, absent for the current active cycle."`
	TimezoneOffset string      `json:"timezone_offset" jsonschema:"User timezone offset for the cycle, such as -05:00."`
	ScoreState     string      `json:"score_state" jsonschema:"Scoring state for the cycle, such as SCORED."`
	Score          *CycleScore `json:"score,omitempty" jsonschema:"Cycle score measurements, present only when scored."`
	StepCount      *int        `json:"step_count,omitempty" jsonschema:"Total steps during the cycle, absent when no step data exists."`
}

type CycleScore struct {
	Strain           float64 `json:"strain" jsonschema:"WHOOP strain score from 0 to 21."`
	Kilojoule        float64 `json:"kilojoule" jsonschema:"Energy expended during the cycle in kilojoules."`
	AverageHeartRate int     `json:"average_heart_rate" jsonschema:"Average heart rate during the cycle in beats per minute."`
	MaxHeartRate     int     `json:"max_heart_rate" jsonschema:"Max heart rate during the cycle in beats per minute."`
}

func registerCycleTool(server *mcp.Server, service *WhoopService) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "get_cycle",
		Description: "Fetch WHOOP physiological cycle records for the authenticated user. " +
			"Supports optional limit, start, end, and next_token pagination inputs. " +
			"Returns cycle timing, score state, optional strain score, heart-rate summary, kilojoules, and optional step count.",
	}, service.getCycle)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_cycle_by_id",
		Description: "Fetch a single WHOOP physiological cycle by cycle ID.",
	}, service.getCycleByID)
}

func (s *WhoopService) getCycle(ctx context.Context, req *mcp.CallToolRequest, input GetCycleInput) (*mcp.CallToolResult, GetCycleOutput, error) {
	accessToken, err := s.TokenManager.AccessToken(ctx)
	if err != nil {
		return nil, GetCycleOutput{}, err
	}

	query, err := cycleQuery(input)
	if err != nil {
		return nil, GetCycleOutput{}, err
	}

	cycles, err := s.WhoopClient.Cycles(ctx, accessToken, query)
	if err != nil {
		return nil, GetCycleOutput{}, err
	}

	return nil, cycleOutput(cycles), nil
}

func (s *WhoopService) getCycleByID(ctx context.Context, req *mcp.CallToolRequest, input GetCycleByIDInput) (*mcp.CallToolResult, CycleRecord, error) {
	accessToken, err := s.TokenManager.AccessToken(ctx)
	if err != nil {
		return nil, CycleRecord{}, err
	}

	cycle, err := s.WhoopClient.Cycle(ctx, accessToken, input.CycleID)
	if err != nil {
		return nil, CycleRecord{}, err
	}

	return nil, cycleRecord(cycle), nil
}

func cycleQuery(input GetCycleInput) (whoop.CycleQuery, error) {
	query := whoop.CycleQuery{
		Limit:     input.Limit,
		NextToken: input.NextToken,
	}

	if input.Start != "" {
		start, err := time.Parse(time.RFC3339Nano, input.Start)
		if err != nil {
			return whoop.CycleQuery{}, err
		}
		query.Start = start
	}

	if input.End != "" {
		end, err := time.Parse(time.RFC3339Nano, input.End)
		if err != nil {
			return whoop.CycleQuery{}, err
		}
		query.End = end
	}

	return query, nil
}

func cycleOutput(collection whoop.CycleCollection) GetCycleOutput {
	out := GetCycleOutput{
		Records:   make([]CycleRecord, 0, len(collection.Records)),
		NextToken: collection.NextToken,
	}

	for _, record := range collection.Records {
		out.Records = append(out.Records, cycleRecord(record))
	}

	return out
}

func cycleRecord(record whoop.Cycle) CycleRecord {
	out := CycleRecord{
		ID:             record.ID,
		UserID:         record.UserID,
		CreatedAt:      record.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:      record.UpdatedAt.Format(time.RFC3339Nano),
		Start:          record.Start.Format(time.RFC3339Nano),
		TimezoneOffset: record.TimezoneOffset,
		ScoreState:     record.ScoreState,
		StepCount:      record.StepCount,
	}

	if record.End != nil {
		out.End = record.End.Format(time.RFC3339Nano)
	}

	if record.Score != nil {
		out.Score = &CycleScore{
			Strain:           record.Score.Strain,
			Kilojoule:        record.Score.Kilojoule,
			AverageHeartRate: record.Score.AverageHeartRate,
			MaxHeartRate:     record.Score.MaxHeartRate,
		}
	}

	return out
}
