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
			mcp_go.WithReadOnlyHintAnnotation(true),
			mcp_go.WithString("dateTime"),
			mcp_go.WithString("timeZone"),
		),
		s.TimeSince)

	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"timeUntil",
			mcp_go.WithDescription("Calculate the time until a given date and time.  The date/time must be in the format YYYY-MM-DD HH:MM:SS or YYYY-MM-DD.	An IANA formatted timezone can be specified (e.g. America/New_York), otherwise UTC is used."),
			mcp_go.WithReadOnlyHintAnnotation(true),
			mcp_go.WithString("dateTime"),
			mcp_go.WithString("timeZone"),
		),
		s.TimeUntil)
	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"timeDifference",
			mcp_go.WithDescription("Calculate the difference between two date and time values.  The date/time must be in the format YYYY-MM-DD HH:MM:SS or YYYY-MM-DD.	An IANA formatted timezone can be specified (e.g. America/New_York), otherwise UTC is used."),
			mcp_go.WithReadOnlyHintAnnotation(true),
			mcp_go.WithString("firstDateTime"),
			mcp_go.WithString("secondDateTime"),
			mcp_go.WithString("firstTimeZone"),
			mcp_go.WithString("secondTimeZone"),
		),
		s.TimeDifference)

	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"isLeapYear",
			mcp_go.WithDescription("Check if a given year is a leap year.  The year must be provided as a number in the format YYYY."),
			mcp_go.WithReadOnlyHintAnnotation(true),
			mcp_go.WithNumber("year"),
		),
		s.IsLeapYear)

	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"dayOfWeek",
			mcp_go.WithDescription("Get the day of the week for a given date.	The date must be in the format YYYY-MM-DD."),
			mcp_go.WithReadOnlyHintAnnotation(true),
			mcp_go.WithString("dateTime"),
		),
		s.DayOfWeek)

	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"nextOccurrence",
			mcp_go.WithDescription("Get the next occurrence of a specified day of the week after a given date. The date must be in the format YYYY-MM-DD. The day of the week must be provided as a string (e.g. 'Monday', 'Tuesday', etc.)."),
			mcp_go.WithReadOnlyHintAnnotation(true),
			mcp_go.WithString("dateTime"),
			mcp_go.WithString("dayOfWeek"),
		),
		s.NextOccurrence)
	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"addDuration",
			mcp_go.WithDescription("Add a duration to a given date and time. The date/time must be in the format YYYY-MM-DD HH:MM:SS or YYYY-MM-DD. The duration must be in the format '1h30m' for 1 hour and 30 minutes."),
			mcp_go.WithReadOnlyHintAnnotation(true),
			mcp_go.WithString("dateTime"),
			mcp_go.WithString("duration"),
		),
		s.AddDuration)
	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"subtractDuration",
			mcp_go.WithDescription("Subtract a duration from a given date and time. The date/time must be in the format YYYY-MM-DD HH:MM:SS or YYYY-MM-DD. The duration must be in the format '1h30m' for 1 hour and 30 minutes."),
			mcp_go.WithReadOnlyHintAnnotation(true),
			mcp_go.WithString("dateTime"),
			mcp_go.WithString("duration"),
		),
		s.SubtractDuration)
	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"previousOccurrence",
			mcp_go.WithDescription("Get the previous occurrence of a specified day of the week before a given date. The date must be in the format YYYY-MM-DD. The day of the week must be provided as a string (e.g. 'Monday', 'Tuesday', etc.)."),
			mcp_go.WithReadOnlyHintAnnotation(true),
			mcp_go.WithString("dateTime"),
			mcp_go.WithString("dayOfWeek"),
		),
		s.PreviousOccurrence)
	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"isWeekend",
			mcp_go.WithDescription("Check if a given date is a weekend (Saturday or Sunday). The date must be in the format YYYY-MM-DD."),
			mcp_go.WithReadOnlyHintAnnotation(true),
			mcp_go.WithString("dateTime"),
		),
		s.IsWeekend)
	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"isWeekday",
			mcp_go.WithDescription("Check if a given date is a weekday (Monday to Friday). The date must be in the format YYYY-MM-DD."),
			mcp_go.WithReadOnlyHintAnnotation(true),
			mcp_go.WithString("dateTime"),
		),
		s.IsWeekday)
	s.MCPServer.AddTool(
		mcp_go.NewTool(
			"daysBetween",
			mcp_go.WithDescription("Calculate the number of days between two dates. The dates must be in the format YYYY-MM-DD."),
			mcp_go.WithReadOnlyHintAnnotation(true),
			mcp_go.WithString("firstDate"),
			mcp_go.WithString("secondDate"),
		),
		s.DaysBetween)

	return s
}

func (s *Server) Ready() bool {
	return s.ready
}
