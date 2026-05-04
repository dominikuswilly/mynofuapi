package utils

import "time"

const DefaultTimeFormat = "2006-01-02 15:04:05"
const DefaultTimeZoneOffset = 7 * time.Hour

// FormatTime formats a time.Time object to the global standard format with UTC+7 adjustment
func FormatTime(t time.Time) string {
	return t.Add(DefaultTimeZoneOffset).Format(DefaultTimeFormat)
}
