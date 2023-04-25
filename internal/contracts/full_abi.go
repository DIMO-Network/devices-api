// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

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

// FullAbiMetaData contains all meta data concerning the FullAbi contract.
var FullAbiMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"UintUtils__InsufficientHexLength\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"moduleAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes4[]\",\"name\":\"selectors\",\"type\":\"bytes4[]\"}],\"name\":\"ModuleAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"moduleAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes4[]\",\"name\":\"selectors\",\"type\":\"bytes4[]\"}],\"name\":\"ModuleRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldImplementation\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes4[]\",\"name\":\"oldSelectors\",\"type\":\"bytes4[]\"},{\"indexed\":false,\"internalType\":\"bytes4[]\",\"name\":\"newSelectors\",\"type\":\"bytes4[]\"}],\"name\":\"ModuleUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"},{\"internalType\":\"bytes4[]\",\"name\":\"selectors\",\"type\":\"bytes4[]\"}],\"name\":\"addModule\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"},{\"internalType\":\"bytes4[]\",\"name\":\"selectors\",\"type\":\"bytes4[]\"}],\"name\":\"removeModule\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"oldImplementation\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes4[]\",\"name\":\"oldSelectors\",\"type\":\"bytes4[]\"},{\"internalType\":\"bytes4[]\",\"name\":\"newSelectors\",\"type\":\"bytes4[]\"}],\"name\":\"updateModule\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"name_\",\"type\":\"string\"}],\"name\":\"NameChanged\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"name_\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"string\",\"name\":\"name_\",\"type\":\"string\"}],\"name\":\"setName\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"dcnManager\",\"type\":\"address\"}],\"name\":\"setDcnManager\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"dcnRegistry\",\"type\":\"address\"}],\"name\":\"setDcnRegistry\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"dimoToken\",\"type\":\"address\"}],\"name\":\"setDimoToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"foundation\",\"type\":\"address\"}],\"name\":\"setFoundationAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"vehicleId_\",\"type\":\"uint256\"}],\"name\":\"VehicleIdChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"proxy\",\"type\":\"address\"}],\"name\":\"VehicleIdProxySet\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"vehicleId_\",\"type\":\"uint256\"}],\"name\":\"nodeByVehicleId\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"vehicleId_\",\"type\":\"uint256\"}],\"name\":\"resetVehicleId\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"vehicleId_\",\"type\":\"uint256\"}],\"name\":\"setVehicleId\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"setVehicleIdProxyAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"name\":\"vehicleId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"vehicleId_\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes[]\",\"name\":\"data\",\"type\":\"bytes[]\"}],\"name\":\"multiDelegateCall\",\"outputs\":[{\"internalType\":\"bytes[]\",\"name\":\"results\",\"type\":\"bytes[]\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes[]\",\"name\":\"data\",\"type\":\"bytes[]\"}],\"name\":\"multiStaticCall\",\"outputs\":[{\"internalType\":\"bytes[]\",\"name\":\"results\",\"type\":\"bytes[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// FullAbiABI is the input ABI used to generate the binding from.
// Deprecated: Use FullAbiMetaData.ABI instead.
var FullAbiABI = FullAbiMetaData.ABI

// FullAbi is an auto generated Go binding around an Ethereum contract.
type FullAbi struct {
	FullAbiCaller     // Read-only binding to the contract
	FullAbiTransactor // Write-only binding to the contract
	FullAbiFilterer   // Log filterer for contract events
}

// FullAbiCaller is an auto generated read-only Go binding around an Ethereum contract.
type FullAbiCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FullAbiTransactor is an auto generated write-only Go binding around an Ethereum contract.
type FullAbiTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FullAbiFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type FullAbiFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FullAbiSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type FullAbiSession struct {
	Contract     *FullAbi          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// FullAbiCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type FullAbiCallerSession struct {
	Contract *FullAbiCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// FullAbiTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type FullAbiTransactorSession struct {
	Contract     *FullAbiTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// FullAbiRaw is an auto generated low-level Go binding around an Ethereum contract.
type FullAbiRaw struct {
	Contract *FullAbi // Generic contract binding to access the raw methods on
}

// FullAbiCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type FullAbiCallerRaw struct {
	Contract *FullAbiCaller // Generic read-only contract binding to access the raw methods on
}

// FullAbiTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type FullAbiTransactorRaw struct {
	Contract *FullAbiTransactor // Generic write-only contract binding to access the raw methods on
}

