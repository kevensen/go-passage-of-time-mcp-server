package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	mcp_go "github.com/mark3labs/mcp-go/mcp"
)

const dateFormat = "2006-01-02"

// YYYY-MM-DD HH:MM:SS
const dateTimeFormat = "2006-01-02 15:04:05"

const dateTimeFormatTimeZone = "2006-01-02 15:04:05 -0700"

type TimeManager interface {
	Now() time.Time
	LoadLocation(name string) (*time.Location, error)
}

// TimeOpts is a struct that holds options for parsing time.
type TimeOpts struct {
	input      string
	timeZone   string
	timeFormat string
}

// ParseTime creates a new TimeOpts instance with the provided input and
// optional time zone.  If TimeZone is not provided, it defaults to UTC.
func ParseTime(opts *TimeOpts) (time.Time, error) {
	if opts == nil {
		return time.Time{}, NewNilTimeOptsError()
	}
	if opts.input == "" {
		return time.Time{}, NewNilInputTime()
	}
	if opts.timeFormat == "" {

	}

	var t time.Time
	var err error
	if opts.timeFormat != "" {
		t, err = time.Parse(opts.timeFormat, opts.input)
		if err != nil {
			return time.Time{}, NewInvalidTimeFormatError(opts.input)
		}
	} else {

		t, err = time.Parse(dateTimeFormat, opts.input)
		if err != nil {
			t, err = time.Parse(dateFormat, opts.input)
			if err != nil {
				return time.Time{}, NewInvalidTimeFormatError(opts.input)
			}
		}
	}

	if opts.timeZone != "" {
		loc, err := time.LoadLocation(opts.timeZone)
		if err != nil {
			return time.Time{}, NewTimeZoneLoadError(opts.timeZone, err)
		}
		t = t.In(loc)
	}
	return t, nil
}

type LiveTimeManager struct{}

func (l *LiveTimeManager) Now() time.Time {
	return time.Now().UTC()
}

func (l *LiveTimeManager) LoadLocation(name string) (*time.Location, error) {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, NewTimeZoneLoadError(name, err)
	}
	return loc, nil
}

func (s *Server) CurrentDateTime(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	tz := request.GetString("timeZone", "UTC")

	now := s.TimeManager.Now()

	// Get the zone name and offset
	zoneName, offsetSeconds := now.Zone()

	loc, err := s.TimeManager.LoadLocation(tz)
	if err != nil {
		return nil, NewTimeZoneLoadError(tz, err)
	}

	t := now.In(loc)

	_, requestedZoneOffset := t.Zone()
	slog.InfoContext(ctx, "CurrentDateTime", slog.String("server_local_tz", zoneName), slog.Int("server_local_offset_seconds", offsetSeconds), slog.String("requested_tz", tz), slog.Int("requested_tz_offset", requestedZoneOffset))

	return &mcp_go.CallToolResult{
		Content: []mcp_go.Content{
			mcp_go.TextContent{
				Type: "text",
				Text: t.Format(dateTimeFormatTimeZone),
			},
		},
	}, nil
}

func (s *Server) TimeSince(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	tz := request.GetString("timeZone", "UTC")
	input := request.GetString("dateTime", "")
	if input == "" {
		return mcp_go.NewToolResultError(NewNilInputTime().Error()), nil
	}

	opts := &TimeOpts{
		input:    input,
		timeZone: tz,
	}

	t, err := ParseTime(opts)
	if err != nil {
		return mcp_go.NewToolResultError(err.Error()), nil
	}
	t = normalizeTimeToUTC(ctx, t)

	now := s.TimeManager.Now()
	now = normalizeTimeToUTC(ctx, now)

	if t.After(now) {
		return mcp_go.NewToolResultError("The specified time is in the future"), nil
	}
	if t.Equal(now) {
		return &mcp_go.CallToolResult{
			Content: []mcp_go.Content{
				mcp_go.TextContent{
					Type: "text",
					Text: "The specified time is now.",
				},
			},
		}, nil
	}

	duration := now.Sub(t)

	return &mcp_go.CallToolResult{
		Content: []mcp_go.Content{
			mcp_go.TextContent{
				Type: "text",
				Text: duration.String(),
			},
		},
	}, nil
}

