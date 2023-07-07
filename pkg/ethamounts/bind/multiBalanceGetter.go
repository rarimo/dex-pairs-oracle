// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bind

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

// MultiBalanceGetterMetaData contains all meta data concerning the MultiBalanceGetter contract.
var MultiBalanceGetterMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"tokens\",\"type\":\"address[]\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"getMultipleBalances\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// MultiBalanceGetterABI is the input ABI used to generate the binding from.
// Deprecated: Use MultiBalanceGetterMetaData.ABI instead.
var MultiBalanceGetterABI = MultiBalanceGetterMetaData.ABI

// MultiBalanceGetter is an auto generated Go binding around an Ethereum contract.
type MultiBalanceGetter struct {
	MultiBalanceGetterCaller     // Read-only binding to the contract
	MultiBalanceGetterTransactor // Write-only binding to the contract
	MultiBalanceGetterFilterer   // Log filterer for contract events
}

// MultiBalanceGetterCaller is an auto generated read-only Go binding around an Ethereum contract.
type MultiBalanceGetterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultiBalanceGetterTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MultiBalanceGetterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultiBalanceGetterFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MultiBalanceGetterFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultiBalanceGetterSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MultiBalanceGetterSession struct {
	Contract     *MultiBalanceGetter // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// MultiBalanceGetterCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MultiBalanceGetterCallerSession struct {
	Contract *MultiBalanceGetterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// MultiBalanceGetterTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MultiBalanceGetterTransactorSession struct {
	Contract     *MultiBalanceGetterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// MultiBalanceGetterRaw is an auto generated low-level Go binding around an Ethereum contract.
type MultiBalanceGetterRaw struct {
	Contract *MultiBalanceGetter // Generic contract binding to access the raw methods on
}

// MultiBalanceGetterCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MultiBalanceGetterCallerRaw struct {
	Contract *MultiBalanceGetterCaller // Generic read-only contract binding to access the raw methods on
}

// MultiBalanceGetterTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MultiBalanceGetterTransactorRaw struct {
	Contract *MultiBalanceGetterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMultiBalanceGetter creates a new instance of MultiBalanceGetter, bound to a specific deployed contract.
func NewMultiBalanceGetter(address common.Address, backend bind.ContractBackend) (*MultiBalanceGetter, error) {
	contract, err := bindMultiBalanceGetter(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MultiBalanceGetter{MultiBalanceGetterCaller: MultiBalanceGetterCaller{contract: contract}, MultiBalanceGetterTransactor: MultiBalanceGetterTransactor{contract: contract}, MultiBalanceGetterFilterer: MultiBalanceGetterFilterer{contract: contract}}, nil
}

// NewMultiBalanceGetterCaller creates a new read-only instance of MultiBalanceGetter, bound to a specific deployed contract.
func NewMultiBalanceGetterCaller(address common.Address, caller bind.ContractCaller) (*MultiBalanceGetterCaller, error) {
	contract, err := bindMultiBalanceGetter(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MultiBalanceGetterCaller{contract: contract}, nil
}

// NewMultiBalanceGetterTransactor creates a new write-only instance of MultiBalanceGetter, bound to a specific deployed contract.
func NewMultiBalanceGetterTransactor(address common.Address, transactor bind.ContractTransactor) (*MultiBalanceGetterTransactor, error) {
	contract, err := bindMultiBalanceGetter(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MultiBalanceGetterTransactor{contract: contract}, nil
}

// NewMultiBalanceGetterFilterer creates a new log filterer instance of MultiBalanceGetter, bound to a specific deployed contract.
func NewMultiBalanceGetterFilterer(address common.Address, filterer bind.ContractFilterer) (*MultiBalanceGetterFilterer, error) {
	contract, err := bindMultiBalanceGetter(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MultiBalanceGetterFilterer{contract: contract}, nil
}

// bindMultiBalanceGetter binds a generic wrapper to an already deployed contract.
func bindMultiBalanceGetter(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := MultiBalanceGetterMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MultiBalanceGetter *MultiBalanceGetterRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MultiBalanceGetter.Contract.MultiBalanceGetterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MultiBalanceGetter *MultiBalanceGetterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MultiBalanceGetter.Contract.MultiBalanceGetterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MultiBalanceGetter *MultiBalanceGetterRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MultiBalanceGetter.Contract.MultiBalanceGetterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MultiBalanceGetter *MultiBalanceGetterCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MultiBalanceGetter.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MultiBalanceGetter *MultiBalanceGetterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MultiBalanceGetter.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MultiBalanceGetter *MultiBalanceGetterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MultiBalanceGetter.Contract.contract.Transact(opts, method, params...)
}

// GetMultipleBalances is a free data retrieval call binding the contract method 0xa55ee84e.
//
// Solidity: function getMultipleBalances(address[] tokens, address account) view returns(uint256, uint256[])
func (_MultiBalanceGetter *MultiBalanceGetterCaller) GetMultipleBalances(opts *bind.CallOpts, tokens []common.Address, account common.Address) (*big.Int, []*big.Int, error) {
	var out []interface{}
	err := _MultiBalanceGetter.contract.Call(opts, &out, "getMultipleBalances", tokens, account)

	if err != nil {
		return *new(*big.Int), *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	out1 := *abi.ConvertType(out[1], new([]*big.Int)).(*[]*big.Int)

	return out0, out1, err

}

// GetMultipleBalances is a free data retrieval call binding the contract method 0xa55ee84e.
//
// Solidity: function getMultipleBalances(address[] tokens, address account) view returns(uint256, uint256[])
func (_MultiBalanceGetter *MultiBalanceGetterSession) GetMultipleBalances(tokens []common.Address, account common.Address) (*big.Int, []*big.Int, error) {
	return _MultiBalanceGetter.Contract.GetMultipleBalances(&_MultiBalanceGetter.CallOpts, tokens, account)
}

// GetMultipleBalances is a free data retrieval call binding the contract method 0xa55ee84e.
//
// Solidity: function getMultipleBalances(address[] tokens, address account) view returns(uint256, uint256[])
func (_MultiBalanceGetter *MultiBalanceGetterCallerSession) GetMultipleBalances(tokens []common.Address, account common.Address) (*big.Int, []*big.Int, error) {
	return _MultiBalanceGetter.Contract.GetMultipleBalances(&_MultiBalanceGetter.CallOpts, tokens, account)
}
