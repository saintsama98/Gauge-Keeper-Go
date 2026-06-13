// Package morpho provides the LIVE signals for MetaMorpho (ERC-4626 curated
// vault) strategies, over two ports:
//
//   - perceive.YieldProvider (YieldReader, yield.go) — the vault's current net
//     APY from Morpho's public Blue GraphQL API
//     (https://blue-api.morpho.org/graphql — NO auth required).
//   - brain.RiskProvider (MorphoBlueReader, onchain.go) — the KEYLESS on-chain
//     risk signal: it walks each vault's withdraw queue and reads per-market
//     supply/borrow/utilization from the Morpho Blue singleton, deriving an
//     expected loss with NO API key and NO subscription.
//
// The on-chain reader is the free alternative to the auth-gated Credora rating
// feed (internal/risk/credora): the public Blue API exposes net APY but NO risk
// rating, and Credora's PSL feed is paid/request-access. Credora is therefore
// retained only as an OPTIONAL paid override — wired when an operator supplies
// an endpoint, and otherwise no longer shadowing this on-chain reader. With
// neither, the brain prices Morpho at the closed-form floor (EL 0) — no risk
// gate — which is the hole this package's risk reader closes.
//
// Grounding rule (same as every adapter here): no fabricated numbers. The APY
// returned traces 1:1 to the API's `vaultByAddress.state.netApy` (the
// depositor-realized rate net of vault fees); any transport/shape/zero failure
// degrades to ok=false — the honest no-data path.
package morpho
