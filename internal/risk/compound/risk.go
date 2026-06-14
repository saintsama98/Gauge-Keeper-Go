package compound

import (
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/vault-router-keeper/internal/bindings/comet"
	"github.com/vault-router-keeper/internal/brain"
	"github.com/vault-router-keeper/pkg/types"
)

// Documented-default curve coefficients, mirrored from internal/risk/aave/chaos.go
// (a Comet base-supply position carries the same utilization/exit-liquidity risk
// posture as an Aave supply position). The INPUT (utilization) is a live on-chain
// read; only the input->risk curve shape is a default (CALIBRATION TODO).
const (
	// utilization -> liquidity-haircut piecewise-linear breakpoints.
	utilHaircutLow  = 0.90 // <= this: no liquidity haircut from utilization.
	utilHaircutHigh = 0.99 // > this: near-100% util, max utilization haircut.
	haircutUtilMid  = 0.50 // haircut at utilHaircutHigh (ramp top of mid band).
	haircutUtilTop  = 0.80 // haircut above utilHaircutHigh.

	// utilization -> modeled-EL (vendor channel) curve.
	vendorELBase = 0.0005 // 5 bps baseline protocol/contract risk.
	vendorELU0   = 0.80   // utilization above which modeled EL starts ramping.
	vendorELK    = 0.05   // EL slope per unit utilization above u0.
	vendorELCap  = 0.05   // 5% modeled-EL ceiling.
)

// utilScale is Comet's 1e18 fixed-point base for getUtilization.
var utilScale = new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

// utilizationCaller is the narrow Comet read surface RiskReader needs.
type utilizationCaller interface {
	GetUtilization(opts *bind.CallOpts) (*big.Int, error)
}

// RiskReader is a LIVE on-chain brain.RiskProvider for Compound V3 supply
// strategies. It reads the Comet market's utilization and maps it to a liquidity
// haircut + modeled EL via the documented-default curves above. Returns ok=false
// when the id is not a compound strategy, the backend is unconfigured, or the
// read fails (brain then falls back to its closed-form EL floor).
type RiskReader struct {
	caller     utilizationCaller // nil => unconfigured backend (no-data path)
	isCompound map[types.StrategyID]bool
}

// NewRiskReader builds a live reader over a read-only backend. market is the
// Comet base-market address; isCompound marks which strategy ids it owns. A nil
// backend or zero market address leaves it unconfigured (Risk returns ok=false).
func NewRiskReader(backend bind.ContractCaller, market common.Address, isCompound map[types.StrategyID]bool) *RiskReader {
	r := &RiskReader{isCompound: isCompound}
	if backend != nil && market != (common.Address{}) {
		if c, err := comet.NewCometCaller(market, backend); err == nil {
			r.caller = c
		}
	}
	return r
}

// newRiskReaderFromCaller is the test seam: inject a utilizationCaller (a mock).
func newRiskReaderFromCaller(caller utilizationCaller, isCompound map[types.StrategyID]bool) *RiskReader {
	return &RiskReader{caller: caller, isCompound: isCompound}
}

// Risk reads the live Comet utilization for a compound strategy and maps it to
// RiskInputs. Returns ok=false on non-compound id / unconfigured / read failure.
func (r *RiskReader) Risk(id types.StrategyID) (brain.RiskInputs, bool) {
	if r == nil || r.caller == nil || !r.isCompound[id] {
		return brain.RiskInputs{}, false
	}
	u, err := r.caller.GetUtilization(nil)
	if err != nil || u == nil {
		return brain.RiskInputs{}, false
	}
	util := utilToFloat(u)

	var out brain.RiskInputs
	out.LiquidityHaircut = utilHaircut(util)
	out.HasVendor = true
	out.VendorEL = vendorEL(util)
	return out, true
}

// utilToFloat converts Comet's 1e18-scaled utilization to a unitless fraction.
func utilToFloat(u *big.Int) float64 {
	f, _ := new(big.Float).Quo(new(big.Float).SetInt(u), utilScale).Float64()
	if f < 0 {
		return 0
	}
	return f
}

func utilHaircut(util float64) float64 {
	switch {
	case util <= utilHaircutLow:
		return 0
	case util <= utilHaircutHigh:
		frac := (util - utilHaircutLow) / (utilHaircutHigh - utilHaircutLow)
		return frac * haircutUtilMid
	default:
		return haircutUtilTop
	}
}

func vendorEL(util float64) float64 {
	el := vendorELBase + vendorELK*math.Max(0, util-vendorELU0)
	if el < 0 {
		return 0
	}
	if el > vendorELCap {
		return vendorELCap
	}
	return el
}
