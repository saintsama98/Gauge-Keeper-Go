// Package vault holds the generated Go bindings for the Vault Router diamond.
//
// Generate them with abigen from the Foundry build output of the contracts repo
// (../vault-router-diamond):
//
//	abigen --abi out/Vault.sol/Vault.abi.json \
//	       --pkg vault --type Vault --out vault.gen.go
//
// The keeper only needs the AllocatorFacet, HarvestFacet, GuardFacet,
// WithdrawQueueFacet and RolesFacet selectors plus the ERC-4626 views. Until the
// bindings exist, perceive.StubReader and execute.LogExecutor stand in so the
// skeleton compiles and runs.
package vault
