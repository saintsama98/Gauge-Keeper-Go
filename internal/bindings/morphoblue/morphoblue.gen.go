// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package morphoblue

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// MorphoBlueMetaData contains all meta data concerning the MorphoBlue contract.
var MorphoBlueMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"market\",\"stateMutability\":\"view\",\"inputs\":[{\"name\":\"id\",\"type\":\"bytes32\"}],\"outputs\":[{\"name\":\"totalSupplyAssets\",\"type\":\"uint128\"},{\"name\":\"totalSupplyShares\",\"type\":\"uint128\"},{\"name\":\"totalBorrowAssets\",\"type\":\"uint128\"},{\"name\":\"totalBorrowShares\",\"type\":\"uint128\"},{\"name\":\"lastUpdate\",\"type\":\"uint128\"},{\"name\":\"fee\",\"type\":\"uint128\"}]},{\"type\":\"function\",\"name\":\"position\",\"stateMutability\":\"view\",\"inputs\":[{\"name\":\"id\",\"type\":\"bytes32\"},{\"name\":\"user\",\"type\":\"address\"}],\"outputs\":[{\"name\":\"supplyShares\",\"type\":\"uint256\"},{\"name\":\"borrowShares\",\"type\":\"uint128\"},{\"name\":\"collateral\",\"type\":\"uint128\"}]}]",
}

// MorphoBlueABI is the input ABI used to generate the binding from.
// Deprecated: Use MorphoBlueMetaData.ABI instead.
var MorphoBlueABI = MorphoBlueMetaData.ABI

// MorphoBlue is an auto generated Go binding around an Ethereum contract.
type MorphoBlue struct {
	MorphoBlueCaller     // Read-only binding to the contract
	MorphoBlueTransactor // Write-only binding to the contract
	MorphoBlueFilterer   // Log filterer for contract events
}

