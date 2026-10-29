// Package risk hosts the adapters that satisfy brain.RiskProvider: the
// per-facet feeds that supply each strategy's expected-loss (EL) signal.
//
// Ports & adapters: the PORT is brain.RiskProvider (defined where it is
// consumed, in internal/brain). The ADAPTERS are the facet subpackages below,
// each turning one heterogeneous source into the common RiskInputs the brain
// understands. Adding a new EL source is purely additive: write a subpackage
// that implements brain.RiskProvider and register it in the composite router;
// no change to the brain or the keeper loop.
//
//	risk/aave      Aave V3 EL via on-chain reads (utilization, paused, supply caps)
//	risk/compound  Compound V3 EL via Comet utilization
//	risk/morpho    Morpho EL via keyless on-chain Morpho Blue reads
//	risk/credora   optional Morpho EL override via the Credora GraphQL feed
//	risk/pendle    Pendle PT gate via the PT oracle and the withdraw-queue liquidity cap
//	composite.go   CompositeProvider: routes a StrategyID to the right adapter
package risk
