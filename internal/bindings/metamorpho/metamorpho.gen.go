// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package metamorpho

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

// MetaMorphoMetaData contains all meta data concerning the MetaMorpho contract.
var MetaMorphoMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"withdrawQueueLength\",\"stateMutability\":\"view\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}]},{\"type\":\"function\",\"name\":\"withdrawQueue\",\"stateMutability\":\"view\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}]},{\"type\":\"function\",\"name\":\"config\",\"stateMutability\":\"view\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"outputs\":[{\"name\":\"cap\",\"type\":\"uint184\"},{\"name\":\"enabled\",\"type\":\"bool\"},{\"name\":\"removableAt\",\"type\":\"uint64\"}]}]",
}

// MetaMorphoABI is the input ABI used to generate the binding from.
// Deprecated: Use MetaMorphoMetaData.ABI instead.
var MetaMorphoABI = MetaMorphoMetaData.ABI

// MetaMorpho is an auto generated Go binding around an Ethereum contract.
type MetaMorpho struct {
	MetaMorphoCaller     // Read-only binding to the contract
	MetaMorphoTransactor // Write-only binding to the contract
	MetaMorphoFilterer   // Log filterer for contract events
}

// MetaMorphoCaller is an auto generated read-only Go binding around an Ethereum contract.
type MetaMorphoCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MetaMorphoTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MetaMorphoTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MetaMorphoFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MetaMorphoFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MetaMorphoSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MetaMorphoSession struct {
	Contract     *MetaMorpho       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MetaMorphoCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MetaMorphoCallerSession struct {
	Contract *MetaMorphoCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// MetaMorphoTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MetaMorphoTransactorSession struct {
	Contract     *MetaMorphoTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// MetaMorphoRaw is an auto generated low-level Go binding around an Ethereum contract.
type MetaMorphoRaw struct {
	Contract *MetaMorpho // Generic contract binding to access the raw methods on
}

// MetaMorphoCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MetaMorphoCallerRaw struct {
	Contract *MetaMorphoCaller // Generic read-only contract binding to access the raw methods on
}

// MetaMorphoTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MetaMorphoTransactorRaw struct {
	Contract *MetaMorphoTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMetaMorpho creates a new instance of MetaMorpho, bound to a specific deployed contract.
func NewMetaMorpho(address common.Address, backend bind.ContractBackend) (*MetaMorpho, error) {
	contract, err := bindMetaMorpho(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MetaMorpho{MetaMorphoCaller: MetaMorphoCaller{contract: contract}, MetaMorphoTransactor: MetaMorphoTransactor{contract: contract}, MetaMorphoFilterer: MetaMorphoFilterer{contract: contract}}, nil
}

// NewMetaMorphoCaller creates a new read-only instance of MetaMorpho, bound to a specific deployed contract.
func NewMetaMorphoCaller(address common.Address, caller bind.ContractCaller) (*MetaMorphoCaller, error) {
	contract, err := bindMetaMorpho(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MetaMorphoCaller{contract: contract}, nil
}

// NewMetaMorphoTransactor creates a new write-only instance of MetaMorpho, bound to a specific deployed contract.
func NewMetaMorphoTransactor(address common.Address, transactor bind.ContractTransactor) (*MetaMorphoTransactor, error) {
	contract, err := bindMetaMorpho(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MetaMorphoTransactor{contract: contract}, nil
}

// NewMetaMorphoFilterer creates a new log filterer instance of MetaMorpho, bound to a specific deployed contract.
func NewMetaMorphoFilterer(address common.Address, filterer bind.ContractFilterer) (*MetaMorphoFilterer, error) {
	contract, err := bindMetaMorpho(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MetaMorphoFilterer{contract: contract}, nil
}

// bindMetaMorpho binds a generic wrapper to an already deployed contract.
func bindMetaMorpho(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := MetaMorphoMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MetaMorpho *MetaMorphoRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MetaMorpho.Contract.MetaMorphoCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MetaMorpho *MetaMorphoRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MetaMorpho.Contract.MetaMorphoTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MetaMorpho *MetaMorphoRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MetaMorpho.Contract.MetaMorphoTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MetaMorpho *MetaMorphoCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MetaMorpho.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MetaMorpho *MetaMorphoTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MetaMorpho.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MetaMorpho *MetaMorphoTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MetaMorpho.Contract.contract.Transact(opts, method, params...)
}

// Config is a free data retrieval call binding the contract method 0xcc718f76.
//
// Solidity: function config(bytes32 ) view returns(uint184 cap, bool enabled, uint64 removableAt)
func (_MetaMorpho *MetaMorphoCaller) Config(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Cap         *big.Int
	Enabled     bool
	RemovableAt uint64
}, error) {
	var out []interface{}
	err := _MetaMorpho.contract.Call(opts, &out, "config", arg0)

	outstruct := new(struct {
		Cap         *big.Int
		Enabled     bool
		RemovableAt uint64
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Cap = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Enabled = *abi.ConvertType(out[1], new(bool)).(*bool)
	outstruct.RemovableAt = *abi.ConvertType(out[2], new(uint64)).(*uint64)

	return *outstruct, err

}

// Config is a free data retrieval call binding the contract method 0xcc718f76.
//
// Solidity: function config(bytes32 ) view returns(uint184 cap, bool enabled, uint64 removableAt)
func (_MetaMorpho *MetaMorphoSession) Config(arg0 [32]byte) (struct {
	Cap         *big.Int
	Enabled     bool
	RemovableAt uint64
}, error) {
	return _MetaMorpho.Contract.Config(&_MetaMorpho.CallOpts, arg0)
}

// Config is a free data retrieval call binding the contract method 0xcc718f76.
//
// Solidity: function config(bytes32 ) view returns(uint184 cap, bool enabled, uint64 removableAt)
func (_MetaMorpho *MetaMorphoCallerSession) Config(arg0 [32]byte) (struct {
	Cap         *big.Int
	Enabled     bool
	RemovableAt uint64
}, error) {
	return _MetaMorpho.Contract.Config(&_MetaMorpho.CallOpts, arg0)
}

// WithdrawQueue is a free data retrieval call binding the contract method 0x62518ddf.
//
// Solidity: function withdrawQueue(uint256 ) view returns(bytes32)
func (_MetaMorpho *MetaMorphoCaller) WithdrawQueue(opts *bind.CallOpts, arg0 *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _MetaMorpho.contract.Call(opts, &out, "withdrawQueue", arg0)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// WithdrawQueue is a free data retrieval call binding the contract method 0x62518ddf.
//
// Solidity: function withdrawQueue(uint256 ) view returns(bytes32)
func (_MetaMorpho *MetaMorphoSession) WithdrawQueue(arg0 *big.Int) ([32]byte, error) {
	return _MetaMorpho.Contract.WithdrawQueue(&_MetaMorpho.CallOpts, arg0)
}

// WithdrawQueue is a free data retrieval call binding the contract method 0x62518ddf.
//
// Solidity: function withdrawQueue(uint256 ) view returns(bytes32)
func (_MetaMorpho *MetaMorphoCallerSession) WithdrawQueue(arg0 *big.Int) ([32]byte, error) {
	return _MetaMorpho.Contract.WithdrawQueue(&_MetaMorpho.CallOpts, arg0)
}

// WithdrawQueueLength is a free data retrieval call binding the contract method 0x33f91ebb.
//
// Solidity: function withdrawQueueLength() view returns(uint256)
func (_MetaMorpho *MetaMorphoCaller) WithdrawQueueLength(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MetaMorpho.contract.Call(opts, &out, "withdrawQueueLength")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// WithdrawQueueLength is a free data retrieval call binding the contract method 0x33f91ebb.
//
// Solidity: function withdrawQueueLength() view returns(uint256)
func (_MetaMorpho *MetaMorphoSession) WithdrawQueueLength() (*big.Int, error) {
	return _MetaMorpho.Contract.WithdrawQueueLength(&_MetaMorpho.CallOpts)
}

// WithdrawQueueLength is a free data retrieval call binding the contract method 0x33f91ebb.
//
// Solidity: function withdrawQueueLength() view returns(uint256)
func (_MetaMorpho *MetaMorphoCallerSession) WithdrawQueueLength() (*big.Int, error) {
	return _MetaMorpho.Contract.WithdrawQueueLength(&_MetaMorpho.CallOpts)
}
