package morpho

import (
	"errors"
	"math"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/vault-router-keeper/pkg/types"
)

// --- mocks --------------------------------------------------------------------

type marketState struct {
	supply, shares, borrow *big.Int
}

type mockBlue struct {
	markets   map[[32]byte]marketState
	positions map[[32]byte]*big.Int // id -> vault supplyShares
	err       error
}

func (m *mockBlue) Market(_ *bind.CallOpts, id [32]byte) (struct {
	TotalSupplyAssets *big.Int
	TotalSupplyShares *big.Int
	TotalBorrowAssets *big.Int
	TotalBorrowShares *big.Int
	LastUpdate        *big.Int
	Fee               *big.Int
}, error) {
	var out struct {
		TotalSupplyAssets *big.Int
		TotalSupplyShares *big.Int
		TotalBorrowAssets *big.Int
		TotalBorrowShares *big.Int
		LastUpdate        *big.Int
		Fee               *big.Int
	}
	if m.err != nil {
		return out, m.err
	}
	s := m.markets[id]
	out.TotalSupplyAssets = s.supply
	out.TotalSupplyShares = s.shares
	out.TotalBorrowAssets = s.borrow
	out.TotalBorrowShares = big.NewInt(0)
	out.LastUpdate = big.NewInt(0)
	out.Fee = big.NewInt(0)
	return out, nil
}

func (m *mockBlue) Position(_ *bind.CallOpts, id [32]byte, _ common.Address) (struct {
	SupplyShares *big.Int
	BorrowShares *big.Int
	Collateral   *big.Int
}, error) {
	var out struct {
		SupplyShares *big.Int
		BorrowShares *big.Int
		Collateral   *big.Int
	}
	if m.err != nil {
		return out, m.err
	}
	ss := m.positions[id]
	if ss == nil {
		ss = big.NewInt(0)
	}
	out.SupplyShares = ss
	out.BorrowShares = big.NewInt(0)
	out.Collateral = big.NewInt(0)
	return out, nil
}

type mockVault struct {
	queue   [][32]byte
	enabled map[[32]byte]bool
	err     error
}

func (v *mockVault) WithdrawQueueLength(*bind.CallOpts) (*big.Int, error) {
	if v.err != nil {
		return nil, v.err
	}
	return big.NewInt(int64(len(v.queue))), nil
}

func (v *mockVault) WithdrawQueue(_ *bind.CallOpts, i *big.Int) ([32]byte, error) {
	if v.err != nil {
		return [32]byte{}, v.err
	}
	return v.queue[i.Int64()], nil
}

func (v *mockVault) Config(_ *bind.CallOpts, id [32]byte) (struct {
	Cap         *big.Int
	Enabled     bool
	RemovableAt uint64
}, error) {
	return struct {
		Cap         *big.Int
		Enabled     bool
		RemovableAt uint64
	}{Cap: big.NewInt(0), Enabled: v.enabled[id], RemovableAt: 0}, nil
}

// --- helpers ------------------------------------------------------------------

func mkID(b byte) [32]byte { var id [32]byte; id[0] = b; return id }

func reader(blue blueCaller, vault vaultCaller) *MorphoBlueReader {
	id := types.StrategyID(mkID('m'))
	return newMorphoBlueReaderFromCallers(blue, map[types.StrategyID]vaultEntry{
		id: {addr: common.HexToAddress("0x1111111111111111111111111111111111111111"), caller: vault},
	})
}

const eps = 1e-9

// --- tests --------------------------------------------------------------------

// A healthy, low-utilization vault: fully withdrawable, util below the ramp =>
// VendorEL at the baseline, no liquidity haircut.
func TestMorphoBlue_Healthy(t *testing.T) {
	m1 := mkID(1)
	blue := &mockBlue{
		markets:   map[[32]byte]marketState{m1: {supply: big.NewInt(1_000_000), shares: big.NewInt(1_000_000), borrow: big.NewInt(500_000)}},
		positions: map[[32]byte]*big.Int{m1: big.NewInt(200_000)},
	}
	vault := &mockVault{queue: [][32]byte{m1}, enabled: map[[32]byte]bool{m1: true}}

	ri, ok := reader(blue, vault).Risk(types.StrategyID(mkID('m')))
	if !ok {
		t.Fatal("expected ok=true for a healthy vault")
	}
	if !ri.Pegged {
		t.Error("expected Pegged=true for a MetaMorpho USDC vault")
	}
	if math.Abs(ri.LiquidityHaircut-0) > eps {
		t.Errorf("expected LiquidityHaircut 0 (fully withdrawable), got %v", ri.LiquidityHaircut)
	}
	if !ri.HasVendor || math.Abs(ri.VendorEL-morphoVendorELBase) > eps {
		t.Errorf("expected VendorEL == base %v, got %v", morphoVendorELBase, ri.VendorEL)
	}
}

