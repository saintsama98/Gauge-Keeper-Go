package brain

import (
	"context"
	"log/slog"
	"sort"

	"github.com/saintsama98/Gauge-Keeper-Go/pkg/types"
)

// WaterFillDecider is the default brain: a risk-gated, yield-maximizing
// water-fill allocator. Yield is the objective; risk is the gate. It maximizes
// effectiveYield(APY) − Lambda·EL per strategy while EL also caps each venue
// (effectiveCapBps) and the liquidity gate caps it again. Idle reserve, owner
// cap, liquidity cap and churn bound are all hard ceilings the fill never
// crosses. It implements brain.Decider; StubDecider stays the default.
type WaterFillDecider struct {
	provider RiskProvider
	model    RiskModel
	log      *slog.Logger
}

func NewWaterFillDecider(r RiskProvider, m RiskModel, log *slog.Logger) *WaterFillDecider {
	if log == nil {
		log = slog.Default()
	}
	return &WaterFillDecider{provider: r, model: m, log: log}
}

// slot is the per-strategy working state through the allocation pipeline.
type slot struct {
	id     types.StrategyID
	cap    types.Bps // risk- and liquidity-adjusted ceiling
	hard   types.Bps // ceiling the vault enforces: owner cap, or 0 when quarantined
	yield  float64
	el     float64
	target types.Bps // current on-chain target, then the base the churn clamp moves from
	fill   types.Bps // water-fill proposal, before hysteresis and the churn clamp
	alloc  types.Bps // proposed new target
}

func (d *WaterFillDecider) Decide(_ context.Context, state *types.VaultState) (*types.Allocation, error) {
	slots := make([]slot, len(state.Strategies))

	// Step 1, per-strategy risk pass: derive each venue's effective cap, EL and
	// scoring yield. Quarantined and killed venues end up with cap 0.
	for i, s := range state.Strategies {
		sl := slot{id: s.ID, target: s.TargetBps, hard: s.CapBps}
		if s.Quarantined {
			sl.hard = 0
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

	// Step 2: water-fill the investable budget in StepBps chunks, always to the
	// eligible strategy with the highest marginal score yield − Lambda·EL.
	budget := types.Bps(0)
	if state.IdleReserveBps < types.BpsDenominator {
		budget = types.BpsDenominator - state.IdleReserveBps
	}
	investable := budget
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

	for i := range slots {
		slots[i].fill = slots[i].alloc
	}

	// Step 3, forced moves: the vault rejects the whole vector if any target is
	// nonzero for a quarantined venue, above the owner cap, or if the sum is
	// above the budget. Current targets that break these are moved first,
	// toward the fill. A quarantined venue's exit is outside the vault's churn
	// bound, as in its rebalance; the other forced moves are charged to it.
	var forced types.Bps
	for i := range slots {
		if slots[i].target > slots[i].hard {
			if slots[i].hard != 0 || !state.Strategies[i].Quarantined {
				forced += slots[i].target - slots[i].hard
			}
			slots[i].target = slots[i].hard
		}
	}
	forced += trimToward(slots, func(s *slot) *types.Bps { return &s.target }, investable)

	// Step 4, hysteresis: ignore sub-threshold moves to avoid churn for noise,
	// unless the target is above the venue's gated cap. Undo snaps that lifted
	// a venue above its fill if together they push the sum over the budget.
	for i := range slots {
		if absDeltaBps(slots[i].alloc, slots[i].target) < d.model.MinDeltaBps && slots[i].target <= slots[i].cap {
			slots[i].alloc = slots[i].target
		}
	}
	trimToward(slots, func(s *slot) *types.Bps { return &s.alloc }, investable)

	// Step 5, churn clamp: if total movement exceeds what is left of the
	// per-call bound, scale every delta proportionally back toward the base.
	if bound := state.MaxRebalanceDeltaBps; bound != 0 {
		if forced >= bound {
			for i := range slots {
				slots[i].alloc = slots[i].target
			}
		} else {
			clampChurn(slots, bound-forced)
		}
	}

	// Step 6: emit every strategy (quarantined explicitly 0; killed 0 unless
	// the churn bound is still draining it).
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

// trimToward lowers the field picked by get toward each slot's fill until the
// slots sum to at most budget, lowest score first, ties to the lower index. It
// returns the amount removed. The fill sums to at most budget, so this always
// succeeds.
func trimToward(slots []slot, get func(*slot) *types.Bps, budget types.Bps) types.Bps {
	var sum int
	for i := range slots {
		sum += int(*get(&slots[i]))
	}
	excess := sum - int(budget)
	if excess <= 0 {
		return 0
	}
	order := make([]int, len(slots))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		return slots[order[a]].yield-slots[order[a]].el < slots[order[b]].yield-slots[order[b]].el
	})
	removed := 0
	for _, i := range order {
		v := get(&slots[i])
		if excess == 0 || *v <= slots[i].fill {
			continue
		}
		cut := min(int(*v-slots[i].fill), excess)
		*v -= types.Bps(cut)
		excess -= cut
		removed += cut
	}
	return types.Bps(removed)
}

// absDeltaBps is |a - b| without unsigned underflow.
func absDeltaBps(a, b types.Bps) types.Bps {
	if a >= b {
		return a - b
	}
	return b - a
}
