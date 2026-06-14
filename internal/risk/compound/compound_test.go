package compound

import (
	"math"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/vault-router-keeper/pkg/types"
)

// mockCaller satisfies both supplyRateCaller and utilizationCaller.
type mockCaller struct {
	util *big.Int
	rate uint64
}

func (m mockCaller) GetUtilization(_ *bind.CallOpts) (*big.Int, error) { return m.util, nil }
func (m mockCaller) GetSupplyRate(_ *bind.CallOpts, _ *big.Int) (uint64, error) {
	return m.rate, nil
}

func compoundID() types.StrategyID {
	var id types.StrategyID
	copy(id[:], []byte("compound"))
	return id
}

func TestYieldReaderAPR(t *testing.T) {
	id := compoundID()
	owned := map[types.StrategyID]bool{id: true}

	// 988405955 per-second (1e18-scaled) is the live mainnet value at ~86.6%
	// utilization; annualized that is ~3.117% APR.
	r := newYieldReaderFromCaller(mockCaller{util: big.NewInt(866_000_000_000_000_000), rate: 988405955}, owned)
	apr, ok := r.APY(id)
	if !ok {
		t.Fatal("expected ok=true for an owned compound strategy")
	}
	want := float64(988405955) / 1e18 * secondsPerYear
	if math.Abs(apr-want) > 1e-9 {
		t.Fatalf("apr = %v, want %v", apr, want)
	}
	if apr < 0.02 || apr > 0.05 {
		t.Fatalf("apr %v outside the believable 2-5%% band", apr)
	}

	// Non-compound id => ok=false.
	if _, ok := r.APY(types.StrategyID{'x'}); ok {
		t.Fatal("expected ok=false for a non-compound id")
	}
	// Zero rate => ok=false (no fabricated yield).
	if _, ok := newYieldReaderFromCaller(mockCaller{util: big.NewInt(0), rate: 0}, owned).APY(id); ok {
		t.Fatal("expected ok=false for a zero supply rate")
	}
	// Unconfigured (nil caller) => ok=false.
	if _, ok := (&YieldReader{isCompound: owned}).APY(id); ok {
		t.Fatal("expected ok=false for an unconfigured reader")
	}
}

func TestRiskReaderUtilizationMapping(t *testing.T) {
	id := compoundID()
	owned := map[types.StrategyID]bool{id: true}
	e18 := func(frac float64) *big.Int {
		f := new(big.Float).Mul(big.NewFloat(frac), utilScale)
		out, _ := f.Int(nil)
		return out
	}

	// Low utilization (50%): no liquidity haircut, baseline EL only.
	low := newRiskReaderFromCaller(mockCaller{util: e18(0.50)}, owned)
	ri, ok := low.Risk(id)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if ri.LiquidityHaircut != 0 {
		t.Fatalf("low util haircut = %v, want 0", ri.LiquidityHaircut)
	}
	if !ri.HasVendor || math.Abs(ri.VendorEL-vendorELBase) > 1e-9 {
		t.Fatalf("low util EL = %v, want baseline %v", ri.VendorEL, vendorELBase)
	}

	// Very high utilization (99.5%): top haircut and EL ramping above u0.
	hi := newRiskReaderFromCaller(mockCaller{util: e18(0.995)}, owned)
	ri, _ = hi.Risk(id)
	if ri.LiquidityHaircut != haircutUtilTop {
		t.Fatalf("high util haircut = %v, want %v", ri.LiquidityHaircut, haircutUtilTop)
	}
	if ri.VendorEL <= vendorELBase {
		t.Fatalf("high util EL = %v, expected a ramp above baseline", ri.VendorEL)
	}

	// Non-compound id => ok=false.
	if _, ok := low.Risk(types.StrategyID{'x'}); ok {
		t.Fatal("expected ok=false for a non-compound id")
	}
}
