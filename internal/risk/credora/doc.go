// Package credora implements brain.RiskProvider for the MORPHO facet. It reads
// Credora-by-RedStone risk ratings over GraphQL and maps PSL (D->A) plus
// Stress / WithdrawLiquidity / Diversification into RiskInputs{HasVendor, VendorEL}.
//
// Delivery: off-chain GraphQL (stdlib net/http + encoding/json, or genqlient).
// The HTTP client sits behind an interface; tests use recorded JSON fixtures,
// never a live endpoint.
package credora
