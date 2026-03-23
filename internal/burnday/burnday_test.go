package burnday

import (
	"testing"
	"time"
)

func TestWindow_AfterSix(t *testing.T) {
	// 2026-03-23 14:32 → burn day started today at 06:00
	now := time.Date(2026, 3, 23, 14, 32, 0, 0, time.Local)
	w := CurrentWindow(now)

	expectStart := time.Date(2026, 3, 23, 6, 0, 0, 0, time.Local)
	if !w.Start.Equal(expectStart) {
		t.Errorf("Start = %v, want %v", w.Start, expectStart)
	}
	if !w.End.Equal(now) {
		t.Errorf("End = %v, want %v", w.End, now)
	}
}

func TestWindow_BeforeSix(t *testing.T) {
	// 2026-03-23 03:00 → still in yesterday's burn day (started 03-22 06:00)
	now := time.Date(2026, 3, 23, 3, 0, 0, 0, time.Local)
	w := CurrentWindow(now)

	expectStart := time.Date(2026, 3, 22, 6, 0, 0, 0, time.Local)
	if !w.Start.Equal(expectStart) {
		t.Errorf("Start = %v, want %v", w.Start, expectStart)
	}
}

func TestWindow_ExactlySix(t *testing.T) {
	// 2026-03-23 06:00 → burn day just started today
	now := time.Date(2026, 3, 23, 6, 0, 0, 0, time.Local)
	w := CurrentWindow(now)

	expectStart := time.Date(2026, 3, 23, 6, 0, 0, 0, time.Local)
	if !w.Start.Equal(expectStart) {
		t.Errorf("Start = %v, want %v", w.Start, expectStart)
	}
}

func TestYesterdayWindow(t *testing.T) {
	// 2026-03-23 14:32 → yesterday's burn day: 03-22 06:00 ~ 03-23 05:59
	now := time.Date(2026, 3, 23, 14, 32, 0, 0, time.Local)
	w := YesterdayWindow(now)

	expectStart := time.Date(2026, 3, 22, 6, 0, 0, 0, time.Local)
	expectEnd := time.Date(2026, 3, 23, 5, 59, 59, 0, time.Local)
	if !w.Start.Equal(expectStart) {
		t.Errorf("Start = %v, want %v", w.Start, expectStart)
	}
	if !w.End.Equal(expectEnd) {
		t.Errorf("End = %v, want %v", w.End, expectEnd)
	}
}

func TestYesterdayWindow_BeforeSix(t *testing.T) {
	// 2026-03-23 03:00 → current burn day is 03-22, yesterday is 03-21
	now := time.Date(2026, 3, 23, 3, 0, 0, 0, time.Local)
	w := YesterdayWindow(now)

	expectStart := time.Date(2026, 3, 21, 6, 0, 0, 0, time.Local)
	expectEnd := time.Date(2026, 3, 22, 5, 59, 59, 0, time.Local)
	if !w.Start.Equal(expectStart) {
		t.Errorf("Start = %v, want %v", w.Start, expectStart)
	}
	if !w.End.Equal(expectEnd) {
		t.Errorf("End = %v, want %v", w.End, expectEnd)
	}
}

func TestDateWindow(t *testing.T) {
	w := DateWindow(2026, 3, 22)

	expectStart := time.Date(2026, 3, 22, 6, 0, 0, 0, time.Local)
	expectEnd := time.Date(2026, 3, 23, 5, 59, 59, 0, time.Local)
	if !w.Start.Equal(expectStart) {
		t.Errorf("Start = %v, want %v", w.Start, expectStart)
	}
	if !w.End.Equal(expectEnd) {
		t.Errorf("End = %v, want %v", w.End, expectEnd)
	}
}

func TestWindow_Contains(t *testing.T) {
	now := time.Date(2026, 3, 23, 14, 32, 0, 0, time.Local)
	w := CurrentWindow(now)

	inside := time.Date(2026, 3, 23, 10, 0, 0, 0, time.Local)
	outside := time.Date(2026, 3, 23, 5, 0, 0, 0, time.Local)

	if !w.Contains(inside) {
		t.Error("Contains(inside) = false, want true")
	}
	if w.Contains(outside) {
		t.Error("Contains(outside) = true, want false")
	}
}
