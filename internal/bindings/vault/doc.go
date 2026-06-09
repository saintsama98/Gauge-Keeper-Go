// Package vault holds the generated Go bindings for the Vault Router diamond.
//
// Generate with abigen from the contracts repo's Foundry output
// (../vault-router-diamond):
//
//	abigen --abi abi/vault.abi.json --pkg vault --type Vault --out internal/bindings/vault/vault.gen.go
//
// The keeper needs the AllocatorFacet, HarvestFacet, GuardFacet,
// WithdrawQueueFacet and RolesFacet selectors plus the ERC-4626 views. Until the
// bindings exist, perceive.StubReader and execute.LogExecutor stand in so the
// skeleton compiles and runs offline.
package vault
