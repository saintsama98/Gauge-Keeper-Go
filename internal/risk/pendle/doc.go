// Package pendle implements brain.RiskProvider for the PENDLE PT facet. It reads
// the PT risk oracle — Chaos killswitch (on-chain getter) and the RedStone
// Dynamic PT mark (push adapter where available) — into mark / time-to-maturity
// / killswitch -> VendorEL, AND computes LiquidityCapBps: the withdraw-queue gate
// that caps PT to what the async queue can cover before maturity.
//
// Language note: RedStone's *pull* model has no Go SDK (the evm-connector is
// JS/TS only). Prefer the Chaos on-chain getter and/or a RedStone *push* adapter
// (Chainlink-compatible, Go-clean); fall back to a thin TS sidecar only if a pull
// feed is unavoidable. The chain client sits behind an interface; tests mock it.
// The feed model of each PT market has to be confirmed before it is configured.
package pendle