func (s *Server) TimeUntil(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	tz := request.GetString("timeZone", "UTC")
	input := request.GetString("dateTime", "")
	if input == "" {
		return mcp_go.NewToolResultError(NewNilInputTime().Error()), nil
	}

	opts := &TimeOpts{
		input:    input,
		timeZone: tz,
	}

	t, err := ParseTime(opts)
	if err != nil {
		return mcp_go.NewToolResultError(err.Error()), nil
	}
	t = normalizeTimeToUTC(ctx, t)

	now := s.TimeManager.Now()
	now = normalizeTimeToUTC(ctx, now)

	if t.Before(now) {
		return mcp_go.NewToolResultError("The specified time is in the past"), nil
	}

	duration := t.Sub(now)

	return &mcp_go.CallToolResult{
		Content: []mcp_go.Content{
			mcp_go.TextContent{
				Type: "text",
				Text: duration.String(),
			},
		},
	}, nil
}

func (s *Server) TimeDifference(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	firstTimeZone := request.GetString("firstTimeZone", "UTC")
	firstDateTime := request.GetString("firstDateTime", "")
	secondTimeZone := request.GetString("secondTimeZone", "UTC")
	secondDateTime := request.GetString("secondDateTime", "")
	if firstDateTime == "" || secondDateTime == "" {
		return mcp_go.NewToolResultError("Both firstDateTime and secondDateTime must be provided"), nil
	}

	firstOpts := &TimeOpts{
		input:    firstDateTime,
		timeZone: firstTimeZone,
	}
	secondOpts := &TimeOpts{
		input:    secondDateTime,
		timeZone: secondTimeZone,
	}
	firstTime, err := ParseTime(firstOpts)
	if err != nil {
		return mcp_go.NewToolResultErrorFromErr("error with first input time", err), nil
	}
	secondTime, err := ParseTime(secondOpts)
	if err != nil {
		return mcp_go.NewToolResultErrorFromErr("error with second input time", err), nil
	}

	firstTime = normalizeTimeToUTC(ctx, firstTime)
	secondTime = normalizeTimeToUTC(ctx, secondTime)

	var result *mcp_go.CallToolResult

	if firstTime.Equal(secondTime) {
		result = &mcp_go.CallToolResult{
			Content: []mcp_go.Content{
				mcp_go.TextContent{
					Type: "text",
					Text: "The two times are equal.",
				},
			},
		}
	}

	if firstTime.Before(secondTime) {
		duration := secondTime.Sub(firstTime)
		result = &mcp_go.CallToolResult{
			Content: []mcp_go.Content{
				mcp_go.TextContent{
					Type: "text",
					Text: "The first time is earlier than the second time by " + duration.String(),
				},
			},
		}
	}
	if firstTime.After(secondTime) {
		duration := firstTime.Sub(secondTime)
		result = &mcp_go.CallToolResult{
			Content: []mcp_go.Content{
				mcp_go.TextContent{
					Type: "text",
					Text: "The first time is later than the second time by " + duration.String(),
				},
			},
		}
	}

	return result, nil

}

func (s *Server) IsLeapYear(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	year := request.GetInt("year", 0)
	if year <= 0 {
		return mcp_go.NewToolResultError("Invalid year provided"), nil
	}
	isLeap := (year%4 == 0 && year%100 != 0) || (year%400 == 0)
	result := "is not a leap year."
	if isLeap {
		result = "is a leap year."
	}
	return &mcp_go.CallToolResult{
		Content: []mcp_go.Content{
			mcp_go.TextContent{
				Type: "text",
				Text: fmt.Sprintf("%d %s", year, result),
			},
		},
	}, nil
}

func (s *Server) DayOfWeek(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	input := request.GetString("dateTime", "")
	if input == "" {
		return mcp_go.NewToolResultError(NewNilInputTime().Error()), nil
	}
	opts := &TimeOpts{
		input: input,
	}

	t, err := ParseTime(opts)
	if err != nil {
		return mcp_go.NewToolResultError(err.Error()), nil
	}
	dayOfWeek := t.Weekday().String()

	return &mcp_go.CallToolResult{
		Content: []mcp_go.Content{
			mcp_go.TextContent{
				Type: "text",
				Text: fmt.Sprintf("The day of the week for %s is %s.", t.Format(dateFormat), dayOfWeek),
			},
		},
	}, nil
}

func (s *Server) AddDuration(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	input := request.GetString("dateTime", "")
	if input == "" {
		return mcp_go.NewToolResultError(NewNilInputTime().Error()), nil
	}

	opts := &TimeOpts{
		input: input,
	}

	t, err := ParseTime(opts)
	if err != nil {
		return mcp_go.NewToolResultError(err.Error()), nil
	}

	durationStr := request.GetString("duration", "")
	if durationStr == "" {
		return mcp_go.NewToolResultError("Duration must be provided"), nil
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return mcp_go.NewToolResultError(fmt.Sprintf("Invalid duration format: %s", durationStr)), nil
	}

	newTime := t.Add(duration)

	return &mcp_go.CallToolResult{
		Content: []mcp_go.Content{
			mcp_go.TextContent{
				Type: "text",
				Text: fmt.Sprintf("New time after adding duration: %s", newTime.Format(dateTimeFormatTimeZone)),
			},
		},
	}, nil
}

