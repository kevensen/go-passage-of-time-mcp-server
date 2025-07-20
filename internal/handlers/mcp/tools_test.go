package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

type mockTmanager struct{}

func (m *mockTmanager) Now() time.Time {
	return time.Date(2023, 10, 1, 12, 30, 0, 0, time.UTC)
}
func (m *mockTmanager) LoadLocation(name string) (*time.Location, error) {
	return time.LoadLocation(name)
}

func TestParseTime(t *testing.T) {
	testCases := []struct {
		desc    string
		opts    *TimeOpts
		want    time.Time
		wanterr bool
	}{
		{
			desc: "Valid date and time with timezone",
			opts: &TimeOpts{
				input:    "2023-10-01 12:30:00",
				timeZone: "America/New_York",
			},
			want:    time.Date(2023, 10, 1, 12, 30, 0, 0, time.FixedZone("EST", 0)),
			wanterr: false,
		},
		{
			desc: "Valid date only",
			opts: &TimeOpts{
				input: "2023-10-01",
			},
			want:    time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			wanterr: false,
		},
		{
			desc: "Valid date and time without timezone",
			opts: &TimeOpts{
				input: "2023-10-01 12:30:00",
			},
			want:    time.Date(2023, 10, 1, 12, 30, 0, 0, time.UTC),
			wanterr: false,
		},
		{
			desc: "Valid date with custom format",
			opts: &TimeOpts{
				input:      "01-10-2023",
				timeFormat: "02-01-2006",
			},
			want:    time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			wanterr: false,
		},
		{
			desc: "Valid date and time with custom format",
			opts: &TimeOpts{
				input:      "01-10-2023 12:30",
				timeFormat: "02-01-2006 15:04",
			},
			want:    time.Date(2023, 10, 1, 12, 30, 0, 0, time.UTC),
			wanterr: false,
		},
		{
			desc: "Bad custom format",
			opts: &TimeOpts{
				input:      "01-10-2023 12:30",
				timeFormat: "02-01-2006 15:04:05",
			},
			want:    time.Time{},
			wanterr: true,
		},
		{
			desc: "Invalid date format",
			opts: &TimeOpts{
				input: "2023-10-01 12:30",
			},
			want:    time.Time{},
			wanterr: true,
		},
		{
			desc: "Empty input",
			opts: &TimeOpts{
				input: "",
			},
			want:    time.Time{},
			wanterr: true,
		},
		{
			desc:    "Nil options",
			opts:    nil,
			want:    time.Time{},
			wanterr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := ParseTime(tc.opts)
			if (err != nil) != tc.wanterr {
				t.Errorf("ParseTime() error = %v, wantErr %v", err, tc.wanterr)
				return
			}
			if !got.Equal(tc.want) {
				t.Errorf("ParseTime() got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCurrentDateTime(t *testing.T) {
	testCases := []struct {
		desc    string
		request *mcp.CallToolRequest
		want    time.Time
		wantErr bool
	}{
		{
			desc: "Valid request with UTC timezone",
			request: &mcp.CallToolRequest{
				Params: mcp.CallToolParams{},
			},
			want:    time.Date(2023, 10, 1, 12, 30, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			desc: "Valid request with specific timezone",
			request: &mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]any{
						"timeZone": "Pacific/Honolulu",
					},
				},
			},
			want:    time.Date(2023, 10, 1, 2, 30, 0, 0, time.FixedZone("HST", -10*60*60)),
			wantErr: false,
		},
	}
	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := &Server{
				TimeManager: &mockTmanager{},
			}

			got, err := s.CurrentDateTime(ctx, *tc.request)
			if (err != nil) != tc.wantErr {
				t.Fatalf("CurrentDateTime() error = %v, wantErr %v", err, tc.wantErr)
			}
			if got == nil || len(got.Content) == 0 {
				t.Fatalf("CurrentDateTime() got = nil or empty content")
			}
			gotTextContent, ok := got.Content[0].(mcp.TextContent)
			if !ok {
				t.Fatalf("CurrentDateTime() got = %+v, want TextContent", got.Content[0])
			}
			gotTime, err := time.Parse(dateTimeFormatTimeZone, gotTextContent.Text)
			if err != nil {
				t.Fatalf("CurrentDateTime() failed to parse time from content: %v", err)
			}
			if !gotTime.Equal(tc.want) {
				t.Errorf("CurrentDateTime() got = %v, want %v", gotTime, tc.want)
			}
			_, gotOffset := gotTime.Zone()
			_, wantOffset := tc.want.Zone()
			if gotOffset != wantOffset {
				t.Errorf("CurrentDateTime() got offset = %d, want %d", gotOffset, wantOffset)
			}
		})
	}
}

