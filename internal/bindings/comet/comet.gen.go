// Package comet is a minimal read-only binding for a Compound III (Comet) base
// market — only the view methods the keeper's yield/risk adapters call. It
// mirrors the abigen caller shape used by internal/bindings/aavedata (a
// bind.MetaData ABI + a *bind.BoundContract caller), trimmed to a caller-only
// surface since the keeper never sends Comet transactions (the diamond facet does).
package comet

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = abi.ConvertType
)

// CometMetaData carries the trimmed Comet ABI: the two curator rate-readers
// (getUtilization, getSupplyRate) plus balanceOf and baseToken.
var CometMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"stateMutability\":\"view\",\"name\":\"getUtilization\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"function\",\"stateMutability\":\"view\",\"name\":\"getSupplyRate\",\"inputs\":[{\"name\":\"utilization\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"type\":\"function\",\"stateMutability\":\"view\",\"name\":\"balanceOf\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"function\",\"stateMutability\":\"view\",\"name\":\"baseToken\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}]}]",
}

// CometCaller is a read-only binding around a Comet market.
type CometCaller struct {
	contract *bind.BoundContract
}

// NewCometCaller binds a Comet market at address over a read-only backend.
func NewCometCaller(address common.Address, caller bind.ContractCaller) (*CometCaller, error) {
	parsed, err := CometMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return &CometCaller{contract: bind.NewBoundContract(address, *parsed, caller, nil, nil)}, nil
}

// GetUtilization returns the market's current utilization, scaled by 1e18.
func (c *CometCaller) GetUtilization(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	if err := c.contract.Call(opts, &out, "getUtilization"); err != nil {
		return new(big.Int), err
	}
	return *abi.ConvertType(out[0], new(*big.Int)).(**big.Int), nil
}

// GetSupplyRate returns the per-second supply rate (scaled 1e18) at utilization.
func (c *CometCaller) GetSupplyRate(opts *bind.CallOpts, utilization *big.Int) (uint64, error) {
	var out []interface{}
	if err := c.contract.Call(opts, &out, "getSupplyRate", utilization); err != nil {
		return *new(uint64), err
	}
	return *abi.ConvertType(out[0], new(uint64)).(*uint64), nil
}

// BalanceOf returns the present value of account's base-asset supply.
func (c *CometCaller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	if err := c.contract.Call(opts, &out, "balanceOf", account); err != nil {
		return new(big.Int), err
	}
	return *abi.ConvertType(out[0], new(*big.Int)).(**big.Int), nil
}

// BaseToken returns the market's base asset.
func (c *CometCaller) BaseToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	if err := c.contract.Call(opts, &out, "baseToken"); err != nil {
		return common.Address{}, err
	}
	return *abi.ConvertType(out[0], new(common.Address)).(*common.Address), nil
}
