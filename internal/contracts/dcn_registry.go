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

// DcnRegistryMetaData contains all meta data concerning the DcnRegistry contract.
var DcnRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"previousAdmin\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"AdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"approved\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"beacon\",\"type\":\"address\"}],\"name\":\"BeaconUpgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"string\",\"name\":\"baseURI\",\"type\":\"string\"}],\"name\":\"NewBaseURI\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"defaultResolver\",\"type\":\"address\"}],\"name\":\"NewDefaultResolver\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"expiration\",\"type\":\"uint256\"}],\"name\":\"NewExpiration\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"NewNode\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"}],\"name\":\"NewResolver\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"ADMIN_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DEFAULT_ADMIN_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MANAGER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"TRANSFERER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"UPGRADER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"baseURI\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"label\",\"type\":\"string\"}],\"name\":\"burnTld\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"resolver_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"name\":\"claim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"defaultResolver\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"name\":\"expires\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"expires_\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getApproved\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name_\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol_\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"baseURI_\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"defaultResolver_\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"string[]\",\"name\":\"labels\",\"type\":\"string[]\"},{\"internalType\":\"address\",\"name\":\"resolver_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"label\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"resolver_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"name\":\"mintTld\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxiableUUID\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"name\":\"record\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"resolver_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"expires_\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"name\":\"recordExists\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"name\":\"records\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"expires\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"name\":\"renew\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"name\":\"resolver\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"resolver_\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"baseURI_\",\"type\":\"string\"}],\"name\":\"setBaseURI\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"defaultResolver_\",\"type\":\"address\"}],\"name\":\"setDefaultResolver\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"name\":\"setExpiration\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"resolver_\",\"type\":\"address\"}],\"name\":\"setResolver\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"tokenURI\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"}],\"name\":\"upgradeTo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"upgradeToAndCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
}

// DcnRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use DcnRegistryMetaData.ABI instead.
var DcnRegistryABI = DcnRegistryMetaData.ABI

// DcnRegistry is an auto generated Go binding around an Ethereum contract.
type DcnRegistry struct {
	DcnRegistryCaller     // Read-only binding to the contract
	DcnRegistryTransactor // Write-only binding to the contract
	DcnRegistryFilterer   // Log filterer for contract events
}

// DcnRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type DcnRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DcnRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DcnRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DcnRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DcnRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DcnRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DcnRegistrySession struct {
	Contract     *DcnRegistry      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DcnRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DcnRegistryCallerSession struct {
	Contract *DcnRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// DcnRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DcnRegistryTransactorSession struct {
	Contract     *DcnRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// DcnRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type DcnRegistryRaw struct {
	Contract *DcnRegistry // Generic contract binding to access the raw methods on
}

// DcnRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DcnRegistryCallerRaw struct {
	Contract *DcnRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// DcnRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DcnRegistryTransactorRaw struct {
	Contract *DcnRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDcnRegistry creates a new instance of DcnRegistry, bound to a specific deployed contract.
func NewDcnRegistry(address common.Address, backend bind.ContractBackend) (*DcnRegistry, error) {
	contract, err := bindDcnRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &DcnRegistry{DcnRegistryCaller: DcnRegistryCaller{contract: contract}, DcnRegistryTransactor: DcnRegistryTransactor{contract: contract}, DcnRegistryFilterer: DcnRegistryFilterer{contract: contract}}, nil
}

// NewDcnRegistryCaller creates a new read-only instance of DcnRegistry, bound to a specific deployed contract.
func NewDcnRegistryCaller(address common.Address, caller bind.ContractCaller) (*DcnRegistryCaller, error) {
	contract, err := bindDcnRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DcnRegistryCaller{contract: contract}, nil
}

// NewDcnRegistryTransactor creates a new write-only instance of DcnRegistry, bound to a specific deployed contract.
func NewDcnRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*DcnRegistryTransactor, error) {
	contract, err := bindDcnRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DcnRegistryTransactor{contract: contract}, nil
}

// NewDcnRegistryFilterer creates a new log filterer instance of DcnRegistry, bound to a specific deployed contract.
func NewDcnRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*DcnRegistryFilterer, error) {
	contract, err := bindDcnRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DcnRegistryFilterer{contract: contract}, nil
}

// bindDcnRegistry binds a generic wrapper to an already deployed contract.
func bindDcnRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := DcnRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DcnRegistry *DcnRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _DcnRegistry.Contract.DcnRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DcnRegistry *DcnRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DcnRegistry.Contract.DcnRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DcnRegistry *DcnRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DcnRegistry.Contract.DcnRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DcnRegistry *DcnRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _DcnRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DcnRegistry *DcnRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DcnRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DcnRegistry *DcnRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DcnRegistry.Contract.contract.Transact(opts, method, params...)
}

// ADMINROLE is a free data retrieval call binding the contract method 0x75b238fc.
//
// Solidity: function ADMIN_ROLE() view returns(bytes32)
func (_DcnRegistry *DcnRegistryCaller) ADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ADMINROLE is a free data retrieval call binding the contract method 0x75b238fc.
//
// Solidity: function ADMIN_ROLE() view returns(bytes32)
func (_DcnRegistry *DcnRegistrySession) ADMINROLE() ([32]byte, error) {
	return _DcnRegistry.Contract.ADMINROLE(&_DcnRegistry.CallOpts)
}