func TestTimeUntil(t *testing.T) {
	testCases := []struct {
		desc    string
		request *mcp.CallToolRequest
		want    string
		wantErr bool
	}{
		{
			desc: "Valid request with past date, time, and same timezone",
			request: &mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]any{
						"dateTime": "2023-09-30 12:00:00",
						"timeZone": "UTC",
					},
				},
			},
			want:    "24h30m0s",
			wantErr: false,
		},
		{
			desc: "Valid request with past date and time without timezone",
			request: &mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]any{
						"dateTime": "2023-09-30 12:00:00",
					},
				},
			},
			want:    "24h30m0s",
			wantErr: false,
		},
		{
			desc: "Valid request with past date and time with different timezone",
			request: &mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]any{
						"dateTime": "2023-09-30 12:00:00",
						"timeZone": "HST",
					},
				},
			},
			want:    "14h30m0s",
			wantErr: false,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := &Server{
				TimeManager: &mockTmanager{},
			}

			got, err := s.TimeSince(ctx, *tc.request)
			if (err != nil) != tc.wantErr {
				t.Fatalf("CurrentDateTime() error = %v, wantErr %v", err, tc.wantErr)
			}
			if got.IsError && !tc.wantErr {
				t.Fatalf("CurrentDateTime() got error = %v, want no error", got.Content[0])
			}
			if tc.wantErr && !got.IsError {
				t.Fatalf("CurrentDateTime() expected error, got no error")
			}
			if got.IsError {
				t.Skip()
			}

			gotTextContent, ok := got.Content[0].(mcp.TextContent)
			if !ok {
				t.Fatalf("CurrentDateTime() got = %+v, want TextContent", got.Content[0])
			}
			if gotTextContent.Text != tc.want {
				t.Errorf("CurrentDateTime() got = %v, want %v", gotTextContent.Text, tc.want)
			}

		})
	}
}

