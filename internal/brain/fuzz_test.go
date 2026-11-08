package brain

import (
	"context"
	"testing"

	"github.com/saintsama98/Gauge-Keeper-Go/internal/validate"
	"github.com/saintsama98/Gauge-Keeper-Go/pkg/types"
)

// fuzzState builds a vault state of up to six strategies from raw fuzz bytes.
// Each strategy takes seven bytes: target (2), cap (2), APY (1), flags (1),
// liquidity cap (1). Targets are drawn so their sum stays within 100%.
func fuzzState(data []byte, idle, maxDelta uint16) (*types.VaultState, map[types.StrategyID]RiskInputs) {
	state := &types.VaultState{
		IdleReserveBps:       types.Bps(idle % 10_001),
		MaxRebalanceDeltaBps: types.Bps(maxDelta % 10_001),
	}
	risk := map[types.StrategyID]RiskInputs{}
	left := int(types.BpsDenominator)
	for i := 0; i+7 <= len(data) && i/7 < 6; i += 7 {
		b := data[i : i+7]
		target := (int(b[0])<<8 | int(b[1])) % (left + 1)
		left -= target
		s := types.StrategyState{
			ID:          sid(byte(i/7 + 1)),
			TargetBps:   types.Bps(target),
			CapBps:      types.Bps((int(b[2])<<8 | int(b[3])) % 10_001),
			Quarantined: b[5]&1 == 1,
			APY:         float64(b[4]) / 1000,
		}
		state.Strategies = append(state.Strategies, s)
		switch {
		case b[5]&2 == 2:
			risk[s.ID] = RiskInputs{HasVendor: true, VendorEL: float64(b[6]) / 1000}
		case b[5]&4 == 4:
			risk[s.ID] = RiskInputs{HasLiquidityCap: true, LiquidityCapBps: types.Bps(int(b[6]) * 40)}
		}
	}
	return state, risk
}

// The cap each strategy may be raised to: the owner cap after the risk and
// liquidity gates, and zero for a quarantined strategy.
func gateCaps(d *WaterFillDecider, state *types.VaultState) map[types.StrategyID]types.Bps {
	caps := make(map[types.StrategyID]types.Bps, len(state.Strategies))
	for _, s := range state.Strategies {
		if s.Quarantined {
			caps[s.ID] = 0
			continue
		}
		el := d.model.UnknownEL
		ri, known := d.provider.Risk(s.ID)
		if known {
			el = d.model.expectedLoss(ri)
		}
		c := d.model.effectiveCapBps(s.CapBps, el)
		if known && ri.HasLiquidityCap && ri.LiquidityCapBps < c {
			c = ri.LiquidityCapBps
		}
		caps[s.ID] = c
	}
	return caps
}

