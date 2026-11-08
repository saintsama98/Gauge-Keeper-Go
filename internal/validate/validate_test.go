package validate

import (
	"strings"
	"testing"

	"github.com/saintsama98/Gauge-Keeper-Go/pkg/types"
)

func vid(b byte) types.StrategyID {
	var id types.StrategyID
	id[0] = b
	return id
}

// baseState: two strategies, cap 5000 each, current targets 3000/2000, 10% idle.
func baseState() *types.VaultState {
	return &types.VaultState{
		IdleReserveBps: 1000,
		Strategies: []types.StrategyState{
			{ID: vid(1), CapBps: 5000, TargetBps: 3000},
			{ID: vid(2), CapBps: 5000, TargetBps: 2000},
		},
	}
}

func alloc(pairs map[byte]types.Bps) *types.Allocation {
	t := make(map[types.StrategyID]types.Bps, len(pairs))
	for b, v := range pairs {
		t[vid(b)] = v
	}
	return &types.Allocation{Targets: t}
}

func TestValidPasses(t *testing.T) {
	v := New(nil)
	res := v.Check(baseState(), alloc(map[byte]types.Bps{1: 5000, 2: 4000})) // sum 9000 ≤ 9000
	if !res.OK {
		t.Errorf("valid allocation rejected: %v", res.Violations)
	}
}

func TestRejectsOverCap(t *testing.T) {
	v := New(nil)
	res := v.Check(baseState(), alloc(map[byte]types.Bps{1: 6000, 2: 1000})) // 6000 > cap 5000
	if res.OK {
		t.Fatal("over-cap allocation accepted")
	}
	if !containsSubstr(res.Violations, "owner cap") {
		t.Errorf("expected owner-cap violation, got %v", res.Violations)
	}
}

func TestRejectsIdleFloorBreach(t *testing.T) {
	v := New(nil)
	res := v.Check(baseState(), alloc(map[byte]types.Bps{1: 5000, 2: 5000})) // sum 10000 > 9000
	if res.OK {
		t.Fatal("idle-floor breach accepted")
	}
	if !containsSubstr(res.Violations, "idle reserve") {
		t.Errorf("expected idle-reserve violation, got %v", res.Violations)
	}
}

func TestRejectsQuarantineViolation(t *testing.T) {
	st := baseState()
	st.Strategies[0].Quarantined = true
	v := New(nil)
	res := v.Check(st, alloc(map[byte]types.Bps{1: 100, 2: 2000}))
	if res.OK {
		t.Fatal("nonzero target on quarantined strategy accepted")
	}
	if !containsSubstr(res.Violations, "quarantined") {
		t.Errorf("expected quarantine violation, got %v", res.Violations)
	}
}

func TestRejectsChurnBreach(t *testing.T) {
	st := baseState()
	st.MaxRebalanceDeltaBps = 500 // current 3000/2000
	v := New(nil)
	// Σ|Δ| = |5000-3000| + |0-2000| = 4000 > 500
	res := v.Check(st, alloc(map[byte]types.Bps{1: 5000, 2: 0}))
	if res.OK {
		t.Fatal("churn breach accepted")
	}
	if !containsSubstr(res.Violations, "churn") {
		t.Errorf("expected churn violation, got %v", res.Violations)
	}
}

func TestRejectsMissingAndUnknown(t *testing.T) {
	v := New(nil)
	// vid(2) missing; vid(9) unknown.
	res := v.Check(baseState(), alloc(map[byte]types.Bps{1: 3000, 9: 1000}))
	if res.OK {
		t.Fatal("missing/unknown allocation accepted")
	}
	if !containsSubstr(res.Violations, "missing") || !containsSubstr(res.Violations, "unknown") {
		t.Errorf("expected missing+unknown violations, got %v", res.Violations)
	}
}

func containsSubstr(ss []string, sub string) bool {
	for _, s := range ss {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// A quarantined strategy's exit is outside the churn bound, as in the vault's
// rebalance, so draining it in one call is valid.
func TestAcceptsQuarantineExitPastChurnBound(t *testing.T) {
	st := baseState()
	st.MaxRebalanceDeltaBps = 100
	st.Strategies[0].Quarantined = true
	res := New(nil).Check(st, alloc(map[byte]types.Bps{1: 0, 2: 2100}))
	if !res.OK {
		t.Fatalf("quarantine exit rejected: %v", res.Violations)
	}
}

// Movement forced by an owner cap below the current target is allowed past
// the bound, but nothing on top of it.
func TestChurnAllowsOnlyForcedMovementPastBound(t *testing.T) {
	st := baseState()
	st.MaxRebalanceDeltaBps = 100
	st.Strategies[0].CapBps = 2500 // forces 500 bps off strategy 1
	if res := New(nil).Check(st, alloc(map[byte]types.Bps{1: 2500, 2: 2000})); !res.OK {
		t.Fatalf("forced move rejected: %v", res.Violations)
	}
	res := New(nil).Check(st, alloc(map[byte]types.Bps{1: 2500, 2: 2200}))
	if res.OK || !containsSubstr(res.Violations, "churn") {
		t.Fatalf("discretionary move on top of a forced one accepted: %v", res.Violations)
	}
}

// Movement forced by an idle reserve raised above the current targets is
// allowed past the bound.
func TestChurnAllowsForcedBudgetCut(t *testing.T) {
	st := baseState()
	st.MaxRebalanceDeltaBps = 100
	st.IdleReserveBps = 6000 // investable 4000 against 5000 held
	if res := New(nil).Check(st, alloc(map[byte]types.Bps{1: 2000, 2: 2000})); !res.OK {
		t.Fatalf("forced budget cut rejected: %v", res.Violations)
	}
}
