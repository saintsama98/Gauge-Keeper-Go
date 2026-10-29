// Package pendle implements brain.RiskProvider for the Pendle PT strategy. It
// reads the PT-to-asset rate, oracle readiness and expiry from the canonical
// PendlePYLpOracle and computes LiquidityCapBps: the withdraw-queue gate that
// caps PT to what the async queue can cover before maturity. A live implied
// fixed-to-maturity APY overlay lives in yield.go. The chain client sits behind
// an interface; tests mock it.
package pendle
