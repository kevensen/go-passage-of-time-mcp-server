package mcp

import (
	mcp_go "github.com/mark3labs/mcp-go/mcp"
	mcp_go_server "github.com/mark3labs/mcp-go/server"
)

type Server struct {
	*mcp_go_server.MCPServer
	TimeManager TimeManager
	ready       bool
}

func NewServer() *Server {
	s := &Server{
		MCPServer: mcp_go_server.NewMCPServer(
			"example-servers/everything",
			"1.0.0",
			mcp_go_server.WithToolCapabilities(true),
			mcp_go_server.WithLogging(),
		),
		TimeManager: &LiveTimeManager{},
	}
	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"currentDateTime",
			mcp_go.WithDescription("Get the current date and time in a specified timezone."),
		),
		s.CurrentDateTime)
	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"timeSince",
			mcp_go.WithDescription("Calculate the time since a given date and time.  The date/time must be in the format YYYY-MM-DD HH:MM:SS or YYYY-MM-DD.  An IANA formatted timezone can be specified (e.g. America/New_York), otherwise UTC is used."),
			mcp_go.WithString("dateTime"),
			mcp_go.WithString("timeZone"),
		),
		s.TimeSince)

	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"timeUntil",
			mcp_go.WithDescription("Calculate the time until a given date and time.  The date/time must be in the format YYYY-MM-DD HH:MM:SS or YYYY-MM-DD.	An IANA formatted timezone can be specified (e.g. America/New_York), otherwise UTC is used."),
			mcp_go.WithString("dateTime"),
			mcp_go.WithString("timeZone"),
		),
		s.TimeUntil)
	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"timeDifference",
			mcp_go.WithDescription("Calculate the difference between two date and time values.  The date/time must be in the format YYYY-MM-DD HH:MM:SS or YYYY-MM-DD.	An IANA formatted timezone can be specified (e.g. America/New_York), otherwise UTC is used."),
			mcp_go.WithString("firstDateTime"),
			mcp_go.WithString("secondDateTime"),
			mcp_go.WithString("firstTimeZone"),
			mcp_go.WithString("secondTimeZone"),
		),
		s.TimeDifference)
	return s
}

func (s *Server) Ready() bool {
	return s.ready
}