// NewFullAbi creates a new instance of FullAbi, bound to a specific deployed contract.
func NewFullAbi(address common.Address, backend bind.ContractBackend) (*FullAbi, error) {
	contract, err := bindFullAbi(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &FullAbi{FullAbiCaller: FullAbiCaller{contract: contract}, FullAbiTransactor: FullAbiTransactor{contract: contract}, FullAbiFilterer: FullAbiFilterer{contract: contract}}, nil
}

// NewFullAbiCaller creates a new read-only instance of FullAbi, bound to a specific deployed contract.
func NewFullAbiCaller(address common.Address, caller bind.ContractCaller) (*FullAbiCaller, error) {
	contract, err := bindFullAbi(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &FullAbiCaller{contract: contract}, nil
}

// NewFullAbiTransactor creates a new write-only instance of FullAbi, bound to a specific deployed contract.
func NewFullAbiTransactor(address common.Address, transactor bind.ContractTransactor) (*FullAbiTransactor, error) {
	contract, err := bindFullAbi(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &FullAbiTransactor{contract: contract}, nil
}

// NewFullAbiFilterer creates a new log filterer instance of FullAbi, bound to a specific deployed contract.
func NewFullAbiFilterer(address common.Address, filterer bind.ContractFilterer) (*FullAbiFilterer, error) {
	contract, err := bindFullAbi(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &FullAbiFilterer{contract: contract}, nil
}

// bindFullAbi binds a generic wrapper to an already deployed contract.
func bindFullAbi(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := FullAbiMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FullAbi *FullAbiRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FullAbi.Contract.FullAbiCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FullAbi *FullAbiRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FullAbi.Contract.FullAbiTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FullAbi *FullAbiRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FullAbi.Contract.FullAbiTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FullAbi *FullAbiCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FullAbi.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FullAbi *FullAbiTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FullAbi.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FullAbi *FullAbiTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FullAbi.Contract.contract.Transact(opts, method, params...)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_FullAbi *FullAbiCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _FullAbi.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_FullAbi *FullAbiSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _FullAbi.Contract.GetRoleAdmin(&_FullAbi.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_FullAbi *FullAbiCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _FullAbi.Contract.GetRoleAdmin(&_FullAbi.CallOpts, role)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_FullAbi *FullAbiCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _FullAbi.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_FullAbi *FullAbiSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _FullAbi.Contract.HasRole(&_FullAbi.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_FullAbi *FullAbiCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _FullAbi.Contract.HasRole(&_FullAbi.CallOpts, role, account)
}

// MultiStaticCall is a free data retrieval call binding the contract method 0x1c0c6e51.
//
// Solidity: function multiStaticCall(bytes[] data) view returns(bytes[] results)
func (_FullAbi *FullAbiCaller) MultiStaticCall(opts *bind.CallOpts, data [][]byte) ([][]byte, error) {
	var out []interface{}
	err := _FullAbi.contract.Call(opts, &out, "multiStaticCall", data)

	if err != nil {
		return *new([][]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][]byte)).(*[][]byte)

	return out0, err

}

// MultiStaticCall is a free data retrieval call binding the contract method 0x1c0c6e51.
//
// Solidity: function multiStaticCall(bytes[] data) view returns(bytes[] results)
func (_FullAbi *FullAbiSession) MultiStaticCall(data [][]byte) ([][]byte, error) {
	return _FullAbi.Contract.MultiStaticCall(&_FullAbi.CallOpts, data)
}

// MultiStaticCall is a free data retrieval call binding the contract method 0x1c0c6e51.
//
// Solidity: function multiStaticCall(bytes[] data) view returns(bytes[] results)
func (_FullAbi *FullAbiCallerSession) MultiStaticCall(data [][]byte) ([][]byte, error) {
	return _FullAbi.Contract.MultiStaticCall(&_FullAbi.CallOpts, data)
}

// Name is a free data retrieval call binding the contract method 0x691f3431.
//
// Solidity: function name(bytes32 node) view returns(string name_)
func (_FullAbi *FullAbiCaller) Name(opts *bind.CallOpts, node [32]byte) (string, error) {
	var out []interface{}
	err := _FullAbi.contract.Call(opts, &out, "name", node)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x691f3431.
//
// Solidity: function name(bytes32 node) view returns(string name_)
func (_FullAbi *FullAbiSession) Name(node [32]byte) (string, error) {
	return _FullAbi.Contract.Name(&_FullAbi.CallOpts, node)
}

// Name is a free data retrieval call binding the contract method 0x691f3431.
//
// Solidity: function name(bytes32 node) view returns(string name_)
func (_FullAbi *FullAbiCallerSession) Name(node [32]byte) (string, error) {
	return _FullAbi.Contract.Name(&_FullAbi.CallOpts, node)
}

// NodeByVehicleId is a free data retrieval call binding the contract method 0x01e11675.
//
// Solidity: function nodeByVehicleId(uint256 vehicleId_) view returns(bytes32 node)
func (_FullAbi *FullAbiCaller) NodeByVehicleId(opts *bind.CallOpts, vehicleId_ *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _FullAbi.contract.Call(opts, &out, "nodeByVehicleId", vehicleId_)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// NodeByVehicleId is a free data retrieval call binding the contract method 0x01e11675.
//
// Solidity: function nodeByVehicleId(uint256 vehicleId_) view returns(bytes32 node)
func (_FullAbi *FullAbiSession) NodeByVehicleId(vehicleId_ *big.Int) ([32]byte, error) {
	return _FullAbi.Contract.NodeByVehicleId(&_FullAbi.CallOpts, vehicleId_)
}

// NodeByVehicleId is a free data retrieval call binding the contract method 0x01e11675.
//
// Solidity: function nodeByVehicleId(uint256 vehicleId_) view returns(bytes32 node)
func (_FullAbi *FullAbiCallerSession) NodeByVehicleId(vehicleId_ *big.Int) ([32]byte, error) {
	return _FullAbi.Contract.NodeByVehicleId(&_FullAbi.CallOpts, vehicleId_)
}

// VehicleId is a free data retrieval call binding the contract method 0x24f9e9da.
//
// Solidity: function vehicleId(bytes32 node) view returns(uint256 vehicleId_)
func (_FullAbi *FullAbiCaller) VehicleId(opts *bind.CallOpts, node [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _FullAbi.contract.Call(opts, &out, "vehicleId", node)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VehicleId is a free data retrieval call binding the contract method 0x24f9e9da.
//
// Solidity: function vehicleId(bytes32 node) view returns(uint256 vehicleId_)
func (_FullAbi *FullAbiSession) VehicleId(node [32]byte) (*big.Int, error) {
	return _FullAbi.Contract.VehicleId(&_FullAbi.CallOpts, node)
}

// VehicleId is a free data retrieval call binding the contract method 0x24f9e9da.
//
// Solidity: function vehicleId(bytes32 node) view returns(uint256 vehicleId_)
func (_FullAbi *FullAbiCallerSession) VehicleId(node [32]byte) (*big.Int, error) {
	return _FullAbi.Contract.VehicleId(&_FullAbi.CallOpts, node)
}

// AddModule is a paid mutator transaction binding the contract method 0x0df5b997.
//
// Solidity: function addModule(address implementation, bytes4[] selectors) returns()
func (_FullAbi *FullAbiTransactor) AddModule(opts *bind.TransactOpts, implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _FullAbi.contract.Transact(opts, "addModule", implementation, selectors)
}

// AddModule is a paid mutator transaction binding the contract method 0x0df5b997.
//
// Solidity: function addModule(address implementation, bytes4[] selectors) returns()
func (_FullAbi *FullAbiSession) AddModule(implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _FullAbi.Contract.AddModule(&_FullAbi.TransactOpts, implementation, selectors)
}

// AddModule is a paid mutator transaction binding the contract method 0x0df5b997.
//
// Solidity: function addModule(address implementation, bytes4[] selectors) returns()
func (_FullAbi *FullAbiTransactorSession) AddModule(implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _FullAbi.Contract.AddModule(&_FullAbi.TransactOpts, implementation, selectors)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_FullAbi *FullAbiTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _FullAbi.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_FullAbi *FullAbiSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _FullAbi.Contract.GrantRole(&_FullAbi.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_FullAbi *FullAbiTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _FullAbi.Contract.GrantRole(&_FullAbi.TransactOpts, role, account)
}

// MultiDelegateCall is a paid mutator transaction binding the contract method 0x415c2d96.
//
// Solidity: function multiDelegateCall(bytes[] data) returns(bytes[] results)
func (_FullAbi *FullAbiTransactor) MultiDelegateCall(opts *bind.TransactOpts, data [][]byte) (*types.Transaction, error) {
	return _FullAbi.contract.Transact(opts, "multiDelegateCall", data)
}

// MultiDelegateCall is a paid mutator transaction binding the contract method 0x415c2d96.
//
// Solidity: function multiDelegateCall(bytes[] data) returns(bytes[] results)
func (_FullAbi *FullAbiSession) MultiDelegateCall(data [][]byte) (*types.Transaction, error) {
	return _FullAbi.Contract.MultiDelegateCall(&_FullAbi.TransactOpts, data)
}

// MultiDelegateCall is a paid mutator transaction binding the contract method 0x415c2d96.
//
// Solidity: function multiDelegateCall(bytes[] data) returns(bytes[] results)
func (_FullAbi *FullAbiTransactorSession) MultiDelegateCall(data [][]byte) (*types.Transaction, error) {
	return _FullAbi.Contract.MultiDelegateCall(&_FullAbi.TransactOpts, data)
}

// RemoveModule is a paid mutator transaction binding the contract method 0x9748a762.
//
// Solidity: function removeModule(address implementation, bytes4[] selectors) returns()
func (_FullAbi *FullAbiTransactor) RemoveModule(opts *bind.TransactOpts, implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _FullAbi.contract.Transact(opts, "removeModule", implementation, selectors)
}

// RemoveModule is a paid mutator transaction binding the contract method 0x9748a762.
//
// Solidity: function removeModule(address implementation, bytes4[] selectors) returns()
func (_FullAbi *FullAbiSession) RemoveModule(implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _FullAbi.Contract.RemoveModule(&_FullAbi.TransactOpts, implementation, selectors)
}

// RemoveModule is a paid mutator transaction binding the contract method 0x9748a762.
//
// Solidity: function removeModule(address implementation, bytes4[] selectors) returns()
func (_FullAbi *FullAbiTransactorSession) RemoveModule(implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _FullAbi.Contract.RemoveModule(&_FullAbi.TransactOpts, implementation, selectors)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x8bb9c5bf.
//
// Solidity: function renounceRole(bytes32 role) returns()
func (_FullAbi *FullAbiTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte) (*types.Transaction, error) {
	return _FullAbi.contract.Transact(opts, "renounceRole", role)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x8bb9c5bf.
//
// Solidity: function renounceRole(bytes32 role) returns()
func (_FullAbi *FullAbiSession) RenounceRole(role [32]byte) (*types.Transaction, error) {
	return _FullAbi.Contract.RenounceRole(&_FullAbi.TransactOpts, role)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x8bb9c5bf.
//
// Solidity: function renounceRole(bytes32 role) returns()
func (_FullAbi *FullAbiTransactorSession) RenounceRole(role [32]byte) (*types.Transaction, error) {
	return _FullAbi.Contract.RenounceRole(&_FullAbi.TransactOpts, role)
}

// ResetVehicleId is a paid mutator transaction binding the contract method 0x1bd59757.
//
// Solidity: function resetVehicleId(bytes32 node, uint256 vehicleId_) returns()
func (_FullAbi *FullAbiTransactor) ResetVehicleId(opts *bind.TransactOpts, node [32]byte, vehicleId_ *big.Int) (*types.Transaction, error) {
	return _FullAbi.contract.Transact(opts, "resetVehicleId", node, vehicleId_)
}

// ResetVehicleId is a paid mutator transaction binding the contract method 0x1bd59757.
//
// Solidity: function resetVehicleId(bytes32 node, uint256 vehicleId_) returns()
func (_FullAbi *FullAbiSession) ResetVehicleId(node [32]byte, vehicleId_ *big.Int) (*types.Transaction, error) {
	return _FullAbi.Contract.ResetVehicleId(&_FullAbi.TransactOpts, node, vehicleId_)
}

// ResetVehicleId is a paid mutator transaction binding the contract method 0x1bd59757.
//
// Solidity: function resetVehicleId(bytes32 node, uint256 vehicleId_) returns()
func (_FullAbi *FullAbiTransactorSession) ResetVehicleId(node [32]byte, vehicleId_ *big.Int) (*types.Transaction, error) {
	return _FullAbi.Contract.ResetVehicleId(&_FullAbi.TransactOpts, node, vehicleId_)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_FullAbi *FullAbiTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _FullAbi.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_FullAbi *FullAbiSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _FullAbi.Contract.RevokeRole(&_FullAbi.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_FullAbi *FullAbiTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _FullAbi.Contract.RevokeRole(&_FullAbi.TransactOpts, role, account)
}

// SetDcnManager is a paid mutator transaction binding the contract method 0x56f98b37.
//
// Solidity: function setDcnManager(address dcnManager) returns()
func (_FullAbi *FullAbiTransactor) SetDcnManager(opts *bind.TransactOpts, dcnManager common.Address) (*types.Transaction, error) {
	return _FullAbi.contract.Transact(opts, "setDcnManager", dcnManager)
}

// SetDcnManager is a paid mutator transaction binding the contract method 0x56f98b37.
//
// Solidity: function setDcnManager(address dcnManager) returns()
func (_FullAbi *FullAbiSession) SetDcnManager(dcnManager common.Address) (*types.Transaction, error) {
	return _FullAbi.Contract.SetDcnManager(&_FullAbi.TransactOpts, dcnManager)
}

// SetDcnManager is a paid mutator transaction binding the contract method 0x56f98b37.
//
// Solidity: function setDcnManager(address dcnManager) returns()
func (_FullAbi *FullAbiTransactorSession) SetDcnManager(dcnManager common.Address) (*types.Transaction, error) {
	return _FullAbi.Contract.SetDcnManager(&_FullAbi.TransactOpts, dcnManager)
}

// SetDcnRegistry is a paid mutator transaction binding the contract method 0xa1caf728.
//
// Solidity: function setDcnRegistry(address dcnRegistry) returns()
func (_FullAbi *FullAbiTransactor) SetDcnRegistry(opts *bind.TransactOpts, dcnRegistry common.Address) (*types.Transaction, error) {
	return _FullAbi.contract.Transact(opts, "setDcnRegistry", dcnRegistry)
}

// SetDcnRegistry is a paid mutator transaction binding the contract method 0xa1caf728.
//
// Solidity: function setDcnRegistry(address dcnRegistry) returns()
func (_FullAbi *FullAbiSession) SetDcnRegistry(dcnRegistry common.Address) (*types.Transaction, error) {
	return _FullAbi.Contract.SetDcnRegistry(&_FullAbi.TransactOpts, dcnRegistry)
}

// SetDcnRegistry is a paid mutator transaction binding the contract method 0xa1caf728.
//
// Solidity: function setDcnRegistry(address dcnRegistry) returns()
func (_FullAbi *FullAbiTransactorSession) SetDcnRegistry(dcnRegistry common.Address) (*types.Transaction, error) {
	return _FullAbi.Contract.SetDcnRegistry(&_FullAbi.TransactOpts, dcnRegistry)
}

// SetDimoToken is a paid mutator transaction binding the contract method 0x5b6c1979.
//
// Solidity: function setDimoToken(address dimoToken) returns()
func (_FullAbi *FullAbiTransactor) SetDimoToken(opts *bind.TransactOpts, dimoToken common.Address) (*types.Transaction, error) {
	return _FullAbi.contract.Transact(opts, "setDimoToken", dimoToken)
}

// SetDimoToken is a paid mutator transaction binding the contract method 0x5b6c1979.
//
// Solidity: function setDimoToken(address dimoToken) returns()
func (_FullAbi *FullAbiSession) SetDimoToken(dimoToken common.Address) (*types.Transaction, error) {
	return _FullAbi.Contract.SetDimoToken(&_FullAbi.TransactOpts, dimoToken)
}

// SetDimoToken is a paid mutator transaction binding the contract method 0x5b6c1979.
//
// Solidity: function setDimoToken(address dimoToken) returns()
func (_FullAbi *FullAbiTransactorSession) SetDimoToken(dimoToken common.Address) (*types.Transaction, error) {
	return _FullAbi.Contract.SetDimoToken(&_FullAbi.TransactOpts, dimoToken)
}

// SetFoundationAddress is a paid mutator transaction binding the contract method 0xf41377ca.
//
// Solidity: function setFoundationAddress(address foundation) returns()
func (_FullAbi *FullAbiTransactor) SetFoundationAddress(opts *bind.TransactOpts, foundation common.Address) (*types.Transaction, error) {
	return _FullAbi.contract.Transact(opts, "setFoundationAddress", foundation)
}

// SetFoundationAddress is a paid mutator transaction binding the contract method 0xf41377ca.
//
// Solidity: function setFoundationAddress(address foundation) returns()
func (_FullAbi *FullAbiSession) SetFoundationAddress(foundation common.Address) (*types.Transaction, error) {
	return _FullAbi.Contract.SetFoundationAddress(&_FullAbi.TransactOpts, foundation)
}

// SetFoundationAddress is a paid mutator transaction binding the contract method 0xf41377ca.
//
// Solidity: function setFoundationAddress(address foundation) returns()
func (_FullAbi *FullAbiTransactorSession) SetFoundationAddress(foundation common.Address) (*types.Transaction, error) {
	return _FullAbi.Contract.SetFoundationAddress(&_FullAbi.TransactOpts, foundation)
}

// SetName is a paid mutator transaction binding the contract method 0x77372213.
//
// Solidity: function setName(bytes32 node, string name_) returns()
func (_FullAbi *FullAbiTransactor) SetName(opts *bind.TransactOpts, node [32]byte, name_ string) (*types.Transaction, error) {
	return _FullAbi.contract.Transact(opts, "setName", node, name_)
}

// SetName is a paid mutator transaction binding the contract method 0x77372213.
//
// Solidity: function setName(bytes32 node, string name_) returns()
func (_FullAbi *FullAbiSession) SetName(node [32]byte, name_ string) (*types.Transaction, error) {
	return _FullAbi.Contract.SetName(&_FullAbi.TransactOpts, node, name_)
}

// SetName is a paid mutator transaction binding the contract method 0x77372213.
//
// Solidity: function setName(bytes32 node, string name_) returns()
func (_FullAbi *FullAbiTransactorSession) SetName(node [32]byte, name_ string) (*types.Transaction, error) {
	return _FullAbi.Contract.SetName(&_FullAbi.TransactOpts, node, name_)
}

// SetVehicleId is a paid mutator transaction binding the contract method 0xf3cdd5c3.
//
// Solidity: function setVehicleId(bytes32 node, uint256 vehicleId_) returns()
func (_FullAbi *FullAbiTransactor) SetVehicleId(opts *bind.TransactOpts, node [32]byte, vehicleId_ *big.Int) (*types.Transaction, error) {
	return _FullAbi.contract.Transact(opts, "setVehicleId", node, vehicleId_)
}

// SetVehicleId is a paid mutator transaction binding the contract method 0xf3cdd5c3.
//
// Solidity: function setVehicleId(bytes32 node, uint256 vehicleId_) returns()
func (_FullAbi *FullAbiSession) SetVehicleId(node [32]byte, vehicleId_ *big.Int) (*types.Transaction, error) {
	return _FullAbi.Contract.SetVehicleId(&_FullAbi.TransactOpts, node, vehicleId_)
}

// SetVehicleId is a paid mutator transaction binding the contract method 0xf3cdd5c3.
//
// Solidity: function setVehicleId(bytes32 node, uint256 vehicleId_) returns()
func (_FullAbi *FullAbiTransactorSession) SetVehicleId(node [32]byte, vehicleId_ *big.Int) (*types.Transaction, error) {
	return _FullAbi.Contract.SetVehicleId(&_FullAbi.TransactOpts, node, vehicleId_)
}

// SetVehicleIdProxyAddress is a paid mutator transaction binding the contract method 0x9bfae6da.
//
// Solidity: function setVehicleIdProxyAddress(address addr) returns()
func (_FullAbi *FullAbiTransactor) SetVehicleIdProxyAddress(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _FullAbi.contract.Transact(opts, "setVehicleIdProxyAddress", addr)
}

// SetVehicleIdProxyAddress is a paid mutator transaction binding the contract method 0x9bfae6da.
//
// Solidity: function setVehicleIdProxyAddress(address addr) returns()
func (_FullAbi *FullAbiSession) SetVehicleIdProxyAddress(addr common.Address) (*types.Transaction, error) {
	return _FullAbi.Contract.SetVehicleIdProxyAddress(&_FullAbi.TransactOpts, addr)
}

// SetVehicleIdProxyAddress is a paid mutator transaction binding the contract method 0x9bfae6da.
//
// Solidity: function setVehicleIdProxyAddress(address addr) returns()
func (_FullAbi *FullAbiTransactorSession) SetVehicleIdProxyAddress(addr common.Address) (*types.Transaction, error) {
	return _FullAbi.Contract.SetVehicleIdProxyAddress(&_FullAbi.TransactOpts, addr)
}

// UpdateModule is a paid mutator transaction binding the contract method 0x06d1d2a1.
//
// Solidity: function updateModule(address oldImplementation, address newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors) returns()
func (_FullAbi *FullAbiTransactor) UpdateModule(opts *bind.TransactOpts, oldImplementation common.Address, newImplementation common.Address, oldSelectors [][4]byte, newSelectors [][4]byte) (*types.Transaction, error) {
	return _FullAbi.contract.Transact(opts, "updateModule", oldImplementation, newImplementation, oldSelectors, newSelectors)
}

// UpdateModule is a paid mutator transaction binding the contract method 0x06d1d2a1.
//
// Solidity: function updateModule(address oldImplementation, address newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors) returns()
func (_FullAbi *FullAbiSession) UpdateModule(oldImplementation common.Address, newImplementation common.Address, oldSelectors [][4]byte, newSelectors [][4]byte) (*types.Transaction, error) {
	return _FullAbi.Contract.UpdateModule(&_FullAbi.TransactOpts, oldImplementation, newImplementation, oldSelectors, newSelectors)
}

// UpdateModule is a paid mutator transaction binding the contract method 0x06d1d2a1.
//
// Solidity: function updateModule(address oldImplementation, address newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors) returns()
func (_FullAbi *FullAbiTransactorSession) UpdateModule(oldImplementation common.Address, newImplementation common.Address, oldSelectors [][4]byte, newSelectors [][4]byte) (*types.Transaction, error) {
	return _FullAbi.Contract.UpdateModule(&_FullAbi.TransactOpts, oldImplementation, newImplementation, oldSelectors, newSelectors)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_FullAbi *FullAbiTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _FullAbi.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_FullAbi *FullAbiSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _FullAbi.Contract.Fallback(&_FullAbi.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_FullAbi *FullAbiTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _FullAbi.Contract.Fallback(&_FullAbi.TransactOpts, calldata)
}

// FullAbiModuleAddedIterator is returned from FilterModuleAdded and is used to iterate over the raw logs and unpacked data for ModuleAdded events raised by the FullAbi contract.
type FullAbiModuleAddedIterator struct {
	Event *FullAbiModuleAdded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FullAbiModuleAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FullAbiModuleAdded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FullAbiModuleAdded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FullAbiModuleAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FullAbiModuleAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FullAbiModuleAdded represents a ModuleAdded event raised by the FullAbi contract.
type FullAbiModuleAdded struct {
	ModuleAddr common.Address
	Selectors  [][4]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterModuleAdded is a free log retrieval operation binding the contract event 0x02d0c334c706cd2f08faf7bc03674fc7f3970dd8921776c655069cde33b7fb29.
//
// Solidity: event ModuleAdded(address indexed moduleAddr, bytes4[] selectors)
func (_FullAbi *FullAbiFilterer) FilterModuleAdded(opts *bind.FilterOpts, moduleAddr []common.Address) (*FullAbiModuleAddedIterator, error) {

	var moduleAddrRule []interface{}
	for _, moduleAddrItem := range moduleAddr {
		moduleAddrRule = append(moduleAddrRule, moduleAddrItem)
	}

	logs, sub, err := _FullAbi.contract.FilterLogs(opts, "ModuleAdded", moduleAddrRule)
	if err != nil {
		return nil, err
	}
	return &FullAbiModuleAddedIterator{contract: _FullAbi.contract, event: "ModuleAdded", logs: logs, sub: sub}, nil
}

// WatchModuleAdded is a free log subscription operation binding the contract event 0x02d0c334c706cd2f08faf7bc03674fc7f3970dd8921776c655069cde33b7fb29.
//
// Solidity: event ModuleAdded(address indexed moduleAddr, bytes4[] selectors)
func (_FullAbi *FullAbiFilterer) WatchModuleAdded(opts *bind.WatchOpts, sink chan<- *FullAbiModuleAdded, moduleAddr []common.Address) (event.Subscription, error) {

	var moduleAddrRule []interface{}
	for _, moduleAddrItem := range moduleAddr {
		moduleAddrRule = append(moduleAddrRule, moduleAddrItem)
	}

	logs, sub, err := _FullAbi.contract.WatchLogs(opts, "ModuleAdded", moduleAddrRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FullAbiModuleAdded)
				if err := _FullAbi.contract.UnpackLog(event, "ModuleAdded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseModuleAdded is a log parse operation binding the contract event 0x02d0c334c706cd2f08faf7bc03674fc7f3970dd8921776c655069cde33b7fb29.
//
// Solidity: event ModuleAdded(address indexed moduleAddr, bytes4[] selectors)
func (_FullAbi *FullAbiFilterer) ParseModuleAdded(log types.Log) (*FullAbiModuleAdded, error) {
	event := new(FullAbiModuleAdded)
	if err := _FullAbi.contract.UnpackLog(event, "ModuleAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FullAbiModuleRemovedIterator is returned from FilterModuleRemoved and is used to iterate over the raw logs and unpacked data for ModuleRemoved events raised by the FullAbi contract.
type FullAbiModuleRemovedIterator struct {
	Event *FullAbiModuleRemoved // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FullAbiModuleRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FullAbiModuleRemoved)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FullAbiModuleRemoved)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FullAbiModuleRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FullAbiModuleRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FullAbiModuleRemoved represents a ModuleRemoved event raised by the FullAbi contract.
type FullAbiModuleRemoved struct {
	ModuleAddr common.Address
	Selectors  [][4]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterModuleRemoved is a free log retrieval operation binding the contract event 0x7c3eb4f9083f75cbed2bd3f703e24b4bbcb77d345d3c50945f3abf3e967755cb.
//
// Solidity: event ModuleRemoved(address indexed moduleAddr, bytes4[] selectors)
func (_FullAbi *FullAbiFilterer) FilterModuleRemoved(opts *bind.FilterOpts, moduleAddr []common.Address) (*FullAbiModuleRemovedIterator, error) {

	var moduleAddrRule []interface{}
	for _, moduleAddrItem := range moduleAddr {
		moduleAddrRule = append(moduleAddrRule, moduleAddrItem)
	}

	logs, sub, err := _FullAbi.contract.FilterLogs(opts, "ModuleRemoved", moduleAddrRule)
	if err != nil {
		return nil, err
	}
	return &FullAbiModuleRemovedIterator{contract: _FullAbi.contract, event: "ModuleRemoved", logs: logs, sub: sub}, nil
}

// WatchModuleRemoved is a free log subscription operation binding the contract event 0x7c3eb4f9083f75cbed2bd3f703e24b4bbcb77d345d3c50945f3abf3e967755cb.
//
// Solidity: event ModuleRemoved(address indexed moduleAddr, bytes4[] selectors)
func (_FullAbi *FullAbiFilterer) WatchModuleRemoved(opts *bind.WatchOpts, sink chan<- *FullAbiModuleRemoved, moduleAddr []common.Address) (event.Subscription, error) {

	var moduleAddrRule []interface{}
	for _, moduleAddrItem := range moduleAddr {
		moduleAddrRule = append(moduleAddrRule, moduleAddrItem)
	}

	logs, sub, err := _FullAbi.contract.WatchLogs(opts, "ModuleRemoved", moduleAddrRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FullAbiModuleRemoved)
				if err := _FullAbi.contract.UnpackLog(event, "ModuleRemoved", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseModuleRemoved is a log parse operation binding the contract event 0x7c3eb4f9083f75cbed2bd3f703e24b4bbcb77d345d3c50945f3abf3e967755cb.
//
// Solidity: event ModuleRemoved(address indexed moduleAddr, bytes4[] selectors)
func (_FullAbi *FullAbiFilterer) ParseModuleRemoved(log types.Log) (*FullAbiModuleRemoved, error) {
	event := new(FullAbiModuleRemoved)
	if err := _FullAbi.contract.UnpackLog(event, "ModuleRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FullAbiModuleUpdatedIterator is returned from FilterModuleUpdated and is used to iterate over the raw logs and unpacked data for ModuleUpdated events raised by the FullAbi contract.
type FullAbiModuleUpdatedIterator struct {
	Event *FullAbiModuleUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FullAbiModuleUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FullAbiModuleUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FullAbiModuleUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FullAbiModuleUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FullAbiModuleUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FullAbiModuleUpdated represents a ModuleUpdated event raised by the FullAbi contract.
type FullAbiModuleUpdated struct {
	OldImplementation common.Address
	NewImplementation common.Address
	OldSelectors      [][4]byte
	NewSelectors      [][4]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterModuleUpdated is a free log retrieval operation binding the contract event 0xa062c2c046aa14dc9284b13bde77061cb034f0aa820f20057af6b164651eaa08.
//
// Solidity: event ModuleUpdated(address indexed oldImplementation, address indexed newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors)
func (_FullAbi *FullAbiFilterer) FilterModuleUpdated(opts *bind.FilterOpts, oldImplementation []common.Address, newImplementation []common.Address) (*FullAbiModuleUpdatedIterator, error) {

	var oldImplementationRule []interface{}
	for _, oldImplementationItem := range oldImplementation {
		oldImplementationRule = append(oldImplementationRule, oldImplementationItem)
	}
	var newImplementationRule []interface{}
	for _, newImplementationItem := range newImplementation {
		newImplementationRule = append(newImplementationRule, newImplementationItem)
	}

	logs, sub, err := _FullAbi.contract.FilterLogs(opts, "ModuleUpdated", oldImplementationRule, newImplementationRule)
	if err != nil {
		return nil, err
	}
	return &FullAbiModuleUpdatedIterator{contract: _FullAbi.contract, event: "ModuleUpdated", logs: logs, sub: sub}, nil
}

// WatchModuleUpdated is a free log subscription operation binding the contract event 0xa062c2c046aa14dc9284b13bde77061cb034f0aa820f20057af6b164651eaa08.
//
// Solidity: event ModuleUpdated(address indexed oldImplementation, address indexed newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors)
func (_FullAbi *FullAbiFilterer) WatchModuleUpdated(opts *bind.WatchOpts, sink chan<- *FullAbiModuleUpdated, oldImplementation []common.Address, newImplementation []common.Address) (event.Subscription, error) {

	var oldImplementationRule []interface{}
	for _, oldImplementationItem := range oldImplementation {
		oldImplementationRule = append(oldImplementationRule, oldImplementationItem)
	}
	var newImplementationRule []interface{}
	for _, newImplementationItem := range newImplementation {
		newImplementationRule = append(newImplementationRule, newImplementationItem)
	}

	logs, sub, err := _FullAbi.contract.WatchLogs(opts, "ModuleUpdated", oldImplementationRule, newImplementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FullAbiModuleUpdated)
				if err := _FullAbi.contract.UnpackLog(event, "ModuleUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseModuleUpdated is a log parse operation binding the contract event 0xa062c2c046aa14dc9284b13bde77061cb034f0aa820f20057af6b164651eaa08.
//
// Solidity: event ModuleUpdated(address indexed oldImplementation, address indexed newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors)
func (_FullAbi *FullAbiFilterer) ParseModuleUpdated(log types.Log) (*FullAbiModuleUpdated, error) {
	event := new(FullAbiModuleUpdated)
	if err := _FullAbi.contract.UnpackLog(event, "ModuleUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FullAbiNameChangedIterator is returned from FilterNameChanged and is used to iterate over the raw logs and unpacked data for NameChanged events raised by the FullAbi contract.
type FullAbiNameChangedIterator struct {
	Event *FullAbiNameChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FullAbiNameChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FullAbiNameChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FullAbiNameChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FullAbiNameChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FullAbiNameChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FullAbiNameChanged represents a NameChanged event raised by the FullAbi contract.
type FullAbiNameChanged struct {
	Node [32]byte
	Name string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterNameChanged is a free log retrieval operation binding the contract event 0xb7d29e911041e8d9b843369e890bcb72c9388692ba48b65ac54e7214c4c348f7.
//
// Solidity: event NameChanged(bytes32 indexed node, string name_)
func (_FullAbi *FullAbiFilterer) FilterNameChanged(opts *bind.FilterOpts, node [][32]byte) (*FullAbiNameChangedIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _FullAbi.contract.FilterLogs(opts, "NameChanged", nodeRule)
	if err != nil {
		return nil, err
	}
	return &FullAbiNameChangedIterator{contract: _FullAbi.contract, event: "NameChanged", logs: logs, sub: sub}, nil
}

// WatchNameChanged is a free log subscription operation binding the contract event 0xb7d29e911041e8d9b843369e890bcb72c9388692ba48b65ac54e7214c4c348f7.
//
// Solidity: event NameChanged(bytes32 indexed node, string name_)
func (_FullAbi *FullAbiFilterer) WatchNameChanged(opts *bind.WatchOpts, sink chan<- *FullAbiNameChanged, node [][32]byte) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _FullAbi.contract.WatchLogs(opts, "NameChanged", nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FullAbiNameChanged)
				if err := _FullAbi.contract.UnpackLog(event, "NameChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNameChanged is a log parse operation binding the contract event 0xb7d29e911041e8d9b843369e890bcb72c9388692ba48b65ac54e7214c4c348f7.
//
// Solidity: event NameChanged(bytes32 indexed node, string name_)
func (_FullAbi *FullAbiFilterer) ParseNameChanged(log types.Log) (*FullAbiNameChanged, error) {
	event := new(FullAbiNameChanged)
	if err := _FullAbi.contract.UnpackLog(event, "NameChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FullAbiRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the FullAbi contract.
type FullAbiRoleAdminChangedIterator struct {
	Event *FullAbiRoleAdminChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FullAbiRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FullAbiRoleAdminChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FullAbiRoleAdminChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FullAbiRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FullAbiRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FullAbiRoleAdminChanged represents a RoleAdminChanged event raised by the FullAbi contract.
type FullAbiRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_FullAbi *FullAbiFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*FullAbiRoleAdminChangedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _FullAbi.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &FullAbiRoleAdminChangedIterator{contract: _FullAbi.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_FullAbi *FullAbiFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *FullAbiRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _FullAbi.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FullAbiRoleAdminChanged)
				if err := _FullAbi.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleAdminChanged is a log parse operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_FullAbi *FullAbiFilterer) ParseRoleAdminChanged(log types.Log) (*FullAbiRoleAdminChanged, error) {
	event := new(FullAbiRoleAdminChanged)
	if err := _FullAbi.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FullAbiRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the FullAbi contract.
type FullAbiRoleGrantedIterator struct {
	Event *FullAbiRoleGranted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FullAbiRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FullAbiRoleGranted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FullAbiRoleGranted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FullAbiRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FullAbiRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FullAbiRoleGranted represents a RoleGranted event raised by the FullAbi contract.
type FullAbiRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_FullAbi *FullAbiFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*FullAbiRoleGrantedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _FullAbi.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &FullAbiRoleGrantedIterator{contract: _FullAbi.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_FullAbi *FullAbiFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *FullAbiRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _FullAbi.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FullAbiRoleGranted)
				if err := _FullAbi.contract.UnpackLog(event, "RoleGranted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleGranted is a log parse operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_FullAbi *FullAbiFilterer) ParseRoleGranted(log types.Log) (*FullAbiRoleGranted, error) {
	event := new(FullAbiRoleGranted)
	if err := _FullAbi.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FullAbiRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the FullAbi contract.
type FullAbiRoleRevokedIterator struct {
	Event *FullAbiRoleRevoked // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FullAbiRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FullAbiRoleRevoked)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FullAbiRoleRevoked)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FullAbiRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FullAbiRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FullAbiRoleRevoked represents a RoleRevoked event raised by the FullAbi contract.
type FullAbiRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_FullAbi *FullAbiFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*FullAbiRoleRevokedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _FullAbi.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &FullAbiRoleRevokedIterator{contract: _FullAbi.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_FullAbi *FullAbiFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *FullAbiRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _FullAbi.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FullAbiRoleRevoked)
				if err := _FullAbi.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleRevoked is a log parse operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_FullAbi *FullAbiFilterer) ParseRoleRevoked(log types.Log) (*FullAbiRoleRevoked, error) {
	event := new(FullAbiRoleRevoked)
	if err := _FullAbi.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FullAbiVehicleIdChangedIterator is returned from FilterVehicleIdChanged and is used to iterate over the raw logs and unpacked data for VehicleIdChanged events raised by the FullAbi contract.
type FullAbiVehicleIdChangedIterator struct {
	Event *FullAbiVehicleIdChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FullAbiVehicleIdChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FullAbiVehicleIdChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FullAbiVehicleIdChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FullAbiVehicleIdChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FullAbiVehicleIdChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FullAbiVehicleIdChanged represents a VehicleIdChanged event raised by the FullAbi contract.
type FullAbiVehicleIdChanged struct {
	Node      [32]byte
	VehicleId *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterVehicleIdChanged is a free log retrieval operation binding the contract event 0x5337d7999bbcb00ab163ed1b63d44da7c6a3edb512a74afdc3dea8e106c5b336.
//
// Solidity: event VehicleIdChanged(bytes32 indexed node, uint256 indexed vehicleId_)
func (_FullAbi *FullAbiFilterer) FilterVehicleIdChanged(opts *bind.FilterOpts, node [][32]byte, vehicleId_ []*big.Int) (*FullAbiVehicleIdChangedIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}
	var vehicleId_Rule []interface{}
	for _, vehicleId_Item := range vehicleId_ {
		vehicleId_Rule = append(vehicleId_Rule, vehicleId_Item)
	}

	logs, sub, err := _FullAbi.contract.FilterLogs(opts, "VehicleIdChanged", nodeRule, vehicleId_Rule)
	if err != nil {
		return nil, err
	}
	return &FullAbiVehicleIdChangedIterator{contract: _FullAbi.contract, event: "VehicleIdChanged", logs: logs, sub: sub}, nil
}

// WatchVehicleIdChanged is a free log subscription operation binding the contract event 0x5337d7999bbcb00ab163ed1b63d44da7c6a3edb512a74afdc3dea8e106c5b336.
//
// Solidity: event VehicleIdChanged(bytes32 indexed node, uint256 indexed vehicleId_)
func (_FullAbi *FullAbiFilterer) WatchVehicleIdChanged(opts *bind.WatchOpts, sink chan<- *FullAbiVehicleIdChanged, node [][32]byte, vehicleId_ []*big.Int) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}
	var vehicleId_Rule []interface{}
	for _, vehicleId_Item := range vehicleId_ {
		vehicleId_Rule = append(vehicleId_Rule, vehicleId_Item)
	}

	logs, sub, err := _FullAbi.contract.WatchLogs(opts, "VehicleIdChanged", nodeRule, vehicleId_Rule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FullAbiVehicleIdChanged)
				if err := _FullAbi.contract.UnpackLog(event, "VehicleIdChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseVehicleIdChanged is a log parse operation binding the contract event 0x5337d7999bbcb00ab163ed1b63d44da7c6a3edb512a74afdc3dea8e106c5b336.
//
// Solidity: event VehicleIdChanged(bytes32 indexed node, uint256 indexed vehicleId_)
func (_FullAbi *FullAbiFilterer) ParseVehicleIdChanged(log types.Log) (*FullAbiVehicleIdChanged, error) {
	event := new(FullAbiVehicleIdChanged)
	if err := _FullAbi.contract.UnpackLog(event, "VehicleIdChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FullAbiVehicleIdProxySetIterator is returned from FilterVehicleIdProxySet and is used to iterate over the raw logs and unpacked data for VehicleIdProxySet events raised by the FullAbi contract.
type FullAbiVehicleIdProxySetIterator struct {
	Event *FullAbiVehicleIdProxySet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FullAbiVehicleIdProxySetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FullAbiVehicleIdProxySet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FullAbiVehicleIdProxySet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FullAbiVehicleIdProxySetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FullAbiVehicleIdProxySetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FullAbiVehicleIdProxySet represents a VehicleIdProxySet event raised by the FullAbi contract.
type FullAbiVehicleIdProxySet struct {
	Proxy common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterVehicleIdProxySet is a free log retrieval operation binding the contract event 0x3e7484c4e57f7d92e9f02eba6cd805d89112e48db8c21aeb8485fcf0020e479d.
//
// Solidity: event VehicleIdProxySet(address indexed proxy)
func (_FullAbi *FullAbiFilterer) FilterVehicleIdProxySet(opts *bind.FilterOpts, proxy []common.Address) (*FullAbiVehicleIdProxySetIterator, error) {

	var proxyRule []interface{}
	for _, proxyItem := range proxy {
		proxyRule = append(proxyRule, proxyItem)
	}

	logs, sub, err := _FullAbi.contract.FilterLogs(opts, "VehicleIdProxySet", proxyRule)
	if err != nil {
		return nil, err
	}
	return &FullAbiVehicleIdProxySetIterator{contract: _FullAbi.contract, event: "VehicleIdProxySet", logs: logs, sub: sub}, nil
}

// WatchVehicleIdProxySet is a free log subscription operation binding the contract event 0x3e7484c4e57f7d92e9f02eba6cd805d89112e48db8c21aeb8485fcf0020e479d.
//
// Solidity: event VehicleIdProxySet(address indexed proxy)
func (_FullAbi *FullAbiFilterer) WatchVehicleIdProxySet(opts *bind.WatchOpts, sink chan<- *FullAbiVehicleIdProxySet, proxy []common.Address) (event.Subscription, error) {

	var proxyRule []interface{}
	for _, proxyItem := range proxy {
		proxyRule = append(proxyRule, proxyItem)
	}

	logs, sub, err := _FullAbi.contract.WatchLogs(opts, "VehicleIdProxySet", proxyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FullAbiVehicleIdProxySet)
				if err := _FullAbi.contract.UnpackLog(event, "VehicleIdProxySet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseVehicleIdProxySet is a log parse operation binding the contract event 0x3e7484c4e57f7d92e9f02eba6cd805d89112e48db8c21aeb8485fcf0020e479d.
//
// Solidity: event VehicleIdProxySet(address indexed proxy)
func (_FullAbi *FullAbiFilterer) ParseVehicleIdProxySet(log types.Log) (*FullAbiVehicleIdProxySet, error) {
	event := new(FullAbiVehicleIdProxySet)
	if err := _FullAbi.contract.UnpackLog(event, "VehicleIdProxySet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
