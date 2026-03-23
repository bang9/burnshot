package burnday

import "time"

const BurnDayStartHour = 6

type Window struct {
	Start time.Time
	End   time.Time
}

func (w Window) Contains(t time.Time) bool {
	return !t.Before(w.Start) && !t.After(w.End)
}

// FormatPeriod returns "HH:MM" strings for the start and end.
func (w Window) FormatPeriod() (string, string) {
	return w.Start.Format("15:04"), w.End.Format("15:04")
}

// burnDayStart returns the start of the burn day containing the given time.
func burnDayStart(now time.Time) time.Time {
	today6am := time.Date(now.Year(), now.Month(), now.Day(), BurnDayStartHour, 0, 0, 0, now.Location())
	if now.Before(today6am) {
		return today6am.AddDate(0, 0, -1)
	}
	return today6am
}

// CurrentWindow returns the time window from the current burn day start to now.
func CurrentWindow(now time.Time) Window {
	return Window{
		Start: burnDayStart(now),
		End:   now,
	}
}

// YesterdayWindow returns the complete burn day immediately before the current one.
func YesterdayWindow(now time.Time) Window {
	currentStart := burnDayStart(now)
	prevStart := currentStart.AddDate(0, 0, -1)
	prevEnd := time.Date(currentStart.Year(), currentStart.Month(), currentStart.Day(), 5, 59, 59, 0, currentStart.Location())
	return Window{
		Start: prevStart,
		End:   prevEnd,
	}
}

// DateWindow returns the complete burn day for a given date (06:00 that day → 05:59 next day).
func DateWindow(year int, month time.Month, day int) Window {
	start := time.Date(year, month, day, BurnDayStartHour, 0, 0, 0, time.Local)
	end := time.Date(year, month, day+1, 5, 59, 59, 0, time.Local)
	return Window{Start: start, End: end}
}
