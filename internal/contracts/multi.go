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
)

// MultiPrivilegeSetPrivilegeData is an auto generated low-level Go binding around an user-defined struct.
type MultiPrivilegeSetPrivilegeData struct {
	TokenId *big.Int
	PrivId  *big.Int
	User    common.Address
	Expires *big.Int
}

// MultiPrivilegeMetaData contains all meta data concerning the MultiPrivilege contract.
var MultiPrivilegeMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"previousAdmin\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"AdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"approved\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"beacon\",\"type\":\"address\"}],\"name\":\"BeaconUpgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"privilegeId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"enabled\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"description\",\"type\":\"string\"}],\"name\":\"PrivilegeCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"privilegeId\",\"type\":\"uint256\"}],\"name\":\"PrivilegeDisabled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"privilegeId\",\"type\":\"uint256\"}],\"name\":\"PrivilegeEnabled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"privId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"expires\",\"type\":\"uint256\"}],\"name\":\"PrivilegeSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"BURNER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DEFAULT_ADMIN_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MINTER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"TRANSFERER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"UPGRADER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"enabled\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"description\",\"type\":\"string\"}],\"name\":\"createPrivilege\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"privId\",\"type\":\"uint256\"}],\"name\":\"disablePrivilege\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"privId\",\"type\":\"uint256\"}],\"name\":\"enablePrivilege\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"exists\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getApproved\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"privId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"}],\"name\":\"hasPrivilege\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"privilegeEntry\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"privId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"}],\"name\":\"privilegeExpiresAt\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"privilegeRecord\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"enabled\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"description\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxiableUUID\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"safeMint\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"uri\",\"type\":\"string\"}],\"name\":\"safeMint\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"baseURI_\",\"type\":\"string\"}],\"name\":\"setBaseURI\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"privId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"expires\",\"type\":\"uint256\"}],\"name\":\"setPrivilege\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"privId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"expires\",\"type\":\"uint256\"}],\"internalType\":\"structMultiPrivilege.SetPrivilegeData[]\",\"name\":\"privData\",\"type\":\"tuple[]\"}],\"name\":\"setPrivileges\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"tokenIdToVersion\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"tokenURI\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"}],\"name\":\"upgradeTo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"upgradeToAndCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
}

// MultiPrivilegeABI is the input ABI used to generate the binding from.
// Deprecated: Use MultiPrivilegeMetaData.ABI instead.
var MultiPrivilegeABI = MultiPrivilegeMetaData.ABI

// MultiPrivilege is an auto generated Go binding around an Ethereum contract.
type MultiPrivilege struct {
	MultiPrivilegeCaller     // Read-only binding to the contract
	MultiPrivilegeTransactor // Write-only binding to the contract
	MultiPrivilegeFilterer   // Log filterer for contract events
}

// MultiPrivilegeCaller is an auto generated read-only Go binding around an Ethereum contract.
type MultiPrivilegeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultiPrivilegeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MultiPrivilegeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultiPrivilegeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MultiPrivilegeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultiPrivilegeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MultiPrivilegeSession struct {
	Contract     *MultiPrivilege   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MultiPrivilegeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MultiPrivilegeCallerSession struct {
	Contract *MultiPrivilegeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// MultiPrivilegeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MultiPrivilegeTransactorSession struct {
	Contract     *MultiPrivilegeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// MultiPrivilegeRaw is an auto generated low-level Go binding around an Ethereum contract.
type MultiPrivilegeRaw struct {
	Contract *MultiPrivilege // Generic contract binding to access the raw methods on
}

// MultiPrivilegeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MultiPrivilegeCallerRaw struct {
	Contract *MultiPrivilegeCaller // Generic read-only contract binding to access the raw methods on
}

// MultiPrivilegeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MultiPrivilegeTransactorRaw struct {
	Contract *MultiPrivilegeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMultiPrivilege creates a new instance of MultiPrivilege, bound to a specific deployed contract.
func NewMultiPrivilege(address common.Address, backend bind.ContractBackend) (*MultiPrivilege, error) {
	contract, err := bindMultiPrivilege(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MultiPrivilege{MultiPrivilegeCaller: MultiPrivilegeCaller{contract: contract}, MultiPrivilegeTransactor: MultiPrivilegeTransactor{contract: contract}, MultiPrivilegeFilterer: MultiPrivilegeFilterer{contract: contract}}, nil
}

// NewMultiPrivilegeCaller creates a new read-only instance of MultiPrivilege, bound to a specific deployed contract.
func NewMultiPrivilegeCaller(address common.Address, caller bind.ContractCaller) (*MultiPrivilegeCaller, error) {
	contract, err := bindMultiPrivilege(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MultiPrivilegeCaller{contract: contract}, nil
}

// NewMultiPrivilegeTransactor creates a new write-only instance of MultiPrivilege, bound to a specific deployed contract.
func NewMultiPrivilegeTransactor(address common.Address, transactor bind.ContractTransactor) (*MultiPrivilegeTransactor, error) {
	contract, err := bindMultiPrivilege(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MultiPrivilegeTransactor{contract: contract}, nil
}

// NewMultiPrivilegeFilterer creates a new log filterer instance of MultiPrivilege, bound to a specific deployed contract.
func NewMultiPrivilegeFilterer(address common.Address, filterer bind.ContractFilterer) (*MultiPrivilegeFilterer, error) {
	contract, err := bindMultiPrivilege(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MultiPrivilegeFilterer{contract: contract}, nil
}

// bindMultiPrivilege binds a generic wrapper to an already deployed contract.
func bindMultiPrivilege(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MultiPrivilegeABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MultiPrivilege *MultiPrivilegeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MultiPrivilege.Contract.MultiPrivilegeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MultiPrivilege *MultiPrivilegeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.MultiPrivilegeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MultiPrivilege *MultiPrivilegeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.MultiPrivilegeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MultiPrivilege *MultiPrivilegeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MultiPrivilege.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MultiPrivilege *MultiPrivilegeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MultiPrivilege *MultiPrivilegeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.contract.Transact(opts, method, params...)
}

// BURNERROLE is a free data retrieval call binding the contract method 0x282c51f3.
//
// Solidity: function BURNER_ROLE() view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeCaller) BURNERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "BURNER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BURNERROLE is a free data retrieval call binding the contract method 0x282c51f3.
//
// Solidity: function BURNER_ROLE() view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeSession) BURNERROLE() ([32]byte, error) {
	return _MultiPrivilege.Contract.BURNERROLE(&_MultiPrivilege.CallOpts)
}

// BURNERROLE is a free data retrieval call binding the contract method 0x282c51f3.
//
// Solidity: function BURNER_ROLE() view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeCallerSession) BURNERROLE() ([32]byte, error) {
	return _MultiPrivilege.Contract.BURNERROLE(&_MultiPrivilege.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _MultiPrivilege.Contract.DEFAULTADMINROLE(&_MultiPrivilege.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _MultiPrivilege.Contract.DEFAULTADMINROLE(&_MultiPrivilege.CallOpts)
}

// MINTERROLE is a free data retrieval call binding the contract method 0xd5391393.
//
// Solidity: function MINTER_ROLE() view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeCaller) MINTERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "MINTER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MINTERROLE is a free data retrieval call binding the contract method 0xd5391393.
//
// Solidity: function MINTER_ROLE() view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeSession) MINTERROLE() ([32]byte, error) {
	return _MultiPrivilege.Contract.MINTERROLE(&_MultiPrivilege.CallOpts)
}

// MINTERROLE is a free data retrieval call binding the contract method 0xd5391393.
//
// Solidity: function MINTER_ROLE() view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeCallerSession) MINTERROLE() ([32]byte, error) {
	return _MultiPrivilege.Contract.MINTERROLE(&_MultiPrivilege.CallOpts)
}

// TRANSFERERROLE is a free data retrieval call binding the contract method 0x0ade7dc1.
//
// Solidity: function TRANSFERER_ROLE() view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeCaller) TRANSFERERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "TRANSFERER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// TRANSFERERROLE is a free data retrieval call binding the contract method 0x0ade7dc1.
//
// Solidity: function TRANSFERER_ROLE() view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeSession) TRANSFERERROLE() ([32]byte, error) {
	return _MultiPrivilege.Contract.TRANSFERERROLE(&_MultiPrivilege.CallOpts)
}

// TRANSFERERROLE is a free data retrieval call binding the contract method 0x0ade7dc1.
//
// Solidity: function TRANSFERER_ROLE() view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeCallerSession) TRANSFERERROLE() ([32]byte, error) {
	return _MultiPrivilege.Contract.TRANSFERERROLE(&_MultiPrivilege.CallOpts)
}

// UPGRADERROLE is a free data retrieval call binding the contract method 0xf72c0d8b.
//
// Solidity: function UPGRADER_ROLE() view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeCaller) UPGRADERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "UPGRADER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// UPGRADERROLE is a free data retrieval call binding the contract method 0xf72c0d8b.
//
// Solidity: function UPGRADER_ROLE() view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeSession) UPGRADERROLE() ([32]byte, error) {
	return _MultiPrivilege.Contract.UPGRADERROLE(&_MultiPrivilege.CallOpts)
}

