package morpho

import (
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/vault-router-keeper/internal/bindings/metamorpho"
	"github.com/vault-router-keeper/internal/bindings/morphoblue"
	"github.com/vault-router-keeper/internal/brain"
	"github.com/vault-router-keeper/pkg/types"
)

// blueCaller is the narrow read surface MorphoBlueReader needs from the Morpho
// Blue singleton. It is satisfied in production by the abigen binding's
// *morphoblue.MorphoBlueCaller and mocked in unit tests so they never touch a
// live chain. The anonymous-struct field names mirror the generated binding
// exactly so the binding satisfies this interface with no adapter.
type blueCaller interface {
	Market(opts *bind.CallOpts, id [32]byte) (struct {
		TotalSupplyAssets *big.Int
		TotalSupplyShares *big.Int
		TotalBorrowAssets *big.Int
		TotalBorrowShares *big.Int
		LastUpdate        *big.Int
		Fee               *big.Int
	}, error)

	Position(opts *bind.CallOpts, id [32]byte, user common.Address) (struct {
		SupplyShares *big.Int
		BorrowShares *big.Int
		Collateral   *big.Int
	}, error)
}

// vaultCaller is the narrow read surface from a MetaMorpho (ERC-4626 curated)
// vault: enough to walk its withdraw queue and read each market's cap/enabled
// flag. Satisfied by *metamorpho.MetaMorphoCaller; mocked in tests.
type vaultCaller interface {
	WithdrawQueueLength(opts *bind.CallOpts) (*big.Int, error)
	WithdrawQueue(opts *bind.CallOpts, arg0 *big.Int) ([32]byte, error)
	Config(opts *bind.CallOpts, arg0 [32]byte) (struct {
		Cap         *big.Int
		Enabled     bool
		RemovableAt uint64
	}, error)
}

// Documented-default curve coefficients (CALIBRATION TODO).
//
// Every INPUT these are applied to is a LIVE on-chain read: per-market
// utilization (Blue.market totalBorrowAssets/totalSupplyAssets), the vault's
// own supply in each market (Blue.position supplyShares converted to assets),
// and market liquidity (totalSupplyAssets-totalBorrowAssets). Only the shape of
// the input->risk mapping is a default; the data is never fabricated. Tune
// against historical MetaMorpho withdrawal-stress data before relying on them.
const (
	// vendor-EL util->EL curve. Blue markets are designed to run at higher
	// utilization than Aave reserves, so the ramp starts later.
	morphoVendorELBase   = 0.0005 // 5 bps Morpho Blue baseline contract risk.
	morphoVendorELU0     = 0.92   // supply-weighted util above which modeled EL ramps.
	morphoVendorELUtilK  = 0.10   // EL slope per unit utilization above u0.
	morphoVendorELIlliqK = 0.10   // EL slope per unit of non-withdrawable vault fraction.
	morphoVendorELCap    = 0.05   // 5% modeled-EL ceiling (matches the Aave facet).

	// Bound the per-tick RPC fan-out: a MetaMorpho withdraw queue is small in
	// practice, but cap the walk so a misconfigured vault can never wedge a tick.
	maxQueueMarkets = 64
)

// MorphoBlueReader is a LIVE, keyless on-chain brain.RiskProvider for MetaMorpho
// (ERC-4626 curated vault) strategies. For each owned strategy it walks the
// vault's withdraw queue, and for every enabled market reads the vault's supply
// position and the market's supply/borrow totals from Morpho Blue, then derives:
//
//   - supply-weighted utilization across the vault's markets -> a vendor-EL curve
//   - the fraction of the vault's assets NOT currently withdrawable (borrowed out)
//     -> a liquidity haircut and an additional EL penalty
//
// It needs NO API key, NO subscription, and NO vendor relationship — only an RPC
// endpoint — which is why it is the free alternative to the auth-gated Credora
// rating feed (see internal/risk/credora). The brain composes VendorEL via max()
// with its closed-form EL, so this can only ever tighten an allocation, never
// loosen it. ok=false (the honest no-data path -> closed-form floor) is returned
// when the id is not an owned Morpho strategy, the backend is unconfigured, any
// read fails, or the vault holds nothing in enabled markets.
type MorphoBlueReader struct {
	blue   blueCaller // nil => unconfigured backend (no-data path)
	vaults map[types.StrategyID]vaultEntry
}

type vaultEntry struct {
	addr   common.Address
	caller vaultCaller
}

// NewMorphoBlueReader builds a live reader over a read-only backend (e.g.
// *ethclient.Client). blue is the Morpho Blue singleton address; vaults maps each
// owned StrategyID to its MetaMorpho ERC-4626 vault address.
//
// If backend is nil or blue is the zero address, the reader is "unconfigured"
// and every Risk call returns ok=false (the brain falls back to its closed-form
// EL floor — the honest no-data path, not a stub).
func NewMorphoBlueReader(backend bind.ContractCaller, blue common.Address, vaults map[types.StrategyID]common.Address) *MorphoBlueReader {
	r := &MorphoBlueReader{vaults: map[types.StrategyID]vaultEntry{}}
	if backend == nil || blue == (common.Address{}) {
		return r
	}
	bc, err := morphoblue.NewMorphoBlueCaller(blue, backend)
	if err != nil {
		return r // malformed embedded ABI is a build-time invariant; degrade.
	}
	r.blue = bc
	for id, addr := range vaults {
		if addr == (common.Address{}) {
			continue
		}
		vc, err := metamorpho.NewMetaMorphoCaller(addr, backend)
		if err != nil {
			continue
		}
		r.vaults[id] = vaultEntry{addr: addr, caller: vc}
	}
	return r
}

