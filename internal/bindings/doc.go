// Package bindings is the home for abigen-generated Go contract bindings. Each
// contract gets its own subpackage; the raw ABIs live in /abi and the bindings
// are regenerated with `make bindings`.
//
//	bindings/vault        Vault Router diamond views (Allocator, Guard, WithdrawQueue, ERC-4626)
//	bindings/aavedata     Aave V3 ProtocolDataProvider reserve getters
//	bindings/aavepool     Aave V3 Pool
//	bindings/comet        Compound V3 Comet (utilization, supply rate)
//	bindings/morphoblue   Morpho Blue singleton market reads
//	bindings/metamorpho   MetaMorpho vault reads
//	bindings/pendleoracle Pendle PYLpOracle
//	bindings/pendlept     Pendle PT
//
// Isolating generated code here keeps the domain (pkg/types) and the ports
// (internal/brain, perceive, trigger, execute) free of chain-specific types.
// Both the ABIs and the generated code are kept in version control so builds
// are reproducible offline.
package bindings
