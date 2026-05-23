package main

import (
	"log/slog"
	"os"
	"stathead-mcp/internals/tools"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	apiBase := os.Getenv("NBA_API_BASE_URL")
	if apiBase == "" {
		apiBase = "http://localhost:8080"
	}

	s := server.NewMCPServer(
		"nba-stats",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	client := tools.NewAPIClient(apiBase)
	tools.Register(s, client)

	slog.Info("starting NBA MCP server (stdio)")
	if err := server.ServeStdio(s); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