func (s *Server) SubtractDuration(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	input := request.GetString("dateTime", "")
	if input == "" {
		return mcp_go.NewToolResultError(NewNilInputTime().Error()), nil
	}

	opts := &TimeOpts{
		input: input,
	}

	t, err := ParseTime(opts)
	if err != nil {
		return mcp_go.NewToolResultError(err.Error()), nil
	}

	durationStr := request.GetString("duration", "")
	if durationStr == "" {
		return mcp_go.NewToolResultError("Duration must be provided"), nil
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return mcp_go.NewToolResultError(fmt.Sprintf("Invalid duration format: %s", durationStr)), nil
	}

	newTime := t.Add(-duration)

	return &mcp_go.CallToolResult{
		Content: []mcp_go.Content{
			mcp_go.TextContent{
				Type: "text",
				Text: fmt.Sprintf("New time after subtracting duration: %s", newTime.Format(dateTimeFormatTimeZone)),
			},
		},
	}, nil
}

// NextOccurrence calculates the next date for a specified day of the week (e.g. "Monday") after a given date and time.
func (s *Server) NextOccurrence(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	input := request.GetString("dateTime", "")
	if input == "" {
		return mcp_go.NewToolResultError(NewNilInputTime().Error()), nil
	}
	opts := &TimeOpts{
		input: input,
	}
	t, err := ParseTime(opts)
	if err != nil {
		return mcp_go.NewToolResultError(err.Error()), nil
	}
	dayOfWeekStr := request.GetString("dayOfWeek", "")
	if dayOfWeekStr == "" {
		return mcp_go.NewToolResultError("Day of week must be provided"), nil
	}
	dayOfWeek, err := parseWeekday(dayOfWeekStr)
	if err != nil {
		return mcp_go.NewToolResultError(fmt.Sprintf("Invalid day of week format: %s", dayOfWeekStr)), nil
	}

	//Don't normalize the time to UTC here, as we want to keep the local time context
	nextOccurrence := t
	if nextOccurrence.Weekday() == dayOfWeek {
		// If the input date is already the requested day pre-increment it by one day
		// to find the next occurrence.
		nextOccurrence = nextOccurrence.AddDate(0, 0, 1)
	}

	for nextOccurrence.Weekday() != dayOfWeek {
		nextOccurrence = nextOccurrence.AddDate(0, 0, 1)
	}
	slog.InfoContext(ctx, "NextOccurrence", slog.String("input_time", t.Format(dateTimeFormatTimeZone)), slog.String("requested_day_of_week", dayOfWeekStr), slog.String("next_occurrence", nextOccurrence.Format(dateTimeFormatTimeZone)))
	return &mcp_go.CallToolResult{
		Content: []mcp_go.Content{
			mcp_go.TextContent{
				Type: "text",
				Text: fmt.Sprintf("The next occurrence of %s after %s is %s.", dayOfWeekStr, t.Format(dateTimeFormatTimeZone), nextOccurrence.Format(dateTimeFormatTimeZone)),
			},
		},
	}, nil
}

func (s *Server) PreviousOccurrence(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	input := request.GetString("dateTime", "")
	if input == "" {
		return mcp_go.NewToolResultError(NewNilInputTime().Error()), nil
	}
	opts := &TimeOpts{
		input: input,
	}
	t, err := ParseTime(opts)
	if err != nil {
		return mcp_go.NewToolResultError(err.Error()), nil
	}
	dayOfWeekStr := request.GetString("dayOfWeek", "")
	if dayOfWeekStr == "" {
		return mcp_go.NewToolResultError("Day of week must be provided"), nil
	}
	dayOfWeek, err := parseWeekday(dayOfWeekStr)
	if err != nil {
		return mcp_go.NewToolResultError(fmt.Sprintf("Invalid day of week format: %s", dayOfWeekStr)), nil
	}

	//Don't normalize the time to UTC here, as we want to keep the local time context
	prevOccurrence := t
	if prevOccurrence.Weekday() == dayOfWeek {
		// If the input date is already the requested day pre-decrement it by one day
		// to find the previous occurrence.
		prevOccurrence = prevOccurrence.AddDate(0, 0, -1)
	}

	for prevOccurrence.Weekday() != dayOfWeek {
		prevOccurrence = prevOccurrence.AddDate(0, 0, -1)
	}
	slog.InfoContext(ctx, "PreviousOccurrence", slog.String("input_time", t.Format(dateTimeFormatTimeZone)), slog.String("requested_day_of_week", dayOfWeekStr), slog.String("previous_occurrence", prevOccurrence.Format(dateTimeFormatTimeZone)))
	return &mcp_go.CallToolResult{
		Content: []mcp_go.Content{
			mcp_go.TextContent{
				Type: "text",
				Text: fmt.Sprintf("The previous occurrence of %s before %s is %s.", dayOfWeekStr, t.Format(dateTimeFormatTimeZone), prevOccurrence.Format(dateTimeFormatTimeZone)),
			},
		},
	}, nil
}

