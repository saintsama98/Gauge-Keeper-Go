// Package validate is the keeper's safety stage, modeled on the Chaos-Agents
// check()→execute() pattern (MIT, https://github.com/ChaosLabsInc/chaos-agents):
// before any allocation reaches the executor, every hard bound is re-derived
// off-chain from the same VaultState the brain saw. It is defense in depth, the
// vault rejects a bad allocation on-chain anyway, but re-checking here turns a
// reverted (gas-burning, alerting) transaction into a skipped one.
//
// On failure the caller skips the reallocation and falls back to the last-good /
// idle targets; a passing Result means every structural invariant holds:
// caps, idle floor, churn bound, quarantine, completeness.
//
// Optional eth_call/Tenderly simulation and expiry/min-delay gating are not
// implemented: they require the live ChainReader/backend, which is not wired
// while the keeper runs against StubReader.
package validate

import (
	"fmt"
	"log/slog"

	"github.com/saintsama98/Gauge-Keeper-Go/pkg/types"
)

// Result is the outcome of a validation pass. OK is true only when Violations
// is empty.
type Result struct {
	OK         bool
	Violations []string
}

// Validator re-checks a proposed allocation against vault state.
type Validator struct {
	log *slog.Logger
}

func New(log *slog.Logger) *Validator {
	if log == nil {
		log = slog.Default()
	}
	return &Validator{log: log}
}

// Check re-derives every hard bound from state and reports any the allocation
// breaks. It is pure: same (state, alloc) → same Result.
func (v *Validator) Check(state *types.VaultState, alloc *types.Allocation) Result {
	var violations []string

	known := make(map[types.StrategyID]bool, len(state.Strategies))
	var sumTargets, sumDelta, forced, keptBase int

	for _, s := range state.Strategies {
		known[s.ID] = true
		t, ok := alloc.Targets[s.ID]
		if !ok {
			violations = append(violations, fmt.Sprintf("strategy %x missing from allocation", s.ID[:4]))
			continue
		}
		if t > types.BpsDenominator {
			violations = append(violations, fmt.Sprintf("strategy %x target %d > 100%%", s.ID[:4], t))
		}
		if s.Quarantined && t != 0 {
			violations = append(violations, fmt.Sprintf("quarantined strategy %x has nonzero target %d", s.ID[:4], t))
		}
		if t > s.CapBps {
			violations = append(violations, fmt.Sprintf("strategy %x target %d > owner cap %d", s.ID[:4], t, s.CapBps))
		}
		sumTargets += int(t)
		// Churn is counted as in the vault's rebalance, which skips quarantined
		// strategies. Movement the vault's own checks force on the current
		// targets is allowed past the bound: excess over the owner cap first,
		// then excess over the investable budget (added below).
		if s.Quarantined {
			continue
		}
		sumDelta += absInt(int(t) - int(s.TargetBps))
		base := min(int(s.TargetBps), int(s.CapBps))
		forced += int(s.TargetBps) - base
		keptBase += base
	}

	// Any id in the allocation that the vault doesn't know about is invalid.
	for id := range alloc.Targets {
		if !known[id] {
			violations = append(violations, fmt.Sprintf("allocation references unknown strategy %x", id[:4]))
		}
	}

	// Idle floor: invested fraction may not exceed 100% − idle reserve.
	investable := int(types.BpsDenominator) - int(state.IdleReserveBps)
	if investable < 0 {
		investable = 0
	}
	if sumTargets > investable {
		violations = append(violations, fmt.Sprintf("sum targets %d > investable %d (idle reserve %d)", sumTargets, investable, state.IdleReserveBps))
	}

	// Churn bound: total movement may not exceed the per-call delta cap, or the
	// forced movement when that is larger.
	forced += max(0, keptBase-investable)
	if state.MaxRebalanceDeltaBps > 0 && sumDelta > max(int(state.MaxRebalanceDeltaBps), forced) {
		violations = append(violations, fmt.Sprintf("churn Σ|Δ| %d > max rebalance delta %d", sumDelta, state.MaxRebalanceDeltaBps))
	}

	return Result{OK: len(violations) == 0, Violations: violations}
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
