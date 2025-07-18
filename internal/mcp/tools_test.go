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
