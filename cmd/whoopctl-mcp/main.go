package main

import (
	"context"
	"log"

	"github.com/Tianbo-Qiu/whoopctl/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	server := mcpserver.New("")

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
