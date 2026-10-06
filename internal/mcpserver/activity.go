package mcpserver

import (
	"context"

	"github.com/Tianbo-Qiu/whoopctl/internal/whoop"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetActivityMappingInput struct {
	ActivityV1ID int64 `json:"activity_v1_id" jsonschema:"Legacy WHOOP v1 activity ID (sleep or workout)."`
}

type ActivityMappingRecord struct {
	V2ActivityID string `json:"v2_activity_id" jsonschema:"WHOOP v2 activity UUID for the given v1 activity ID."`
}

func registerActivityTool(server *mcp.Server, service *WhoopService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_activity_mapping",
		Description: "Look up the WHOOP v2 activity UUID for a legacy v1 activity ID. Use the returned UUID with get_sleep_by_id or get_workout_by_id.",
	}, service.getActivityMapping)
}

func (s *WhoopService) getActivityMapping(ctx context.Context, req *mcp.CallToolRequest, input GetActivityMappingInput) (*mcp.CallToolResult, ActivityMappingRecord, error) {
	accessToken, err := s.TokenManager.AccessToken(ctx)
	if err != nil {
		return nil, ActivityMappingRecord{}, err
	}

	mapping, err := s.WhoopClient.ActivityMapping(ctx, accessToken, input.ActivityV1ID)
	if err != nil {
		return nil, ActivityMappingRecord{}, err
	}

	return nil, activityMappingRecord(mapping), nil
}

func activityMappingRecord(mapping whoop.ActivityMapping) ActivityMappingRecord {
	return ActivityMappingRecord{
		V2ActivityID: mapping.V2ActivityID,
	}
}
