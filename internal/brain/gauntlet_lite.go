package brain

import (
	"context"
	"log/slog"

	"github.com/vault-router-keeper/pkg/types"
)

// GauntletLiteDecider is the default brain: a risk-gated, yield-maximizing
// water-fill allocator. Yield is the objective; risk is the gate. It maximizes
// effectiveYield(APY) − Lambda·EL per strategy while EL also caps each venue
// (effectiveCapBps) and the liquidity gate caps it again. Idle reserve, owner
// cap, liquidity cap and churn bound are all hard ceilings the fill never
// crosses. It implements brain.Decider; StubDecider stays the default.
type GauntletLiteDecider struct {
	provider RiskProvider
	model    RiskModel
	log      *slog.Logger
}

func NewGauntletLiteDecider(r RiskProvider, m RiskModel, log *slog.Logger) *GauntletLiteDecider {
	if log == nil {
		log = slog.Default()
	}
	return &GauntletLiteDecider{provider: r, model: m, log: log}
}

// slot is the per-strategy working state through the allocation pipeline.
type slot struct {
	id     types.StrategyID
	cap    types.Bps // risk- and liquidity-adjusted ceiling
	yield  float64
	el     float64
	target types.Bps // current on-chain target (for hysteresis / churn)
	alloc  types.Bps // proposed new target
}

func (d *GauntletLiteDecider) Decide(_ context.Context, state *types.VaultState) (*types.Allocation, error) {
	slots := make([]slot, len(state.Strategies))

	// Step 1 — per-strategy risk pass: derive each venue's effective cap, EL and
	// scoring yield. Quarantined and killed venues end up with cap 0.
	for i, s := range state.Strategies {
		sl := slot{id: s.ID, target: s.TargetBps}
		if s.Quarantined {
			sl.el = 1.0
			sl.cap = 0
		} else {
			ri, known := d.provider.Risk(s.ID)
			if known {
				sl.el = d.model.expectedLoss(ri)
			} else {
				sl.el = d.model.UnknownEL
			}
			sl.cap = d.model.effectiveCapBps(s.CapBps, sl.el)
			// Liquidity / Pendle queue gate: an illiquid venue may hold only
			// what the async withdraw queue can cover before maturity.
			if known && ri.HasLiquidityCap && ri.LiquidityCapBps < sl.cap {
				sl.cap = ri.LiquidityCapBps
			}
			sl.yield = d.model.effectiveYield(s.APY)
		}
		slots[i] = sl
	}

	// Step 2 — water-fill the investable budget in StepBps chunks, always to the
	// eligible strategy with the highest marginal score yield − Lambda·EL.
	budget := types.Bps(0)
	if state.IdleReserveBps < types.BpsDenominator {
		budget = types.BpsDenominator - state.IdleReserveBps
	}
	step := d.model.StepBps
	if step == 0 {
		step = 1 // guard: never spin forever on a zero step
	}
	for budget > 0 {
		best := -1
		var bestScore float64
		for i := range slots {
			if slots[i].alloc >= slots[i].cap {
				continue
			}
			score := slots[i].yield - d.model.Lambda*slots[i].el
			if best == -1 || score > bestScore {
				best, bestScore = i, score
			}
		}
		if best == -1 || bestScore < d.model.MinScore {
			break
		}
		chunk := step
		if chunk > budget {
			chunk = budget
		}
		if room := slots[best].cap - slots[best].alloc; chunk > room {
			chunk = room // never overshoot the hard cap
		}
		slots[best].alloc += chunk
		budget -= chunk
	}

	// Step 3 — hysteresis: ignore sub-threshold moves to avoid churn for noise.
	for i := range slots {
		if absDeltaBps(slots[i].alloc, slots[i].target) < d.model.MinDeltaBps {
			slots[i].alloc = slots[i].target
		}
	}

	// Step 4 — churn clamp: if total movement exceeds the per-call bound, scale
	// every delta proportionally back toward the current targets.
	clampChurn(slots, state.MaxRebalanceDeltaBps)

	// Step 5 — emit every strategy (killed/quarantined explicitly 0).
	targets := make(map[types.StrategyID]types.Bps, len(slots))
	for i := range slots {
		targets[slots[i].id] = slots[i].alloc
	}
	return &types.Allocation{Targets: targets}, nil
}

// clampChurn scales every delta proportionally back toward the current targets
// when total movement exceeds maxDelta. A zero maxDelta means no bound.
//
// Increases and decreases are scaled as two separate budgets in integer bps.
// The increase budget is rounded down and the decrease budget takes the rest of
// the bound, so total movement never exceeds maxDelta and the invested sum never
// rises above what an exact proportional scale would give. Each budget is then
// split across its slots by largest remainder, so no slot moves past its own
// proposal.
func clampChurn(slots []slot, maxDelta types.Bps) {
	if maxDelta == 0 {
		return
	}
	var up, down int
	for i := range slots {
		if slots[i].alloc >= slots[i].target {
			up += int(slots[i].alloc - slots[i].target)
		} else {
			down += int(slots[i].target - slots[i].alloc)
		}
	}
	total := up + down
	if total <= int(maxDelta) {
		return
	}
	upBudget := up * int(maxDelta) / total
	downBudget := min(down, int(maxDelta)-upBudget)

	upMoves := splitBudget(slots, upBudget, up, true)
	downMoves := splitBudget(slots, downBudget, down, false)
	for i := range slots {
		if slots[i].alloc >= slots[i].target {
			slots[i].alloc = slots[i].target + types.Bps(upMoves[i])
		} else {
			slots[i].alloc = slots[i].target - types.Bps(downMoves[i])
		}
	}
}

// splitBudget divides budget across the slots moving in one direction (up when
// increasing, otherwise down) in proportion to each slot's delta, by largest
// remainder. Ties go to the lower index, so the result is deterministic.
func splitBudget(slots []slot, budget, total int, increasing bool) []int {
	moves := make([]int, len(slots))
	if total == 0 {
		return moves
	}
	rem := make([]int, len(slots))
	given := 0
	for i := range slots {
		var delta int
		switch {
		case increasing && slots[i].alloc >= slots[i].target:
			delta = int(slots[i].alloc - slots[i].target)
		case !increasing && slots[i].alloc < slots[i].target:
			delta = int(slots[i].target - slots[i].alloc)
		default:
			continue
		}
		moves[i] = delta * budget / total
		rem[i] = delta * budget % total
		given += moves[i]
	}
	for ; given < budget; given++ {
		best := -1
		for i := range slots {
			if rem[i] > 0 && (best == -1 || rem[i] > rem[best]) {
				best = i
			}
		}
		moves[best]++
		rem[best] = 0
	}
	return moves
}

// absDeltaBps is |a - b| without unsigned underflow.
func absDeltaBps(a, b types.Bps) types.Bps {
	if a >= b {
		return a - b
	}
	return b - a
}