func TestTimeDifference(t *testing.T) {
	testCases := []struct {
		desc    string
		request *mcp.CallToolRequest
		want    string
		wantErr bool
	}{
		{
			desc: "Valid request with two dates and times in the same timezone with first earlier",
			request: &mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]any{
						"firstDateTime":  "2023-09-30 12:00:00",
						"firstTimeZonw":  "UTC",
						"secondDateTime": "2023-10-01 12:30:00",
						"secondTimeZone": "UTC",
					},
				},
			},
			want:    "The first time is earlier than the second time by 24h30m0s",
			wantErr: false,
		},
		{
			desc: "Valid request with two dates and times in the same timezone with first later",
			request: &mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]any{
						"firstDateTime":  "2023-10-01 12:30:00",
						"firstTimeZone":  "UTC",
						"secondDateTime": "2023-09-30 12:00:00",
						"secondTimeZone": "UTC",
					},
				},
			},
			want:    "The first time is later than the second time by 24h30m0s",
			wantErr: false,
		},
		{
			desc: "Valid request with two dates and times in the same timezone with equal times",
			request: &mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]any{
						"firstDateTime":  "2023-10-01 12:30:00",
						"firstTimeZone":  "UTC",
						"secondDateTime": "2023-10-01 12:30:00",
						"secondTimeZone": "UTC",
					},
				},
			},
			want: "The two times are equal.",
		},
		{
			desc: "Valid request with two dates and times in different timezones with second earlier",
			request: &mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]any{
						"firstDateTime":  "2023-10-01 12:30:00",
						"secondDateTime": "2023-10-01 01:30:00",
						"secondTimeZone": "Pacific/Honolulu",
					},
				},
			},
			want:    "The first time is later than the second time by 1h0m0s",
			wantErr: false,
		},
	}
	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := &Server{
				TimeManager: &mockTmanager{},
			}
			got, err := s.TimeDifference(ctx, *tc.request)
			if (err != nil) != tc.wantErr {
				t.Fatalf("TimeDifference() error = %v, wantErr %v", err, tc.wantErr)
			}
			if got.IsError && !tc.wantErr {
				t.Fatalf("TimeDifference() got error = %v, want no error", got.Content[0])
			}
			if tc.wantErr && !got.IsError {
				t.Fatalf("TimeDifference() expected error, got no error")
			}
			if got.IsError {
				t.Skip()
			}

			gotTextContent, ok := got.Content[0].(mcp.TextContent)
			if !ok {
				t.Fatalf("TimeDifference() got = %+v, want TextContent", got.Content[0])
			}
			if gotTextContent.Text != tc.want {
				t.Errorf("TimeDifference() got = %v, want %v", gotTextContent.Text, tc.want)
			}
		})
	}
}
func TestIsLeapYear(t *testing.T) {
	testCases := []struct {
		desc    string
		year    int
		want    string
		wantErr bool
	}{
		{
			desc: "Leap year divisible by 4 but not 100",
			year: 2024,
			want: "2024 is a leap year.",
		},
		{
			desc: "Non-leap year not divisible by 4",
			year: 2023,
			want: "2023 is not a leap year.",
		},
		{
			desc: "Non-leap year divisible by 100 but not 400",
			year: 1900,
			want: "1900 is not a leap year.",
		},
		{
			desc: "Leap year divisible by 400",
			year: 2000,
			want: "2000 is a leap year.",
		},
		{
			desc:    "Invalid year zero",
			year:    0,
			wantErr: true,
		},
		{
			desc:    "Invalid negative year",
			year:    -2020,
			wantErr: true,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := &Server{
				TimeManager: &mockTmanager{},
			}
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]any{
						"year": tc.year,
					},
				},
			}
			got, _ := s.IsLeapYear(ctx, req)

			if tc.wantErr {
				if got == nil || !got.IsError {
					t.Errorf("IsLeapYear() expected error, got = %+v", got)
				}
				return
			}
			if got == nil || len(got.Content) == 0 {
				t.Fatalf("IsLeapYear() got = nil or empty content")
			}
			gotTextContent, ok := got.Content[0].(mcp.TextContent)
			if !ok {
				t.Fatalf("IsLeapYear() got = %+v, want TextContent", got.Content[0])
			}
			if gotTextContent.Text != tc.want {
				t.Errorf("IsLeapYear() got = %v, want %v", gotTextContent.Text, tc.want)
			}
		})
	}
}
func TestDayOfWeek(t *testing.T) {
	testCases := []struct {
		desc    string
		input   string
		want    string
		wantErr bool
	}{
		{
			desc:  "Valid date returns correct weekday",
			input: "2023-10-01",
			want:  "The day of the week for 2023-10-01 is Sunday.",
		},
		{
			desc:  "Valid date and time returns correct weekday",
			input: "2023-10-02 15:00:00",
			want:  "The day of the week for 2023-10-02 is Monday.",
		},
		{
			desc:    "Empty input returns error",
			input:   "",
			wantErr: true,
		},
		{
			desc:    "Invalid date format returns error",
			input:   "10/01/2023",
			wantErr: true,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := &Server{
				TimeManager: &mockTmanager{},
			}
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]any{
						"dateTime": tc.input,
					},
				},
			}
			got, _ := s.DayOfWeek(ctx, req)
			if tc.wantErr {
				if got == nil || !got.IsError {
					t.Errorf("DayOfWeek() expected error, got = %+v", got)
				}
				return
			}
			if got == nil || len(got.Content) == 0 {
				t.Fatalf("DayOfWeek() got = nil or empty content")
			}
			gotTextContent, ok := got.Content[0].(mcp.TextContent)
			if !ok {
				t.Fatalf("DayOfWeek() got = %+v, want TextContent", got.Content[0])
			}
			if gotTextContent.Text != tc.want {
				t.Errorf("DayOfWeek() got = %v, want %v", gotTextContent.Text, tc.want)
			}
		})
	}
}
func TestAddDuration(t *testing.T) {
	testCases := []struct {
		desc     string
		dateTime string
		duration string
		want     string
		wantErr  bool
	}{
		{
			desc:     "Add 1 hour to valid date and time",
			dateTime: "2023-10-01 12:30:00",
			duration: "1h",
			want:     "New time after adding duration: 2023-10-01 13:30:00 +0000",
		},
		{
			desc:     "Add 30 minutes to valid date and time",
			dateTime: "2023-10-01 12:30:00",
			duration: "30m",
			want:     "New time after adding duration: 2023-10-01 13:00:00 +0000",
		},
		{
			desc:     "Add negative duration",
			dateTime: "2023-10-01 12:30:00",
			duration: "-1h",
			want:     "New time after adding duration: 2023-10-01 11:30:00 +0000",
		},
		{
			desc:     "Missing dateTime",
			dateTime: "",
			duration: "1h",
			wantErr:  true,
		},
		{
			desc:     "Missing duration",
			dateTime: "2023-10-01 12:30:00",
			duration: "",
			wantErr:  true,
		},
		{
			desc:     "Invalid duration format",
			dateTime: "2023-10-01 12:30:00",
			duration: "abc",
			wantErr:  true,
		},
		{
			desc:     "Invalid dateTime format",
			dateTime: "not-a-date",
			duration: "1h",
			wantErr:  true,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := &Server{
				TimeManager: &mockTmanager{},
			}
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]any{
						"dateTime": tc.dateTime,
						"duration": tc.duration,
					},
				},
			}
			got, _ := s.AddDuration(ctx, req)
			if tc.wantErr {
				if got == nil || !got.IsError {
					t.Errorf("AddDuration() expected error, got = %+v", got)
				}
				return
			}
			if got == nil || len(got.Content) == 0 {
				t.Fatalf("AddDuration() got = nil or empty content")
			}
			gotTextContent, ok := got.Content[0].(mcp.TextContent)
			if !ok {
				t.Fatalf("AddDuration() got = %+v, want TextContent", got.Content[0])
			}
			if gotTextContent.Text != tc.want {
				t.Errorf("AddDuration() got = %v, want %v", gotTextContent.Text, tc.want)
			}
		})
	}
}

