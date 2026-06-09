// Package bindings is the home for abigen-generated Go contract bindings. Each
// contract gets its own subpackage; the raw ABIs live in /abi and the bindings
// are regenerated with `make bindings`.
//
//	bindings/vault     — Vault Router diamond views (Allocator/Guard/WithdrawQueue, ERC-4626)
//	bindings/aavepool  — Aave V3 UiPoolDataProvider
//	bindings/chaosrisk — Chaos Labs RiskOracle getter
//	bindings/redstone  — RedStone push price adapter (AggregatorV3)
//
// Isolating generated code here keeps the domain (pkg/types) and the ports
// (internal/brain, perceive, trigger, execute) free of chain-specific types —
// the ports-and-adapters boundary. Keep both the ABIs and the generated code in
// version control so builds are reproducible offline.
package bindings
