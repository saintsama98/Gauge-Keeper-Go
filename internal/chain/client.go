// Package chain wraps the Ethereum RPC connection used by the perceive and
// execute layers.
package chain

// Client is a thin handle around the chain connection parameters.
//
// TODO: back this with github.com/ethereum/go-ethereum/ethclient once external
// deps are added (`go get github.com/ethereum/go-ethereum`). Kept dependency-free
// for now so the skeleton builds and runs offline.
type Client struct {
	RPCURL  string
	ChainID int64
}

// New returns a chain client handle. (No dial yet — see TODO above.)
func New(rpcURL string, chainID int64) *Client {
	return &Client{RPCURL: rpcURL, ChainID: chainID}
}
