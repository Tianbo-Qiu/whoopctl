package mcpserver

import (
	"context"

	"github.com/Tianbo-Qiu/whoopctl/internal/whoop"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type BasicProfileRecord struct {
	UserID    int64  `json:"user_id" jsonschema:"WHOOP user ID."`
	Email     string `json:"email" jsonschema:"Email address for the WHOOP user."`
	FirstName string `json:"first_name" jsonschema:"First name for the WHOOP user."`
	LastName  string `json:"last_name" jsonschema:"Last name for the WHOOP user."`
}

func registerProfileTool(server *mcp.Server, service *WhoopService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_profile_basic",
		Description: "Fetch the authenticated WHOOP user's basic profile, including user ID, email, first name, and last name.",
	}, service.getProfileBasic)
}

func (s *WhoopService) getProfileBasic(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, BasicProfileRecord, error) {
	accessToken, err := s.TokenManager.AccessToken(ctx)
	if err != nil {
		return nil, BasicProfileRecord{}, err
	}

	profile, err := s.WhoopClient.BasicProfile(ctx, accessToken)
	if err != nil {
		return nil, BasicProfileRecord{}, err
	}

	return nil, basicProfileRecord(profile), nil
}

func basicProfileRecord(profile whoop.BasicProfile) BasicProfileRecord {
	return BasicProfileRecord{
		UserID:    profile.UserID,
		Email:     profile.Email,
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
	}
}