func TestSubtractDuration(t *testing.T) {
	testCases := []struct {
		desc     string
		dateTime string
		duration string
		want     string
		wantErr  bool
	}{
		{
			desc:     "Subtract 1 hour from valid date and time",
			dateTime: "2023-10-01 12:30:00",
			duration: "1h",
			want:     "New time after subtracting duration: 2023-10-01 11:30:00 +0000",
		},
		{
			desc:     "Subtract 30 minutes from valid date and time",
			dateTime: "2023-10-01 12:30:00",
			duration: "30m",
			want:     "New time after subtracting duration: 2023-10-01 12:00:00 +0000",
		},
		{
			desc:     "Subtract negative duration",
			dateTime: "2023-10-01 12:30:00",
			duration: "-1h",
			want:     "New time after subtracting duration: 2023-10-01 13:30:00 +0000",
		},
		{
			desc:     "Missing dateTime",
			dateTime: "",
			duration: "1h",
			wantErr:  true,
		},
		{
			desc:     "Missing duration",
			dateTime: "2023-10-01 12:30:00",
			duration: "",
			wantErr:  true,
		},
		{
			desc:     "Invalid duration format",
			dateTime: "2023-10-01 12:30:00",
			duration: "abc",
			wantErr:  true,
		},
		{
			desc:     "Invalid dateTime format",
			dateTime: "not-a-date",
			duration: "1h",
			wantErr:  true,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := &Server{
				TimeManager: &mockTmanager{},
			}
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]any{
						"dateTime": tc.dateTime,
						"duration": tc.duration,
					},
				},
			}
			got, _ := s.SubtractDuration(ctx, req)
			if tc.wantErr {
				if got == nil || !got.IsError {
					t.Errorf("SubtractDuration() expected error, got = %+v", got)
				}
				return
			}
			if got == nil || len(got.Content) == 0 {
				t.Fatalf("SubtractDuration() got = nil or empty content")
			}
			gotTextContent, ok := got.Content[0].(mcp.TextContent)
			if !ok {
				t.Fatalf("SubtractDuration() got = %+v, want TextContent", got.Content[0])
			}
			if gotTextContent.Text != tc.want {
				t.Errorf("SubtractDuration() got = %v, want %v", gotTextContent.Text, tc.want)
			}
		})
	}
}
func TestNextOccurrence(t *testing.T) {
	testCases := []struct {
		desc      string
		dateTime  string
		dayOfWeek string
		want      string
		wantErr   bool
	}{
		{
			desc:      "Next Monday after Sunday",
			dateTime:  "2023-10-01 12:30:00", // Sunday
			dayOfWeek: "Monday",
			want:      "The next occurrence of Monday after 2023-10-01 12:30:00 +0000 is 2023-10-02 12:30:00 +0000.",
		},
		{
			desc:      "Next Sunday after Sunday (should be next week)",
			dateTime:  "2023-10-01 12:30:00", // Sunday
			dayOfWeek: "Sunday",
			want:      "The next occurrence of Sunday after 2023-10-01 12:30:00 +0000 is 2023-10-08 12:30:00 +0000.",
		},
		{
			desc:      "Next Wednesday after Monday",
			dateTime:  "2023-10-02 09:00:00", // Monday
			dayOfWeek: "Wednesday",
			want:      "The next occurrence of Wednesday after 2023-10-02 09:00:00 +0000 is 2023-10-04 09:00:00 +0000.",
		},
		{
			desc:      "Missing dateTime",
			dateTime:  "",
			dayOfWeek: "Monday",
			wantErr:   true,
		},
		{
			desc:      "Missing dayOfWeek",
			dateTime:  "2023-10-01 12:30:00",
			dayOfWeek: "",
			wantErr:   true,
		},
		{
			desc:      "Invalid dayOfWeek",
			dateTime:  "2023-10-01 12:30:00",
			dayOfWeek: "Funday",
			wantErr:   true,
		},
		{
			desc:      "Invalid dateTime format",
			dateTime:  "not-a-date",
			dayOfWeek: "Monday",
			wantErr:   true,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := &Server{
				TimeManager: &mockTmanager{},
			}
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]any{
						"dateTime":  tc.dateTime,
						"dayOfWeek": tc.dayOfWeek,
					},
				},
			}
			got, _ := s.NextOccurrence(ctx, req)
			if tc.wantErr {
				if got == nil || !got.IsError {
					t.Errorf("NextOccurrence() expected error, got = %+v", got)
				}
				return
			}
			if got == nil || len(got.Content) == 0 {
				t.Fatalf("NextOccurrence() got = nil or empty content")
			}
			gotTextContent, ok := got.Content[0].(mcp.TextContent)
			if !ok {
				t.Fatalf("NextOccurrence() got = %+v, want TextContent", got.Content[0])
			}
			if gotTextContent.Text != tc.want {
				t.Errorf("NextOccurrence() got = %v, want %v", gotTextContent.Text, tc.want)
			}
		})
	}
}
func TestPreviousOccurrence(t *testing.T) {
	testCases := []struct {
		desc      string
		dateTime  string
		dayOfWeek string
		want      string
		wantErr   bool
	}{
		{
			desc:      "Previous Monday before Sunday",
			dateTime:  "2023-10-01 12:30:00", // Sunday
			dayOfWeek: "Monday",
			want:      "The previous occurrence of Monday before 2023-10-01 12:30:00 +0000 is 2023-09-25 12:30:00 +0000.",
		},
		{
			desc:      "Previous Sunday before Sunday (should be previous week)",
			dateTime:  "2023-10-01 12:30:00", // Sunday
			dayOfWeek: "Sunday",
			want:      "The previous occurrence of Sunday before 2023-10-01 12:30:00 +0000 is 2023-09-24 12:30:00 +0000.",
		},
		{
			desc:      "Previous Wednesday before Monday",
			dateTime:  "2023-10-02 09:00:00", // Monday
			dayOfWeek: "Wednesday",
			want:      "The previous occurrence of Wednesday before 2023-10-02 09:00:00 +0000 is 2023-09-27 09:00:00 +0000.",
		},
		{
			desc:      "Missing dateTime",
			dateTime:  "",
			dayOfWeek: "Monday",
			wantErr:   true,
		},
		{
			desc:      "Missing dayOfWeek",
			dateTime:  "2023-10-01 12:30:00",
			dayOfWeek: "",
			wantErr:   true,
		},
		{
			desc:      "Invalid dayOfWeek",
			dateTime:  "2023-10-01 12:30:00",
			dayOfWeek: "Funday",
			wantErr:   true,
		},
		{
			desc:      "Invalid dateTime format",
			dateTime:  "not-a-date",
			dayOfWeek: "Monday",
			wantErr:   true,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := &Server{
				TimeManager: &mockTmanager{},
			}
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]any{
						"dateTime":  tc.dateTime,
						"dayOfWeek": tc.dayOfWeek,
					},
				},
			}
			got, _ := s.PreviousOccurrence(ctx, req)
			if tc.wantErr {
				if got == nil || !got.IsError {
					t.Errorf("PreviousOccurrence() expected error, got = %+v", got)
				}
				return
			}
			if got == nil || len(got.Content) == 0 {
				t.Fatalf("PreviousOccurrence() got = nil or empty content")
			}
			gotTextContent, ok := got.Content[0].(mcp.TextContent)
			if !ok {
				t.Fatalf("PreviousOccurrence() got = %+v, want TextContent", got.Content[0])
			}
			if gotTextContent.Text != tc.want {
				t.Errorf("PreviousOccurrence() got = %v, want %v", gotTextContent.Text, tc.want)
			}
		})
	}
}