// UPGRADERROLE is a free data retrieval call binding the contract method 0xf72c0d8b.
//
// Solidity: function UPGRADER_ROLE() view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeCallerSession) UPGRADERROLE() ([32]byte, error) {
	return _MultiPrivilege.Contract.UPGRADERROLE(&_MultiPrivilege.CallOpts)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_MultiPrivilege *MultiPrivilegeCaller) BalanceOf(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "balanceOf", owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_MultiPrivilege *MultiPrivilegeSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _MultiPrivilege.Contract.BalanceOf(&_MultiPrivilege.CallOpts, owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_MultiPrivilege *MultiPrivilegeCallerSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _MultiPrivilege.Contract.BalanceOf(&_MultiPrivilege.CallOpts, owner)
}

// Exists is a free data retrieval call binding the contract method 0x4f558e79.
//
// Solidity: function exists(uint256 tokenId) view returns(bool)
func (_MultiPrivilege *MultiPrivilegeCaller) Exists(opts *bind.CallOpts, tokenId *big.Int) (bool, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "exists", tokenId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Exists is a free data retrieval call binding the contract method 0x4f558e79.
//
// Solidity: function exists(uint256 tokenId) view returns(bool)
func (_MultiPrivilege *MultiPrivilegeSession) Exists(tokenId *big.Int) (bool, error) {
	return _MultiPrivilege.Contract.Exists(&_MultiPrivilege.CallOpts, tokenId)
}

// Exists is a free data retrieval call binding the contract method 0x4f558e79.
//
// Solidity: function exists(uint256 tokenId) view returns(bool)
func (_MultiPrivilege *MultiPrivilegeCallerSession) Exists(tokenId *big.Int) (bool, error) {
	return _MultiPrivilege.Contract.Exists(&_MultiPrivilege.CallOpts, tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_MultiPrivilege *MultiPrivilegeCaller) GetApproved(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "getApproved", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_MultiPrivilege *MultiPrivilegeSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _MultiPrivilege.Contract.GetApproved(&_MultiPrivilege.CallOpts, tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_MultiPrivilege *MultiPrivilegeCallerSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _MultiPrivilege.Contract.GetApproved(&_MultiPrivilege.CallOpts, tokenId)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _MultiPrivilege.Contract.GetRoleAdmin(&_MultiPrivilege.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _MultiPrivilege.Contract.GetRoleAdmin(&_MultiPrivilege.CallOpts, role)
}

// HasPrivilege is a free data retrieval call binding the contract method 0x05d80b00.
//
// Solidity: function hasPrivilege(uint256 tokenId, uint256 privId, address user) view returns(bool)
func (_MultiPrivilege *MultiPrivilegeCaller) HasPrivilege(opts *bind.CallOpts, tokenId *big.Int, privId *big.Int, user common.Address) (bool, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "hasPrivilege", tokenId, privId, user)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasPrivilege is a free data retrieval call binding the contract method 0x05d80b00.
//
// Solidity: function hasPrivilege(uint256 tokenId, uint256 privId, address user) view returns(bool)
func (_MultiPrivilege *MultiPrivilegeSession) HasPrivilege(tokenId *big.Int, privId *big.Int, user common.Address) (bool, error) {
	return _MultiPrivilege.Contract.HasPrivilege(&_MultiPrivilege.CallOpts, tokenId, privId, user)
}

// HasPrivilege is a free data retrieval call binding the contract method 0x05d80b00.
//
// Solidity: function hasPrivilege(uint256 tokenId, uint256 privId, address user) view returns(bool)
func (_MultiPrivilege *MultiPrivilegeCallerSession) HasPrivilege(tokenId *big.Int, privId *big.Int, user common.Address) (bool, error) {
	return _MultiPrivilege.Contract.HasPrivilege(&_MultiPrivilege.CallOpts, tokenId, privId, user)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_MultiPrivilege *MultiPrivilegeCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_MultiPrivilege *MultiPrivilegeSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _MultiPrivilege.Contract.HasRole(&_MultiPrivilege.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_MultiPrivilege *MultiPrivilegeCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _MultiPrivilege.Contract.HasRole(&_MultiPrivilege.CallOpts, role, account)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_MultiPrivilege *MultiPrivilegeCaller) IsApprovedForAll(opts *bind.CallOpts, owner common.Address, operator common.Address) (bool, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "isApprovedForAll", owner, operator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_MultiPrivilege *MultiPrivilegeSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _MultiPrivilege.Contract.IsApprovedForAll(&_MultiPrivilege.CallOpts, owner, operator)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_MultiPrivilege *MultiPrivilegeCallerSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _MultiPrivilege.Contract.IsApprovedForAll(&_MultiPrivilege.CallOpts, owner, operator)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_MultiPrivilege *MultiPrivilegeCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_MultiPrivilege *MultiPrivilegeSession) Name() (string, error) {
	return _MultiPrivilege.Contract.Name(&_MultiPrivilege.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_MultiPrivilege *MultiPrivilegeCallerSession) Name() (string, error) {
	return _MultiPrivilege.Contract.Name(&_MultiPrivilege.CallOpts)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_MultiPrivilege *MultiPrivilegeCaller) OwnerOf(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "ownerOf", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_MultiPrivilege *MultiPrivilegeSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _MultiPrivilege.Contract.OwnerOf(&_MultiPrivilege.CallOpts, tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_MultiPrivilege *MultiPrivilegeCallerSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _MultiPrivilege.Contract.OwnerOf(&_MultiPrivilege.CallOpts, tokenId)
}

// PrivilegeEntry is a free data retrieval call binding the contract method 0x48db4640.
//
// Solidity: function privilegeEntry(uint256 , uint256 , uint256 , address ) view returns(uint256)
func (_MultiPrivilege *MultiPrivilegeCaller) PrivilegeEntry(opts *bind.CallOpts, arg0 *big.Int, arg1 *big.Int, arg2 *big.Int, arg3 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "privilegeEntry", arg0, arg1, arg2, arg3)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PrivilegeEntry is a free data retrieval call binding the contract method 0x48db4640.
//
// Solidity: function privilegeEntry(uint256 , uint256 , uint256 , address ) view returns(uint256)
func (_MultiPrivilege *MultiPrivilegeSession) PrivilegeEntry(arg0 *big.Int, arg1 *big.Int, arg2 *big.Int, arg3 common.Address) (*big.Int, error) {
	return _MultiPrivilege.Contract.PrivilegeEntry(&_MultiPrivilege.CallOpts, arg0, arg1, arg2, arg3)
}

// PrivilegeEntry is a free data retrieval call binding the contract method 0x48db4640.
//
// Solidity: function privilegeEntry(uint256 , uint256 , uint256 , address ) view returns(uint256)
func (_MultiPrivilege *MultiPrivilegeCallerSession) PrivilegeEntry(arg0 *big.Int, arg1 *big.Int, arg2 *big.Int, arg3 common.Address) (*big.Int, error) {
	return _MultiPrivilege.Contract.PrivilegeEntry(&_MultiPrivilege.CallOpts, arg0, arg1, arg2, arg3)
}

// PrivilegeExpiresAt is a free data retrieval call binding the contract method 0xd0f8f5f6.
//
// Solidity: function privilegeExpiresAt(uint256 tokenId, uint256 privId, address user) view returns(uint256)
func (_MultiPrivilege *MultiPrivilegeCaller) PrivilegeExpiresAt(opts *bind.CallOpts, tokenId *big.Int, privId *big.Int, user common.Address) (*big.Int, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "privilegeExpiresAt", tokenId, privId, user)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PrivilegeExpiresAt is a free data retrieval call binding the contract method 0xd0f8f5f6.
//
// Solidity: function privilegeExpiresAt(uint256 tokenId, uint256 privId, address user) view returns(uint256)
func (_MultiPrivilege *MultiPrivilegeSession) PrivilegeExpiresAt(tokenId *big.Int, privId *big.Int, user common.Address) (*big.Int, error) {
	return _MultiPrivilege.Contract.PrivilegeExpiresAt(&_MultiPrivilege.CallOpts, tokenId, privId, user)
}

// PrivilegeExpiresAt is a free data retrieval call binding the contract method 0xd0f8f5f6.
//
// Solidity: function privilegeExpiresAt(uint256 tokenId, uint256 privId, address user) view returns(uint256)
func (_MultiPrivilege *MultiPrivilegeCallerSession) PrivilegeExpiresAt(tokenId *big.Int, privId *big.Int, user common.Address) (*big.Int, error) {
	return _MultiPrivilege.Contract.PrivilegeExpiresAt(&_MultiPrivilege.CallOpts, tokenId, privId, user)
}

// PrivilegeRecord is a free data retrieval call binding the contract method 0xf9ad3efe.
//
// Solidity: function privilegeRecord(uint256 ) view returns(bool enabled, string description)
func (_MultiPrivilege *MultiPrivilegeCaller) PrivilegeRecord(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Enabled     bool
	Description string
}, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "privilegeRecord", arg0)

	outstruct := new(struct {
		Enabled     bool
		Description string
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Enabled = *abi.ConvertType(out[0], new(bool)).(*bool)
	outstruct.Description = *abi.ConvertType(out[1], new(string)).(*string)

	return *outstruct, err

}

// PrivilegeRecord is a free data retrieval call binding the contract method 0xf9ad3efe.
//
// Solidity: function privilegeRecord(uint256 ) view returns(bool enabled, string description)
func (_MultiPrivilege *MultiPrivilegeSession) PrivilegeRecord(arg0 *big.Int) (struct {
	Enabled     bool
	Description string
}, error) {
	return _MultiPrivilege.Contract.PrivilegeRecord(&_MultiPrivilege.CallOpts, arg0)
}

// PrivilegeRecord is a free data retrieval call binding the contract method 0xf9ad3efe.
//
// Solidity: function privilegeRecord(uint256 ) view returns(bool enabled, string description)
func (_MultiPrivilege *MultiPrivilegeCallerSession) PrivilegeRecord(arg0 *big.Int) (struct {
	Enabled     bool
	Description string
}, error) {
	return _MultiPrivilege.Contract.PrivilegeRecord(&_MultiPrivilege.CallOpts, arg0)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeCaller) ProxiableUUID(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "proxiableUUID")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeSession) ProxiableUUID() ([32]byte, error) {
	return _MultiPrivilege.Contract.ProxiableUUID(&_MultiPrivilege.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_MultiPrivilege *MultiPrivilegeCallerSession) ProxiableUUID() ([32]byte, error) {
	return _MultiPrivilege.Contract.ProxiableUUID(&_MultiPrivilege.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_MultiPrivilege *MultiPrivilegeCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_MultiPrivilege *MultiPrivilegeSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _MultiPrivilege.Contract.SupportsInterface(&_MultiPrivilege.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_MultiPrivilege *MultiPrivilegeCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _MultiPrivilege.Contract.SupportsInterface(&_MultiPrivilege.CallOpts, interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_MultiPrivilege *MultiPrivilegeCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_MultiPrivilege *MultiPrivilegeSession) Symbol() (string, error) {
	return _MultiPrivilege.Contract.Symbol(&_MultiPrivilege.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_MultiPrivilege *MultiPrivilegeCallerSession) Symbol() (string, error) {
	return _MultiPrivilege.Contract.Symbol(&_MultiPrivilege.CallOpts)
}

// TokenIdToVersion is a free data retrieval call binding the contract method 0xf1a9d41c.
//
// Solidity: function tokenIdToVersion(uint256 ) view returns(uint256)
func (_MultiPrivilege *MultiPrivilegeCaller) TokenIdToVersion(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "tokenIdToVersion", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TokenIdToVersion is a free data retrieval call binding the contract method 0xf1a9d41c.
//
// Solidity: function tokenIdToVersion(uint256 ) view returns(uint256)
func (_MultiPrivilege *MultiPrivilegeSession) TokenIdToVersion(arg0 *big.Int) (*big.Int, error) {
	return _MultiPrivilege.Contract.TokenIdToVersion(&_MultiPrivilege.CallOpts, arg0)
}

// TokenIdToVersion is a free data retrieval call binding the contract method 0xf1a9d41c.
//
// Solidity: function tokenIdToVersion(uint256 ) view returns(uint256)
func (_MultiPrivilege *MultiPrivilegeCallerSession) TokenIdToVersion(arg0 *big.Int) (*big.Int, error) {
	return _MultiPrivilege.Contract.TokenIdToVersion(&_MultiPrivilege.CallOpts, arg0)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_MultiPrivilege *MultiPrivilegeCaller) TokenURI(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var out []interface{}
	err := _MultiPrivilege.contract.Call(opts, &out, "tokenURI", tokenId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_MultiPrivilege *MultiPrivilegeSession) TokenURI(tokenId *big.Int) (string, error) {
	return _MultiPrivilege.Contract.TokenURI(&_MultiPrivilege.CallOpts, tokenId)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_MultiPrivilege *MultiPrivilegeCallerSession) TokenURI(tokenId *big.Int) (string, error) {
	return _MultiPrivilege.Contract.TokenURI(&_MultiPrivilege.CallOpts, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_MultiPrivilege *MultiPrivilegeTransactor) Approve(opts *bind.TransactOpts, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.contract.Transact(opts, "approve", to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_MultiPrivilege *MultiPrivilegeSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.Approve(&_MultiPrivilege.TransactOpts, to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_MultiPrivilege *MultiPrivilegeTransactorSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.Approve(&_MultiPrivilege.TransactOpts, to, tokenId)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 tokenId) returns()
func (_MultiPrivilege *MultiPrivilegeTransactor) Burn(opts *bind.TransactOpts, tokenId *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.contract.Transact(opts, "burn", tokenId)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 tokenId) returns()
func (_MultiPrivilege *MultiPrivilegeSession) Burn(tokenId *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.Burn(&_MultiPrivilege.TransactOpts, tokenId)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 tokenId) returns()
func (_MultiPrivilege *MultiPrivilegeTransactorSession) Burn(tokenId *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.Burn(&_MultiPrivilege.TransactOpts, tokenId)
}

// CreatePrivilege is a paid mutator transaction binding the contract method 0xc1d58b3b.
//
// Solidity: function createPrivilege(bool enabled, string description) returns()
func (_MultiPrivilege *MultiPrivilegeTransactor) CreatePrivilege(opts *bind.TransactOpts, enabled bool, description string) (*types.Transaction, error) {
	return _MultiPrivilege.contract.Transact(opts, "createPrivilege", enabled, description)
}

// CreatePrivilege is a paid mutator transaction binding the contract method 0xc1d58b3b.
//
// Solidity: function createPrivilege(bool enabled, string description) returns()
func (_MultiPrivilege *MultiPrivilegeSession) CreatePrivilege(enabled bool, description string) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.CreatePrivilege(&_MultiPrivilege.TransactOpts, enabled, description)
}

// CreatePrivilege is a paid mutator transaction binding the contract method 0xc1d58b3b.
//
// Solidity: function createPrivilege(bool enabled, string description) returns()
func (_MultiPrivilege *MultiPrivilegeTransactorSession) CreatePrivilege(enabled bool, description string) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.CreatePrivilege(&_MultiPrivilege.TransactOpts, enabled, description)
}

// DisablePrivilege is a paid mutator transaction binding the contract method 0x1a153ed0.
//
// Solidity: function disablePrivilege(uint256 privId) returns()
func (_MultiPrivilege *MultiPrivilegeTransactor) DisablePrivilege(opts *bind.TransactOpts, privId *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.contract.Transact(opts, "disablePrivilege", privId)
}

// DisablePrivilege is a paid mutator transaction binding the contract method 0x1a153ed0.
//
// Solidity: function disablePrivilege(uint256 privId) returns()
func (_MultiPrivilege *MultiPrivilegeSession) DisablePrivilege(privId *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.DisablePrivilege(&_MultiPrivilege.TransactOpts, privId)
}

// DisablePrivilege is a paid mutator transaction binding the contract method 0x1a153ed0.
//
// Solidity: function disablePrivilege(uint256 privId) returns()
func (_MultiPrivilege *MultiPrivilegeTransactorSession) DisablePrivilege(privId *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.DisablePrivilege(&_MultiPrivilege.TransactOpts, privId)
}

// EnablePrivilege is a paid mutator transaction binding the contract method 0x831ba696.
//
// Solidity: function enablePrivilege(uint256 privId) returns()
func (_MultiPrivilege *MultiPrivilegeTransactor) EnablePrivilege(opts *bind.TransactOpts, privId *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.contract.Transact(opts, "enablePrivilege", privId)
}

// EnablePrivilege is a paid mutator transaction binding the contract method 0x831ba696.
//
// Solidity: function enablePrivilege(uint256 privId) returns()
func (_MultiPrivilege *MultiPrivilegeSession) EnablePrivilege(privId *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.EnablePrivilege(&_MultiPrivilege.TransactOpts, privId)
}

// EnablePrivilege is a paid mutator transaction binding the contract method 0x831ba696.
//
// Solidity: function enablePrivilege(uint256 privId) returns()
func (_MultiPrivilege *MultiPrivilegeTransactorSession) EnablePrivilege(privId *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.EnablePrivilege(&_MultiPrivilege.TransactOpts, privId)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_MultiPrivilege *MultiPrivilegeTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _MultiPrivilege.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_MultiPrivilege *MultiPrivilegeSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.GrantRole(&_MultiPrivilege.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_MultiPrivilege *MultiPrivilegeTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.GrantRole(&_MultiPrivilege.TransactOpts, role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_MultiPrivilege *MultiPrivilegeTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _MultiPrivilege.contract.Transact(opts, "renounceRole", role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_MultiPrivilege *MultiPrivilegeSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.RenounceRole(&_MultiPrivilege.TransactOpts, role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_MultiPrivilege *MultiPrivilegeTransactorSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.RenounceRole(&_MultiPrivilege.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_MultiPrivilege *MultiPrivilegeTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _MultiPrivilege.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_MultiPrivilege *MultiPrivilegeSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.RevokeRole(&_MultiPrivilege.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_MultiPrivilege *MultiPrivilegeTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.RevokeRole(&_MultiPrivilege.TransactOpts, role, account)
}

// SafeMint is a paid mutator transaction binding the contract method 0x40d097c3.
//
// Solidity: function safeMint(address to) returns(uint256 tokenId)
func (_MultiPrivilege *MultiPrivilegeTransactor) SafeMint(opts *bind.TransactOpts, to common.Address) (*types.Transaction, error) {
	return _MultiPrivilege.contract.Transact(opts, "safeMint", to)
}

// SafeMint is a paid mutator transaction binding the contract method 0x40d097c3.
//
// Solidity: function safeMint(address to) returns(uint256 tokenId)
func (_MultiPrivilege *MultiPrivilegeSession) SafeMint(to common.Address) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.SafeMint(&_MultiPrivilege.TransactOpts, to)
}

// SafeMint is a paid mutator transaction binding the contract method 0x40d097c3.
//
// Solidity: function safeMint(address to) returns(uint256 tokenId)
func (_MultiPrivilege *MultiPrivilegeTransactorSession) SafeMint(to common.Address) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.SafeMint(&_MultiPrivilege.TransactOpts, to)
}

// SafeMint0 is a paid mutator transaction binding the contract method 0xd204c45e.
//
// Solidity: function safeMint(address to, string uri) returns(uint256 tokenId)
func (_MultiPrivilege *MultiPrivilegeTransactor) SafeMint0(opts *bind.TransactOpts, to common.Address, uri string) (*types.Transaction, error) {
	return _MultiPrivilege.contract.Transact(opts, "safeMint0", to, uri)
}

// SafeMint0 is a paid mutator transaction binding the contract method 0xd204c45e.
//
// Solidity: function safeMint(address to, string uri) returns(uint256 tokenId)
func (_MultiPrivilege *MultiPrivilegeSession) SafeMint0(to common.Address, uri string) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.SafeMint0(&_MultiPrivilege.TransactOpts, to, uri)
}

// SafeMint0 is a paid mutator transaction binding the contract method 0xd204c45e.
//
// Solidity: function safeMint(address to, string uri) returns(uint256 tokenId)
func (_MultiPrivilege *MultiPrivilegeTransactorSession) SafeMint0(to common.Address, uri string) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.SafeMint0(&_MultiPrivilege.TransactOpts, to, uri)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_MultiPrivilege *MultiPrivilegeTransactor) SafeTransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.contract.Transact(opts, "safeTransferFrom", from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_MultiPrivilege *MultiPrivilegeSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.SafeTransferFrom(&_MultiPrivilege.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_MultiPrivilege *MultiPrivilegeTransactorSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.SafeTransferFrom(&_MultiPrivilege.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_MultiPrivilege *MultiPrivilegeTransactor) SafeTransferFrom0(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _MultiPrivilege.contract.Transact(opts, "safeTransferFrom0", from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_MultiPrivilege *MultiPrivilegeSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.SafeTransferFrom0(&_MultiPrivilege.TransactOpts, from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_MultiPrivilege *MultiPrivilegeTransactorSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.SafeTransferFrom0(&_MultiPrivilege.TransactOpts, from, to, tokenId, data)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_MultiPrivilege *MultiPrivilegeTransactor) SetApprovalForAll(opts *bind.TransactOpts, operator common.Address, approved bool) (*types.Transaction, error) {
	return _MultiPrivilege.contract.Transact(opts, "setApprovalForAll", operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_MultiPrivilege *MultiPrivilegeSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.SetApprovalForAll(&_MultiPrivilege.TransactOpts, operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_MultiPrivilege *MultiPrivilegeTransactorSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.SetApprovalForAll(&_MultiPrivilege.TransactOpts, operator, approved)
}

// SetBaseURI is a paid mutator transaction binding the contract method 0x55f804b3.
//
// Solidity: function setBaseURI(string baseURI_) returns()
func (_MultiPrivilege *MultiPrivilegeTransactor) SetBaseURI(opts *bind.TransactOpts, baseURI_ string) (*types.Transaction, error) {
	return _MultiPrivilege.contract.Transact(opts, "setBaseURI", baseURI_)
}

// SetBaseURI is a paid mutator transaction binding the contract method 0x55f804b3.
//
// Solidity: function setBaseURI(string baseURI_) returns()
func (_MultiPrivilege *MultiPrivilegeSession) SetBaseURI(baseURI_ string) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.SetBaseURI(&_MultiPrivilege.TransactOpts, baseURI_)
}

// SetBaseURI is a paid mutator transaction binding the contract method 0x55f804b3.
//
// Solidity: function setBaseURI(string baseURI_) returns()
func (_MultiPrivilege *MultiPrivilegeTransactorSession) SetBaseURI(baseURI_ string) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.SetBaseURI(&_MultiPrivilege.TransactOpts, baseURI_)
}

// SetPrivilege is a paid mutator transaction binding the contract method 0xeca3221a.
//
// Solidity: function setPrivilege(uint256 tokenId, uint256 privId, address user, uint256 expires) returns()
func (_MultiPrivilege *MultiPrivilegeTransactor) SetPrivilege(opts *bind.TransactOpts, tokenId *big.Int, privId *big.Int, user common.Address, expires *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.contract.Transact(opts, "setPrivilege", tokenId, privId, user, expires)
}

// SetPrivilege is a paid mutator transaction binding the contract method 0xeca3221a.
//
// Solidity: function setPrivilege(uint256 tokenId, uint256 privId, address user, uint256 expires) returns()
func (_MultiPrivilege *MultiPrivilegeSession) SetPrivilege(tokenId *big.Int, privId *big.Int, user common.Address, expires *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.SetPrivilege(&_MultiPrivilege.TransactOpts, tokenId, privId, user, expires)
}

// SetPrivilege is a paid mutator transaction binding the contract method 0xeca3221a.
//
// Solidity: function setPrivilege(uint256 tokenId, uint256 privId, address user, uint256 expires) returns()
func (_MultiPrivilege *MultiPrivilegeTransactorSession) SetPrivilege(tokenId *big.Int, privId *big.Int, user common.Address, expires *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.SetPrivilege(&_MultiPrivilege.TransactOpts, tokenId, privId, user, expires)
}

// SetPrivileges is a paid mutator transaction binding the contract method 0x57ae9754.
//
// Solidity: function setPrivileges((uint256,uint256,address,uint256)[] privData) returns()
func (_MultiPrivilege *MultiPrivilegeTransactor) SetPrivileges(opts *bind.TransactOpts, privData []MultiPrivilegeSetPrivilegeData) (*types.Transaction, error) {
	return _MultiPrivilege.contract.Transact(opts, "setPrivileges", privData)
}

// SetPrivileges is a paid mutator transaction binding the contract method 0x57ae9754.
//
// Solidity: function setPrivileges((uint256,uint256,address,uint256)[] privData) returns()
func (_MultiPrivilege *MultiPrivilegeSession) SetPrivileges(privData []MultiPrivilegeSetPrivilegeData) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.SetPrivileges(&_MultiPrivilege.TransactOpts, privData)
}

// SetPrivileges is a paid mutator transaction binding the contract method 0x57ae9754.
//
// Solidity: function setPrivileges((uint256,uint256,address,uint256)[] privData) returns()
func (_MultiPrivilege *MultiPrivilegeTransactorSession) SetPrivileges(privData []MultiPrivilegeSetPrivilegeData) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.SetPrivileges(&_MultiPrivilege.TransactOpts, privData)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_MultiPrivilege *MultiPrivilegeTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.contract.Transact(opts, "transferFrom", from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_MultiPrivilege *MultiPrivilegeSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.TransferFrom(&_MultiPrivilege.TransactOpts, from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_MultiPrivilege *MultiPrivilegeTransactorSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.TransferFrom(&_MultiPrivilege.TransactOpts, from, to, tokenId)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_MultiPrivilege *MultiPrivilegeTransactor) UpgradeTo(opts *bind.TransactOpts, newImplementation common.Address) (*types.Transaction, error) {
	return _MultiPrivilege.contract.Transact(opts, "upgradeTo", newImplementation)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_MultiPrivilege *MultiPrivilegeSession) UpgradeTo(newImplementation common.Address) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.UpgradeTo(&_MultiPrivilege.TransactOpts, newImplementation)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_MultiPrivilege *MultiPrivilegeTransactorSession) UpgradeTo(newImplementation common.Address) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.UpgradeTo(&_MultiPrivilege.TransactOpts, newImplementation)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_MultiPrivilege *MultiPrivilegeTransactor) UpgradeToAndCall(opts *bind.TransactOpts, newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _MultiPrivilege.contract.Transact(opts, "upgradeToAndCall", newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_MultiPrivilege *MultiPrivilegeSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.UpgradeToAndCall(&_MultiPrivilege.TransactOpts, newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_MultiPrivilege *MultiPrivilegeTransactorSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _MultiPrivilege.Contract.UpgradeToAndCall(&_MultiPrivilege.TransactOpts, newImplementation, data)
}

// MultiPrivilegeAdminChangedIterator is returned from FilterAdminChanged and is used to iterate over the raw logs and unpacked data for AdminChanged events raised by the MultiPrivilege contract.
type MultiPrivilegeAdminChangedIterator struct {
	Event *MultiPrivilegeAdminChanged // Event containing the contract specifics and raw log

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
func (it *MultiPrivilegeAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiPrivilegeAdminChanged)
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
		it.Event = new(MultiPrivilegeAdminChanged)
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
func (it *MultiPrivilegeAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiPrivilegeAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiPrivilegeAdminChanged represents a AdminChanged event raised by the MultiPrivilege contract.
type MultiPrivilegeAdminChanged struct {
	PreviousAdmin common.Address
	NewAdmin      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterAdminChanged is a free log retrieval operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_MultiPrivilege *MultiPrivilegeFilterer) FilterAdminChanged(opts *bind.FilterOpts) (*MultiPrivilegeAdminChangedIterator, error) {

	logs, sub, err := _MultiPrivilege.contract.FilterLogs(opts, "AdminChanged")
	if err != nil {
		return nil, err
	}
	return &MultiPrivilegeAdminChangedIterator{contract: _MultiPrivilege.contract, event: "AdminChanged", logs: logs, sub: sub}, nil
}

// WatchAdminChanged is a free log subscription operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_MultiPrivilege *MultiPrivilegeFilterer) WatchAdminChanged(opts *bind.WatchOpts, sink chan<- *MultiPrivilegeAdminChanged) (event.Subscription, error) {

	logs, sub, err := _MultiPrivilege.contract.WatchLogs(opts, "AdminChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiPrivilegeAdminChanged)
				if err := _MultiPrivilege.contract.UnpackLog(event, "AdminChanged", log); err != nil {
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
func (_MultiPrivilege *MultiPrivilegeFilterer) ParseAdminChanged(log types.Log) (*MultiPrivilegeAdminChanged, error) {
	event := new(MultiPrivilegeAdminChanged)
	if err := _MultiPrivilege.contract.UnpackLog(event, "AdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiPrivilegeApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the MultiPrivilege contract.
type MultiPrivilegeApprovalIterator struct {
	Event *MultiPrivilegeApproval // Event containing the contract specifics and raw log

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
func (it *MultiPrivilegeApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiPrivilegeApproval)
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
		it.Event = new(MultiPrivilegeApproval)
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
func (it *MultiPrivilegeApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiPrivilegeApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiPrivilegeApproval represents a Approval event raised by the MultiPrivilege contract.
type MultiPrivilegeApproval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_MultiPrivilege *MultiPrivilegeFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, approved []common.Address, tokenId []*big.Int) (*MultiPrivilegeApprovalIterator, error) {

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

	logs, sub, err := _MultiPrivilege.contract.FilterLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &MultiPrivilegeApprovalIterator{contract: _MultiPrivilege.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_MultiPrivilege *MultiPrivilegeFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *MultiPrivilegeApproval, owner []common.Address, approved []common.Address, tokenId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _MultiPrivilege.contract.WatchLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiPrivilegeApproval)
				if err := _MultiPrivilege.contract.UnpackLog(event, "Approval", log); err != nil {
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
func (_MultiPrivilege *MultiPrivilegeFilterer) ParseApproval(log types.Log) (*MultiPrivilegeApproval, error) {
	event := new(MultiPrivilegeApproval)
	if err := _MultiPrivilege.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiPrivilegeApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the MultiPrivilege contract.
type MultiPrivilegeApprovalForAllIterator struct {
	Event *MultiPrivilegeApprovalForAll // Event containing the contract specifics and raw log

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
func (it *MultiPrivilegeApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiPrivilegeApprovalForAll)
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
		it.Event = new(MultiPrivilegeApprovalForAll)
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
func (it *MultiPrivilegeApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiPrivilegeApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiPrivilegeApprovalForAll represents a ApprovalForAll event raised by the MultiPrivilege contract.
type MultiPrivilegeApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_MultiPrivilege *MultiPrivilegeFilterer) FilterApprovalForAll(opts *bind.FilterOpts, owner []common.Address, operator []common.Address) (*MultiPrivilegeApprovalForAllIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _MultiPrivilege.contract.FilterLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return &MultiPrivilegeApprovalForAllIterator{contract: _MultiPrivilege.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_MultiPrivilege *MultiPrivilegeFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *MultiPrivilegeApprovalForAll, owner []common.Address, operator []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _MultiPrivilege.contract.WatchLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiPrivilegeApprovalForAll)
				if err := _MultiPrivilege.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
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
func (_MultiPrivilege *MultiPrivilegeFilterer) ParseApprovalForAll(log types.Log) (*MultiPrivilegeApprovalForAll, error) {
	event := new(MultiPrivilegeApprovalForAll)
	if err := _MultiPrivilege.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiPrivilegeBeaconUpgradedIterator is returned from FilterBeaconUpgraded and is used to iterate over the raw logs and unpacked data for BeaconUpgraded events raised by the MultiPrivilege contract.
type MultiPrivilegeBeaconUpgradedIterator struct {
	Event *MultiPrivilegeBeaconUpgraded // Event containing the contract specifics and raw log

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
func (it *MultiPrivilegeBeaconUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiPrivilegeBeaconUpgraded)
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
		it.Event = new(MultiPrivilegeBeaconUpgraded)
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
func (it *MultiPrivilegeBeaconUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiPrivilegeBeaconUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiPrivilegeBeaconUpgraded represents a BeaconUpgraded event raised by the MultiPrivilege contract.
type MultiPrivilegeBeaconUpgraded struct {
	Beacon common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterBeaconUpgraded is a free log retrieval operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_MultiPrivilege *MultiPrivilegeFilterer) FilterBeaconUpgraded(opts *bind.FilterOpts, beacon []common.Address) (*MultiPrivilegeBeaconUpgradedIterator, error) {

	var beaconRule []interface{}
	for _, beaconItem := range beacon {
		beaconRule = append(beaconRule, beaconItem)
	}

	logs, sub, err := _MultiPrivilege.contract.FilterLogs(opts, "BeaconUpgraded", beaconRule)
	if err != nil {
		return nil, err
	}
	return &MultiPrivilegeBeaconUpgradedIterator{contract: _MultiPrivilege.contract, event: "BeaconUpgraded", logs: logs, sub: sub}, nil
}

// WatchBeaconUpgraded is a free log subscription operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_MultiPrivilege *MultiPrivilegeFilterer) WatchBeaconUpgraded(opts *bind.WatchOpts, sink chan<- *MultiPrivilegeBeaconUpgraded, beacon []common.Address) (event.Subscription, error) {

	var beaconRule []interface{}
	for _, beaconItem := range beacon {
		beaconRule = append(beaconRule, beaconItem)
	}

	logs, sub, err := _MultiPrivilege.contract.WatchLogs(opts, "BeaconUpgraded", beaconRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiPrivilegeBeaconUpgraded)
				if err := _MultiPrivilege.contract.UnpackLog(event, "BeaconUpgraded", log); err != nil {
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
func (_MultiPrivilege *MultiPrivilegeFilterer) ParseBeaconUpgraded(log types.Log) (*MultiPrivilegeBeaconUpgraded, error) {
	event := new(MultiPrivilegeBeaconUpgraded)
	if err := _MultiPrivilege.contract.UnpackLog(event, "BeaconUpgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiPrivilegeInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the MultiPrivilege contract.
type MultiPrivilegeInitializedIterator struct {
	Event *MultiPrivilegeInitialized // Event containing the contract specifics and raw log

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
func (it *MultiPrivilegeInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiPrivilegeInitialized)
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
		it.Event = new(MultiPrivilegeInitialized)
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
func (it *MultiPrivilegeInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiPrivilegeInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiPrivilegeInitialized represents a Initialized event raised by the MultiPrivilege contract.
type MultiPrivilegeInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_MultiPrivilege *MultiPrivilegeFilterer) FilterInitialized(opts *bind.FilterOpts) (*MultiPrivilegeInitializedIterator, error) {

	logs, sub, err := _MultiPrivilege.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &MultiPrivilegeInitializedIterator{contract: _MultiPrivilege.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_MultiPrivilege *MultiPrivilegeFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *MultiPrivilegeInitialized) (event.Subscription, error) {

	logs, sub, err := _MultiPrivilege.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiPrivilegeInitialized)
				if err := _MultiPrivilege.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_MultiPrivilege *MultiPrivilegeFilterer) ParseInitialized(log types.Log) (*MultiPrivilegeInitialized, error) {
	event := new(MultiPrivilegeInitialized)
	if err := _MultiPrivilege.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiPrivilegePrivilegeCreatedIterator is returned from FilterPrivilegeCreated and is used to iterate over the raw logs and unpacked data for PrivilegeCreated events raised by the MultiPrivilege contract.
type MultiPrivilegePrivilegeCreatedIterator struct {
	Event *MultiPrivilegePrivilegeCreated // Event containing the contract specifics and raw log

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
func (it *MultiPrivilegePrivilegeCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiPrivilegePrivilegeCreated)
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
		it.Event = new(MultiPrivilegePrivilegeCreated)
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
func (it *MultiPrivilegePrivilegeCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiPrivilegePrivilegeCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiPrivilegePrivilegeCreated represents a PrivilegeCreated event raised by the MultiPrivilege contract.
type MultiPrivilegePrivilegeCreated struct {
	PrivilegeId *big.Int
	Enabled     bool
	Description string
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterPrivilegeCreated is a free log retrieval operation binding the contract event 0x6b1e285f7a3f24ce57e873fa5f9150dfc8a8f24f4feb1e581d42f678fa1b40b3.
//
// Solidity: event PrivilegeCreated(uint256 indexed privilegeId, bool enabled, string description)
func (_MultiPrivilege *MultiPrivilegeFilterer) FilterPrivilegeCreated(opts *bind.FilterOpts, privilegeId []*big.Int) (*MultiPrivilegePrivilegeCreatedIterator, error) {

	var privilegeIdRule []interface{}
	for _, privilegeIdItem := range privilegeId {
		privilegeIdRule = append(privilegeIdRule, privilegeIdItem)
	}

	logs, sub, err := _MultiPrivilege.contract.FilterLogs(opts, "PrivilegeCreated", privilegeIdRule)
	if err != nil {
		return nil, err
	}
	return &MultiPrivilegePrivilegeCreatedIterator{contract: _MultiPrivilege.contract, event: "PrivilegeCreated", logs: logs, sub: sub}, nil
}

// WatchPrivilegeCreated is a free log subscription operation binding the contract event 0x6b1e285f7a3f24ce57e873fa5f9150dfc8a8f24f4feb1e581d42f678fa1b40b3.
//
// Solidity: event PrivilegeCreated(uint256 indexed privilegeId, bool enabled, string description)
func (_MultiPrivilege *MultiPrivilegeFilterer) WatchPrivilegeCreated(opts *bind.WatchOpts, sink chan<- *MultiPrivilegePrivilegeCreated, privilegeId []*big.Int) (event.Subscription, error) {

	var privilegeIdRule []interface{}
	for _, privilegeIdItem := range privilegeId {
		privilegeIdRule = append(privilegeIdRule, privilegeIdItem)
	}

	logs, sub, err := _MultiPrivilege.contract.WatchLogs(opts, "PrivilegeCreated", privilegeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiPrivilegePrivilegeCreated)
				if err := _MultiPrivilege.contract.UnpackLog(event, "PrivilegeCreated", log); err != nil {
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

// ParsePrivilegeCreated is a log parse operation binding the contract event 0x6b1e285f7a3f24ce57e873fa5f9150dfc8a8f24f4feb1e581d42f678fa1b40b3.
//
// Solidity: event PrivilegeCreated(uint256 indexed privilegeId, bool enabled, string description)
func (_MultiPrivilege *MultiPrivilegeFilterer) ParsePrivilegeCreated(log types.Log) (*MultiPrivilegePrivilegeCreated, error) {
	event := new(MultiPrivilegePrivilegeCreated)
	if err := _MultiPrivilege.contract.UnpackLog(event, "PrivilegeCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiPrivilegePrivilegeDisabledIterator is returned from FilterPrivilegeDisabled and is used to iterate over the raw logs and unpacked data for PrivilegeDisabled events raised by the MultiPrivilege contract.
type MultiPrivilegePrivilegeDisabledIterator struct {
	Event *MultiPrivilegePrivilegeDisabled // Event containing the contract specifics and raw log

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
func (it *MultiPrivilegePrivilegeDisabledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiPrivilegePrivilegeDisabled)
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
		it.Event = new(MultiPrivilegePrivilegeDisabled)
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
func (it *MultiPrivilegePrivilegeDisabledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiPrivilegePrivilegeDisabledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiPrivilegePrivilegeDisabled represents a PrivilegeDisabled event raised by the MultiPrivilege contract.
type MultiPrivilegePrivilegeDisabled struct {
	PrivilegeId *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterPrivilegeDisabled is a free log retrieval operation binding the contract event 0xd0a5bf4add1e8e17a93542ca7a6aa2a541ed2da0cd3f76b8b799e7b698f8a633.
//
// Solidity: event PrivilegeDisabled(uint256 indexed privilegeId)
func (_MultiPrivilege *MultiPrivilegeFilterer) FilterPrivilegeDisabled(opts *bind.FilterOpts, privilegeId []*big.Int) (*MultiPrivilegePrivilegeDisabledIterator, error) {

	var privilegeIdRule []interface{}
	for _, privilegeIdItem := range privilegeId {
		privilegeIdRule = append(privilegeIdRule, privilegeIdItem)
	}

	logs, sub, err := _MultiPrivilege.contract.FilterLogs(opts, "PrivilegeDisabled", privilegeIdRule)
	if err != nil {
		return nil, err
	}
	return &MultiPrivilegePrivilegeDisabledIterator{contract: _MultiPrivilege.contract, event: "PrivilegeDisabled", logs: logs, sub: sub}, nil
}

// WatchPrivilegeDisabled is a free log subscription operation binding the contract event 0xd0a5bf4add1e8e17a93542ca7a6aa2a541ed2da0cd3f76b8b799e7b698f8a633.
//
// Solidity: event PrivilegeDisabled(uint256 indexed privilegeId)
func (_MultiPrivilege *MultiPrivilegeFilterer) WatchPrivilegeDisabled(opts *bind.WatchOpts, sink chan<- *MultiPrivilegePrivilegeDisabled, privilegeId []*big.Int) (event.Subscription, error) {

	var privilegeIdRule []interface{}
	for _, privilegeIdItem := range privilegeId {
		privilegeIdRule = append(privilegeIdRule, privilegeIdItem)
	}

	logs, sub, err := _MultiPrivilege.contract.WatchLogs(opts, "PrivilegeDisabled", privilegeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiPrivilegePrivilegeDisabled)
				if err := _MultiPrivilege.contract.UnpackLog(event, "PrivilegeDisabled", log); err != nil {
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

// ParsePrivilegeDisabled is a log parse operation binding the contract event 0xd0a5bf4add1e8e17a93542ca7a6aa2a541ed2da0cd3f76b8b799e7b698f8a633.
//
// Solidity: event PrivilegeDisabled(uint256 indexed privilegeId)
func (_MultiPrivilege *MultiPrivilegeFilterer) ParsePrivilegeDisabled(log types.Log) (*MultiPrivilegePrivilegeDisabled, error) {
	event := new(MultiPrivilegePrivilegeDisabled)
	if err := _MultiPrivilege.contract.UnpackLog(event, "PrivilegeDisabled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiPrivilegePrivilegeEnabledIterator is returned from FilterPrivilegeEnabled and is used to iterate over the raw logs and unpacked data for PrivilegeEnabled events raised by the MultiPrivilege contract.
type MultiPrivilegePrivilegeEnabledIterator struct {
	Event *MultiPrivilegePrivilegeEnabled // Event containing the contract specifics and raw log

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
func (it *MultiPrivilegePrivilegeEnabledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiPrivilegePrivilegeEnabled)
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
		it.Event = new(MultiPrivilegePrivilegeEnabled)
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
func (it *MultiPrivilegePrivilegeEnabledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiPrivilegePrivilegeEnabledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiPrivilegePrivilegeEnabled represents a PrivilegeEnabled event raised by the MultiPrivilege contract.
type MultiPrivilegePrivilegeEnabled struct {
	PrivilegeId *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterPrivilegeEnabled is a free log retrieval operation binding the contract event 0xdb76175b0679b36e2d84207a86ca82a64ed1fd1af72c1a48eb3ff9369bbc4c83.
//
// Solidity: event PrivilegeEnabled(uint256 indexed privilegeId)
func (_MultiPrivilege *MultiPrivilegeFilterer) FilterPrivilegeEnabled(opts *bind.FilterOpts, privilegeId []*big.Int) (*MultiPrivilegePrivilegeEnabledIterator, error) {

	var privilegeIdRule []interface{}
	for _, privilegeIdItem := range privilegeId {
		privilegeIdRule = append(privilegeIdRule, privilegeIdItem)
	}

	logs, sub, err := _MultiPrivilege.contract.FilterLogs(opts, "PrivilegeEnabled", privilegeIdRule)
	if err != nil {
		return nil, err
	}
	return &MultiPrivilegePrivilegeEnabledIterator{contract: _MultiPrivilege.contract, event: "PrivilegeEnabled", logs: logs, sub: sub}, nil
}

// WatchPrivilegeEnabled is a free log subscription operation binding the contract event 0xdb76175b0679b36e2d84207a86ca82a64ed1fd1af72c1a48eb3ff9369bbc4c83.
//
// Solidity: event PrivilegeEnabled(uint256 indexed privilegeId)
func (_MultiPrivilege *MultiPrivilegeFilterer) WatchPrivilegeEnabled(opts *bind.WatchOpts, sink chan<- *MultiPrivilegePrivilegeEnabled, privilegeId []*big.Int) (event.Subscription, error) {

	var privilegeIdRule []interface{}
	for _, privilegeIdItem := range privilegeId {
		privilegeIdRule = append(privilegeIdRule, privilegeIdItem)
	}

	logs, sub, err := _MultiPrivilege.contract.WatchLogs(opts, "PrivilegeEnabled", privilegeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiPrivilegePrivilegeEnabled)
				if err := _MultiPrivilege.contract.UnpackLog(event, "PrivilegeEnabled", log); err != nil {
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

// ParsePrivilegeEnabled is a log parse operation binding the contract event 0xdb76175b0679b36e2d84207a86ca82a64ed1fd1af72c1a48eb3ff9369bbc4c83.
//
// Solidity: event PrivilegeEnabled(uint256 indexed privilegeId)
func (_MultiPrivilege *MultiPrivilegeFilterer) ParsePrivilegeEnabled(log types.Log) (*MultiPrivilegePrivilegeEnabled, error) {
	event := new(MultiPrivilegePrivilegeEnabled)
	if err := _MultiPrivilege.contract.UnpackLog(event, "PrivilegeEnabled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiPrivilegePrivilegeSetIterator is returned from FilterPrivilegeSet and is used to iterate over the raw logs and unpacked data for PrivilegeSet events raised by the MultiPrivilege contract.
type MultiPrivilegePrivilegeSetIterator struct {
	Event *MultiPrivilegePrivilegeSet // Event containing the contract specifics and raw log

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
func (it *MultiPrivilegePrivilegeSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiPrivilegePrivilegeSet)
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
		it.Event = new(MultiPrivilegePrivilegeSet)
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
func (it *MultiPrivilegePrivilegeSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiPrivilegePrivilegeSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiPrivilegePrivilegeSet represents a PrivilegeSet event raised by the MultiPrivilege contract.
type MultiPrivilegePrivilegeSet struct {
	TokenId *big.Int
	Version *big.Int
	PrivId  *big.Int
	User    common.Address
	Expires *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPrivilegeSet is a free log retrieval operation binding the contract event 0x61a24679288162b799d80b2bb2b8b0fcdd5c5f53ac19e9246cc190b60196c359.
//
// Solidity: event PrivilegeSet(uint256 indexed tokenId, uint256 version, uint256 indexed privId, address indexed user, uint256 expires)
func (_MultiPrivilege *MultiPrivilegeFilterer) FilterPrivilegeSet(opts *bind.FilterOpts, tokenId []*big.Int, privId []*big.Int, user []common.Address) (*MultiPrivilegePrivilegeSetIterator, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	var privIdRule []interface{}
	for _, privIdItem := range privId {
		privIdRule = append(privIdRule, privIdItem)
	}
	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _MultiPrivilege.contract.FilterLogs(opts, "PrivilegeSet", tokenIdRule, privIdRule, userRule)
	if err != nil {
		return nil, err
	}
	return &MultiPrivilegePrivilegeSetIterator{contract: _MultiPrivilege.contract, event: "PrivilegeSet", logs: logs, sub: sub}, nil
}

// WatchPrivilegeSet is a free log subscription operation binding the contract event 0x61a24679288162b799d80b2bb2b8b0fcdd5c5f53ac19e9246cc190b60196c359.
//
// Solidity: event PrivilegeSet(uint256 indexed tokenId, uint256 version, uint256 indexed privId, address indexed user, uint256 expires)
func (_MultiPrivilege *MultiPrivilegeFilterer) WatchPrivilegeSet(opts *bind.WatchOpts, sink chan<- *MultiPrivilegePrivilegeSet, tokenId []*big.Int, privId []*big.Int, user []common.Address) (event.Subscription, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	var privIdRule []interface{}
	for _, privIdItem := range privId {
		privIdRule = append(privIdRule, privIdItem)
	}
	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _MultiPrivilege.contract.WatchLogs(opts, "PrivilegeSet", tokenIdRule, privIdRule, userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiPrivilegePrivilegeSet)
				if err := _MultiPrivilege.contract.UnpackLog(event, "PrivilegeSet", log); err != nil {
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

// ParsePrivilegeSet is a log parse operation binding the contract event 0x61a24679288162b799d80b2bb2b8b0fcdd5c5f53ac19e9246cc190b60196c359.
//
// Solidity: event PrivilegeSet(uint256 indexed tokenId, uint256 version, uint256 indexed privId, address indexed user, uint256 expires)
func (_MultiPrivilege *MultiPrivilegeFilterer) ParsePrivilegeSet(log types.Log) (*MultiPrivilegePrivilegeSet, error) {
	event := new(MultiPrivilegePrivilegeSet)
	if err := _MultiPrivilege.contract.UnpackLog(event, "PrivilegeSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiPrivilegeRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the MultiPrivilege contract.
type MultiPrivilegeRoleAdminChangedIterator struct {
	Event *MultiPrivilegeRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *MultiPrivilegeRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiPrivilegeRoleAdminChanged)
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
		it.Event = new(MultiPrivilegeRoleAdminChanged)
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
func (it *MultiPrivilegeRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiPrivilegeRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiPrivilegeRoleAdminChanged represents a RoleAdminChanged event raised by the MultiPrivilege contract.
type MultiPrivilegeRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_MultiPrivilege *MultiPrivilegeFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*MultiPrivilegeRoleAdminChangedIterator, error) {

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

	logs, sub, err := _MultiPrivilege.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &MultiPrivilegeRoleAdminChangedIterator{contract: _MultiPrivilege.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_MultiPrivilege *MultiPrivilegeFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *MultiPrivilegeRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

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

	logs, sub, err := _MultiPrivilege.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiPrivilegeRoleAdminChanged)
				if err := _MultiPrivilege.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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
func (_MultiPrivilege *MultiPrivilegeFilterer) ParseRoleAdminChanged(log types.Log) (*MultiPrivilegeRoleAdminChanged, error) {
	event := new(MultiPrivilegeRoleAdminChanged)
	if err := _MultiPrivilege.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiPrivilegeRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the MultiPrivilege contract.
type MultiPrivilegeRoleGrantedIterator struct {
	Event *MultiPrivilegeRoleGranted // Event containing the contract specifics and raw log

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
func (it *MultiPrivilegeRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiPrivilegeRoleGranted)
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
		it.Event = new(MultiPrivilegeRoleGranted)
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
func (it *MultiPrivilegeRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiPrivilegeRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiPrivilegeRoleGranted represents a RoleGranted event raised by the MultiPrivilege contract.
type MultiPrivilegeRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_MultiPrivilege *MultiPrivilegeFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*MultiPrivilegeRoleGrantedIterator, error) {

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

	logs, sub, err := _MultiPrivilege.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &MultiPrivilegeRoleGrantedIterator{contract: _MultiPrivilege.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_MultiPrivilege *MultiPrivilegeFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *MultiPrivilegeRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _MultiPrivilege.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiPrivilegeRoleGranted)
				if err := _MultiPrivilege.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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
func (_MultiPrivilege *MultiPrivilegeFilterer) ParseRoleGranted(log types.Log) (*MultiPrivilegeRoleGranted, error) {
	event := new(MultiPrivilegeRoleGranted)
	if err := _MultiPrivilege.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiPrivilegeRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the MultiPrivilege contract.
type MultiPrivilegeRoleRevokedIterator struct {
	Event *MultiPrivilegeRoleRevoked // Event containing the contract specifics and raw log

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
func (it *MultiPrivilegeRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiPrivilegeRoleRevoked)
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
		it.Event = new(MultiPrivilegeRoleRevoked)
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
func (it *MultiPrivilegeRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiPrivilegeRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiPrivilegeRoleRevoked represents a RoleRevoked event raised by the MultiPrivilege contract.
type MultiPrivilegeRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_MultiPrivilege *MultiPrivilegeFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*MultiPrivilegeRoleRevokedIterator, error) {

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

	logs, sub, err := _MultiPrivilege.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &MultiPrivilegeRoleRevokedIterator{contract: _MultiPrivilege.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_MultiPrivilege *MultiPrivilegeFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *MultiPrivilegeRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _MultiPrivilege.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiPrivilegeRoleRevoked)
				if err := _MultiPrivilege.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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
func (_MultiPrivilege *MultiPrivilegeFilterer) ParseRoleRevoked(log types.Log) (*MultiPrivilegeRoleRevoked, error) {
	event := new(MultiPrivilegeRoleRevoked)
	if err := _MultiPrivilege.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiPrivilegeTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the MultiPrivilege contract.
type MultiPrivilegeTransferIterator struct {
	Event *MultiPrivilegeTransfer // Event containing the contract specifics and raw log

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
func (it *MultiPrivilegeTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiPrivilegeTransfer)
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
		it.Event = new(MultiPrivilegeTransfer)
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
func (it *MultiPrivilegeTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiPrivilegeTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiPrivilegeTransfer represents a Transfer event raised by the MultiPrivilege contract.
type MultiPrivilegeTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_MultiPrivilege *MultiPrivilegeFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address, tokenId []*big.Int) (*MultiPrivilegeTransferIterator, error) {

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

	logs, sub, err := _MultiPrivilege.contract.FilterLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &MultiPrivilegeTransferIterator{contract: _MultiPrivilege.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_MultiPrivilege *MultiPrivilegeFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *MultiPrivilegeTransfer, from []common.Address, to []common.Address, tokenId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _MultiPrivilege.contract.WatchLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiPrivilegeTransfer)
				if err := _MultiPrivilege.contract.UnpackLog(event, "Transfer", log); err != nil {
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
func (_MultiPrivilege *MultiPrivilegeFilterer) ParseTransfer(log types.Log) (*MultiPrivilegeTransfer, error) {
	event := new(MultiPrivilegeTransfer)
	if err := _MultiPrivilege.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MultiPrivilegeUpgradedIterator is returned from FilterUpgraded and is used to iterate over the raw logs and unpacked data for Upgraded events raised by the MultiPrivilege contract.
type MultiPrivilegeUpgradedIterator struct {
	Event *MultiPrivilegeUpgraded // Event containing the contract specifics and raw log

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
func (it *MultiPrivilegeUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiPrivilegeUpgraded)
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
		it.Event = new(MultiPrivilegeUpgraded)
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
func (it *MultiPrivilegeUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiPrivilegeUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiPrivilegeUpgraded represents a Upgraded event raised by the MultiPrivilege contract.
type MultiPrivilegeUpgraded struct {
	Implementation common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpgraded is a free log retrieval operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_MultiPrivilege *MultiPrivilegeFilterer) FilterUpgraded(opts *bind.FilterOpts, implementation []common.Address) (*MultiPrivilegeUpgradedIterator, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _MultiPrivilege.contract.FilterLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return &MultiPrivilegeUpgradedIterator{contract: _MultiPrivilege.contract, event: "Upgraded", logs: logs, sub: sub}, nil
}

// WatchUpgraded is a free log subscription operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_MultiPrivilege *MultiPrivilegeFilterer) WatchUpgraded(opts *bind.WatchOpts, sink chan<- *MultiPrivilegeUpgraded, implementation []common.Address) (event.Subscription, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _MultiPrivilege.contract.WatchLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiPrivilegeUpgraded)
				if err := _MultiPrivilege.contract.UnpackLog(event, "Upgraded", log); err != nil {
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
func (_MultiPrivilege *MultiPrivilegeFilterer) ParseUpgraded(log types.Log) (*MultiPrivilegeUpgraded, error) {
	event := new(MultiPrivilegeUpgraded)
	if err := _MultiPrivilege.contract.UnpackLog(event, "Upgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
