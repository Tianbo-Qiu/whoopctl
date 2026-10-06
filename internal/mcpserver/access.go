package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type RevokeAccessRecord struct {
	Revoked bool `json:"revoked" jsonschema:"Whether WHOOP access was revoked and the local token removed."`
}

func registerAccessTool(server *mcp.Server, service *WhoopService) {
	destructive := true
	mcp.AddTool(server, &mcp.Tool{
		Name:        "revoke_user_access",
		Description: "Revoke whoopctl's access to the authenticated WHOOP user's data and remove the local token. After this, the user must run `whoopctl auth login` again before any WHOOP tool works. Only call this when the user explicitly asks to revoke access.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Revoke WHOOP access",
			DestructiveHint: &destructive,
		},
	}, service.revokeUserAccess)
}

func (s *WhoopService) revokeUserAccess(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, RevokeAccessRecord, error) {
	accessToken, err := s.TokenManager.AccessToken(ctx)
	if err != nil {
		return nil, RevokeAccessRecord{}, err
	}

	if err := s.WhoopClient.RevokeAccess(ctx, accessToken); err != nil {
		return nil, RevokeAccessRecord{}, err
	}

	if err := s.TokenManager.ClearToken(ctx); err != nil {
		return nil, RevokeAccessRecord{}, err
	}

	return nil, RevokeAccessRecord{Revoked: true}, nil
}