// MorphoBlueCaller is an auto generated read-only Go binding around an Ethereum contract.
type MorphoBlueCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MorphoBlueTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MorphoBlueTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MorphoBlueFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MorphoBlueFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MorphoBlueSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MorphoBlueSession struct {
	Contract     *MorphoBlue       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MorphoBlueCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MorphoBlueCallerSession struct {
	Contract *MorphoBlueCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// MorphoBlueTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MorphoBlueTransactorSession struct {
	Contract     *MorphoBlueTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// MorphoBlueRaw is an auto generated low-level Go binding around an Ethereum contract.
type MorphoBlueRaw struct {
	Contract *MorphoBlue // Generic contract binding to access the raw methods on
}

// MorphoBlueCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MorphoBlueCallerRaw struct {
	Contract *MorphoBlueCaller // Generic read-only contract binding to access the raw methods on
}

// MorphoBlueTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MorphoBlueTransactorRaw struct {
	Contract *MorphoBlueTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMorphoBlue creates a new instance of MorphoBlue, bound to a specific deployed contract.
func NewMorphoBlue(address common.Address, backend bind.ContractBackend) (*MorphoBlue, error) {
	contract, err := bindMorphoBlue(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MorphoBlue{MorphoBlueCaller: MorphoBlueCaller{contract: contract}, MorphoBlueTransactor: MorphoBlueTransactor{contract: contract}, MorphoBlueFilterer: MorphoBlueFilterer{contract: contract}}, nil
}

// NewMorphoBlueCaller creates a new read-only instance of MorphoBlue, bound to a specific deployed contract.
func NewMorphoBlueCaller(address common.Address, caller bind.ContractCaller) (*MorphoBlueCaller, error) {
	contract, err := bindMorphoBlue(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MorphoBlueCaller{contract: contract}, nil
}

// NewMorphoBlueTransactor creates a new write-only instance of MorphoBlue, bound to a specific deployed contract.
func NewMorphoBlueTransactor(address common.Address, transactor bind.ContractTransactor) (*MorphoBlueTransactor, error) {
	contract, err := bindMorphoBlue(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MorphoBlueTransactor{contract: contract}, nil
}

// NewMorphoBlueFilterer creates a new log filterer instance of MorphoBlue, bound to a specific deployed contract.
func NewMorphoBlueFilterer(address common.Address, filterer bind.ContractFilterer) (*MorphoBlueFilterer, error) {
	contract, err := bindMorphoBlue(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MorphoBlueFilterer{contract: contract}, nil
}

// bindMorphoBlue binds a generic wrapper to an already deployed contract.
func bindMorphoBlue(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := MorphoBlueMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MorphoBlue *MorphoBlueRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MorphoBlue.Contract.MorphoBlueCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MorphoBlue *MorphoBlueRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MorphoBlue.Contract.MorphoBlueTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MorphoBlue *MorphoBlueRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MorphoBlue.Contract.MorphoBlueTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MorphoBlue *MorphoBlueCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MorphoBlue.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MorphoBlue *MorphoBlueTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MorphoBlue.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MorphoBlue *MorphoBlueTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MorphoBlue.Contract.contract.Transact(opts, method, params...)
}

// Market is a free data retrieval call binding the contract method 0x5c60e39a.
//
// Solidity: function market(bytes32 id) view returns(uint128 totalSupplyAssets, uint128 totalSupplyShares, uint128 totalBorrowAssets, uint128 totalBorrowShares, uint128 lastUpdate, uint128 fee)
func (_MorphoBlue *MorphoBlueCaller) Market(opts *bind.CallOpts, id [32]byte) (struct {
	TotalSupplyAssets *big.Int
	TotalSupplyShares *big.Int
	TotalBorrowAssets *big.Int
	TotalBorrowShares *big.Int
	LastUpdate        *big.Int
	Fee               *big.Int
}, error) {
	var out []interface{}
	err := _MorphoBlue.contract.Call(opts, &out, "market", id)

	outstruct := new(struct {
		TotalSupplyAssets *big.Int
		TotalSupplyShares *big.Int
		TotalBorrowAssets *big.Int
		TotalBorrowShares *big.Int
		LastUpdate        *big.Int
		Fee               *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.TotalSupplyAssets = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.TotalSupplyShares = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.TotalBorrowAssets = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.TotalBorrowShares = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.LastUpdate = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.Fee = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Market is a free data retrieval call binding the contract method 0x5c60e39a.
//
// Solidity: function market(bytes32 id) view returns(uint128 totalSupplyAssets, uint128 totalSupplyShares, uint128 totalBorrowAssets, uint128 totalBorrowShares, uint128 lastUpdate, uint128 fee)
func (_MorphoBlue *MorphoBlueSession) Market(id [32]byte) (struct {
	TotalSupplyAssets *big.Int
	TotalSupplyShares *big.Int
	TotalBorrowAssets *big.Int
	TotalBorrowShares *big.Int
	LastUpdate        *big.Int
	Fee               *big.Int
}, error) {
	return _MorphoBlue.Contract.Market(&_MorphoBlue.CallOpts, id)
}

// Market is a free data retrieval call binding the contract method 0x5c60e39a.
//
// Solidity: function market(bytes32 id) view returns(uint128 totalSupplyAssets, uint128 totalSupplyShares, uint128 totalBorrowAssets, uint128 totalBorrowShares, uint128 lastUpdate, uint128 fee)
func (_MorphoBlue *MorphoBlueCallerSession) Market(id [32]byte) (struct {
	TotalSupplyAssets *big.Int
	TotalSupplyShares *big.Int
	TotalBorrowAssets *big.Int
	TotalBorrowShares *big.Int
	LastUpdate        *big.Int
	Fee               *big.Int
}, error) {
	return _MorphoBlue.Contract.Market(&_MorphoBlue.CallOpts, id)
}

// Position is a free data retrieval call binding the contract method 0x93c52062.
//
// Solidity: function position(bytes32 id, address user) view returns(uint256 supplyShares, uint128 borrowShares, uint128 collateral)
func (_MorphoBlue *MorphoBlueCaller) Position(opts *bind.CallOpts, id [32]byte, user common.Address) (struct {
	SupplyShares *big.Int
	BorrowShares *big.Int
	Collateral   *big.Int
}, error) {
	var out []interface{}
	err := _MorphoBlue.contract.Call(opts, &out, "position", id, user)

	outstruct := new(struct {
		SupplyShares *big.Int
		BorrowShares *big.Int
		Collateral   *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.SupplyShares = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.BorrowShares = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Collateral = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Position is a free data retrieval call binding the contract method 0x93c52062.
//
// Solidity: function position(bytes32 id, address user) view returns(uint256 supplyShares, uint128 borrowShares, uint128 collateral)
func (_MorphoBlue *MorphoBlueSession) Position(id [32]byte, user common.Address) (struct {
	SupplyShares *big.Int
	BorrowShares *big.Int
	Collateral   *big.Int
}, error) {
	return _MorphoBlue.Contract.Position(&_MorphoBlue.CallOpts, id, user)
}

// Position is a free data retrieval call binding the contract method 0x93c52062.
//
// Solidity: function position(bytes32 id, address user) view returns(uint256 supplyShares, uint128 borrowShares, uint128 collateral)
func (_MorphoBlue *MorphoBlueCallerSession) Position(id [32]byte, user common.Address) (struct {
	SupplyShares *big.Int
	BorrowShares *big.Int
	Collateral   *big.Int
}, error) {
	return _MorphoBlue.Contract.Position(&_MorphoBlue.CallOpts, id, user)
}
