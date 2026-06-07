// Package trigger is the heartbeat layer: it decides WHEN cadence-based actions
// are due. Threshold/event triggers (APY-drift, hysteresis) belong here too —
// add them as additional Scheduler implementations or wrap this one.
package trigger

import "time"

// Scheduler answers whether a named periodic task is due, and records when one
// runs. The keeper uses it to throttle harvest, guard-checkpoint and rebalance.
type Scheduler interface {
	Due(task string, now time.Time) bool
	Mark(task string, now time.Time)
}

// IntervalScheduler fires each task at most once per configured interval.
type IntervalScheduler struct {
	intervals map[string]time.Duration
	lastRun   map[string]time.Time
}

func NewIntervalScheduler(intervals map[string]time.Duration) *IntervalScheduler {
	return &IntervalScheduler{intervals: intervals, lastRun: make(map[string]time.Time)}
}

func (s *IntervalScheduler) Due(task string, now time.Time) bool {
	iv, ok := s.intervals[task]
	if !ok {
		return false
	}
	last, ran := s.lastRun[task]
	if !ran {
		return true // never run → due immediately
	}
	return now.Sub(last) >= iv
}

func (s *IntervalScheduler) Mark(task string, now time.Time) {
	s.lastRun[task] = now
}
