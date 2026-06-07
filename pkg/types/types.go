// Package types holds the shared domain vocabulary for the keeper. Every type
// maps directly onto the Vault Router diamond's curator/keeper surface so the
// perceive, brain, trigger and execute layers speak the same language.
package types

import "math/big"

// StrategyID maps to the Solidity bytes32 strategy identifier used across
// AllocatorFacet / HarvestFacet (e.g. keccak256("morpho")).
type StrategyID [32]byte

// Bps is a basis-point value in [0, 10000]; 10000 == 100%.
type Bps uint16

// BpsDenominator is 100% in basis points.
const BpsDenominator Bps = 10_000

// StrategyState is the per-strategy view the keeper reads from the vault, plus
// any off-chain signal (APY) a real brain would use.
type StrategyState struct {
	ID            StrategyID // bytes32 id
	TargetBps     Bps        // on-chain target  (AllocatorFacet.targetAllocation)
	CurrentAssets *big.Int   // on-chain holdings (AllocatorFacet.strategyTotalAssets)
	CapBps        Bps        // effective cap     (AllocatorFacet.strategyCap)
	Quarantined   bool       // AllocatorFacet.isQuarantined
	APY           float64    // off-chain signal (subgraph/API); 0 if unknown
}

// WithdrawRequest is a pending async-queue exit (WithdrawQueueFacet).
type WithdrawRequest struct {
	ID     *big.Int
	Shares *big.Int
}

// VaultState is one consistent snapshot of everything needed to decide + plan.
type VaultState struct {
	TotalAssets          *big.Int
	IdleAssets           *big.Int
	IdleReserveBps       Bps
	MaxRebalanceDeltaBps Bps
	Paused               bool // GuardFacet.paused — no curator txns while true
	Strategies           []StrategyState
	PendingWithdraws     []WithdrawRequest
}

// Allocation is the brain's output: the target weight vector pushed via
// AllocatorFacet.setAllocation.
type Allocation struct {
	Targets map[StrategyID]Bps
}

// ActionKind enumerates the curator/keeper-callable vault operations.
type ActionKind int

const (
	ActionSetAllocation   ActionKind = iota // AllocatorFacet.setAllocation   (curator)
	ActionRebalance                         // AllocatorFacet.rebalance        (curator)
	ActionHarvestAll                        // HarvestFacet.harvestAll          (curator)
	ActionFulfillWithdraw                   // Vault.fulfillWithdraw            (curator)
	ActionGuardCheckpoint                   // GuardFacet.guardCheckpoint       (permissionless)
)

func (k ActionKind) String() string {
	switch k {
	case ActionSetAllocation:
		return "SetAllocation"
	case ActionRebalance:
		return "Rebalance"
	case ActionHarvestAll:
		return "HarvestAll"
	case ActionFulfillWithdraw:
		return "FulfillWithdraw"
	case ActionGuardCheckpoint:
		return "GuardCheckpoint"
	default:
		return "Unknown"
	}
}

// Action is one unit of work handed to the Executor.
type Action struct {
	Kind       ActionKind
	Allocation *Allocation // set for ActionSetAllocation
	WithdrawID *big.Int    // set for ActionFulfillWithdraw
}
