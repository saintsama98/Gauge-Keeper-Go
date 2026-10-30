package keeper

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"math/big"
	"testing"
	"time"

	"github.com/saintsama98/Gauge-Keeper-Go/internal/trigger"
	"github.com/saintsama98/Gauge-Keeper-Go/internal/validate"
	"github.com/saintsama98/Gauge-Keeper-Go/pkg/types"
)

type fixedReader struct{ state *types.VaultState }

func (r fixedReader) Snapshot(context.Context) (*types.VaultState, error) { return r.state, nil }

type fixedDecider struct{ alloc *types.Allocation }

func (d fixedDecider) Decide(context.Context, *types.VaultState) (*types.Allocation, error) {
	return d.alloc, nil
}

type recordingExecutor struct {
	actions []types.Action
	failOn  types.ActionKind
	fail    bool
}

func (e *recordingExecutor) Execute(_ context.Context, a types.Action) (string, error) {
	e.actions = append(e.actions, a)
	if e.fail && a.Kind == e.failOn {
		return "", errors.New("boom")
	}
	return "0x01", nil
}

var (
	idA = types.StrategyID{1}
	idB = types.StrategyID{2}
)

func state() *types.VaultState {
	return &types.VaultState{
		TotalAssets:          big.NewInt(1_000_000),
		IdleAssets:           big.NewInt(0),
		IdleReserveBps:       0,
		MaxRebalanceDeltaBps: 10_000,
		Strategies: []types.StrategyState{
			{ID: idA, TargetBps: 5_000, CurrentAssets: big.NewInt(500_000), CapBps: 10_000},
			{ID: idB, TargetBps: 5_000, CurrentAssets: big.NewInt(500_000), CapBps: 10_000},
		},
	}
}

func alloc(a, b types.Bps) *types.Allocation {
	return &types.Allocation{Targets: map[types.StrategyID]types.Bps{idA: a, idB: b}}
}

func newKeeper(s *types.VaultState, al *types.Allocation, e *recordingExecutor) *Keeper {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	sched := trigger.NewIntervalScheduler(map[string]time.Duration{
		taskGuard: time.Minute, taskHarvest: time.Hour, taskRebalance: time.Hour,
	})
	return New(fixedReader{s}, fixedDecider{al}, validate.New(log), sched, e, time.Second, log)
}

func kinds(as []types.Action) []types.ActionKind {
	out := make([]types.ActionKind, len(as))
	for i, a := range as {
		out[i] = a.Kind
	}
	return out
}

func equal(a, b []types.ActionKind) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestTick_newTargetsSetAndRebalance(t *testing.T) {
	e := &recordingExecutor{}
	if err := newKeeper(state(), alloc(7_000, 3_000), e).tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	want := []types.ActionKind{types.ActionGuardCheckpoint, types.ActionSetAllocation, types.ActionRebalance, types.ActionHarvestAll}
	if !equal(kinds(e.actions), want) {
		t.Fatalf("actions %v, want %v", kinds(e.actions), want)
	}
	if e.actions[1].Allocation.Targets[idA] != 7_000 {
		t.Fatal("SetAllocation does not carry the brain's targets")
	}
}

func TestTick_rejectedAllocationFallsBackToCurrentTargets(t *testing.T) {
	e := &recordingExecutor{}
	// 9000 + 9000 exceeds the investable budget, so the validator rejects it
	if err := newKeeper(state(), alloc(9_000, 9_000), e).tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, a := range e.actions {
		if a.Kind == types.ActionSetAllocation {
			t.Fatal("a rejected allocation was pushed on chain")
		}
	}
}

func TestTick_driftRebalancesWithoutNewTargets(t *testing.T) {
	s := state()
	s.Strategies[0].CurrentAssets = big.NewInt(900_000)
	s.Strategies[1].CurrentAssets = big.NewInt(100_000)
	e := &recordingExecutor{}
	if err := newKeeper(s, alloc(5_000, 5_000), e).tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	want := []types.ActionKind{types.ActionGuardCheckpoint, types.ActionRebalance, types.ActionHarvestAll}
	if !equal(kinds(e.actions), want) {
		t.Fatalf("actions %v, want %v", kinds(e.actions), want)
	}
}

func TestTick_pausedVaultOnlyCheckpoints(t *testing.T) {
	s := state()
	s.Paused = true
	s.PendingWithdraws = []types.WithdrawRequest{{ID: big.NewInt(1), Shares: big.NewInt(1)}}
	e := &recordingExecutor{}
	if err := newKeeper(s, alloc(7_000, 3_000), e).tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if want := []types.ActionKind{types.ActionGuardCheckpoint}; !equal(kinds(e.actions), want) {
		t.Fatalf("actions %v, want %v", kinds(e.actions), want)
	}
}

func TestTick_throttlesAndFulfillsWithdrawals(t *testing.T) {
	s := state()
	s.PendingWithdraws = []types.WithdrawRequest{{ID: big.NewInt(7), Shares: big.NewInt(1)}}
	e := &recordingExecutor{}
	k := newKeeper(s, alloc(5_000, 5_000), e)
	_ = k.tick(context.Background())
	e.actions = nil
	// second tick inside every interval: only the withdrawal is fulfilled again
	_ = k.tick(context.Background())
	if len(e.actions) != 1 || e.actions[0].Kind != types.ActionFulfillWithdraw || e.actions[0].WithdrawID.Int64() != 7 {
		t.Fatalf("actions %v, want only FulfillWithdraw(7)", kinds(e.actions))
	}
}

func TestTick_executorFailureDoesNotStopTheTick(t *testing.T) {
	e := &recordingExecutor{fail: true, failOn: types.ActionGuardCheckpoint}
	if err := newKeeper(state(), alloc(7_000, 3_000), e).tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(e.actions) != 4 {
		t.Fatalf("a failed action stopped the tick after %d actions", len(e.actions))
	}
}

func TestDeploymentDrifts(t *testing.T) {
	s := state()
	if deploymentDrifts(s) {
		t.Fatal("on-target holdings reported as drift")
	}
	s.Strategies[0].CurrentAssets = big.NewInt(502_000) // 20 bps, inside tolerance
	if deploymentDrifts(s) {
		t.Fatal("dust reported as drift")
	}
	s.Strategies[0].CurrentAssets = big.NewInt(510_000) // 100 bps
	if !deploymentDrifts(s) {
		t.Fatal("real drift missed")
	}
	s.TotalAssets = big.NewInt(0)
	if deploymentDrifts(s) {
		t.Fatal("an empty vault cannot drift")
	}
}
