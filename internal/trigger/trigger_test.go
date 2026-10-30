package trigger

import (
	"testing"
	"time"
)

func TestIntervalScheduler(t *testing.T) {
	t0 := time.Unix(1_700_000_000, 0)
	s := NewIntervalScheduler(map[string]time.Duration{"harvest": time.Hour})

	if !s.Due("harvest", t0) {
		t.Fatal("a task that never ran must be due immediately")
	}
	s.Mark("harvest", t0)
	if s.Due("harvest", t0.Add(59*time.Minute)) {
		t.Fatal("task due before its interval elapsed")
	}
	if !s.Due("harvest", t0.Add(time.Hour)) {
		t.Fatal("task not due once its interval elapsed")
	}
	if s.Due("unknown", t0) {
		t.Fatal("a task without an interval must never be due")
	}
}
