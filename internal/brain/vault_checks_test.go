package brain

import (
	"context"
	"testing"

	"github.com/saintsama98/Gauge-Keeper-Go/internal/validate"
	"github.com/saintsama98/Gauge-Keeper-Go/pkg/types"
)

// The vault's setAllocation rejects the whole vector when any target is nonzero
// for a quarantined strategy, above the owner cap, or when the sum exceeds the
// investable budget. A rejected vector leaves the targets unchanged, so the next
// cycle proposes the same vector again. These tests pin each case.

func decideOK(t *testing.T, d *WaterFillDecider, state *types.VaultState) *types.Allocation {
	t.Helper()
	a, err := d.Decide(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if res := validate.New(nil).Check(state, a); !res.OK {
		t.Fatalf("validator rejected the vector: %v", res.Violations)
	}
	budget := types.BpsDenominator - state.IdleReserveBps
	if s := sumTargets(a); s > budget {
		t.Fatalf("target sum %d above budget %d", s, budget)
	}
	for _, s := range state.Strategies {
		got := a.Targets[s.ID]
		if s.Quarantined && got != 0 {
			t.Fatalf("quarantined strategy %x at %d", s.ID[0], got)
		}
		if got > s.CapBps {
			t.Fatalf("strategy %x at %d above owner cap %d", s.ID[0], got, s.CapBps)
		}
	}
	return a
}

// A quarantined strategy holding less than MinDeltaBps must still go to zero.
func TestDecideDrainsSmallQuarantinedTarget(t *testing.T) {
	decideOK(t, newDecider(nil), &types.VaultState{
		IdleReserveBps: 1000,
		Strategies: []types.StrategyState{
			{ID: sid(1), TargetBps: 40, CapBps: 5000, Quarantined: true, APY: 0.08},
			{ID: sid(2), TargetBps: 4000, CapBps: 5000, APY: 0.05},
		},
	})
}

// An owner cap lowered by less than MinDeltaBps below the target must hold.
func TestDecideHonorsSmallOwnerCapCut(t *testing.T) {
	decideOK(t, newDecider(nil), &types.VaultState{
		IdleReserveBps: 1000,
		Strategies: []types.StrategyState{
			{ID: sid(1), TargetBps: 5000, CapBps: 4970, APY: 0.08},
			{ID: sid(2), TargetBps: 4000, CapBps: 5000, APY: 0.05},
		},
	})
}

// The churn bound must not leave a quarantined strategy partly allocated. Its
// exit is outside the bound, as in the vault's rebalance, so the other moves
// still get the whole bound.
func TestDecideQuarantineOverridesChurnBound(t *testing.T) {
	state := &types.VaultState{
		IdleReserveBps:       1000,
		MaxRebalanceDeltaBps: 100,
		Strategies: []types.StrategyState{
			{ID: sid(1), TargetBps: 3000, CapBps: 5000, Quarantined: true, APY: 0.08},
			{ID: sid(2), TargetBps: 3000, CapBps: 9000, APY: 0.05},
		},
	}
	a := decideOK(t, newDecider(nil), state)
	if got := a.Targets[sid(2)]; got != 3100 {
		t.Errorf("B = %d, want 3100: the quarantined exit is not charged to the bound", got)
	}
}

// Raising the idle reserve above the standing targets under a churn bound must
// still give a vector within the new budget.
func TestDecideIdleReserveRaiseUnderChurnBound(t *testing.T) {
	decideOK(t, newDecider(nil), &types.VaultState{
		IdleReserveBps:       2000,
		MaxRebalanceDeltaBps: 100,
		Strategies: []types.StrategyState{
			{ID: sid(1), TargetBps: 4500, CapBps: 10000, APY: 0.05},
			{ID: sid(2), TargetBps: 4500, CapBps: 10000, APY: 0.04},
		},
	})
}

// Hysteresis must not snap small decreases back while a larger increase is
// kept, which would push the sum over the budget.
func TestDecideHysteresisKeepsBudget(t *testing.T) {
	d := newDecider(map[types.StrategyID]RiskInputs{
		sid(1): {HasLiquidityCap: true, LiquidityCapBps: 2960},
	})
	decideOK(t, d, &types.VaultState{
		IdleReserveBps: 1000,
		Strategies: []types.StrategyState{
			{ID: sid(1), TargetBps: 3000, CapBps: 10000, APY: 0.04}, // fill 2960, -40
			{ID: sid(2), TargetBps: 3000, CapBps: 3060, APY: 0.05},  // fill 3060, +60
			{ID: sid(3), TargetBps: 3000, CapBps: 10000, APY: 0.03}, // fill 2980, -20
		},
	})
}