// newMorphoBlueReaderFromCallers is the test seam: it injects the callers
// directly (mocks), bypassing live binding construction. All owned strategies
// share one blue caller; each id supplies its own vault caller + address.
func newMorphoBlueReaderFromCallers(blue blueCaller, vaults map[types.StrategyID]vaultEntry) *MorphoBlueReader {
	if vaults == nil {
		vaults = map[types.StrategyID]vaultEntry{}
	}
	return &MorphoBlueReader{blue: blue, vaults: vaults}
}

// Risk reads the vault's live per-market state and maps it to RiskInputs.
func (r *MorphoBlueReader) Risk(id types.StrategyID) (brain.RiskInputs, bool) {
	if r == nil || r.blue == nil {
		return brain.RiskInputs{}, false
	}
	v, owned := r.vaults[id]
	if !owned || v.caller == nil {
		return brain.RiskInputs{}, false
	}

	nBig, err := v.caller.WithdrawQueueLength(nil)
	if err != nil || nBig == nil || nBig.Sign() <= 0 {
		return brain.RiskInputs{}, false
	}
	n := nBig.Int64()
	if n > maxQueueMarkets {
		n = maxQueueMarkets
	}

	var totalAssets, totalWithdrawable, weightedUtilNumer float64
	for i := int64(0); i < n; i++ {
		mid, err := v.caller.WithdrawQueue(nil, big.NewInt(i))
		if err != nil {
			return brain.RiskInputs{}, false
		}
		cfg, err := v.caller.Config(nil, mid)
		if err != nil {
			return brain.RiskInputs{}, false
		}
		if !cfg.Enabled {
			continue
		}
		m, err := r.blue.Market(nil, mid)
		if err != nil {
			return brain.RiskInputs{}, false
		}
		supply := m.TotalSupplyAssets
		shares := m.TotalSupplyShares
		if supply == nil || supply.Sign() <= 0 || shares == nil || shares.Sign() <= 0 {
			continue // empty market: no vault assets, no contention.
		}
		pos, err := r.blue.Position(nil, mid, v.addr)
		if err != nil {
			return brain.RiskInputs{}, false
		}
		if pos.SupplyShares == nil || pos.SupplyShares.Sign() <= 0 {
			continue // vault holds nothing here.
		}

		// vaultAssets = supplyShares * totalSupplyAssets / totalSupplyShares.
		vaultAssetsInt := new(big.Int).Mul(pos.SupplyShares, supply)
		vaultAssetsInt.Div(vaultAssetsInt, shares)
		vaultAssets := bigToFloat(vaultAssetsInt)
		if vaultAssets <= 0 {
			continue
		}

		borrow := bigToFloat(m.TotalBorrowAssets)
		supplyF := bigToFloat(supply)
		liquidity := supplyF - borrow // assets currently free to withdraw, market-wide.
		if liquidity < 0 {
			liquidity = 0
		}
		withdrawable := math.Min(vaultAssets, liquidity)

		util := 0.0
		if supplyF > 0 {
			util = borrow / supplyF
		}

		totalAssets += vaultAssets
		totalWithdrawable += withdrawable
		weightedUtilNumer += util * vaultAssets
	}

	if totalAssets <= 0 {
		return brain.RiskInputs{}, false // nothing allocated in enabled markets.
	}

	avgUtil := weightedUtilNumer / totalAssets
	illiquidFrac := clamp01(1 - totalWithdrawable/totalAssets)

	out := brain.RiskInputs{
		Pegged:           true, // MetaMorpho USDC vault: stablecoin loan asset.
		LiquidityHaircut: illiquidFrac,
		HasVendor:        true,
		VendorEL:         morphoVendorEL(avgUtil, illiquidFrac),
	}
	return out, true
}

// morphoVendorEL is the documented-default modeled expected-loss curve
// (CALIBRATION TODO): clamp(base + utilK*max(0,util-u0) + illiqK*illiquid, 0, cap).
// It raises modeled EL as the vault's markets approach full utilization and as a
// larger share of the vault's assets becomes non-withdrawable.
func morphoVendorEL(util, illiquidFrac float64) float64 {
	el := morphoVendorELBase +
		morphoVendorELUtilK*math.Max(0, util-morphoVendorELU0) +
		morphoVendorELIlliqK*clamp01(illiquidFrac)
	if el < 0 {
		return 0
	}
	if el > morphoVendorELCap {
		return morphoVendorELCap
	}
	return el
}

// bigToFloat converts a big.Int to float64 (nil -> 0). Used for ratios where
// float precision is ample; never for value-moving math.
func bigToFloat(x *big.Int) float64 {
	if x == nil {
		return 0
	}
	f, _ := new(big.Float).SetInt(x).Float64()
	return f
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