func (s *Server) IsWeekend(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	input := request.GetString("dateTime", "")
	if input == "" {
		return mcp_go.NewToolResultError(NewNilInputTime().Error()), nil
	}
	opts := &TimeOpts{
		input: input,
	}
	t, err := ParseTime(opts)
	if err != nil {
		return mcp_go.NewToolResultError(err.Error()), nil
	}

	isWeekend := t.Weekday() == time.Saturday || t.Weekday() == time.Sunday

	result := "is not a weekend."
	if isWeekend {
		result = "is a weekend."
	}

	return &mcp_go.CallToolResult{
		Content: []mcp_go.Content{
			mcp_go.TextContent{
				Type: "text",
				Text: fmt.Sprintf("%s %s", t.Format(dateFormat), result),
			},
		},
	}, nil
}

func (s *Server) IsWeekday(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	input := request.GetString("dateTime", "")
	if input == "" {
		return mcp_go.NewToolResultError(NewNilInputTime().Error()), nil
	}
	opts := &TimeOpts{
		input: input,
	}
	t, err := ParseTime(opts)
	if err != nil {
		return mcp_go.NewToolResultError(err.Error()), nil
	}

	isWeekday := t.Weekday() >= time.Monday && t.Weekday() <= time.Friday

	result := "is not a weekday."
	if isWeekday {
		result = "is a weekday."
	}

	return &mcp_go.CallToolResult{
		Content: []mcp_go.Content{
			mcp_go.TextContent{
				Type: "text",
				Text: fmt.Sprintf("%s %s", t.Format(dateFormat), result),
			},
		},
	}, nil
}

func (s *Server) DaysBetween(ctx context.Context, request mcp_go.CallToolRequest) (*mcp_go.CallToolResult, error) {
	firstInput := request.GetString("firstDateTime", "")
	if firstInput == "" {
		return mcp_go.NewToolResultError(NewNilInputTime().Error()), nil
	}
	secondInput := request.GetString("secondDateTime", "")
	if secondInput == "" {
		return mcp_go.NewToolResultError(NewNilInputTime().Error()), nil
	}

	firstOpts := &TimeOpts{
		input: firstInput,
	}
	secondOpts := &TimeOpts{
		input: secondInput,
	}

	firstTime, err := ParseTime(firstOpts)
	if err != nil {
		return mcp_go.NewToolResultError(err.Error()), nil
	}
	secondTime, err := ParseTime(secondOpts)
	if err != nil {
		return mcp_go.NewToolResultError(err.Error()), nil
	}

	firstTime = normalizeTimeToUTC(ctx, firstTime)
	secondTime = normalizeTimeToUTC(ctx, secondTime)

	daysBetween := int(secondTime.Sub(firstTime).Hours() / 24)

	return &mcp_go.CallToolResult{
		Content: []mcp_go.Content{
			mcp_go.TextContent{
				Type: "text",
				Text: fmt.Sprintf("There are %d days between %s and %s.", daysBetween, firstTime.Format(dateFormat), secondTime.Format(dateFormat)),
			},
		},
	}, nil
}

func normalizeTimeToUTC(ctx context.Context, t time.Time) time.Time {
	zone, offset := t.Zone()
	slog.InfoContext(ctx, "normalizeTimeToUTC", slog.String("input_time", t.Format(dateTimeFormatTimeZone)), slog.String("time_zone", zone), slog.Int("offset_seconds", offset))

	// Normalize the time to UTC by subtracting the offset
	return t.Add(-time.Duration(offset) * time.Second)
}

// parseWeekday maps a string to a time.Weekday value.
func parseWeekday(s string) (time.Weekday, error) {
	switch normalized := normalizeWeekdayString(s); normalized {
	case "sunday":
		return time.Sunday, nil
	case "monday":
		return time.Monday, nil
	case "tuesday":
		return time.Tuesday, nil
	case "wednesday":
		return time.Wednesday, nil
	case "thursday":
		return time.Thursday, nil
	case "friday":
		return time.Friday, nil
	case "saturday":
		return time.Saturday, nil
	default:
		return time.Sunday, fmt.Errorf("invalid weekday: %s", s)
	}
}

func normalizeWeekdayString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
