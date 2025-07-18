package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/kevensen/go-passage-of-time-mcp-server/internal/handlers"
	"github.com/kevensen/go-passage-of-time-mcp-server/internal/handlers/mcp"
	"golang.org/x/sync/errgroup"

	mcp_go_server "github.com/mark3labs/mcp-go/server"
)

var port = flag.Int("port", -1, "Port to run the server on")
var host = flag.String("host", "0.0.0.0", "Host to run the server on")

func main() {
	flag.Parse()
	ctx := context.Background()
	eg := errgroup.Group{}

	if *port >= 0 {
		addr := fmt.Sprintf("%s:%d", *host, *port)
		slog.InfoContext(ctx, "Starting server", slog.String("address", addr))
		eg.Go(func() error {
			return handlers.Start(addr)
		})
		if err := eg.Wait(); err != nil {
			slog.ErrorContext(ctx, "Error starting server", slog.Any("error", err))
			os.Exit(1)
		}
		os.Exit(0)
	}

	router := handlers.RouterByName("/mcp")
	if router == nil {
		slog.ErrorContext(ctx, "Router not found", slog.String("name", "/mcp"))
		os.Exit(1)
	}
	server, ok := router.(*mcp.Server)
	if !ok {
		slog.ErrorContext(ctx, "Router is not a MCP Server", slog.String("name", "/mcp"))
		os.Exit(1)
	}
	if err := mcp_go_server.ServeStdio(server.MCPServer); err != nil {
		slog.ErrorContext(ctx, "Error starting MCP server", slog.Any("error", err))
		os.Exit(1)
	}
	os.Exit(0)
}