func TestIsWeekend(t *testing.T) {
	testCases := []struct {
		desc    string
		date    string
		want    string
		wantErr bool
	}{
		{
			desc: "Sunday is weekend",
			date: "2023-10-01",
			want: "2023-10-01 is a weekend.",
		},
		{
			desc: "Saturday is weekend",
			date: "2023-09-30",
			want: "2023-09-30 is a weekend.",
		},
		{
			desc: "Monday is not weekend",
			date: "2023-10-02",
			want: "2023-10-02 is not a weekend.",
		},
		{
			desc:    "Empty input returns error",
			date:    "",
			wantErr: true,
		},
		{
			desc:    "Invalid date format returns error",
			date:    "not-a-date",
			wantErr: true,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := &Server{
				TimeManager: &mockTmanager{},
			}
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]any{
						"dateTime": tc.date,
					},
				},
			}
			got, _ := s.IsWeekend(ctx, req)
			if tc.wantErr {
				if got == nil || !got.IsError {
					t.Errorf("IsWeekend() expected error, got = %+v", got)
				}
				return
			}
			if got == nil || len(got.Content) == 0 {
				t.Fatalf("IsWeekend() got = nil or empty content")
			}
			gotTextContent, ok := got.Content[0].(mcp.TextContent)
			if !ok {
				t.Fatalf("IsWeekend() got = %+v, want TextContent", got.Content[0])
			}
			if gotTextContent.Text != tc.want {
				t.Errorf("IsWeekend() got = %v, want %v", gotTextContent.Text, tc.want)
			}
		})
	}
}
func TestIsWeekday(t *testing.T) {
	testCases := []struct {
		desc    string
		date    string
		want    string
		wantErr bool
	}{
		{
			desc: "Monday is a weekday",
			date: "2023-10-02",
			want: "2023-10-02 is a weekday.",
		},
		{
			desc: "Friday is a weekday",
			date: "2023-10-06",
			want: "2023-10-06 is a weekday.",
		},
		{
			desc: "Sunday is not a weekday",
			date: "2023-10-01",
			want: "2023-10-01 is not a weekday.",
		},
		{
			desc: "Saturday is not a weekday",
			date: "2023-09-30",
			want: "2023-09-30 is not a weekday.",
		},
		{
			desc:    "Empty input returns error",
			date:    "",
			wantErr: true,
		},
		{
			desc:    "Invalid date format returns error",
			date:    "not-a-date",
			wantErr: true,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := &Server{
				TimeManager: &mockTmanager{},
			}
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]any{
						"dateTime": tc.date,
					},
				},
			}
			got, _ := s.IsWeekday(ctx, req)
			if tc.wantErr {
				if got == nil || !got.IsError {
					t.Errorf("IsWeekday() expected error, got = %+v", got)
				}
				return
			}
			if got == nil || len(got.Content) == 0 {
				t.Fatalf("IsWeekday() got = nil or empty content")
			}
			gotTextContent, ok := got.Content[0].(mcp.TextContent)
			if !ok {
				t.Fatalf("IsWeekday() got = %+v, want TextContent", got.Content[0])
			}
			if gotTextContent.Text != tc.want {
				t.Errorf("IsWeekday() got = %v, want %v", gotTextContent.Text, tc.want)
			}
		})
	}
}

