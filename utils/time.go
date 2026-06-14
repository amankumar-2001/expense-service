package utils

import "time"

// MonthBounds returns the half-open interval [from, to) covering the calendar
// month that contains t: the first instant of the month and the first instant of
// the next month.
func MonthBounds(t time.Time) (from, to time.Time) {
	y, m, _ := t.Date()
	from = time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
	to = from.AddDate(0, 1, 0)
	return from, to
}

// ParseMonth parses a "YYYY-MM" string into the first instant of that month in
// loc. An empty string returns the month containing now.
func ParseMonth(s string, now time.Time, loc *time.Location) (time.Time, error) {
	if s == "" {
		from, _ := MonthBounds(now.In(loc))
		return from, nil
	}
	return time.ParseInLocation("2006-01", s, loc)
}

// MonthLabel formats the month for display, e.g. "June 2026".
func MonthLabel(t time.Time) string { return t.Format("January 2006") }

// NextDeductionInDays returns how many whole days from `from` until the next
// occurrence of a monthly deduction on day-of-month `deductDay`. If the deduction
// day is today it returns 0; if it has already passed this month it rolls to next
// month. deductDay values past the end of a short month clamp to the last day.
func NextDeductionInDays(from time.Time, deductDay int) int {
	y, m, d := from.Date()
	loc := from.Location()
	today := time.Date(y, m, d, 0, 0, 0, 0, loc)

	next := clampedDate(y, int(m), deductDay, loc)
	if next.Before(today) {
		next = clampedDate(y, int(m)+1, deductDay, loc)
	}
	return int(next.Sub(today).Hours() / 24)
}

// clampedDate builds a date for the given year/month/day, clamping day to the
// last valid day of that month (e.g. day 31 in February becomes the 28th/29th).
func clampedDate(year, month, day int, loc *time.Location) time.Time {
	// Normalise month overflow (month 13 → next January).
	first := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
	lastDay := first.AddDate(0, 1, -1).Day()
	if day > lastDay {
		day = lastDay
	}
	if day < 1 {
		day = 1
	}
	return time.Date(first.Year(), first.Month(), day, 0, 0, 0, 0, loc)
}
