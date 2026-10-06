package mcpserver

import (
	"context"

	"github.com/Tianbo-Qiu/whoopctl/internal/whoop"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type BodyMeasurementRecord struct {
	HeightMeter    float64 `json:"height_meter" jsonschema:"User height in meters."`
	WeightKilogram float64 `json:"weight_kilogram" jsonschema:"User weight in kilograms."`
	MaxHeartRate   int     `json:"max_heart_rate" jsonschema:"WHOOP-calculated max heart rate."`
}

func registerBodyTool(server *mcp.Server, service *WhoopService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_body_measurement",
		Description: "Fetch the authenticated WHOOP user's body measurements, including height, weight, and max heart rate.",
	}, service.getBodyMeasurement)
}

func (s *WhoopService) getBodyMeasurement(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, BodyMeasurementRecord, error) {
	accessToken, err := s.TokenManager.AccessToken(ctx)
	if err != nil {
		return nil, BodyMeasurementRecord{}, err
	}

	measurement, err := s.WhoopClient.BodyMeasurement(ctx, accessToken)
	if err != nil {
		return nil, BodyMeasurementRecord{}, err
	}

	return nil, bodyMeasurementRecord(measurement), nil
}

func bodyMeasurementRecord(measurement whoop.BodyMeasurement) BodyMeasurementRecord {
	return BodyMeasurementRecord{
		HeightMeter:    measurement.HeightMeter,
		WeightKilogram: measurement.WeightKilogram,
		MaxHeartRate:   measurement.MaxHeartRate,
	}
}