func FuzzDecide(f *testing.F) {
	f.Add([]byte{0, 40, 0x13, 0x88, 80, 1, 0, 0x0f, 0xa0, 0x13, 0x88, 50, 0, 0}, uint16(1000), uint16(0))
	f.Add([]byte{0x13, 0x88, 0x0f, 0xa0, 30, 0, 0, 0x07, 0xd0, 0x27, 0x10, 90, 4, 10}, uint16(500), uint16(300))
	f.Add([]byte{0, 30, 0, 20, 10, 0, 0, 0, 0, 0x27, 0x10, 70, 2, 200}, uint16(0), uint16(0))
	// Three venues at the full budget; the fill moves 40 and 20 bps out of two
	// and 60 bps into the third, and hysteresis snaps back only the two small ones.
	f.Add([]byte{0x0b, 0xb8, 0x27, 0x10, 50, 0, 0, 0x0b, 0xb8, 0x0b, 0x8a, 40, 0, 0, 0x0b, 0xb8, 0x27, 0x10, 30, 0, 0}, uint16(1000), uint16(0))
	// The owner raised the idle reserve above the standing targets, under a churn bound.
	f.Add([]byte{0x11, 0x94, 0x27, 0x10, 50, 0, 0, 0x11, 0x94, 0x27, 0x10, 40, 0, 0}, uint16(2000), uint16(100))
	f.Fuzz(func(t *testing.T, data []byte, idle, maxDelta uint16) {
		state, risk := fuzzState(data, idle, maxDelta)
		d := newDecider(risk)
		a, err := d.Decide(context.Background(), state)
		if err != nil {
			t.Fatal(err)
		}
		if len(a.Targets) != len(state.Strategies) {
			t.Fatalf("emitted %d targets for %d strategies", len(a.Targets), len(state.Strategies))
		}
		// The keeper's own safety stage must accept every vector the brain emits.
		if res := validate.New(nil).Check(state, a); !res.OK {
			t.Fatalf("validator rejected the vector: %v", res.Violations)
		}
		caps := gateCaps(d, state)
		budget := int(types.BpsDenominator) - int(state.IdleReserveBps)
		var sumNew, moved, forced, keptBase int
		for _, s := range state.Strategies {
			got := a.Targets[s.ID]
			sumNew += int(got)
			if !s.Quarantined {
				moved += int(absDeltaBps(got, s.TargetBps))
			}
			// The vault rejects a nonzero target for a quarantined strategy and
			// any target above the owner cap, so these hold on every output.
			hard := s.CapBps
			if s.Quarantined {
				hard = 0
			}
			if got > hard {
				t.Fatalf("strategy %x at %d above on-chain cap %d (quarantined=%v, was %d)",
					s.ID[0], got, hard, s.Quarantined, s.TargetBps)
			}
			// No strategy is raised above its risk-gated cap.
			if got > s.TargetBps && got > caps[s.ID] {
				t.Fatalf("strategy %x raised to %d above gated cap %d", s.ID[0], got, caps[s.ID])
			}
			// Without a churn bound every strategy ends within its gated cap,
			// so a killed venue is always emitted as zero.
			if state.MaxRebalanceDeltaBps == 0 && got > caps[s.ID] {
				t.Fatalf("strategy %x left at %d above gated cap %d (was %d)", s.ID[0], got, caps[s.ID], s.TargetBps)
			}
			base := min(s.TargetBps, hard)
			if !s.Quarantined {
				forced += int(s.TargetBps - base)
			}
			keptBase += int(base)
		}
		// The vault rejects a target sum above the investable budget.
		if sumNew > budget {
			t.Fatalf("target sum %d above budget %d", sumNew, budget)
		}
		// Churn, counted as in the vault's rebalance without quarantined venues,
		// may exceed its bound only by the movement the vault's own checks
		// force: excess over the owner cap, then excess over the budget.
		forced += max(0, keptBase-budget)
		if bound := int(state.MaxRebalanceDeltaBps); bound != 0 && moved > max(bound, forced) {
			t.Fatalf("moved %d bps over the churn bound %d (forced %d)", moved, bound, forced)
		}
	})
}

func FuzzClampChurn(f *testing.F) {
	f.Add([]byte{0, 100, 0, 0, 0, 0, 0, 100}, uint16(30))
	f.Add([]byte{0x13, 0x88, 0, 0, 0, 1, 0, 3, 0, 2, 0, 0}, uint16(1))
	f.Fuzz(func(t *testing.T, data []byte, maxDelta uint16) {
		var slots []slot
		for i := 0; i+4 <= len(data) && i/4 < 8; i += 4 {
			slots = append(slots, slot{
				alloc:  types.Bps((int(data[i])<<8 | int(data[i+1])) % 10_001),
				target: types.Bps((int(data[i+2])<<8 | int(data[i+3])) % 10_001),
			})
		}
		bound := types.Bps(maxDelta % 10_001)
		proposed := make([]slot, len(slots))
		copy(proposed, slots)
		clampChurn(slots, bound)

		var total, moved, upProposed, upGot int
		for i, p := range proposed {
			total += int(absDeltaBps(p.alloc, p.target))
			moved += int(absDeltaBps(slots[i].alloc, p.target))
			lo, hi := p.target, p.alloc
			if lo > hi {
				lo, hi = hi, lo
			}
			// Each slot stays between its current target and its proposal.
			if slots[i].alloc < lo || slots[i].alloc > hi {
				t.Fatalf("slot %d: %d outside [%d, %d]", i, slots[i].alloc, lo, hi)
			}
			if p.alloc > p.target {
				upProposed += int(p.alloc - p.target)
				upGot += int(slots[i].alloc - p.target)
			}
		}
		if bound == 0 || total <= int(bound) {
			for i := range slots {
				if slots[i].alloc != proposed[i].alloc {
					t.Fatalf("slot %d changed with no clamp needed", i)
				}
			}
			return
		}
		if moved > int(bound) {
			t.Fatalf("moved %d over bound %d", moved, bound)
		}
		// Increases never exceed their exact proportional share of the bound.
		if upGot*total > upProposed*int(bound) {
			t.Fatalf("increase %d above proportional share %d*%d/%d", upGot, upProposed, bound, total)
		}
	})
}