// ADMINROLE is a free data retrieval call binding the contract method 0x75b238fc.
//
// Solidity: function ADMIN_ROLE() view returns(bytes32)
func (_DcnRegistry *DcnRegistryCallerSession) ADMINROLE() ([32]byte, error) {
	return _DcnRegistry.Contract.ADMINROLE(&_DcnRegistry.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_DcnRegistry *DcnRegistryCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_DcnRegistry *DcnRegistrySession) DEFAULTADMINROLE() ([32]byte, error) {
	return _DcnRegistry.Contract.DEFAULTADMINROLE(&_DcnRegistry.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_DcnRegistry *DcnRegistryCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _DcnRegistry.Contract.DEFAULTADMINROLE(&_DcnRegistry.CallOpts)
}

// MANAGERROLE is a free data retrieval call binding the contract method 0xec87621c.
//
// Solidity: function MANAGER_ROLE() view returns(bytes32)
func (_DcnRegistry *DcnRegistryCaller) MANAGERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "MANAGER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MANAGERROLE is a free data retrieval call binding the contract method 0xec87621c.
//
// Solidity: function MANAGER_ROLE() view returns(bytes32)
func (_DcnRegistry *DcnRegistrySession) MANAGERROLE() ([32]byte, error) {
	return _DcnRegistry.Contract.MANAGERROLE(&_DcnRegistry.CallOpts)
}

// MANAGERROLE is a free data retrieval call binding the contract method 0xec87621c.
//
// Solidity: function MANAGER_ROLE() view returns(bytes32)
func (_DcnRegistry *DcnRegistryCallerSession) MANAGERROLE() ([32]byte, error) {
	return _DcnRegistry.Contract.MANAGERROLE(&_DcnRegistry.CallOpts)
}

// TRANSFERERROLE is a free data retrieval call binding the contract method 0x0ade7dc1.
//
// Solidity: function TRANSFERER_ROLE() view returns(bytes32)
func (_DcnRegistry *DcnRegistryCaller) TRANSFERERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "TRANSFERER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// TRANSFERERROLE is a free data retrieval call binding the contract method 0x0ade7dc1.
//
// Solidity: function TRANSFERER_ROLE() view returns(bytes32)
func (_DcnRegistry *DcnRegistrySession) TRANSFERERROLE() ([32]byte, error) {
	return _DcnRegistry.Contract.TRANSFERERROLE(&_DcnRegistry.CallOpts)
}

// TRANSFERERROLE is a free data retrieval call binding the contract method 0x0ade7dc1.
//
// Solidity: function TRANSFERER_ROLE() view returns(bytes32)
func (_DcnRegistry *DcnRegistryCallerSession) TRANSFERERROLE() ([32]byte, error) {
	return _DcnRegistry.Contract.TRANSFERERROLE(&_DcnRegistry.CallOpts)
}

// UPGRADERROLE is a free data retrieval call binding the contract method 0xf72c0d8b.
//
// Solidity: function UPGRADER_ROLE() view returns(bytes32)
func (_DcnRegistry *DcnRegistryCaller) UPGRADERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "UPGRADER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// UPGRADERROLE is a free data retrieval call binding the contract method 0xf72c0d8b.
//
// Solidity: function UPGRADER_ROLE() view returns(bytes32)
func (_DcnRegistry *DcnRegistrySession) UPGRADERROLE() ([32]byte, error) {
	return _DcnRegistry.Contract.UPGRADERROLE(&_DcnRegistry.CallOpts)
}

// UPGRADERROLE is a free data retrieval call binding the contract method 0xf72c0d8b.
//
// Solidity: function UPGRADER_ROLE() view returns(bytes32)
func (_DcnRegistry *DcnRegistryCallerSession) UPGRADERROLE() ([32]byte, error) {
	return _DcnRegistry.Contract.UPGRADERROLE(&_DcnRegistry.CallOpts)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_DcnRegistry *DcnRegistryCaller) BalanceOf(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "balanceOf", owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_DcnRegistry *DcnRegistrySession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _DcnRegistry.Contract.BalanceOf(&_DcnRegistry.CallOpts, owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_DcnRegistry *DcnRegistryCallerSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _DcnRegistry.Contract.BalanceOf(&_DcnRegistry.CallOpts, owner)
}

// BaseURI is a free data retrieval call binding the contract method 0x6c0360eb.
//
// Solidity: function baseURI() view returns(string)
func (_DcnRegistry *DcnRegistryCaller) BaseURI(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "baseURI")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// BaseURI is a free data retrieval call binding the contract method 0x6c0360eb.
//
// Solidity: function baseURI() view returns(string)
func (_DcnRegistry *DcnRegistrySession) BaseURI() (string, error) {
	return _DcnRegistry.Contract.BaseURI(&_DcnRegistry.CallOpts)
}

// BaseURI is a free data retrieval call binding the contract method 0x6c0360eb.
//
// Solidity: function baseURI() view returns(string)
func (_DcnRegistry *DcnRegistryCallerSession) BaseURI() (string, error) {
	return _DcnRegistry.Contract.BaseURI(&_DcnRegistry.CallOpts)
}

// DefaultResolver is a free data retrieval call binding the contract method 0x828eab0e.
//
// Solidity: function defaultResolver() view returns(address)
func (_DcnRegistry *DcnRegistryCaller) DefaultResolver(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "defaultResolver")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// DefaultResolver is a free data retrieval call binding the contract method 0x828eab0e.
//
// Solidity: function defaultResolver() view returns(address)
func (_DcnRegistry *DcnRegistrySession) DefaultResolver() (common.Address, error) {
	return _DcnRegistry.Contract.DefaultResolver(&_DcnRegistry.CallOpts)
}

// DefaultResolver is a free data retrieval call binding the contract method 0x828eab0e.
//
// Solidity: function defaultResolver() view returns(address)
func (_DcnRegistry *DcnRegistryCallerSession) DefaultResolver() (common.Address, error) {
	return _DcnRegistry.Contract.DefaultResolver(&_DcnRegistry.CallOpts)
}

// Expires is a free data retrieval call binding the contract method 0x44a60bbb.
//
// Solidity: function expires(bytes32 node) view returns(uint256 expires_)
func (_DcnRegistry *DcnRegistryCaller) Expires(opts *bind.CallOpts, node [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "expires", node)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Expires is a free data retrieval call binding the contract method 0x44a60bbb.
//
// Solidity: function expires(bytes32 node) view returns(uint256 expires_)
func (_DcnRegistry *DcnRegistrySession) Expires(node [32]byte) (*big.Int, error) {
	return _DcnRegistry.Contract.Expires(&_DcnRegistry.CallOpts, node)
}

// Expires is a free data retrieval call binding the contract method 0x44a60bbb.
//
// Solidity: function expires(bytes32 node) view returns(uint256 expires_)
func (_DcnRegistry *DcnRegistryCallerSession) Expires(node [32]byte) (*big.Int, error) {
	return _DcnRegistry.Contract.Expires(&_DcnRegistry.CallOpts, node)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_DcnRegistry *DcnRegistryCaller) GetApproved(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "getApproved", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_DcnRegistry *DcnRegistrySession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _DcnRegistry.Contract.GetApproved(&_DcnRegistry.CallOpts, tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_DcnRegistry *DcnRegistryCallerSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _DcnRegistry.Contract.GetApproved(&_DcnRegistry.CallOpts, tokenId)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_DcnRegistry *DcnRegistryCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_DcnRegistry *DcnRegistrySession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _DcnRegistry.Contract.GetRoleAdmin(&_DcnRegistry.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_DcnRegistry *DcnRegistryCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _DcnRegistry.Contract.GetRoleAdmin(&_DcnRegistry.CallOpts, role)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_DcnRegistry *DcnRegistryCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_DcnRegistry *DcnRegistrySession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _DcnRegistry.Contract.HasRole(&_DcnRegistry.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_DcnRegistry *DcnRegistryCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _DcnRegistry.Contract.HasRole(&_DcnRegistry.CallOpts, role, account)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_DcnRegistry *DcnRegistryCaller) IsApprovedForAll(opts *bind.CallOpts, owner common.Address, operator common.Address) (bool, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "isApprovedForAll", owner, operator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_DcnRegistry *DcnRegistrySession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _DcnRegistry.Contract.IsApprovedForAll(&_DcnRegistry.CallOpts, owner, operator)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_DcnRegistry *DcnRegistryCallerSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _DcnRegistry.Contract.IsApprovedForAll(&_DcnRegistry.CallOpts, owner, operator)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_DcnRegistry *DcnRegistryCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_DcnRegistry *DcnRegistrySession) Name() (string, error) {
	return _DcnRegistry.Contract.Name(&_DcnRegistry.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_DcnRegistry *DcnRegistryCallerSession) Name() (string, error) {
	return _DcnRegistry.Contract.Name(&_DcnRegistry.CallOpts)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_DcnRegistry *DcnRegistryCaller) OwnerOf(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "ownerOf", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_DcnRegistry *DcnRegistrySession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _DcnRegistry.Contract.OwnerOf(&_DcnRegistry.CallOpts, tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_DcnRegistry *DcnRegistryCallerSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _DcnRegistry.Contract.OwnerOf(&_DcnRegistry.CallOpts, tokenId)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_DcnRegistry *DcnRegistryCaller) ProxiableUUID(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "proxiableUUID")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_DcnRegistry *DcnRegistrySession) ProxiableUUID() ([32]byte, error) {
	return _DcnRegistry.Contract.ProxiableUUID(&_DcnRegistry.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_DcnRegistry *DcnRegistryCallerSession) ProxiableUUID() ([32]byte, error) {
	return _DcnRegistry.Contract.ProxiableUUID(&_DcnRegistry.CallOpts)
}

// Record is a free data retrieval call binding the contract method 0xb5c645bd.
//
// Solidity: function record(bytes32 node) view returns(address resolver_, uint256 expires_)
func (_DcnRegistry *DcnRegistryCaller) Record(opts *bind.CallOpts, node [32]byte) (struct {
	Resolver common.Address
	Expires  *big.Int
}, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "record", node)

	outstruct := new(struct {
		Resolver common.Address
		Expires  *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Resolver = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Expires = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Record is a free data retrieval call binding the contract method 0xb5c645bd.
//
// Solidity: function record(bytes32 node) view returns(address resolver_, uint256 expires_)
func (_DcnRegistry *DcnRegistrySession) Record(node [32]byte) (struct {
	Resolver common.Address
	Expires  *big.Int
}, error) {
	return _DcnRegistry.Contract.Record(&_DcnRegistry.CallOpts, node)
}

// Record is a free data retrieval call binding the contract method 0xb5c645bd.
//
// Solidity: function record(bytes32 node) view returns(address resolver_, uint256 expires_)
func (_DcnRegistry *DcnRegistryCallerSession) Record(node [32]byte) (struct {
	Resolver common.Address
	Expires  *big.Int
}, error) {
	return _DcnRegistry.Contract.Record(&_DcnRegistry.CallOpts, node)
}

// RecordExists is a free data retrieval call binding the contract method 0xf79fe538.
//
// Solidity: function recordExists(bytes32 node) view returns(bool)
func (_DcnRegistry *DcnRegistryCaller) RecordExists(opts *bind.CallOpts, node [32]byte) (bool, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "recordExists", node)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// RecordExists is a free data retrieval call binding the contract method 0xf79fe538.
//
// Solidity: function recordExists(bytes32 node) view returns(bool)
func (_DcnRegistry *DcnRegistrySession) RecordExists(node [32]byte) (bool, error) {
	return _DcnRegistry.Contract.RecordExists(&_DcnRegistry.CallOpts, node)
}

// RecordExists is a free data retrieval call binding the contract method 0xf79fe538.
//
// Solidity: function recordExists(bytes32 node) view returns(bool)
func (_DcnRegistry *DcnRegistryCallerSession) RecordExists(node [32]byte) (bool, error) {
	return _DcnRegistry.Contract.RecordExists(&_DcnRegistry.CallOpts, node)
}

// Records is a free data retrieval call binding the contract method 0x01e64725.
//
// Solidity: function records(bytes32 node) view returns(address resolver, uint256 expires)
func (_DcnRegistry *DcnRegistryCaller) Records(opts *bind.CallOpts, node [32]byte) (struct {
	Resolver common.Address
	Expires  *big.Int
}, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "records", node)

	outstruct := new(struct {
		Resolver common.Address
		Expires  *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Resolver = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Expires = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Records is a free data retrieval call binding the contract method 0x01e64725.
//
// Solidity: function records(bytes32 node) view returns(address resolver, uint256 expires)
func (_DcnRegistry *DcnRegistrySession) Records(node [32]byte) (struct {
	Resolver common.Address
	Expires  *big.Int
}, error) {
	return _DcnRegistry.Contract.Records(&_DcnRegistry.CallOpts, node)
}

// Records is a free data retrieval call binding the contract method 0x01e64725.
//
// Solidity: function records(bytes32 node) view returns(address resolver, uint256 expires)
func (_DcnRegistry *DcnRegistryCallerSession) Records(node [32]byte) (struct {
	Resolver common.Address
	Expires  *big.Int
}, error) {
	return _DcnRegistry.Contract.Records(&_DcnRegistry.CallOpts, node)
}

// Resolver is a free data retrieval call binding the contract method 0x0178b8bf.
//
// Solidity: function resolver(bytes32 node) view returns(address resolver_)
func (_DcnRegistry *DcnRegistryCaller) Resolver(opts *bind.CallOpts, node [32]byte) (common.Address, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "resolver", node)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Resolver is a free data retrieval call binding the contract method 0x0178b8bf.
//
// Solidity: function resolver(bytes32 node) view returns(address resolver_)
func (_DcnRegistry *DcnRegistrySession) Resolver(node [32]byte) (common.Address, error) {
	return _DcnRegistry.Contract.Resolver(&_DcnRegistry.CallOpts, node)
}

// Resolver is a free data retrieval call binding the contract method 0x0178b8bf.
//
// Solidity: function resolver(bytes32 node) view returns(address resolver_)
func (_DcnRegistry *DcnRegistryCallerSession) Resolver(node [32]byte) (common.Address, error) {
	return _DcnRegistry.Contract.Resolver(&_DcnRegistry.CallOpts, node)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_DcnRegistry *DcnRegistryCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_DcnRegistry *DcnRegistrySession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _DcnRegistry.Contract.SupportsInterface(&_DcnRegistry.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_DcnRegistry *DcnRegistryCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _DcnRegistry.Contract.SupportsInterface(&_DcnRegistry.CallOpts, interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_DcnRegistry *DcnRegistryCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_DcnRegistry *DcnRegistrySession) Symbol() (string, error) {
	return _DcnRegistry.Contract.Symbol(&_DcnRegistry.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_DcnRegistry *DcnRegistryCallerSession) Symbol() (string, error) {
	return _DcnRegistry.Contract.Symbol(&_DcnRegistry.CallOpts)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_DcnRegistry *DcnRegistryCaller) TokenURI(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var out []interface{}
	err := _DcnRegistry.contract.Call(opts, &out, "tokenURI", tokenId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_DcnRegistry *DcnRegistrySession) TokenURI(tokenId *big.Int) (string, error) {
	return _DcnRegistry.Contract.TokenURI(&_DcnRegistry.CallOpts, tokenId)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_DcnRegistry *DcnRegistryCallerSession) TokenURI(tokenId *big.Int) (string, error) {
	return _DcnRegistry.Contract.TokenURI(&_DcnRegistry.CallOpts, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_DcnRegistry *DcnRegistryTransactor) Approve(opts *bind.TransactOpts, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "approve", to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_DcnRegistry *DcnRegistrySession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.Contract.Approve(&_DcnRegistry.TransactOpts, to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_DcnRegistry *DcnRegistryTransactorSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.Contract.Approve(&_DcnRegistry.TransactOpts, to, tokenId)
}

// BurnTld is a paid mutator transaction binding the contract method 0x27cbcf23.
//
// Solidity: function burnTld(string label) returns()
func (_DcnRegistry *DcnRegistryTransactor) BurnTld(opts *bind.TransactOpts, label string) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "burnTld", label)
}

// BurnTld is a paid mutator transaction binding the contract method 0x27cbcf23.
//
// Solidity: function burnTld(string label) returns()
func (_DcnRegistry *DcnRegistrySession) BurnTld(label string) (*types.Transaction, error) {
	return _DcnRegistry.Contract.BurnTld(&_DcnRegistry.TransactOpts, label)
}

// BurnTld is a paid mutator transaction binding the contract method 0x27cbcf23.
//
// Solidity: function burnTld(string label) returns()
func (_DcnRegistry *DcnRegistryTransactorSession) BurnTld(label string) (*types.Transaction, error) {
	return _DcnRegistry.Contract.BurnTld(&_DcnRegistry.TransactOpts, label)
}

// Claim is a paid mutator transaction binding the contract method 0x42554b3c.
//
// Solidity: function claim(address to, bytes32 node, address resolver_, uint256 duration) returns()
func (_DcnRegistry *DcnRegistryTransactor) Claim(opts *bind.TransactOpts, to common.Address, node [32]byte, resolver_ common.Address, duration *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "claim", to, node, resolver_, duration)
}

// Claim is a paid mutator transaction binding the contract method 0x42554b3c.
//
// Solidity: function claim(address to, bytes32 node, address resolver_, uint256 duration) returns()
func (_DcnRegistry *DcnRegistrySession) Claim(to common.Address, node [32]byte, resolver_ common.Address, duration *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.Contract.Claim(&_DcnRegistry.TransactOpts, to, node, resolver_, duration)
}

// Claim is a paid mutator transaction binding the contract method 0x42554b3c.
//
// Solidity: function claim(address to, bytes32 node, address resolver_, uint256 duration) returns()
func (_DcnRegistry *DcnRegistryTransactorSession) Claim(to common.Address, node [32]byte, resolver_ common.Address, duration *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.Contract.Claim(&_DcnRegistry.TransactOpts, to, node, resolver_, duration)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_DcnRegistry *DcnRegistryTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_DcnRegistry *DcnRegistrySession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _DcnRegistry.Contract.GrantRole(&_DcnRegistry.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_DcnRegistry *DcnRegistryTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _DcnRegistry.Contract.GrantRole(&_DcnRegistry.TransactOpts, role, account)
}

// Initialize is a paid mutator transaction binding the contract method 0x5c6d8da1.
//
// Solidity: function initialize(string name_, string symbol_, string baseURI_, address defaultResolver_) returns()
func (_DcnRegistry *DcnRegistryTransactor) Initialize(opts *bind.TransactOpts, name_ string, symbol_ string, baseURI_ string, defaultResolver_ common.Address) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "initialize", name_, symbol_, baseURI_, defaultResolver_)
}

// Initialize is a paid mutator transaction binding the contract method 0x5c6d8da1.
//
// Solidity: function initialize(string name_, string symbol_, string baseURI_, address defaultResolver_) returns()
func (_DcnRegistry *DcnRegistrySession) Initialize(name_ string, symbol_ string, baseURI_ string, defaultResolver_ common.Address) (*types.Transaction, error) {
	return _DcnRegistry.Contract.Initialize(&_DcnRegistry.TransactOpts, name_, symbol_, baseURI_, defaultResolver_)
}

// Initialize is a paid mutator transaction binding the contract method 0x5c6d8da1.
//
// Solidity: function initialize(string name_, string symbol_, string baseURI_, address defaultResolver_) returns()
func (_DcnRegistry *DcnRegistryTransactorSession) Initialize(name_ string, symbol_ string, baseURI_ string, defaultResolver_ common.Address) (*types.Transaction, error) {
	return _DcnRegistry.Contract.Initialize(&_DcnRegistry.TransactOpts, name_, symbol_, baseURI_, defaultResolver_)
}

// Mint is a paid mutator transaction binding the contract method 0x7b145b9c.
//
// Solidity: function mint(address to, string[] labels, address resolver_, uint256 duration) returns(bytes32 node)
func (_DcnRegistry *DcnRegistryTransactor) Mint(opts *bind.TransactOpts, to common.Address, labels []string, resolver_ common.Address, duration *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "mint", to, labels, resolver_, duration)
}

// Mint is a paid mutator transaction binding the contract method 0x7b145b9c.
//
// Solidity: function mint(address to, string[] labels, address resolver_, uint256 duration) returns(bytes32 node)
func (_DcnRegistry *DcnRegistrySession) Mint(to common.Address, labels []string, resolver_ common.Address, duration *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.Contract.Mint(&_DcnRegistry.TransactOpts, to, labels, resolver_, duration)
}

// Mint is a paid mutator transaction binding the contract method 0x7b145b9c.
//
// Solidity: function mint(address to, string[] labels, address resolver_, uint256 duration) returns(bytes32 node)
func (_DcnRegistry *DcnRegistryTransactorSession) Mint(to common.Address, labels []string, resolver_ common.Address, duration *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.Contract.Mint(&_DcnRegistry.TransactOpts, to, labels, resolver_, duration)
}

// MintTld is a paid mutator transaction binding the contract method 0x48ba15d0.
//
// Solidity: function mintTld(address to, string label, address resolver_, uint256 duration) returns(bytes32 node)
func (_DcnRegistry *DcnRegistryTransactor) MintTld(opts *bind.TransactOpts, to common.Address, label string, resolver_ common.Address, duration *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "mintTld", to, label, resolver_, duration)
}

// MintTld is a paid mutator transaction binding the contract method 0x48ba15d0.
//
// Solidity: function mintTld(address to, string label, address resolver_, uint256 duration) returns(bytes32 node)
func (_DcnRegistry *DcnRegistrySession) MintTld(to common.Address, label string, resolver_ common.Address, duration *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.Contract.MintTld(&_DcnRegistry.TransactOpts, to, label, resolver_, duration)
}

// MintTld is a paid mutator transaction binding the contract method 0x48ba15d0.
//
// Solidity: function mintTld(address to, string label, address resolver_, uint256 duration) returns(bytes32 node)
func (_DcnRegistry *DcnRegistryTransactorSession) MintTld(to common.Address, label string, resolver_ common.Address, duration *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.Contract.MintTld(&_DcnRegistry.TransactOpts, to, label, resolver_, duration)
}

// Renew is a paid mutator transaction binding the contract method 0xf544d82f.
//
// Solidity: function renew(bytes32 node, uint256 duration) returns()
func (_DcnRegistry *DcnRegistryTransactor) Renew(opts *bind.TransactOpts, node [32]byte, duration *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "renew", node, duration)
}

// Renew is a paid mutator transaction binding the contract method 0xf544d82f.
//
// Solidity: function renew(bytes32 node, uint256 duration) returns()
func (_DcnRegistry *DcnRegistrySession) Renew(node [32]byte, duration *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.Contract.Renew(&_DcnRegistry.TransactOpts, node, duration)
}

// Renew is a paid mutator transaction binding the contract method 0xf544d82f.
//
// Solidity: function renew(bytes32 node, uint256 duration) returns()
func (_DcnRegistry *DcnRegistryTransactorSession) Renew(node [32]byte, duration *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.Contract.Renew(&_DcnRegistry.TransactOpts, node, duration)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_DcnRegistry *DcnRegistryTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "renounceRole", role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_DcnRegistry *DcnRegistrySession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _DcnRegistry.Contract.RenounceRole(&_DcnRegistry.TransactOpts, role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_DcnRegistry *DcnRegistryTransactorSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _DcnRegistry.Contract.RenounceRole(&_DcnRegistry.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_DcnRegistry *DcnRegistryTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_DcnRegistry *DcnRegistrySession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _DcnRegistry.Contract.RevokeRole(&_DcnRegistry.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_DcnRegistry *DcnRegistryTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _DcnRegistry.Contract.RevokeRole(&_DcnRegistry.TransactOpts, role, account)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_DcnRegistry *DcnRegistryTransactor) SafeTransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "safeTransferFrom", from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_DcnRegistry *DcnRegistrySession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.Contract.SafeTransferFrom(&_DcnRegistry.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_DcnRegistry *DcnRegistryTransactorSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.Contract.SafeTransferFrom(&_DcnRegistry.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_DcnRegistry *DcnRegistryTransactor) SafeTransferFrom0(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "safeTransferFrom0", from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_DcnRegistry *DcnRegistrySession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _DcnRegistry.Contract.SafeTransferFrom0(&_DcnRegistry.TransactOpts, from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_DcnRegistry *DcnRegistryTransactorSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _DcnRegistry.Contract.SafeTransferFrom0(&_DcnRegistry.TransactOpts, from, to, tokenId, data)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_DcnRegistry *DcnRegistryTransactor) SetApprovalForAll(opts *bind.TransactOpts, operator common.Address, approved bool) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "setApprovalForAll", operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_DcnRegistry *DcnRegistrySession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _DcnRegistry.Contract.SetApprovalForAll(&_DcnRegistry.TransactOpts, operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_DcnRegistry *DcnRegistryTransactorSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _DcnRegistry.Contract.SetApprovalForAll(&_DcnRegistry.TransactOpts, operator, approved)
}

// SetBaseURI is a paid mutator transaction binding the contract method 0x55f804b3.
//
// Solidity: function setBaseURI(string baseURI_) returns()
func (_DcnRegistry *DcnRegistryTransactor) SetBaseURI(opts *bind.TransactOpts, baseURI_ string) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "setBaseURI", baseURI_)
}

// SetBaseURI is a paid mutator transaction binding the contract method 0x55f804b3.
//
// Solidity: function setBaseURI(string baseURI_) returns()
func (_DcnRegistry *DcnRegistrySession) SetBaseURI(baseURI_ string) (*types.Transaction, error) {
	return _DcnRegistry.Contract.SetBaseURI(&_DcnRegistry.TransactOpts, baseURI_)
}

// SetBaseURI is a paid mutator transaction binding the contract method 0x55f804b3.
//
// Solidity: function setBaseURI(string baseURI_) returns()
func (_DcnRegistry *DcnRegistryTransactorSession) SetBaseURI(baseURI_ string) (*types.Transaction, error) {
	return _DcnRegistry.Contract.SetBaseURI(&_DcnRegistry.TransactOpts, baseURI_)
}

// SetDefaultResolver is a paid mutator transaction binding the contract method 0xc66485b2.
//
// Solidity: function setDefaultResolver(address defaultResolver_) returns()
func (_DcnRegistry *DcnRegistryTransactor) SetDefaultResolver(opts *bind.TransactOpts, defaultResolver_ common.Address) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "setDefaultResolver", defaultResolver_)
}

// SetDefaultResolver is a paid mutator transaction binding the contract method 0xc66485b2.
//
// Solidity: function setDefaultResolver(address defaultResolver_) returns()
func (_DcnRegistry *DcnRegistrySession) SetDefaultResolver(defaultResolver_ common.Address) (*types.Transaction, error) {
	return _DcnRegistry.Contract.SetDefaultResolver(&_DcnRegistry.TransactOpts, defaultResolver_)
}

// SetDefaultResolver is a paid mutator transaction binding the contract method 0xc66485b2.
//
// Solidity: function setDefaultResolver(address defaultResolver_) returns()
func (_DcnRegistry *DcnRegistryTransactorSession) SetDefaultResolver(defaultResolver_ common.Address) (*types.Transaction, error) {
	return _DcnRegistry.Contract.SetDefaultResolver(&_DcnRegistry.TransactOpts, defaultResolver_)
}

// SetExpiration is a paid mutator transaction binding the contract method 0x636c726a.
//
// Solidity: function setExpiration(bytes32 node, uint256 duration) returns()
func (_DcnRegistry *DcnRegistryTransactor) SetExpiration(opts *bind.TransactOpts, node [32]byte, duration *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "setExpiration", node, duration)
}

// SetExpiration is a paid mutator transaction binding the contract method 0x636c726a.
//
// Solidity: function setExpiration(bytes32 node, uint256 duration) returns()
func (_DcnRegistry *DcnRegistrySession) SetExpiration(node [32]byte, duration *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.Contract.SetExpiration(&_DcnRegistry.TransactOpts, node, duration)
}

// SetExpiration is a paid mutator transaction binding the contract method 0x636c726a.
//
// Solidity: function setExpiration(bytes32 node, uint256 duration) returns()
func (_DcnRegistry *DcnRegistryTransactorSession) SetExpiration(node [32]byte, duration *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.Contract.SetExpiration(&_DcnRegistry.TransactOpts, node, duration)
}

// SetResolver is a paid mutator transaction binding the contract method 0x1896f70a.
//
// Solidity: function setResolver(bytes32 node, address resolver_) returns()
func (_DcnRegistry *DcnRegistryTransactor) SetResolver(opts *bind.TransactOpts, node [32]byte, resolver_ common.Address) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "setResolver", node, resolver_)
}

// SetResolver is a paid mutator transaction binding the contract method 0x1896f70a.
//
// Solidity: function setResolver(bytes32 node, address resolver_) returns()
func (_DcnRegistry *DcnRegistrySession) SetResolver(node [32]byte, resolver_ common.Address) (*types.Transaction, error) {
	return _DcnRegistry.Contract.SetResolver(&_DcnRegistry.TransactOpts, node, resolver_)
}

// SetResolver is a paid mutator transaction binding the contract method 0x1896f70a.
//
// Solidity: function setResolver(bytes32 node, address resolver_) returns()
func (_DcnRegistry *DcnRegistryTransactorSession) SetResolver(node [32]byte, resolver_ common.Address) (*types.Transaction, error) {
	return _DcnRegistry.Contract.SetResolver(&_DcnRegistry.TransactOpts, node, resolver_)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_DcnRegistry *DcnRegistryTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "transferFrom", from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_DcnRegistry *DcnRegistrySession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.Contract.TransferFrom(&_DcnRegistry.TransactOpts, from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_DcnRegistry *DcnRegistryTransactorSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _DcnRegistry.Contract.TransferFrom(&_DcnRegistry.TransactOpts, from, to, tokenId)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_DcnRegistry *DcnRegistryTransactor) UpgradeTo(opts *bind.TransactOpts, newImplementation common.Address) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "upgradeTo", newImplementation)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_DcnRegistry *DcnRegistrySession) UpgradeTo(newImplementation common.Address) (*types.Transaction, error) {
	return _DcnRegistry.Contract.UpgradeTo(&_DcnRegistry.TransactOpts, newImplementation)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_DcnRegistry *DcnRegistryTransactorSession) UpgradeTo(newImplementation common.Address) (*types.Transaction, error) {
	return _DcnRegistry.Contract.UpgradeTo(&_DcnRegistry.TransactOpts, newImplementation)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_DcnRegistry *DcnRegistryTransactor) UpgradeToAndCall(opts *bind.TransactOpts, newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _DcnRegistry.contract.Transact(opts, "upgradeToAndCall", newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_DcnRegistry *DcnRegistrySession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _DcnRegistry.Contract.UpgradeToAndCall(&_DcnRegistry.TransactOpts, newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_DcnRegistry *DcnRegistryTransactorSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _DcnRegistry.Contract.UpgradeToAndCall(&_DcnRegistry.TransactOpts, newImplementation, data)
}

// DcnRegistryAdminChangedIterator is returned from FilterAdminChanged and is used to iterate over the raw logs and unpacked data for AdminChanged events raised by the DcnRegistry contract.
type DcnRegistryAdminChangedIterator struct {
	Event *DcnRegistryAdminChanged // Event containing the contract specifics and raw log

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
func (it *DcnRegistryAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DcnRegistryAdminChanged)
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
		it.Event = new(DcnRegistryAdminChanged)
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
func (it *DcnRegistryAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DcnRegistryAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DcnRegistryAdminChanged represents a AdminChanged event raised by the DcnRegistry contract.
type DcnRegistryAdminChanged struct {
	PreviousAdmin common.Address
	NewAdmin      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterAdminChanged is a free log retrieval operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_DcnRegistry *DcnRegistryFilterer) FilterAdminChanged(opts *bind.FilterOpts) (*DcnRegistryAdminChangedIterator, error) {

	logs, sub, err := _DcnRegistry.contract.FilterLogs(opts, "AdminChanged")
	if err != nil {
		return nil, err
	}
	return &DcnRegistryAdminChangedIterator{contract: _DcnRegistry.contract, event: "AdminChanged", logs: logs, sub: sub}, nil
}

// WatchAdminChanged is a free log subscription operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_DcnRegistry *DcnRegistryFilterer) WatchAdminChanged(opts *bind.WatchOpts, sink chan<- *DcnRegistryAdminChanged) (event.Subscription, error) {

	logs, sub, err := _DcnRegistry.contract.WatchLogs(opts, "AdminChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DcnRegistryAdminChanged)
				if err := _DcnRegistry.contract.UnpackLog(event, "AdminChanged", log); err != nil {
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

// ParseAdminChanged is a log parse operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_DcnRegistry *DcnRegistryFilterer) ParseAdminChanged(log types.Log) (*DcnRegistryAdminChanged, error) {
	event := new(DcnRegistryAdminChanged)
	if err := _DcnRegistry.contract.UnpackLog(event, "AdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DcnRegistryApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the DcnRegistry contract.
type DcnRegistryApprovalIterator struct {
	Event *DcnRegistryApproval // Event containing the contract specifics and raw log

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
func (it *DcnRegistryApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DcnRegistryApproval)
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
		it.Event = new(DcnRegistryApproval)
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
func (it *DcnRegistryApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DcnRegistryApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DcnRegistryApproval represents a Approval event raised by the DcnRegistry contract.
type DcnRegistryApproval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_DcnRegistry *DcnRegistryFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, approved []common.Address, tokenId []*big.Int) (*DcnRegistryApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var approvedRule []interface{}
	for _, approvedItem := range approved {
		approvedRule = append(approvedRule, approvedItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _DcnRegistry.contract.FilterLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &DcnRegistryApprovalIterator{contract: _DcnRegistry.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_DcnRegistry *DcnRegistryFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *DcnRegistryApproval, owner []common.Address, approved []common.Address, tokenId []*big.Int) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var approvedRule []interface{}
	for _, approvedItem := range approved {
		approvedRule = append(approvedRule, approvedItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _DcnRegistry.contract.WatchLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DcnRegistryApproval)
				if err := _DcnRegistry.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_DcnRegistry *DcnRegistryFilterer) ParseApproval(log types.Log) (*DcnRegistryApproval, error) {
	event := new(DcnRegistryApproval)
	if err := _DcnRegistry.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DcnRegistryApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the DcnRegistry contract.
type DcnRegistryApprovalForAllIterator struct {
	Event *DcnRegistryApprovalForAll // Event containing the contract specifics and raw log

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
func (it *DcnRegistryApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DcnRegistryApprovalForAll)
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
		it.Event = new(DcnRegistryApprovalForAll)
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
func (it *DcnRegistryApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DcnRegistryApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DcnRegistryApprovalForAll represents a ApprovalForAll event raised by the DcnRegistry contract.
type DcnRegistryApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_DcnRegistry *DcnRegistryFilterer) FilterApprovalForAll(opts *bind.FilterOpts, owner []common.Address, operator []common.Address) (*DcnRegistryApprovalForAllIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _DcnRegistry.contract.FilterLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return &DcnRegistryApprovalForAllIterator{contract: _DcnRegistry.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_DcnRegistry *DcnRegistryFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *DcnRegistryApprovalForAll, owner []common.Address, operator []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _DcnRegistry.contract.WatchLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DcnRegistryApprovalForAll)
				if err := _DcnRegistry.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
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

// ParseApprovalForAll is a log parse operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_DcnRegistry *DcnRegistryFilterer) ParseApprovalForAll(log types.Log) (*DcnRegistryApprovalForAll, error) {
	event := new(DcnRegistryApprovalForAll)
	if err := _DcnRegistry.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DcnRegistryBeaconUpgradedIterator is returned from FilterBeaconUpgraded and is used to iterate over the raw logs and unpacked data for BeaconUpgraded events raised by the DcnRegistry contract.
type DcnRegistryBeaconUpgradedIterator struct {
	Event *DcnRegistryBeaconUpgraded // Event containing the contract specifics and raw log

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
func (it *DcnRegistryBeaconUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DcnRegistryBeaconUpgraded)
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
		it.Event = new(DcnRegistryBeaconUpgraded)
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
func (it *DcnRegistryBeaconUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DcnRegistryBeaconUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DcnRegistryBeaconUpgraded represents a BeaconUpgraded event raised by the DcnRegistry contract.
type DcnRegistryBeaconUpgraded struct {
	Beacon common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterBeaconUpgraded is a free log retrieval operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_DcnRegistry *DcnRegistryFilterer) FilterBeaconUpgraded(opts *bind.FilterOpts, beacon []common.Address) (*DcnRegistryBeaconUpgradedIterator, error) {

	var beaconRule []interface{}
	for _, beaconItem := range beacon {
		beaconRule = append(beaconRule, beaconItem)
	}

	logs, sub, err := _DcnRegistry.contract.FilterLogs(opts, "BeaconUpgraded", beaconRule)
	if err != nil {
		return nil, err
	}
	return &DcnRegistryBeaconUpgradedIterator{contract: _DcnRegistry.contract, event: "BeaconUpgraded", logs: logs, sub: sub}, nil
}

// WatchBeaconUpgraded is a free log subscription operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_DcnRegistry *DcnRegistryFilterer) WatchBeaconUpgraded(opts *bind.WatchOpts, sink chan<- *DcnRegistryBeaconUpgraded, beacon []common.Address) (event.Subscription, error) {

	var beaconRule []interface{}
	for _, beaconItem := range beacon {
		beaconRule = append(beaconRule, beaconItem)
	}

	logs, sub, err := _DcnRegistry.contract.WatchLogs(opts, "BeaconUpgraded", beaconRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DcnRegistryBeaconUpgraded)
				if err := _DcnRegistry.contract.UnpackLog(event, "BeaconUpgraded", log); err != nil {
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

// ParseBeaconUpgraded is a log parse operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_DcnRegistry *DcnRegistryFilterer) ParseBeaconUpgraded(log types.Log) (*DcnRegistryBeaconUpgraded, error) {
	event := new(DcnRegistryBeaconUpgraded)
	if err := _DcnRegistry.contract.UnpackLog(event, "BeaconUpgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DcnRegistryInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the DcnRegistry contract.
type DcnRegistryInitializedIterator struct {
	Event *DcnRegistryInitialized // Event containing the contract specifics and raw log

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
func (it *DcnRegistryInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DcnRegistryInitialized)
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
		it.Event = new(DcnRegistryInitialized)
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
func (it *DcnRegistryInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DcnRegistryInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DcnRegistryInitialized represents a Initialized event raised by the DcnRegistry contract.
type DcnRegistryInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_DcnRegistry *DcnRegistryFilterer) FilterInitialized(opts *bind.FilterOpts) (*DcnRegistryInitializedIterator, error) {

	logs, sub, err := _DcnRegistry.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &DcnRegistryInitializedIterator{contract: _DcnRegistry.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_DcnRegistry *DcnRegistryFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *DcnRegistryInitialized) (event.Subscription, error) {

	logs, sub, err := _DcnRegistry.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DcnRegistryInitialized)
				if err := _DcnRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
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

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_DcnRegistry *DcnRegistryFilterer) ParseInitialized(log types.Log) (*DcnRegistryInitialized, error) {
	event := new(DcnRegistryInitialized)
	if err := _DcnRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DcnRegistryNewBaseURIIterator is returned from FilterNewBaseURI and is used to iterate over the raw logs and unpacked data for NewBaseURI events raised by the DcnRegistry contract.
type DcnRegistryNewBaseURIIterator struct {
	Event *DcnRegistryNewBaseURI // Event containing the contract specifics and raw log

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
func (it *DcnRegistryNewBaseURIIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DcnRegistryNewBaseURI)
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
		it.Event = new(DcnRegistryNewBaseURI)
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
func (it *DcnRegistryNewBaseURIIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DcnRegistryNewBaseURIIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DcnRegistryNewBaseURI represents a NewBaseURI event raised by the DcnRegistry contract.
type DcnRegistryNewBaseURI struct {
	BaseURI common.Hash
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNewBaseURI is a free log retrieval operation binding the contract event 0x325d37e8fb549c86966f09bc3e6f62eb3afa93b255d6e3234338001f3d80bd86.
//
// Solidity: event NewBaseURI(string indexed baseURI)
func (_DcnRegistry *DcnRegistryFilterer) FilterNewBaseURI(opts *bind.FilterOpts, baseURI []string) (*DcnRegistryNewBaseURIIterator, error) {

	var baseURIRule []interface{}
	for _, baseURIItem := range baseURI {
		baseURIRule = append(baseURIRule, baseURIItem)
	}

	logs, sub, err := _DcnRegistry.contract.FilterLogs(opts, "NewBaseURI", baseURIRule)
	if err != nil {
		return nil, err
	}
	return &DcnRegistryNewBaseURIIterator{contract: _DcnRegistry.contract, event: "NewBaseURI", logs: logs, sub: sub}, nil
}

// WatchNewBaseURI is a free log subscription operation binding the contract event 0x325d37e8fb549c86966f09bc3e6f62eb3afa93b255d6e3234338001f3d80bd86.
//
// Solidity: event NewBaseURI(string indexed baseURI)
func (_DcnRegistry *DcnRegistryFilterer) WatchNewBaseURI(opts *bind.WatchOpts, sink chan<- *DcnRegistryNewBaseURI, baseURI []string) (event.Subscription, error) {

	var baseURIRule []interface{}
	for _, baseURIItem := range baseURI {
		baseURIRule = append(baseURIRule, baseURIItem)
	}

	logs, sub, err := _DcnRegistry.contract.WatchLogs(opts, "NewBaseURI", baseURIRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DcnRegistryNewBaseURI)
				if err := _DcnRegistry.contract.UnpackLog(event, "NewBaseURI", log); err != nil {
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

// ParseNewBaseURI is a log parse operation binding the contract event 0x325d37e8fb549c86966f09bc3e6f62eb3afa93b255d6e3234338001f3d80bd86.
//
// Solidity: event NewBaseURI(string indexed baseURI)
func (_DcnRegistry *DcnRegistryFilterer) ParseNewBaseURI(log types.Log) (*DcnRegistryNewBaseURI, error) {
	event := new(DcnRegistryNewBaseURI)
	if err := _DcnRegistry.contract.UnpackLog(event, "NewBaseURI", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DcnRegistryNewDefaultResolverIterator is returned from FilterNewDefaultResolver and is used to iterate over the raw logs and unpacked data for NewDefaultResolver events raised by the DcnRegistry contract.
type DcnRegistryNewDefaultResolverIterator struct {
	Event *DcnRegistryNewDefaultResolver // Event containing the contract specifics and raw log

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
func (it *DcnRegistryNewDefaultResolverIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DcnRegistryNewDefaultResolver)
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
		it.Event = new(DcnRegistryNewDefaultResolver)
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
func (it *DcnRegistryNewDefaultResolverIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DcnRegistryNewDefaultResolverIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DcnRegistryNewDefaultResolver represents a NewDefaultResolver event raised by the DcnRegistry contract.
type DcnRegistryNewDefaultResolver struct {
	DefaultResolver common.Address
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterNewDefaultResolver is a free log retrieval operation binding the contract event 0x3fea9780619f8b9e765fccf1e175478c75e6df0dc30650125e1ec1e035756b17.
//
// Solidity: event NewDefaultResolver(address indexed defaultResolver)
func (_DcnRegistry *DcnRegistryFilterer) FilterNewDefaultResolver(opts *bind.FilterOpts, defaultResolver []common.Address) (*DcnRegistryNewDefaultResolverIterator, error) {

	var defaultResolverRule []interface{}
	for _, defaultResolverItem := range defaultResolver {
		defaultResolverRule = append(defaultResolverRule, defaultResolverItem)
	}

	logs, sub, err := _DcnRegistry.contract.FilterLogs(opts, "NewDefaultResolver", defaultResolverRule)
	if err != nil {
		return nil, err
	}
	return &DcnRegistryNewDefaultResolverIterator{contract: _DcnRegistry.contract, event: "NewDefaultResolver", logs: logs, sub: sub}, nil
}

// WatchNewDefaultResolver is a free log subscription operation binding the contract event 0x3fea9780619f8b9e765fccf1e175478c75e6df0dc30650125e1ec1e035756b17.
//
// Solidity: event NewDefaultResolver(address indexed defaultResolver)
func (_DcnRegistry *DcnRegistryFilterer) WatchNewDefaultResolver(opts *bind.WatchOpts, sink chan<- *DcnRegistryNewDefaultResolver, defaultResolver []common.Address) (event.Subscription, error) {

	var defaultResolverRule []interface{}
	for _, defaultResolverItem := range defaultResolver {
		defaultResolverRule = append(defaultResolverRule, defaultResolverItem)
	}

	logs, sub, err := _DcnRegistry.contract.WatchLogs(opts, "NewDefaultResolver", defaultResolverRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DcnRegistryNewDefaultResolver)
				if err := _DcnRegistry.contract.UnpackLog(event, "NewDefaultResolver", log); err != nil {
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

// ParseNewDefaultResolver is a log parse operation binding the contract event 0x3fea9780619f8b9e765fccf1e175478c75e6df0dc30650125e1ec1e035756b17.
//
// Solidity: event NewDefaultResolver(address indexed defaultResolver)
func (_DcnRegistry *DcnRegistryFilterer) ParseNewDefaultResolver(log types.Log) (*DcnRegistryNewDefaultResolver, error) {
	event := new(DcnRegistryNewDefaultResolver)
	if err := _DcnRegistry.contract.UnpackLog(event, "NewDefaultResolver", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DcnRegistryNewExpirationIterator is returned from FilterNewExpiration and is used to iterate over the raw logs and unpacked data for NewExpiration events raised by the DcnRegistry contract.
type DcnRegistryNewExpirationIterator struct {
	Event *DcnRegistryNewExpiration // Event containing the contract specifics and raw log

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
func (it *DcnRegistryNewExpirationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DcnRegistryNewExpiration)
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
		it.Event = new(DcnRegistryNewExpiration)
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
func (it *DcnRegistryNewExpirationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DcnRegistryNewExpirationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DcnRegistryNewExpiration represents a NewExpiration event raised by the DcnRegistry contract.
type DcnRegistryNewExpiration struct {
	Node       [32]byte
	Expiration *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterNewExpiration is a free log retrieval operation binding the contract event 0x3675f1c37f2f7104e081dcec231e0fe4466579164fd14d1fda438f72720f9a3d.
//
// Solidity: event NewExpiration(bytes32 indexed node, uint256 expiration)
func (_DcnRegistry *DcnRegistryFilterer) FilterNewExpiration(opts *bind.FilterOpts, node [][32]byte) (*DcnRegistryNewExpirationIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _DcnRegistry.contract.FilterLogs(opts, "NewExpiration", nodeRule)
	if err != nil {
		return nil, err
	}
	return &DcnRegistryNewExpirationIterator{contract: _DcnRegistry.contract, event: "NewExpiration", logs: logs, sub: sub}, nil
}

// WatchNewExpiration is a free log subscription operation binding the contract event 0x3675f1c37f2f7104e081dcec231e0fe4466579164fd14d1fda438f72720f9a3d.
//
// Solidity: event NewExpiration(bytes32 indexed node, uint256 expiration)
func (_DcnRegistry *DcnRegistryFilterer) WatchNewExpiration(opts *bind.WatchOpts, sink chan<- *DcnRegistryNewExpiration, node [][32]byte) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _DcnRegistry.contract.WatchLogs(opts, "NewExpiration", nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DcnRegistryNewExpiration)
				if err := _DcnRegistry.contract.UnpackLog(event, "NewExpiration", log); err != nil {
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

// ParseNewExpiration is a log parse operation binding the contract event 0x3675f1c37f2f7104e081dcec231e0fe4466579164fd14d1fda438f72720f9a3d.
//
// Solidity: event NewExpiration(bytes32 indexed node, uint256 expiration)
func (_DcnRegistry *DcnRegistryFilterer) ParseNewExpiration(log types.Log) (*DcnRegistryNewExpiration, error) {
	event := new(DcnRegistryNewExpiration)
	if err := _DcnRegistry.contract.UnpackLog(event, "NewExpiration", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DcnRegistryNewNodeIterator is returned from FilterNewNode and is used to iterate over the raw logs and unpacked data for NewNode events raised by the DcnRegistry contract.
type DcnRegistryNewNodeIterator struct {
	Event *DcnRegistryNewNode // Event containing the contract specifics and raw log

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
func (it *DcnRegistryNewNodeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DcnRegistryNewNode)
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
		it.Event = new(DcnRegistryNewNode)
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
func (it *DcnRegistryNewNodeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DcnRegistryNewNodeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DcnRegistryNewNode represents a NewNode event raised by the DcnRegistry contract.
type DcnRegistryNewNode struct {
	Node  [32]byte
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterNewNode is a free log retrieval operation binding the contract event 0xf56af60bf9a9e7db86e26247d8f180353650027464ea5d0b83fb069738b8b741.
//
// Solidity: event NewNode(bytes32 indexed node, address indexed owner)
func (_DcnRegistry *DcnRegistryFilterer) FilterNewNode(opts *bind.FilterOpts, node [][32]byte, owner []common.Address) (*DcnRegistryNewNodeIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _DcnRegistry.contract.FilterLogs(opts, "NewNode", nodeRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &DcnRegistryNewNodeIterator{contract: _DcnRegistry.contract, event: "NewNode", logs: logs, sub: sub}, nil
}

// WatchNewNode is a free log subscription operation binding the contract event 0xf56af60bf9a9e7db86e26247d8f180353650027464ea5d0b83fb069738b8b741.
//
// Solidity: event NewNode(bytes32 indexed node, address indexed owner)
func (_DcnRegistry *DcnRegistryFilterer) WatchNewNode(opts *bind.WatchOpts, sink chan<- *DcnRegistryNewNode, node [][32]byte, owner []common.Address) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _DcnRegistry.contract.WatchLogs(opts, "NewNode", nodeRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DcnRegistryNewNode)
				if err := _DcnRegistry.contract.UnpackLog(event, "NewNode", log); err != nil {
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

// ParseNewNode is a log parse operation binding the contract event 0xf56af60bf9a9e7db86e26247d8f180353650027464ea5d0b83fb069738b8b741.
//
// Solidity: event NewNode(bytes32 indexed node, address indexed owner)
func (_DcnRegistry *DcnRegistryFilterer) ParseNewNode(log types.Log) (*DcnRegistryNewNode, error) {
	event := new(DcnRegistryNewNode)
	if err := _DcnRegistry.contract.UnpackLog(event, "NewNode", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DcnRegistryNewResolverIterator is returned from FilterNewResolver and is used to iterate over the raw logs and unpacked data for NewResolver events raised by the DcnRegistry contract.
type DcnRegistryNewResolverIterator struct {
	Event *DcnRegistryNewResolver // Event containing the contract specifics and raw log

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
func (it *DcnRegistryNewResolverIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DcnRegistryNewResolver)
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
		it.Event = new(DcnRegistryNewResolver)
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
func (it *DcnRegistryNewResolverIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DcnRegistryNewResolverIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DcnRegistryNewResolver represents a NewResolver event raised by the DcnRegistry contract.
type DcnRegistryNewResolver struct {
	Node     [32]byte
	Resolver common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterNewResolver is a free log retrieval operation binding the contract event 0x335721b01866dc23fbee8b6b2c7b1e14d6f05c28cd35a2c934239f94095602a0.
//
// Solidity: event NewResolver(bytes32 indexed node, address resolver)
func (_DcnRegistry *DcnRegistryFilterer) FilterNewResolver(opts *bind.FilterOpts, node [][32]byte) (*DcnRegistryNewResolverIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _DcnRegistry.contract.FilterLogs(opts, "NewResolver", nodeRule)
	if err != nil {
		return nil, err
	}
	return &DcnRegistryNewResolverIterator{contract: _DcnRegistry.contract, event: "NewResolver", logs: logs, sub: sub}, nil
}

// WatchNewResolver is a free log subscription operation binding the contract event 0x335721b01866dc23fbee8b6b2c7b1e14d6f05c28cd35a2c934239f94095602a0.
//
// Solidity: event NewResolver(bytes32 indexed node, address resolver)
func (_DcnRegistry *DcnRegistryFilterer) WatchNewResolver(opts *bind.WatchOpts, sink chan<- *DcnRegistryNewResolver, node [][32]byte) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _DcnRegistry.contract.WatchLogs(opts, "NewResolver", nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DcnRegistryNewResolver)
				if err := _DcnRegistry.contract.UnpackLog(event, "NewResolver", log); err != nil {
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

// ParseNewResolver is a log parse operation binding the contract event 0x335721b01866dc23fbee8b6b2c7b1e14d6f05c28cd35a2c934239f94095602a0.
//
// Solidity: event NewResolver(bytes32 indexed node, address resolver)
func (_DcnRegistry *DcnRegistryFilterer) ParseNewResolver(log types.Log) (*DcnRegistryNewResolver, error) {
	event := new(DcnRegistryNewResolver)
	if err := _DcnRegistry.contract.UnpackLog(event, "NewResolver", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DcnRegistryRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the DcnRegistry contract.
type DcnRegistryRoleAdminChangedIterator struct {
	Event *DcnRegistryRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *DcnRegistryRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DcnRegistryRoleAdminChanged)
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
		it.Event = new(DcnRegistryRoleAdminChanged)
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
func (it *DcnRegistryRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DcnRegistryRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DcnRegistryRoleAdminChanged represents a RoleAdminChanged event raised by the DcnRegistry contract.
type DcnRegistryRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_DcnRegistry *DcnRegistryFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*DcnRegistryRoleAdminChangedIterator, error) {

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

	logs, sub, err := _DcnRegistry.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &DcnRegistryRoleAdminChangedIterator{contract: _DcnRegistry.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_DcnRegistry *DcnRegistryFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *DcnRegistryRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

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

	logs, sub, err := _DcnRegistry.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DcnRegistryRoleAdminChanged)
				if err := _DcnRegistry.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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
func (_DcnRegistry *DcnRegistryFilterer) ParseRoleAdminChanged(log types.Log) (*DcnRegistryRoleAdminChanged, error) {
	event := new(DcnRegistryRoleAdminChanged)
	if err := _DcnRegistry.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DcnRegistryRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the DcnRegistry contract.
type DcnRegistryRoleGrantedIterator struct {
	Event *DcnRegistryRoleGranted // Event containing the contract specifics and raw log

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
func (it *DcnRegistryRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DcnRegistryRoleGranted)
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
		it.Event = new(DcnRegistryRoleGranted)
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
func (it *DcnRegistryRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DcnRegistryRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DcnRegistryRoleGranted represents a RoleGranted event raised by the DcnRegistry contract.
type DcnRegistryRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_DcnRegistry *DcnRegistryFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*DcnRegistryRoleGrantedIterator, error) {

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

	logs, sub, err := _DcnRegistry.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &DcnRegistryRoleGrantedIterator{contract: _DcnRegistry.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_DcnRegistry *DcnRegistryFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *DcnRegistryRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _DcnRegistry.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DcnRegistryRoleGranted)
				if err := _DcnRegistry.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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
func (_DcnRegistry *DcnRegistryFilterer) ParseRoleGranted(log types.Log) (*DcnRegistryRoleGranted, error) {
	event := new(DcnRegistryRoleGranted)
	if err := _DcnRegistry.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DcnRegistryRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the DcnRegistry contract.
type DcnRegistryRoleRevokedIterator struct {
	Event *DcnRegistryRoleRevoked // Event containing the contract specifics and raw log

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
func (it *DcnRegistryRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DcnRegistryRoleRevoked)
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
		it.Event = new(DcnRegistryRoleRevoked)
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
func (it *DcnRegistryRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DcnRegistryRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DcnRegistryRoleRevoked represents a RoleRevoked event raised by the DcnRegistry contract.
type DcnRegistryRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_DcnRegistry *DcnRegistryFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*DcnRegistryRoleRevokedIterator, error) {

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

	logs, sub, err := _DcnRegistry.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &DcnRegistryRoleRevokedIterator{contract: _DcnRegistry.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_DcnRegistry *DcnRegistryFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *DcnRegistryRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _DcnRegistry.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DcnRegistryRoleRevoked)
				if err := _DcnRegistry.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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
func (_DcnRegistry *DcnRegistryFilterer) ParseRoleRevoked(log types.Log) (*DcnRegistryRoleRevoked, error) {
	event := new(DcnRegistryRoleRevoked)
	if err := _DcnRegistry.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DcnRegistryTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the DcnRegistry contract.
type DcnRegistryTransferIterator struct {
	Event *DcnRegistryTransfer // Event containing the contract specifics and raw log

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
func (it *DcnRegistryTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DcnRegistryTransfer)
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
		it.Event = new(DcnRegistryTransfer)
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
func (it *DcnRegistryTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DcnRegistryTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DcnRegistryTransfer represents a Transfer event raised by the DcnRegistry contract.
type DcnRegistryTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_DcnRegistry *DcnRegistryFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address, tokenId []*big.Int) (*DcnRegistryTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _DcnRegistry.contract.FilterLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &DcnRegistryTransferIterator{contract: _DcnRegistry.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_DcnRegistry *DcnRegistryFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *DcnRegistryTransfer, from []common.Address, to []common.Address, tokenId []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _DcnRegistry.contract.WatchLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DcnRegistryTransfer)
				if err := _DcnRegistry.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_DcnRegistry *DcnRegistryFilterer) ParseTransfer(log types.Log) (*DcnRegistryTransfer, error) {
	event := new(DcnRegistryTransfer)
	if err := _DcnRegistry.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DcnRegistryUpgradedIterator is returned from FilterUpgraded and is used to iterate over the raw logs and unpacked data for Upgraded events raised by the DcnRegistry contract.
type DcnRegistryUpgradedIterator struct {
	Event *DcnRegistryUpgraded // Event containing the contract specifics and raw log

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
func (it *DcnRegistryUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DcnRegistryUpgraded)
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
		it.Event = new(DcnRegistryUpgraded)
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
func (it *DcnRegistryUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DcnRegistryUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DcnRegistryUpgraded represents a Upgraded event raised by the DcnRegistry contract.
type DcnRegistryUpgraded struct {
	Implementation common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpgraded is a free log retrieval operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_DcnRegistry *DcnRegistryFilterer) FilterUpgraded(opts *bind.FilterOpts, implementation []common.Address) (*DcnRegistryUpgradedIterator, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _DcnRegistry.contract.FilterLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return &DcnRegistryUpgradedIterator{contract: _DcnRegistry.contract, event: "Upgraded", logs: logs, sub: sub}, nil
}

// WatchUpgraded is a free log subscription operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_DcnRegistry *DcnRegistryFilterer) WatchUpgraded(opts *bind.WatchOpts, sink chan<- *DcnRegistryUpgraded, implementation []common.Address) (event.Subscription, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _DcnRegistry.contract.WatchLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DcnRegistryUpgraded)
				if err := _DcnRegistry.contract.UnpackLog(event, "Upgraded", log); err != nil {
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

// ParseUpgraded is a log parse operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_DcnRegistry *DcnRegistryFilterer) ParseUpgraded(log types.Log) (*DcnRegistryUpgraded, error) {
	event := new(DcnRegistryUpgraded)
	if err := _DcnRegistry.contract.UnpackLog(event, "Upgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
