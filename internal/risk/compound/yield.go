// Package compound provides the keeper's live yield + risk adapters for a
// Compound III (Comet) base-supply strategy. Both read the Comet market directly
// over a read-only backend (keyless, no oracle deployed) and degrade to ok=false
// when unconfigured — the honest no-data path, never a fabricated number.
package compound

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/vault-router-keeper/internal/bindings/comet"
	"github.com/vault-router-keeper/pkg/types"
)

// secondsPerYear annualizes Comet's per-second supply rate. Unlike Aave's
// pre-annualized RAY liquidityRate, getSupplyRate returns a spot per-second rate
// (scaled 1e18), so APR = rate/1e18 * secondsPerYear.
const secondsPerYear = 31_536_000

// supplyRateCaller is the narrow Comet read surface the YieldReader needs. It is
// satisfied in production by *comet.CometCaller (method names/signatures match
// exactly) and mocked directly in tests so they never touch a live chain.
type supplyRateCaller interface {
	GetUtilization(opts *bind.CallOpts) (*big.Int, error)
	GetSupplyRate(opts *bind.CallOpts, utilization *big.Int) (uint64, error)
}

// YieldReader implements perceive.YieldProvider for Compound V3 supply
// strategies: APR = getSupplyRate(getUtilization()) / 1e18 * secondsPerYear,
// read live from the Comet market. Returns ok=false when the id is not a
// compound strategy, the backend/market is absent, or the read fails / is
// non-positive (the OverlayReader then leaves APY at 0). Never fabricates yield.
type YieldReader struct {
	caller     supplyRateCaller // nil => unconfigured backend (no-data path)
	isCompound map[types.StrategyID]bool
}

// NewYieldReader builds a live Compound supply-APR provider over a read-only
// backend. market is the Comet base-market address (cUSDCv3); isCompound marks
// which strategy ids this adapter owns. A nil backend or zero market address
// leaves the reader unconfigured (every APY call returns ok=false).
func NewYieldReader(backend bind.ContractCaller, market common.Address, isCompound map[types.StrategyID]bool) *YieldReader {
	r := &YieldReader{isCompound: isCompound}
	if backend != nil && market != (common.Address{}) {
		if c, err := comet.NewCometCaller(market, backend); err == nil {
			r.caller = c
		}
	}
	return r
}

// newYieldReaderFromCaller is the test seam: inject a supplyRateCaller (a mock).
func newYieldReaderFromCaller(caller supplyRateCaller, isCompound map[types.StrategyID]bool) *YieldReader {
	return &YieldReader{caller: caller, isCompound: isCompound}
}

// APY returns the live Compound supply APR for an owned compound strategy.
func (r *YieldReader) APY(id types.StrategyID) (float64, bool) {
	if r == nil || r.caller == nil || !r.isCompound[id] {
		return 0, false
	}
	util, err := r.caller.GetUtilization(nil)
	if err != nil || util == nil {
		return 0, false
	}
	rate, err := r.caller.GetSupplyRate(nil, util)
	if err != nil || rate == 0 {
		return 0, false
	}
	apr := float64(rate) / 1e18 * secondsPerYear
	if apr <= 0 {
		return 0, false
	}
	return apr, true
}
