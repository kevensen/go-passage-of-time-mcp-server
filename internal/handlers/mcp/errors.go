package mcp

type NilTimeOptsError struct{}

func (e *NilTimeOptsError) Error() string {
	return "input time options cannot be nil"
}

func NewNilTimeOptsError() *NilTimeOptsError {
	return &NilTimeOptsError{}
}

type NilInputTime struct{}

func (e *NilInputTime) Error() string {
	return "input time cannot be empty"
}

func NewNilInputTime() *NilInputTime {
	return &NilInputTime{}
}

type InvalidTimeFormatError struct {
	Input string
}

func (e *InvalidTimeFormatError) Error() string {
	return "failed to parse time: \"" + e.Input + "\". Format must be YYYY-MM-DD HH:MM:SS for date/time or YYYY-MM-DD for date only"
}

func NewInvalidTimeFormatError(input string) *InvalidTimeFormatError {
	return &InvalidTimeFormatError{
		Input: input,
	}
}

type TimeZoneLoadError struct {
	TimeZone string
	Err      error
}

func (e *TimeZoneLoadError) Error() string {
	return "failed to load time zone \"" + e.TimeZone + "\": " + e.Err.Error()
}

func NewTimeZoneLoadError(timeZone string, err error) *TimeZoneLoadError {
	return &TimeZoneLoadError{
		TimeZone: timeZone,
		Err:      err,
	}
}
