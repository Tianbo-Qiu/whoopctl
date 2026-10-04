package mcpserver

import (
	"context"

	"github.com/Tianbo-Qiu/whoopctl/internal/session"
	"github.com/Tianbo-Qiu/whoopctl/internal/whoop"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type tokenManager interface {
	AccessToken(ctx context.Context) (string, error)
}

type whoopClient interface {
	Recovery(ctx context.Context, accessToken string, query whoop.RecoveryQuery) (whoop.RecoveryCollection, error)
}

type WhoopService struct {
	TokenManager tokenManager
	WhoopClient  whoopClient
}

func New(configDir string) *mcp.Server {
	service := &WhoopService{
		TokenManager: &session.TokenManager{ConfigDir: configDir},
		WhoopClient:  &whoop.Client{},
	}

	return NewWithService(service)
}

func NewWithService(service *WhoopService) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "whoopctl",
		Version: "dev",
	}, nil)

	registerRecoveryTool(server, service)

	return server
}
