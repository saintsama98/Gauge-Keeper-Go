# `abi/` — raw contract ABIs (source for code generation)

Drop each contract's `*.abi.json` here, then regenerate the Go bindings under
`internal/bindings/<name>/` with `make bindings` (requires `abigen` on PATH).

| ABI file | → bindings package | Source |
|---|---|---|
| `vault.abi.json` | `internal/bindings/vault` | vault-router-diamond Foundry `out/` |
| `aave-uipool.abi.json` | `internal/bindings/aavepool` | Aave V3 periphery (UiPoolDataProvider) |
| `chaos-riskoracle.abi.json` | `internal/bindings/chaosrisk` | ChaosLabsInc/chaos-agents (IRiskOracle) |
| `redstone-adapter.abi.json` | `internal/bindings/redstone` | RedStone push price-feed adapter (AggregatorV3-style) |

Keep both the ABIs here and the generated code in version control so the project
builds reproducibly without network access. ABIs are read-only inputs — never
hand-edit the generated `*.gen.go` files.