// A stressed vault: util 0.99 and most of the vault's assets borrowed out so they
// cannot be withdrawn => VendorEL saturates at the cap and the liquidity haircut
// reflects the non-withdrawable fraction.
func TestMorphoBlue_Stressed(t *testing.T) {
	m1 := mkID(1)
	blue := &mockBlue{
		markets:   map[[32]byte]marketState{m1: {supply: big.NewInt(1_000_000), shares: big.NewInt(1_000_000), borrow: big.NewInt(990_000)}},
		positions: map[[32]byte]*big.Int{m1: big.NewInt(200_000)},
	}
	vault := &mockVault{queue: [][32]byte{m1}, enabled: map[[32]byte]bool{m1: true}}

	ri, ok := reader(blue, vault).Risk(types.StrategyID(mkID('m')))
	if !ok {
		t.Fatal("expected ok=true")
	}
	// withdrawable = min(200k, supply-borrow=10k) = 10k => illiquid 1-10k/200k = 0.95.
	if math.Abs(ri.LiquidityHaircut-0.95) > 1e-6 {
		t.Errorf("expected LiquidityHaircut 0.95, got %v", ri.LiquidityHaircut)
	}
	if math.Abs(ri.VendorEL-morphoVendorELCap) > eps {
		t.Errorf("expected VendorEL saturated at cap %v, got %v", morphoVendorELCap, ri.VendorEL)
	}
}

// Two markets, one disabled: the disabled market is skipped entirely (its
// liquidity must not count toward withdrawability).
func TestMorphoBlue_DisabledMarketSkipped(t *testing.T) {
	good, bad := mkID(1), mkID(2)
	blue := &mockBlue{
		markets: map[[32]byte]marketState{
			good: {supply: big.NewInt(1_000_000), shares: big.NewInt(1_000_000), borrow: big.NewInt(0)},
			bad:  {supply: big.NewInt(1_000_000), shares: big.NewInt(1_000_000), borrow: big.NewInt(1_000_000)},
		},
		positions: map[[32]byte]*big.Int{good: big.NewInt(100_000), bad: big.NewInt(100_000)},
	}
	vault := &mockVault{queue: [][32]byte{good, bad}, enabled: map[[32]byte]bool{good: true, bad: false}}

	ri, ok := reader(blue, vault).Risk(types.StrategyID(mkID('m')))
	if !ok {
		t.Fatal("expected ok=true")
	}
	// Only the good (0% util, fully liquid) market counts => no haircut, base EL.
	if math.Abs(ri.LiquidityHaircut-0) > eps {
		t.Errorf("disabled market leaked into the haircut: got %v", ri.LiquidityHaircut)
	}
	if math.Abs(ri.VendorEL-morphoVendorELBase) > eps {
		t.Errorf("expected base VendorEL, got %v", ri.VendorEL)
	}
}

func TestMorphoBlue_NoDataPaths(t *testing.T) {
	m1 := mkID(1)
	healthy := &mockBlue{
		markets:   map[[32]byte]marketState{m1: {supply: big.NewInt(1_000_000), shares: big.NewInt(1_000_000), borrow: big.NewInt(0)}},
		positions: map[[32]byte]*big.Int{m1: big.NewInt(100_000)},
	}

	t.Run("unowned id", func(t *testing.T) {
		v := &mockVault{queue: [][32]byte{m1}, enabled: map[[32]byte]bool{m1: true}}
		if _, ok := reader(healthy, v).Risk(types.StrategyID(mkID('x'))); ok {
			t.Error("expected ok=false for an unowned strategy id")
		}
	})
	t.Run("unconfigured (nil blue)", func(t *testing.T) {
		if _, ok := reader(nil, &mockVault{}).Risk(types.StrategyID(mkID('m'))); ok {
			t.Error("expected ok=false when the backend is unconfigured")
		}
	})
	t.Run("empty queue", func(t *testing.T) {
		v := &mockVault{queue: nil, enabled: map[[32]byte]bool{}}
		if _, ok := reader(healthy, v).Risk(types.StrategyID(mkID('m'))); ok {
			t.Error("expected ok=false for an empty withdraw queue")
		}
	})
	t.Run("vault holds nothing in enabled markets", func(t *testing.T) {
		v := &mockVault{queue: [][32]byte{m1}, enabled: map[[32]byte]bool{m1: true}}
		empty := &mockBlue{
			markets:   healthy.markets,
			positions: map[[32]byte]*big.Int{m1: big.NewInt(0)},
		}
		if _, ok := reader(empty, v).Risk(types.StrategyID(mkID('m'))); ok {
			t.Error("expected ok=false when the vault has no supply in enabled markets")
		}
	})
	t.Run("read error degrades to no-data", func(t *testing.T) {
		v := &mockVault{queue: [][32]byte{m1}, enabled: map[[32]byte]bool{m1: true}}
		boom := &mockBlue{err: errors.New("rpc down")}
		if _, ok := reader(boom, v).Risk(types.StrategyID(mkID('m'))); ok {
			t.Error("expected ok=false when an on-chain read fails")
		}
	})
}