func TestDaysBetween(t *testing.T) {
	testCases := []struct {
		desc       string
		firstDate  string
		secondDate string
		want       string
		wantErr    bool
	}{
		{
			desc:       "1 day between consecutive dates",
			firstDate:  "2023-10-01",
			secondDate: "2023-10-02",
			want:       "There are 1 days between 2023-10-01 and 2023-10-02.",
		},
		{
			desc:       "0 days between same date",
			firstDate:  "2023-10-01",
			secondDate: "2023-10-01",
			want:       "There are 0 days between 2023-10-01 and 2023-10-01.",
		},
		{
			desc:       "Negative days when second date before first",
			firstDate:  "2023-10-02",
			secondDate: "2023-10-01",
			want:       "There are -1 days between 2023-10-02 and 2023-10-01.",
		},
		{
			desc:       "Multiple days between dates",
			firstDate:  "2023-09-28",
			secondDate: "2023-10-01",
			want:       "There are 3 days between 2023-09-28 and 2023-10-01.",
		},
		{
			desc:       "Missing firstDateTime returns error",
			firstDate:  "",
			secondDate: "2023-10-01",
			wantErr:    true,
		},
		{
			desc:       "Missing secondDateTime returns error",
			firstDate:  "2023-10-01",
			secondDate: "",
			wantErr:    true,
		},
		{
			desc:       "Invalid firstDateTime format returns error",
			firstDate:  "not-a-date",
			secondDate: "2023-10-01",
			wantErr:    true,
		},
		{
			desc:       "Invalid secondDateTime format returns error",
			firstDate:  "2023-10-01",
			secondDate: "not-a-date",
			wantErr:    true,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := &Server{
				TimeManager: &mockTmanager{},
			}
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]any{
						"firstDateTime":  tc.firstDate,
						"secondDateTime": tc.secondDate,
					},
				},
			}
			got, _ := s.DaysBetween(ctx, req)
			if tc.wantErr {
				if got == nil || !got.IsError {
					t.Errorf("DaysBetween() expected error, got = %+v", got)
				}
				return
			}
			if got == nil || len(got.Content) == 0 {
				t.Fatalf("DaysBetween() got = nil or empty content")
			}
			gotTextContent, ok := got.Content[0].(mcp.TextContent)
			if !ok {
				t.Fatalf("DaysBetween() got = %+v, want TextContent", got.Content[0])
			}
			if gotTextContent.Text != tc.want {
				t.Errorf("DaysBetween() got = %v, want %v", gotTextContent.Text, tc.want)
			}
		})
	}
}
