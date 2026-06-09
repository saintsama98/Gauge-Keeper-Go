// Package aave implements brain.RiskProvider for the AAVE V3 facet via on-chain
// reads: UiPoolDataProvider gives utilization, available liquidity and
// paused/frozen state, plus the caps Chaos Labs' risk oracle has set. These map
// to RiskInputs (no vendor EL; a utilization/paused-driven floor composed with
// the closed-form).
//
// Delivery: on-chain getters via internal/chain + internal/bindings/aavepool
// (go-ethereum / abigen). The chain client sits behind an interface; tests mock
// it.
package aave
