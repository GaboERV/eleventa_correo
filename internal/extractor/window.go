package extractor

import (
	"time"
)

// CalculateWindow determines the reporting window (Monday to Monday).
// Returns start (inclusive), end (exclusive), and whether there's work to do.
//
// The window always goes from one Monday 00:00 to the next Monday 00:00.
// If lastReportEnd is empty (first run), we look back one week from the most recent Monday.
func CalculateWindow(lastReportEnd string) (start, end time.Time, hasWork bool) {
	loc, err := time.LoadLocation("America/Mexico_City")
	if err != nil {
		// Fallback to UTC-6 if timezone data is not available
		loc = time.FixedZone("CST", -6*3600)
	}

	now := time.Now().In(loc)

	// Find the most recent Monday at 00:00
	end = mostRecentMonday(now)

	// If today is Monday and it's before the scheduled time, use the previous Monday
	if now.Weekday() == time.Monday && now.Hour() < 4 {
		end = end.AddDate(0, 0, -7)
	}

	// Determine start of window
	if lastReportEnd == "" {
		// First run: look back one week
		start = end.AddDate(0, 0, -7)
	} else {
		parsed, err := time.ParseInLocation(time.RFC3339, lastReportEnd, loc)
		if err != nil {
			// If we can't parse the last report end, look back one week
			start = end.AddDate(0, 0, -7)
		} else {
			start = parsed
		}
	}

	// No work if window is empty or negative
	hasWork = start.Before(end)
	return start, end, hasWork
}

// mostRecentMonday returns the most recent Monday at 00:00:00 in the given location.
func mostRecentMonday(t time.Time) time.Time {
	// Go's time.Monday = 1
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	daysBack := int(weekday) - int(time.Monday)
	monday := t.AddDate(0, 0, -daysBack)
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, t.Location())
}
