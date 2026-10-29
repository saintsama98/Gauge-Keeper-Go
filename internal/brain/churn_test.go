package brain

import (
	"math/rand"
	"testing"

	"github.com/saintsama98/Gauge-Keeper-Go/internal/validate"
	"github.com/saintsama98/Gauge-Keeper-Go/pkg/types"
)

// clampCase runs clampChurn over the given targets and proposed allocations and
// returns the clamped allocations.
func clampCase(targets, allocs []types.Bps, maxDelta types.Bps) []types.Bps {
	slots := make([]slot, len(targets))
	for i := range targets {
		slots[i] = slot{id: sid(byte(i + 1)), target: targets[i], alloc: allocs[i]}
	}
	clampChurn(slots, maxDelta)
	out := make([]types.Bps, len(slots))
	for i := range slots {
		out[i] = slots[i].alloc
	}
	return out
}

// validateClamped checks the clamped allocation with the same validator the
// keeper runs before it sends anything.
func validateClamped(t *testing.T, targets, clamped []types.Bps, idle, maxDelta types.Bps) {
	t.Helper()
	state := &types.VaultState{IdleReserveBps: idle, MaxRebalanceDeltaBps: maxDelta}
	alloc := &types.Allocation{Targets: map[types.StrategyID]types.Bps{}}
	for i := range targets {
		id := sid(byte(i + 1))
		state.Strategies = append(state.Strategies, types.StrategyState{ID: id, CapBps: types.BpsDenominator, TargetBps: targets[i]})
		alloc.Targets[id] = clamped[i]
	}
	if r := validate.New(nil).Check(state, alloc); !r.OK {
		t.Fatalf("clamped allocation %v rejected: %v", clamped, r.Violations)
	}
}

// Deltas of +65, +65, +65 and −195 against a bound of 100 scale to 16.67 and
// −50. Rounding each one on its own gives 17, 17, 17, −50: 101 of movement.
func TestClampChurnRoundingStaysWithinBound(t *testing.T) {
	targets := []types.Bps{2000, 2000, 2000, 2000}
	allocs := []types.Bps{2065, 2065, 2065, 1805}
	got := clampCase(targets, allocs, 100)
	validateClamped(t, targets, got, 1000, 100)
}

// The mirror image, with the targets already at the investable limit: rounding
// the three small decreases toward zero would push the sum over the idle floor.
func TestClampChurnRoundingKeepsIdleFloor(t *testing.T) {
	targets := []types.Bps{2250, 2250, 2250, 2250}
	allocs := []types.Bps{2185, 2185, 2185, 2445}
	got := clampCase(targets, allocs, 100)
	validateClamped(t, targets, got, 1000, 100)
}

// Random allocations: the clamp must stay within its bound, move every slot only
// between its target and its proposal, and never raise the invested sum above
// the larger of the two sums it started from.
func TestClampChurnProperties(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for iter := 0; iter < 20000; iter++ {
		n := 1 + rng.Intn(6)
		targets := randomSplit(rng, n)
		allocs := randomSplit(rng, n)
		maxDelta := types.Bps(1 + rng.Intn(2000))
		got := clampCase(targets, allocs, maxDelta)

		var moved, sumTarget, sumAlloc, sumGot int
		for i := range got {
			lo, hi := targets[i], allocs[i]
			if lo > hi {
				lo, hi = hi, lo
			}
			if got[i] < lo || got[i] > hi {
				t.Fatalf("slot %d moved outside [%d, %d]: %d", i, lo, hi, got[i])
			}
			moved += int(absDeltaBps(got[i], targets[i]))
			sumTarget += int(targets[i])
			sumAlloc += int(allocs[i])
			sumGot += int(got[i])
		}
		if moved > int(maxDelta) {
			t.Fatalf("moved %d > bound %d (targets %v, allocs %v, got %v)", moved, maxDelta, targets, allocs, got)
		}
		if sumGot > max(sumTarget, sumAlloc) {
			t.Fatalf("sum %d above both %d and %d (targets %v, allocs %v, got %v)", sumGot, sumTarget, sumAlloc, targets, allocs, got)
		}
	}
}

// randomSplit returns n targets that sum to at most 9000 bps.
func randomSplit(rng *rand.Rand, n int) []types.Bps {
	out := make([]types.Bps, n)
	left := 9000
	for i := range out {
		v := rng.Intn(left + 1)
		out[i] = types.Bps(v)
		left -= v
	}
	rng.Shuffle(n, func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}
