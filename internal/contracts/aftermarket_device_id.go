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

// AftermarketDeviceIdMetaData contains all meta data concerning the AftermarketDeviceId contract.
var AftermarketDeviceIdMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"previousAdmin\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"AdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"approved\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"beacon\",\"type\":\"address\"}],\"name\":\"BeaconUpgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"privilegeId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"enabled\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"description\",\"type\":\"string\"}],\"name\":\"PrivilegeCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"privilegeId\",\"type\":\"uint256\"}],\"name\":\"PrivilegeDisabled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"privilegeId\",\"type\":\"uint256\"}],\"name\":\"PrivilegeEnabled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"privId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"expires\",\"type\":\"uint256\"}],\"name\":\"PrivilegeSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"BURNER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DEFAULT_ADMIN_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MINTER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"TRANSFERER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"UPGRADER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"enabled\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"description\",\"type\":\"string\"}],\"name\":\"createPrivilege\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"privId\",\"type\":\"uint256\"}],\"name\":\"disablePrivilege\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"privId\",\"type\":\"uint256\"}],\"name\":\"enablePrivilege\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"exists\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getApproved\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"privId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"}],\"name\":\"hasPrivilege\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name_\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol_\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"baseUri_\",\"type\":\"string\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"privilegeEntry\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"privId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"}],\"name\":\"privilegeExpiresAt\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"privilegeRecord\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"enabled\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"description\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxiableUUID\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"safeMint\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"uri\",\"type\":\"string\"}],\"name\":\"safeMint\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"baseURI_\",\"type\":\"string\"}],\"name\":\"setBaseURI\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"setDimoRegistryAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"privId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"expires\",\"type\":\"uint256\"}],\"name\":\"setPrivilege\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"privId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"expires\",\"type\":\"uint256\"}],\"internalType\":\"structMultiPrivilege.SetPrivilegeData[]\",\"name\":\"privData\",\"type\":\"tuple[]\"}],\"name\":\"setPrivileges\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"tokenIdToVersion\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"tokenURI\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"}],\"name\":\"upgradeTo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"upgradeToAndCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
}

// AftermarketDeviceIdABI is the input ABI used to generate the binding from.
// Deprecated: Use AftermarketDeviceIdMetaData.ABI instead.
var AftermarketDeviceIdABI = AftermarketDeviceIdMetaData.ABI

// AftermarketDeviceId is an auto generated Go binding around an Ethereum contract.
type AftermarketDeviceId struct {
	AftermarketDeviceIdCaller     // Read-only binding to the contract
	AftermarketDeviceIdTransactor // Write-only binding to the contract
	AftermarketDeviceIdFilterer   // Log filterer for contract events
}

// AftermarketDeviceIdCaller is an auto generated read-only Go binding around an Ethereum contract.
type AftermarketDeviceIdCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AftermarketDeviceIdTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AftermarketDeviceIdTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AftermarketDeviceIdFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AftermarketDeviceIdFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AftermarketDeviceIdSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AftermarketDeviceIdSession struct {
	Contract     *AftermarketDeviceId // Generic contract binding to set the session for
	CallOpts     bind.CallOpts        // Call options to use throughout this session
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// AftermarketDeviceIdCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AftermarketDeviceIdCallerSession struct {
	Contract *AftermarketDeviceIdCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts              // Call options to use throughout this session
}

// AftermarketDeviceIdTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AftermarketDeviceIdTransactorSession struct {
	Contract     *AftermarketDeviceIdTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts              // Transaction auth options to use throughout this session
}

// AftermarketDeviceIdRaw is an auto generated low-level Go binding around an Ethereum contract.
type AftermarketDeviceIdRaw struct {
	Contract *AftermarketDeviceId // Generic contract binding to access the raw methods on
}

// AftermarketDeviceIdCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AftermarketDeviceIdCallerRaw struct {
	Contract *AftermarketDeviceIdCaller // Generic read-only contract binding to access the raw methods on
}

// AftermarketDeviceIdTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AftermarketDeviceIdTransactorRaw struct {
	Contract *AftermarketDeviceIdTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAftermarketDeviceId creates a new instance of AftermarketDeviceId, bound to a specific deployed contract.
func NewAftermarketDeviceId(address common.Address, backend bind.ContractBackend) (*AftermarketDeviceId, error) {
	contract, err := bindAftermarketDeviceId(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AftermarketDeviceId{AftermarketDeviceIdCaller: AftermarketDeviceIdCaller{contract: contract}, AftermarketDeviceIdTransactor: AftermarketDeviceIdTransactor{contract: contract}, AftermarketDeviceIdFilterer: AftermarketDeviceIdFilterer{contract: contract}}, nil
}

// NewAftermarketDeviceIdCaller creates a new read-only instance of AftermarketDeviceId, bound to a specific deployed contract.
func NewAftermarketDeviceIdCaller(address common.Address, caller bind.ContractCaller) (*AftermarketDeviceIdCaller, error) {
	contract, err := bindAftermarketDeviceId(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AftermarketDeviceIdCaller{contract: contract}, nil
}

// NewAftermarketDeviceIdTransactor creates a new write-only instance of AftermarketDeviceId, bound to a specific deployed contract.
func NewAftermarketDeviceIdTransactor(address common.Address, transactor bind.ContractTransactor) (*AftermarketDeviceIdTransactor, error) {
	contract, err := bindAftermarketDeviceId(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AftermarketDeviceIdTransactor{contract: contract}, nil
}

// NewAftermarketDeviceIdFilterer creates a new log filterer instance of AftermarketDeviceId, bound to a specific deployed contract.
func NewAftermarketDeviceIdFilterer(address common.Address, filterer bind.ContractFilterer) (*AftermarketDeviceIdFilterer, error) {
	contract, err := bindAftermarketDeviceId(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AftermarketDeviceIdFilterer{contract: contract}, nil
}

// bindAftermarketDeviceId binds a generic wrapper to an already deployed contract.
func bindAftermarketDeviceId(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := AftermarketDeviceIdMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AftermarketDeviceId *AftermarketDeviceIdRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AftermarketDeviceId.Contract.AftermarketDeviceIdCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AftermarketDeviceId *AftermarketDeviceIdRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.AftermarketDeviceIdTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AftermarketDeviceId *AftermarketDeviceIdRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.AftermarketDeviceIdTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AftermarketDeviceId *AftermarketDeviceIdCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AftermarketDeviceId.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.contract.Transact(opts, method, params...)
}

// BURNERROLE is a free data retrieval call binding the contract method 0x282c51f3.
//
// Solidity: function BURNER_ROLE() view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) BURNERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "BURNER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BURNERROLE is a free data retrieval call binding the contract method 0x282c51f3.
//
// Solidity: function BURNER_ROLE() view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) BURNERROLE() ([32]byte, error) {
	return _AftermarketDeviceId.Contract.BURNERROLE(&_AftermarketDeviceId.CallOpts)
}

// BURNERROLE is a free data retrieval call binding the contract method 0x282c51f3.
//
// Solidity: function BURNER_ROLE() view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) BURNERROLE() ([32]byte, error) {
	return _AftermarketDeviceId.Contract.BURNERROLE(&_AftermarketDeviceId.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _AftermarketDeviceId.Contract.DEFAULTADMINROLE(&_AftermarketDeviceId.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _AftermarketDeviceId.Contract.DEFAULTADMINROLE(&_AftermarketDeviceId.CallOpts)
}

// MINTERROLE is a free data retrieval call binding the contract method 0xd5391393.
//
// Solidity: function MINTER_ROLE() view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) MINTERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "MINTER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MINTERROLE is a free data retrieval call binding the contract method 0xd5391393.
//
// Solidity: function MINTER_ROLE() view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) MINTERROLE() ([32]byte, error) {
	return _AftermarketDeviceId.Contract.MINTERROLE(&_AftermarketDeviceId.CallOpts)
}

// MINTERROLE is a free data retrieval call binding the contract method 0xd5391393.
//
// Solidity: function MINTER_ROLE() view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) MINTERROLE() ([32]byte, error) {
	return _AftermarketDeviceId.Contract.MINTERROLE(&_AftermarketDeviceId.CallOpts)
}

// TRANSFERERROLE is a free data retrieval call binding the contract method 0x0ade7dc1.
//
// Solidity: function TRANSFERER_ROLE() view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) TRANSFERERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "TRANSFERER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// TRANSFERERROLE is a free data retrieval call binding the contract method 0x0ade7dc1.
//
// Solidity: function TRANSFERER_ROLE() view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) TRANSFERERROLE() ([32]byte, error) {
	return _AftermarketDeviceId.Contract.TRANSFERERROLE(&_AftermarketDeviceId.CallOpts)
}

// TRANSFERERROLE is a free data retrieval call binding the contract method 0x0ade7dc1.
//
// Solidity: function TRANSFERER_ROLE() view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) TRANSFERERROLE() ([32]byte, error) {
	return _AftermarketDeviceId.Contract.TRANSFERERROLE(&_AftermarketDeviceId.CallOpts)
}

// UPGRADERROLE is a free data retrieval call binding the contract method 0xf72c0d8b.
//
// Solidity: function UPGRADER_ROLE() view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) UPGRADERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "UPGRADER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// UPGRADERROLE is a free data retrieval call binding the contract method 0xf72c0d8b.
//
// Solidity: function UPGRADER_ROLE() view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) UPGRADERROLE() ([32]byte, error) {
	return _AftermarketDeviceId.Contract.UPGRADERROLE(&_AftermarketDeviceId.CallOpts)
}

// UPGRADERROLE is a free data retrieval call binding the contract method 0xf72c0d8b.
//
// Solidity: function UPGRADER_ROLE() view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) UPGRADERROLE() ([32]byte, error) {
	return _AftermarketDeviceId.Contract.UPGRADERROLE(&_AftermarketDeviceId.CallOpts)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) BalanceOf(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "balanceOf", owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _AftermarketDeviceId.Contract.BalanceOf(&_AftermarketDeviceId.CallOpts, owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _AftermarketDeviceId.Contract.BalanceOf(&_AftermarketDeviceId.CallOpts, owner)
}

// Exists is a free data retrieval call binding the contract method 0x4f558e79.
//
// Solidity: function exists(uint256 tokenId) view returns(bool)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) Exists(opts *bind.CallOpts, tokenId *big.Int) (bool, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "exists", tokenId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Exists is a free data retrieval call binding the contract method 0x4f558e79.
//
// Solidity: function exists(uint256 tokenId) view returns(bool)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) Exists(tokenId *big.Int) (bool, error) {
	return _AftermarketDeviceId.Contract.Exists(&_AftermarketDeviceId.CallOpts, tokenId)
}

// Exists is a free data retrieval call binding the contract method 0x4f558e79.
//
// Solidity: function exists(uint256 tokenId) view returns(bool)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) Exists(tokenId *big.Int) (bool, error) {
	return _AftermarketDeviceId.Contract.Exists(&_AftermarketDeviceId.CallOpts, tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) GetApproved(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "getApproved", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _AftermarketDeviceId.Contract.GetApproved(&_AftermarketDeviceId.CallOpts, tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _AftermarketDeviceId.Contract.GetApproved(&_AftermarketDeviceId.CallOpts, tokenId)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _AftermarketDeviceId.Contract.GetRoleAdmin(&_AftermarketDeviceId.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _AftermarketDeviceId.Contract.GetRoleAdmin(&_AftermarketDeviceId.CallOpts, role)
}

// HasPrivilege is a free data retrieval call binding the contract method 0x05d80b00.
//
// Solidity: function hasPrivilege(uint256 tokenId, uint256 privId, address user) view returns(bool)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) HasPrivilege(opts *bind.CallOpts, tokenId *big.Int, privId *big.Int, user common.Address) (bool, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "hasPrivilege", tokenId, privId, user)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasPrivilege is a free data retrieval call binding the contract method 0x05d80b00.
//
// Solidity: function hasPrivilege(uint256 tokenId, uint256 privId, address user) view returns(bool)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) HasPrivilege(tokenId *big.Int, privId *big.Int, user common.Address) (bool, error) {
	return _AftermarketDeviceId.Contract.HasPrivilege(&_AftermarketDeviceId.CallOpts, tokenId, privId, user)
}

// HasPrivilege is a free data retrieval call binding the contract method 0x05d80b00.
//
// Solidity: function hasPrivilege(uint256 tokenId, uint256 privId, address user) view returns(bool)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) HasPrivilege(tokenId *big.Int, privId *big.Int, user common.Address) (bool, error) {
	return _AftermarketDeviceId.Contract.HasPrivilege(&_AftermarketDeviceId.CallOpts, tokenId, privId, user)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _AftermarketDeviceId.Contract.HasRole(&_AftermarketDeviceId.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _AftermarketDeviceId.Contract.HasRole(&_AftermarketDeviceId.CallOpts, role, account)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) IsApprovedForAll(opts *bind.CallOpts, owner common.Address, operator common.Address) (bool, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "isApprovedForAll", owner, operator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _AftermarketDeviceId.Contract.IsApprovedForAll(&_AftermarketDeviceId.CallOpts, owner, operator)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _AftermarketDeviceId.Contract.IsApprovedForAll(&_AftermarketDeviceId.CallOpts, owner, operator)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) Name() (string, error) {
	return _AftermarketDeviceId.Contract.Name(&_AftermarketDeviceId.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) Name() (string, error) {
	return _AftermarketDeviceId.Contract.Name(&_AftermarketDeviceId.CallOpts)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) OwnerOf(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "ownerOf", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _AftermarketDeviceId.Contract.OwnerOf(&_AftermarketDeviceId.CallOpts, tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _AftermarketDeviceId.Contract.OwnerOf(&_AftermarketDeviceId.CallOpts, tokenId)
}

// PrivilegeEntry is a free data retrieval call binding the contract method 0x48db4640.
//
// Solidity: function privilegeEntry(uint256 , uint256 , uint256 , address ) view returns(uint256)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) PrivilegeEntry(opts *bind.CallOpts, arg0 *big.Int, arg1 *big.Int, arg2 *big.Int, arg3 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "privilegeEntry", arg0, arg1, arg2, arg3)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PrivilegeEntry is a free data retrieval call binding the contract method 0x48db4640.
//
// Solidity: function privilegeEntry(uint256 , uint256 , uint256 , address ) view returns(uint256)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) PrivilegeEntry(arg0 *big.Int, arg1 *big.Int, arg2 *big.Int, arg3 common.Address) (*big.Int, error) {
	return _AftermarketDeviceId.Contract.PrivilegeEntry(&_AftermarketDeviceId.CallOpts, arg0, arg1, arg2, arg3)
}

// PrivilegeEntry is a free data retrieval call binding the contract method 0x48db4640.
//
// Solidity: function privilegeEntry(uint256 , uint256 , uint256 , address ) view returns(uint256)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) PrivilegeEntry(arg0 *big.Int, arg1 *big.Int, arg2 *big.Int, arg3 common.Address) (*big.Int, error) {
	return _AftermarketDeviceId.Contract.PrivilegeEntry(&_AftermarketDeviceId.CallOpts, arg0, arg1, arg2, arg3)
}

// PrivilegeExpiresAt is a free data retrieval call binding the contract method 0xd0f8f5f6.
//
// Solidity: function privilegeExpiresAt(uint256 tokenId, uint256 privId, address user) view returns(uint256)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) PrivilegeExpiresAt(opts *bind.CallOpts, tokenId *big.Int, privId *big.Int, user common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "privilegeExpiresAt", tokenId, privId, user)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PrivilegeExpiresAt is a free data retrieval call binding the contract method 0xd0f8f5f6.
//
// Solidity: function privilegeExpiresAt(uint256 tokenId, uint256 privId, address user) view returns(uint256)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) PrivilegeExpiresAt(tokenId *big.Int, privId *big.Int, user common.Address) (*big.Int, error) {
	return _AftermarketDeviceId.Contract.PrivilegeExpiresAt(&_AftermarketDeviceId.CallOpts, tokenId, privId, user)
}

// PrivilegeExpiresAt is a free data retrieval call binding the contract method 0xd0f8f5f6.
//
// Solidity: function privilegeExpiresAt(uint256 tokenId, uint256 privId, address user) view returns(uint256)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) PrivilegeExpiresAt(tokenId *big.Int, privId *big.Int, user common.Address) (*big.Int, error) {
	return _AftermarketDeviceId.Contract.PrivilegeExpiresAt(&_AftermarketDeviceId.CallOpts, tokenId, privId, user)
}

// PrivilegeRecord is a free data retrieval call binding the contract method 0xf9ad3efe.
//
// Solidity: function privilegeRecord(uint256 ) view returns(bool enabled, string description)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) PrivilegeRecord(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Enabled     bool
	Description string
}, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "privilegeRecord", arg0)

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
func (_AftermarketDeviceId *AftermarketDeviceIdSession) PrivilegeRecord(arg0 *big.Int) (struct {
	Enabled     bool
	Description string
}, error) {
	return _AftermarketDeviceId.Contract.PrivilegeRecord(&_AftermarketDeviceId.CallOpts, arg0)
}

// PrivilegeRecord is a free data retrieval call binding the contract method 0xf9ad3efe.
//
// Solidity: function privilegeRecord(uint256 ) view returns(bool enabled, string description)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) PrivilegeRecord(arg0 *big.Int) (struct {
	Enabled     bool
	Description string
}, error) {
	return _AftermarketDeviceId.Contract.PrivilegeRecord(&_AftermarketDeviceId.CallOpts, arg0)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) ProxiableUUID(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "proxiableUUID")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) ProxiableUUID() ([32]byte, error) {
	return _AftermarketDeviceId.Contract.ProxiableUUID(&_AftermarketDeviceId.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) ProxiableUUID() ([32]byte, error) {
	return _AftermarketDeviceId.Contract.ProxiableUUID(&_AftermarketDeviceId.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _AftermarketDeviceId.Contract.SupportsInterface(&_AftermarketDeviceId.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _AftermarketDeviceId.Contract.SupportsInterface(&_AftermarketDeviceId.CallOpts, interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) Symbol() (string, error) {
	return _AftermarketDeviceId.Contract.Symbol(&_AftermarketDeviceId.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) Symbol() (string, error) {
	return _AftermarketDeviceId.Contract.Symbol(&_AftermarketDeviceId.CallOpts)
}

// TokenIdToVersion is a free data retrieval call binding the contract method 0xf1a9d41c.
//
// Solidity: function tokenIdToVersion(uint256 ) view returns(uint256)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) TokenIdToVersion(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "tokenIdToVersion", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TokenIdToVersion is a free data retrieval call binding the contract method 0xf1a9d41c.
//
// Solidity: function tokenIdToVersion(uint256 ) view returns(uint256)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) TokenIdToVersion(arg0 *big.Int) (*big.Int, error) {
	return _AftermarketDeviceId.Contract.TokenIdToVersion(&_AftermarketDeviceId.CallOpts, arg0)
}

// TokenIdToVersion is a free data retrieval call binding the contract method 0xf1a9d41c.
//
// Solidity: function tokenIdToVersion(uint256 ) view returns(uint256)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) TokenIdToVersion(arg0 *big.Int) (*big.Int, error) {
	return _AftermarketDeviceId.Contract.TokenIdToVersion(&_AftermarketDeviceId.CallOpts, arg0)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_AftermarketDeviceId *AftermarketDeviceIdCaller) TokenURI(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var out []interface{}
	err := _AftermarketDeviceId.contract.Call(opts, &out, "tokenURI", tokenId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) TokenURI(tokenId *big.Int) (string, error) {
	return _AftermarketDeviceId.Contract.TokenURI(&_AftermarketDeviceId.CallOpts, tokenId)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_AftermarketDeviceId *AftermarketDeviceIdCallerSession) TokenURI(tokenId *big.Int) (string, error) {
	return _AftermarketDeviceId.Contract.TokenURI(&_AftermarketDeviceId.CallOpts, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) Approve(opts *bind.TransactOpts, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "approve", to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.Approve(&_AftermarketDeviceId.TransactOpts, to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.Approve(&_AftermarketDeviceId.TransactOpts, to, tokenId)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 tokenId) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) Burn(opts *bind.TransactOpts, tokenId *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "burn", tokenId)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 tokenId) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdSession) Burn(tokenId *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.Burn(&_AftermarketDeviceId.TransactOpts, tokenId)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 tokenId) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) Burn(tokenId *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.Burn(&_AftermarketDeviceId.TransactOpts, tokenId)
}

// CreatePrivilege is a paid mutator transaction binding the contract method 0xc1d58b3b.
//
// Solidity: function createPrivilege(bool enabled, string description) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) CreatePrivilege(opts *bind.TransactOpts, enabled bool, description string) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "createPrivilege", enabled, description)
}

// CreatePrivilege is a paid mutator transaction binding the contract method 0xc1d58b3b.
//
// Solidity: function createPrivilege(bool enabled, string description) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdSession) CreatePrivilege(enabled bool, description string) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.CreatePrivilege(&_AftermarketDeviceId.TransactOpts, enabled, description)
}

// CreatePrivilege is a paid mutator transaction binding the contract method 0xc1d58b3b.
//
// Solidity: function createPrivilege(bool enabled, string description) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) CreatePrivilege(enabled bool, description string) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.CreatePrivilege(&_AftermarketDeviceId.TransactOpts, enabled, description)
}

// DisablePrivilege is a paid mutator transaction binding the contract method 0x1a153ed0.
//
// Solidity: function disablePrivilege(uint256 privId) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) DisablePrivilege(opts *bind.TransactOpts, privId *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "disablePrivilege", privId)
}

// DisablePrivilege is a paid mutator transaction binding the contract method 0x1a153ed0.
//
// Solidity: function disablePrivilege(uint256 privId) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdSession) DisablePrivilege(privId *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.DisablePrivilege(&_AftermarketDeviceId.TransactOpts, privId)
}

// DisablePrivilege is a paid mutator transaction binding the contract method 0x1a153ed0.
//
// Solidity: function disablePrivilege(uint256 privId) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) DisablePrivilege(privId *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.DisablePrivilege(&_AftermarketDeviceId.TransactOpts, privId)
}

// EnablePrivilege is a paid mutator transaction binding the contract method 0x831ba696.
//
// Solidity: function enablePrivilege(uint256 privId) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) EnablePrivilege(opts *bind.TransactOpts, privId *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "enablePrivilege", privId)
}

// EnablePrivilege is a paid mutator transaction binding the contract method 0x831ba696.
//
// Solidity: function enablePrivilege(uint256 privId) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdSession) EnablePrivilege(privId *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.EnablePrivilege(&_AftermarketDeviceId.TransactOpts, privId)
}

// EnablePrivilege is a paid mutator transaction binding the contract method 0x831ba696.
//
// Solidity: function enablePrivilege(uint256 privId) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) EnablePrivilege(privId *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.EnablePrivilege(&_AftermarketDeviceId.TransactOpts, privId)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.GrantRole(&_AftermarketDeviceId.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.GrantRole(&_AftermarketDeviceId.TransactOpts, role, account)
}

// Initialize is a paid mutator transaction binding the contract method 0xa6487c53.
//
// Solidity: function initialize(string name_, string symbol_, string baseUri_) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) Initialize(opts *bind.TransactOpts, name_ string, symbol_ string, baseUri_ string) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "initialize", name_, symbol_, baseUri_)
}

// Initialize is a paid mutator transaction binding the contract method 0xa6487c53.
//
// Solidity: function initialize(string name_, string symbol_, string baseUri_) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdSession) Initialize(name_ string, symbol_ string, baseUri_ string) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.Initialize(&_AftermarketDeviceId.TransactOpts, name_, symbol_, baseUri_)
}

// Initialize is a paid mutator transaction binding the contract method 0xa6487c53.
//
// Solidity: function initialize(string name_, string symbol_, string baseUri_) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) Initialize(name_ string, symbol_ string, baseUri_ string) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.Initialize(&_AftermarketDeviceId.TransactOpts, name_, symbol_, baseUri_)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "renounceRole", role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.RenounceRole(&_AftermarketDeviceId.TransactOpts, role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.RenounceRole(&_AftermarketDeviceId.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.RevokeRole(&_AftermarketDeviceId.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.RevokeRole(&_AftermarketDeviceId.TransactOpts, role, account)
}

// SafeMint is a paid mutator transaction binding the contract method 0x40d097c3.
//
// Solidity: function safeMint(address to) returns(uint256 tokenId)
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) SafeMint(opts *bind.TransactOpts, to common.Address) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "safeMint", to)
}

// SafeMint is a paid mutator transaction binding the contract method 0x40d097c3.
//
// Solidity: function safeMint(address to) returns(uint256 tokenId)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) SafeMint(to common.Address) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.SafeMint(&_AftermarketDeviceId.TransactOpts, to)
}

// SafeMint is a paid mutator transaction binding the contract method 0x40d097c3.
//
// Solidity: function safeMint(address to) returns(uint256 tokenId)
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) SafeMint(to common.Address) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.SafeMint(&_AftermarketDeviceId.TransactOpts, to)
}

// SafeMint0 is a paid mutator transaction binding the contract method 0xd204c45e.
//
// Solidity: function safeMint(address to, string uri) returns(uint256 tokenId)
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) SafeMint0(opts *bind.TransactOpts, to common.Address, uri string) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "safeMint0", to, uri)
}

// SafeMint0 is a paid mutator transaction binding the contract method 0xd204c45e.
//
// Solidity: function safeMint(address to, string uri) returns(uint256 tokenId)
func (_AftermarketDeviceId *AftermarketDeviceIdSession) SafeMint0(to common.Address, uri string) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.SafeMint0(&_AftermarketDeviceId.TransactOpts, to, uri)
}

// SafeMint0 is a paid mutator transaction binding the contract method 0xd204c45e.
//
// Solidity: function safeMint(address to, string uri) returns(uint256 tokenId)
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) SafeMint0(to common.Address, uri string) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.SafeMint0(&_AftermarketDeviceId.TransactOpts, to, uri)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) SafeTransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "safeTransferFrom", from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.SafeTransferFrom(&_AftermarketDeviceId.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.SafeTransferFrom(&_AftermarketDeviceId.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) SafeTransferFrom0(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "safeTransferFrom0", from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.SafeTransferFrom0(&_AftermarketDeviceId.TransactOpts, from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.SafeTransferFrom0(&_AftermarketDeviceId.TransactOpts, from, to, tokenId, data)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) SetApprovalForAll(opts *bind.TransactOpts, operator common.Address, approved bool) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "setApprovalForAll", operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.SetApprovalForAll(&_AftermarketDeviceId.TransactOpts, operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.SetApprovalForAll(&_AftermarketDeviceId.TransactOpts, operator, approved)
}

// SetBaseURI is a paid mutator transaction binding the contract method 0x55f804b3.
//
// Solidity: function setBaseURI(string baseURI_) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) SetBaseURI(opts *bind.TransactOpts, baseURI_ string) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "setBaseURI", baseURI_)
}

// SetBaseURI is a paid mutator transaction binding the contract method 0x55f804b3.
//
// Solidity: function setBaseURI(string baseURI_) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdSession) SetBaseURI(baseURI_ string) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.SetBaseURI(&_AftermarketDeviceId.TransactOpts, baseURI_)
}

// SetBaseURI is a paid mutator transaction binding the contract method 0x55f804b3.
//
// Solidity: function setBaseURI(string baseURI_) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) SetBaseURI(baseURI_ string) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.SetBaseURI(&_AftermarketDeviceId.TransactOpts, baseURI_)
}

// SetDimoRegistryAddress is a paid mutator transaction binding the contract method 0x0db857ea.
//
// Solidity: function setDimoRegistryAddress(address addr) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) SetDimoRegistryAddress(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "setDimoRegistryAddress", addr)
}

// SetDimoRegistryAddress is a paid mutator transaction binding the contract method 0x0db857ea.
//
// Solidity: function setDimoRegistryAddress(address addr) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdSession) SetDimoRegistryAddress(addr common.Address) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.SetDimoRegistryAddress(&_AftermarketDeviceId.TransactOpts, addr)
}

// SetDimoRegistryAddress is a paid mutator transaction binding the contract method 0x0db857ea.
//
// Solidity: function setDimoRegistryAddress(address addr) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) SetDimoRegistryAddress(addr common.Address) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.SetDimoRegistryAddress(&_AftermarketDeviceId.TransactOpts, addr)
}

// SetPrivilege is a paid mutator transaction binding the contract method 0xeca3221a.
//
// Solidity: function setPrivilege(uint256 tokenId, uint256 privId, address user, uint256 expires) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) SetPrivilege(opts *bind.TransactOpts, tokenId *big.Int, privId *big.Int, user common.Address, expires *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "setPrivilege", tokenId, privId, user, expires)
}

// SetPrivilege is a paid mutator transaction binding the contract method 0xeca3221a.
//
// Solidity: function setPrivilege(uint256 tokenId, uint256 privId, address user, uint256 expires) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdSession) SetPrivilege(tokenId *big.Int, privId *big.Int, user common.Address, expires *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.SetPrivilege(&_AftermarketDeviceId.TransactOpts, tokenId, privId, user, expires)
}

// SetPrivilege is a paid mutator transaction binding the contract method 0xeca3221a.
//
// Solidity: function setPrivilege(uint256 tokenId, uint256 privId, address user, uint256 expires) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) SetPrivilege(tokenId *big.Int, privId *big.Int, user common.Address, expires *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.SetPrivilege(&_AftermarketDeviceId.TransactOpts, tokenId, privId, user, expires)
}

// SetPrivileges is a paid mutator transaction binding the contract method 0x57ae9754.
//
// Solidity: function setPrivileges((uint256,uint256,address,uint256)[] privData) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) SetPrivileges(opts *bind.TransactOpts, privData []MultiPrivilegeSetPrivilegeData) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "setPrivileges", privData)
}

// SetPrivileges is a paid mutator transaction binding the contract method 0x57ae9754.
//
// Solidity: function setPrivileges((uint256,uint256,address,uint256)[] privData) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdSession) SetPrivileges(privData []MultiPrivilegeSetPrivilegeData) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.SetPrivileges(&_AftermarketDeviceId.TransactOpts, privData)
}

// SetPrivileges is a paid mutator transaction binding the contract method 0x57ae9754.
//
// Solidity: function setPrivileges((uint256,uint256,address,uint256)[] privData) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) SetPrivileges(privData []MultiPrivilegeSetPrivilegeData) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.SetPrivileges(&_AftermarketDeviceId.TransactOpts, privData)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "transferFrom", from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.TransferFrom(&_AftermarketDeviceId.TransactOpts, from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.TransferFrom(&_AftermarketDeviceId.TransactOpts, from, to, tokenId)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) UpgradeTo(opts *bind.TransactOpts, newImplementation common.Address) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "upgradeTo", newImplementation)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdSession) UpgradeTo(newImplementation common.Address) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.UpgradeTo(&_AftermarketDeviceId.TransactOpts, newImplementation)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) UpgradeTo(newImplementation common.Address) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.UpgradeTo(&_AftermarketDeviceId.TransactOpts, newImplementation)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactor) UpgradeToAndCall(opts *bind.TransactOpts, newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _AftermarketDeviceId.contract.Transact(opts, "upgradeToAndCall", newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_AftermarketDeviceId *AftermarketDeviceIdSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.UpgradeToAndCall(&_AftermarketDeviceId.TransactOpts, newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_AftermarketDeviceId *AftermarketDeviceIdTransactorSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _AftermarketDeviceId.Contract.UpgradeToAndCall(&_AftermarketDeviceId.TransactOpts, newImplementation, data)
}

// AftermarketDeviceIdAdminChangedIterator is returned from FilterAdminChanged and is used to iterate over the raw logs and unpacked data for AdminChanged events raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdAdminChangedIterator struct {
	Event *AftermarketDeviceIdAdminChanged // Event containing the contract specifics and raw log

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
func (it *AftermarketDeviceIdAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AftermarketDeviceIdAdminChanged)
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
		it.Event = new(AftermarketDeviceIdAdminChanged)
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
func (it *AftermarketDeviceIdAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AftermarketDeviceIdAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AftermarketDeviceIdAdminChanged represents a AdminChanged event raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdAdminChanged struct {
	PreviousAdmin common.Address
	NewAdmin      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterAdminChanged is a free log retrieval operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) FilterAdminChanged(opts *bind.FilterOpts) (*AftermarketDeviceIdAdminChangedIterator, error) {

	logs, sub, err := _AftermarketDeviceId.contract.FilterLogs(opts, "AdminChanged")
	if err != nil {
		return nil, err
	}
	return &AftermarketDeviceIdAdminChangedIterator{contract: _AftermarketDeviceId.contract, event: "AdminChanged", logs: logs, sub: sub}, nil
}

// WatchAdminChanged is a free log subscription operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) WatchAdminChanged(opts *bind.WatchOpts, sink chan<- *AftermarketDeviceIdAdminChanged) (event.Subscription, error) {

	logs, sub, err := _AftermarketDeviceId.contract.WatchLogs(opts, "AdminChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AftermarketDeviceIdAdminChanged)
				if err := _AftermarketDeviceId.contract.UnpackLog(event, "AdminChanged", log); err != nil {
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
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) ParseAdminChanged(log types.Log) (*AftermarketDeviceIdAdminChanged, error) {
	event := new(AftermarketDeviceIdAdminChanged)
	if err := _AftermarketDeviceId.contract.UnpackLog(event, "AdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AftermarketDeviceIdApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdApprovalIterator struct {
	Event *AftermarketDeviceIdApproval // Event containing the contract specifics and raw log

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
func (it *AftermarketDeviceIdApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AftermarketDeviceIdApproval)
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
		it.Event = new(AftermarketDeviceIdApproval)
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
func (it *AftermarketDeviceIdApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AftermarketDeviceIdApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AftermarketDeviceIdApproval represents a Approval event raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdApproval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, approved []common.Address, tokenId []*big.Int) (*AftermarketDeviceIdApprovalIterator, error) {

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

	logs, sub, err := _AftermarketDeviceId.contract.FilterLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &AftermarketDeviceIdApprovalIterator{contract: _AftermarketDeviceId.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *AftermarketDeviceIdApproval, owner []common.Address, approved []common.Address, tokenId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _AftermarketDeviceId.contract.WatchLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AftermarketDeviceIdApproval)
				if err := _AftermarketDeviceId.contract.UnpackLog(event, "Approval", log); err != nil {
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
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) ParseApproval(log types.Log) (*AftermarketDeviceIdApproval, error) {
	event := new(AftermarketDeviceIdApproval)
	if err := _AftermarketDeviceId.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AftermarketDeviceIdApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdApprovalForAllIterator struct {
	Event *AftermarketDeviceIdApprovalForAll // Event containing the contract specifics and raw log

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
func (it *AftermarketDeviceIdApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AftermarketDeviceIdApprovalForAll)
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
		it.Event = new(AftermarketDeviceIdApprovalForAll)
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
func (it *AftermarketDeviceIdApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AftermarketDeviceIdApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AftermarketDeviceIdApprovalForAll represents a ApprovalForAll event raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) FilterApprovalForAll(opts *bind.FilterOpts, owner []common.Address, operator []common.Address) (*AftermarketDeviceIdApprovalForAllIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _AftermarketDeviceId.contract.FilterLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return &AftermarketDeviceIdApprovalForAllIterator{contract: _AftermarketDeviceId.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *AftermarketDeviceIdApprovalForAll, owner []common.Address, operator []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _AftermarketDeviceId.contract.WatchLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AftermarketDeviceIdApprovalForAll)
				if err := _AftermarketDeviceId.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
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
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) ParseApprovalForAll(log types.Log) (*AftermarketDeviceIdApprovalForAll, error) {
	event := new(AftermarketDeviceIdApprovalForAll)
	if err := _AftermarketDeviceId.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AftermarketDeviceIdBeaconUpgradedIterator is returned from FilterBeaconUpgraded and is used to iterate over the raw logs and unpacked data for BeaconUpgraded events raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdBeaconUpgradedIterator struct {
	Event *AftermarketDeviceIdBeaconUpgraded // Event containing the contract specifics and raw log

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
func (it *AftermarketDeviceIdBeaconUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AftermarketDeviceIdBeaconUpgraded)
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
		it.Event = new(AftermarketDeviceIdBeaconUpgraded)
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
func (it *AftermarketDeviceIdBeaconUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AftermarketDeviceIdBeaconUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AftermarketDeviceIdBeaconUpgraded represents a BeaconUpgraded event raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdBeaconUpgraded struct {
	Beacon common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterBeaconUpgraded is a free log retrieval operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) FilterBeaconUpgraded(opts *bind.FilterOpts, beacon []common.Address) (*AftermarketDeviceIdBeaconUpgradedIterator, error) {

	var beaconRule []interface{}
	for _, beaconItem := range beacon {
		beaconRule = append(beaconRule, beaconItem)
	}

	logs, sub, err := _AftermarketDeviceId.contract.FilterLogs(opts, "BeaconUpgraded", beaconRule)
	if err != nil {
		return nil, err
	}
	return &AftermarketDeviceIdBeaconUpgradedIterator{contract: _AftermarketDeviceId.contract, event: "BeaconUpgraded", logs: logs, sub: sub}, nil
}

// WatchBeaconUpgraded is a free log subscription operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) WatchBeaconUpgraded(opts *bind.WatchOpts, sink chan<- *AftermarketDeviceIdBeaconUpgraded, beacon []common.Address) (event.Subscription, error) {

	var beaconRule []interface{}
	for _, beaconItem := range beacon {
		beaconRule = append(beaconRule, beaconItem)
	}

	logs, sub, err := _AftermarketDeviceId.contract.WatchLogs(opts, "BeaconUpgraded", beaconRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AftermarketDeviceIdBeaconUpgraded)
				if err := _AftermarketDeviceId.contract.UnpackLog(event, "BeaconUpgraded", log); err != nil {
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
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) ParseBeaconUpgraded(log types.Log) (*AftermarketDeviceIdBeaconUpgraded, error) {
	event := new(AftermarketDeviceIdBeaconUpgraded)
	if err := _AftermarketDeviceId.contract.UnpackLog(event, "BeaconUpgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AftermarketDeviceIdInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdInitializedIterator struct {
	Event *AftermarketDeviceIdInitialized // Event containing the contract specifics and raw log

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
func (it *AftermarketDeviceIdInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AftermarketDeviceIdInitialized)
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
		it.Event = new(AftermarketDeviceIdInitialized)
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
func (it *AftermarketDeviceIdInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AftermarketDeviceIdInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AftermarketDeviceIdInitialized represents a Initialized event raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) FilterInitialized(opts *bind.FilterOpts) (*AftermarketDeviceIdInitializedIterator, error) {

	logs, sub, err := _AftermarketDeviceId.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &AftermarketDeviceIdInitializedIterator{contract: _AftermarketDeviceId.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *AftermarketDeviceIdInitialized) (event.Subscription, error) {

	logs, sub, err := _AftermarketDeviceId.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AftermarketDeviceIdInitialized)
				if err := _AftermarketDeviceId.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) ParseInitialized(log types.Log) (*AftermarketDeviceIdInitialized, error) {
	event := new(AftermarketDeviceIdInitialized)
	if err := _AftermarketDeviceId.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AftermarketDeviceIdPrivilegeCreatedIterator is returned from FilterPrivilegeCreated and is used to iterate over the raw logs and unpacked data for PrivilegeCreated events raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdPrivilegeCreatedIterator struct {
	Event *AftermarketDeviceIdPrivilegeCreated // Event containing the contract specifics and raw log

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
func (it *AftermarketDeviceIdPrivilegeCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AftermarketDeviceIdPrivilegeCreated)
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
		it.Event = new(AftermarketDeviceIdPrivilegeCreated)
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
func (it *AftermarketDeviceIdPrivilegeCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AftermarketDeviceIdPrivilegeCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AftermarketDeviceIdPrivilegeCreated represents a PrivilegeCreated event raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdPrivilegeCreated struct {
	PrivilegeId *big.Int
	Enabled     bool
	Description string
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterPrivilegeCreated is a free log retrieval operation binding the contract event 0x6b1e285f7a3f24ce57e873fa5f9150dfc8a8f24f4feb1e581d42f678fa1b40b3.
//
// Solidity: event PrivilegeCreated(uint256 indexed privilegeId, bool enabled, string description)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) FilterPrivilegeCreated(opts *bind.FilterOpts, privilegeId []*big.Int) (*AftermarketDeviceIdPrivilegeCreatedIterator, error) {

	var privilegeIdRule []interface{}
	for _, privilegeIdItem := range privilegeId {
		privilegeIdRule = append(privilegeIdRule, privilegeIdItem)
	}

	logs, sub, err := _AftermarketDeviceId.contract.FilterLogs(opts, "PrivilegeCreated", privilegeIdRule)
	if err != nil {
		return nil, err
	}
	return &AftermarketDeviceIdPrivilegeCreatedIterator{contract: _AftermarketDeviceId.contract, event: "PrivilegeCreated", logs: logs, sub: sub}, nil
}

// WatchPrivilegeCreated is a free log subscription operation binding the contract event 0x6b1e285f7a3f24ce57e873fa5f9150dfc8a8f24f4feb1e581d42f678fa1b40b3.
//
// Solidity: event PrivilegeCreated(uint256 indexed privilegeId, bool enabled, string description)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) WatchPrivilegeCreated(opts *bind.WatchOpts, sink chan<- *AftermarketDeviceIdPrivilegeCreated, privilegeId []*big.Int) (event.Subscription, error) {

	var privilegeIdRule []interface{}
	for _, privilegeIdItem := range privilegeId {
		privilegeIdRule = append(privilegeIdRule, privilegeIdItem)
	}

	logs, sub, err := _AftermarketDeviceId.contract.WatchLogs(opts, "PrivilegeCreated", privilegeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AftermarketDeviceIdPrivilegeCreated)
				if err := _AftermarketDeviceId.contract.UnpackLog(event, "PrivilegeCreated", log); err != nil {
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
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) ParsePrivilegeCreated(log types.Log) (*AftermarketDeviceIdPrivilegeCreated, error) {
	event := new(AftermarketDeviceIdPrivilegeCreated)
	if err := _AftermarketDeviceId.contract.UnpackLog(event, "PrivilegeCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AftermarketDeviceIdPrivilegeDisabledIterator is returned from FilterPrivilegeDisabled and is used to iterate over the raw logs and unpacked data for PrivilegeDisabled events raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdPrivilegeDisabledIterator struct {
	Event *AftermarketDeviceIdPrivilegeDisabled // Event containing the contract specifics and raw log

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
func (it *AftermarketDeviceIdPrivilegeDisabledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AftermarketDeviceIdPrivilegeDisabled)
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
		it.Event = new(AftermarketDeviceIdPrivilegeDisabled)
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
func (it *AftermarketDeviceIdPrivilegeDisabledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AftermarketDeviceIdPrivilegeDisabledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AftermarketDeviceIdPrivilegeDisabled represents a PrivilegeDisabled event raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdPrivilegeDisabled struct {
	PrivilegeId *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterPrivilegeDisabled is a free log retrieval operation binding the contract event 0xd0a5bf4add1e8e17a93542ca7a6aa2a541ed2da0cd3f76b8b799e7b698f8a633.
//
// Solidity: event PrivilegeDisabled(uint256 indexed privilegeId)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) FilterPrivilegeDisabled(opts *bind.FilterOpts, privilegeId []*big.Int) (*AftermarketDeviceIdPrivilegeDisabledIterator, error) {

	var privilegeIdRule []interface{}
	for _, privilegeIdItem := range privilegeId {
		privilegeIdRule = append(privilegeIdRule, privilegeIdItem)
	}

	logs, sub, err := _AftermarketDeviceId.contract.FilterLogs(opts, "PrivilegeDisabled", privilegeIdRule)
	if err != nil {
		return nil, err
	}
	return &AftermarketDeviceIdPrivilegeDisabledIterator{contract: _AftermarketDeviceId.contract, event: "PrivilegeDisabled", logs: logs, sub: sub}, nil
}

// WatchPrivilegeDisabled is a free log subscription operation binding the contract event 0xd0a5bf4add1e8e17a93542ca7a6aa2a541ed2da0cd3f76b8b799e7b698f8a633.
//
// Solidity: event PrivilegeDisabled(uint256 indexed privilegeId)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) WatchPrivilegeDisabled(opts *bind.WatchOpts, sink chan<- *AftermarketDeviceIdPrivilegeDisabled, privilegeId []*big.Int) (event.Subscription, error) {

	var privilegeIdRule []interface{}
	for _, privilegeIdItem := range privilegeId {
		privilegeIdRule = append(privilegeIdRule, privilegeIdItem)
	}

	logs, sub, err := _AftermarketDeviceId.contract.WatchLogs(opts, "PrivilegeDisabled", privilegeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AftermarketDeviceIdPrivilegeDisabled)
				if err := _AftermarketDeviceId.contract.UnpackLog(event, "PrivilegeDisabled", log); err != nil {
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
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) ParsePrivilegeDisabled(log types.Log) (*AftermarketDeviceIdPrivilegeDisabled, error) {
	event := new(AftermarketDeviceIdPrivilegeDisabled)
	if err := _AftermarketDeviceId.contract.UnpackLog(event, "PrivilegeDisabled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AftermarketDeviceIdPrivilegeEnabledIterator is returned from FilterPrivilegeEnabled and is used to iterate over the raw logs and unpacked data for PrivilegeEnabled events raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdPrivilegeEnabledIterator struct {
	Event *AftermarketDeviceIdPrivilegeEnabled // Event containing the contract specifics and raw log

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
func (it *AftermarketDeviceIdPrivilegeEnabledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AftermarketDeviceIdPrivilegeEnabled)
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
		it.Event = new(AftermarketDeviceIdPrivilegeEnabled)
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
func (it *AftermarketDeviceIdPrivilegeEnabledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AftermarketDeviceIdPrivilegeEnabledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AftermarketDeviceIdPrivilegeEnabled represents a PrivilegeEnabled event raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdPrivilegeEnabled struct {
	PrivilegeId *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterPrivilegeEnabled is a free log retrieval operation binding the contract event 0xdb76175b0679b36e2d84207a86ca82a64ed1fd1af72c1a48eb3ff9369bbc4c83.
//
// Solidity: event PrivilegeEnabled(uint256 indexed privilegeId)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) FilterPrivilegeEnabled(opts *bind.FilterOpts, privilegeId []*big.Int) (*AftermarketDeviceIdPrivilegeEnabledIterator, error) {

	var privilegeIdRule []interface{}
	for _, privilegeIdItem := range privilegeId {
		privilegeIdRule = append(privilegeIdRule, privilegeIdItem)
	}

	logs, sub, err := _AftermarketDeviceId.contract.FilterLogs(opts, "PrivilegeEnabled", privilegeIdRule)
	if err != nil {
		return nil, err
	}
	return &AftermarketDeviceIdPrivilegeEnabledIterator{contract: _AftermarketDeviceId.contract, event: "PrivilegeEnabled", logs: logs, sub: sub}, nil
}

// WatchPrivilegeEnabled is a free log subscription operation binding the contract event 0xdb76175b0679b36e2d84207a86ca82a64ed1fd1af72c1a48eb3ff9369bbc4c83.
//
// Solidity: event PrivilegeEnabled(uint256 indexed privilegeId)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) WatchPrivilegeEnabled(opts *bind.WatchOpts, sink chan<- *AftermarketDeviceIdPrivilegeEnabled, privilegeId []*big.Int) (event.Subscription, error) {

	var privilegeIdRule []interface{}
	for _, privilegeIdItem := range privilegeId {
		privilegeIdRule = append(privilegeIdRule, privilegeIdItem)
	}

	logs, sub, err := _AftermarketDeviceId.contract.WatchLogs(opts, "PrivilegeEnabled", privilegeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AftermarketDeviceIdPrivilegeEnabled)
				if err := _AftermarketDeviceId.contract.UnpackLog(event, "PrivilegeEnabled", log); err != nil {
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
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) ParsePrivilegeEnabled(log types.Log) (*AftermarketDeviceIdPrivilegeEnabled, error) {
	event := new(AftermarketDeviceIdPrivilegeEnabled)
	if err := _AftermarketDeviceId.contract.UnpackLog(event, "PrivilegeEnabled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AftermarketDeviceIdPrivilegeSetIterator is returned from FilterPrivilegeSet and is used to iterate over the raw logs and unpacked data for PrivilegeSet events raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdPrivilegeSetIterator struct {
	Event *AftermarketDeviceIdPrivilegeSet // Event containing the contract specifics and raw log

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
func (it *AftermarketDeviceIdPrivilegeSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AftermarketDeviceIdPrivilegeSet)
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
		it.Event = new(AftermarketDeviceIdPrivilegeSet)
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
func (it *AftermarketDeviceIdPrivilegeSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AftermarketDeviceIdPrivilegeSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AftermarketDeviceIdPrivilegeSet represents a PrivilegeSet event raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdPrivilegeSet struct {
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
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) FilterPrivilegeSet(opts *bind.FilterOpts, tokenId []*big.Int, privId []*big.Int, user []common.Address) (*AftermarketDeviceIdPrivilegeSetIterator, error) {

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

	logs, sub, err := _AftermarketDeviceId.contract.FilterLogs(opts, "PrivilegeSet", tokenIdRule, privIdRule, userRule)
	if err != nil {
		return nil, err
	}
	return &AftermarketDeviceIdPrivilegeSetIterator{contract: _AftermarketDeviceId.contract, event: "PrivilegeSet", logs: logs, sub: sub}, nil
}

// WatchPrivilegeSet is a free log subscription operation binding the contract event 0x61a24679288162b799d80b2bb2b8b0fcdd5c5f53ac19e9246cc190b60196c359.
//
// Solidity: event PrivilegeSet(uint256 indexed tokenId, uint256 version, uint256 indexed privId, address indexed user, uint256 expires)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) WatchPrivilegeSet(opts *bind.WatchOpts, sink chan<- *AftermarketDeviceIdPrivilegeSet, tokenId []*big.Int, privId []*big.Int, user []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _AftermarketDeviceId.contract.WatchLogs(opts, "PrivilegeSet", tokenIdRule, privIdRule, userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AftermarketDeviceIdPrivilegeSet)
				if err := _AftermarketDeviceId.contract.UnpackLog(event, "PrivilegeSet", log); err != nil {
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
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) ParsePrivilegeSet(log types.Log) (*AftermarketDeviceIdPrivilegeSet, error) {
	event := new(AftermarketDeviceIdPrivilegeSet)
	if err := _AftermarketDeviceId.contract.UnpackLog(event, "PrivilegeSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AftermarketDeviceIdRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdRoleAdminChangedIterator struct {
	Event *AftermarketDeviceIdRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *AftermarketDeviceIdRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AftermarketDeviceIdRoleAdminChanged)
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
		it.Event = new(AftermarketDeviceIdRoleAdminChanged)
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
func (it *AftermarketDeviceIdRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AftermarketDeviceIdRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AftermarketDeviceIdRoleAdminChanged represents a RoleAdminChanged event raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*AftermarketDeviceIdRoleAdminChangedIterator, error) {

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

	logs, sub, err := _AftermarketDeviceId.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &AftermarketDeviceIdRoleAdminChangedIterator{contract: _AftermarketDeviceId.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *AftermarketDeviceIdRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

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

	logs, sub, err := _AftermarketDeviceId.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AftermarketDeviceIdRoleAdminChanged)
				if err := _AftermarketDeviceId.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) ParseRoleAdminChanged(log types.Log) (*AftermarketDeviceIdRoleAdminChanged, error) {
	event := new(AftermarketDeviceIdRoleAdminChanged)
	if err := _AftermarketDeviceId.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AftermarketDeviceIdRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdRoleGrantedIterator struct {
	Event *AftermarketDeviceIdRoleGranted // Event containing the contract specifics and raw log

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
func (it *AftermarketDeviceIdRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AftermarketDeviceIdRoleGranted)
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
		it.Event = new(AftermarketDeviceIdRoleGranted)
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
func (it *AftermarketDeviceIdRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AftermarketDeviceIdRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AftermarketDeviceIdRoleGranted represents a RoleGranted event raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*AftermarketDeviceIdRoleGrantedIterator, error) {

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

	logs, sub, err := _AftermarketDeviceId.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &AftermarketDeviceIdRoleGrantedIterator{contract: _AftermarketDeviceId.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *AftermarketDeviceIdRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _AftermarketDeviceId.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AftermarketDeviceIdRoleGranted)
				if err := _AftermarketDeviceId.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) ParseRoleGranted(log types.Log) (*AftermarketDeviceIdRoleGranted, error) {
	event := new(AftermarketDeviceIdRoleGranted)
	if err := _AftermarketDeviceId.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AftermarketDeviceIdRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdRoleRevokedIterator struct {
	Event *AftermarketDeviceIdRoleRevoked // Event containing the contract specifics and raw log

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
func (it *AftermarketDeviceIdRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AftermarketDeviceIdRoleRevoked)
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
		it.Event = new(AftermarketDeviceIdRoleRevoked)
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
func (it *AftermarketDeviceIdRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AftermarketDeviceIdRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AftermarketDeviceIdRoleRevoked represents a RoleRevoked event raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*AftermarketDeviceIdRoleRevokedIterator, error) {

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

	logs, sub, err := _AftermarketDeviceId.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &AftermarketDeviceIdRoleRevokedIterator{contract: _AftermarketDeviceId.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *AftermarketDeviceIdRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _AftermarketDeviceId.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AftermarketDeviceIdRoleRevoked)
				if err := _AftermarketDeviceId.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) ParseRoleRevoked(log types.Log) (*AftermarketDeviceIdRoleRevoked, error) {
	event := new(AftermarketDeviceIdRoleRevoked)
	if err := _AftermarketDeviceId.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AftermarketDeviceIdTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdTransferIterator struct {
	Event *AftermarketDeviceIdTransfer // Event containing the contract specifics and raw log

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
func (it *AftermarketDeviceIdTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AftermarketDeviceIdTransfer)
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
		it.Event = new(AftermarketDeviceIdTransfer)
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
func (it *AftermarketDeviceIdTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AftermarketDeviceIdTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AftermarketDeviceIdTransfer represents a Transfer event raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address, tokenId []*big.Int) (*AftermarketDeviceIdTransferIterator, error) {

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

	logs, sub, err := _AftermarketDeviceId.contract.FilterLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &AftermarketDeviceIdTransferIterator{contract: _AftermarketDeviceId.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *AftermarketDeviceIdTransfer, from []common.Address, to []common.Address, tokenId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _AftermarketDeviceId.contract.WatchLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AftermarketDeviceIdTransfer)
				if err := _AftermarketDeviceId.contract.UnpackLog(event, "Transfer", log); err != nil {
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
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) ParseTransfer(log types.Log) (*AftermarketDeviceIdTransfer, error) {
	event := new(AftermarketDeviceIdTransfer)
	if err := _AftermarketDeviceId.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AftermarketDeviceIdUpgradedIterator is returned from FilterUpgraded and is used to iterate over the raw logs and unpacked data for Upgraded events raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdUpgradedIterator struct {
	Event *AftermarketDeviceIdUpgraded // Event containing the contract specifics and raw log

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
func (it *AftermarketDeviceIdUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AftermarketDeviceIdUpgraded)
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
		it.Event = new(AftermarketDeviceIdUpgraded)
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
func (it *AftermarketDeviceIdUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AftermarketDeviceIdUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AftermarketDeviceIdUpgraded represents a Upgraded event raised by the AftermarketDeviceId contract.
type AftermarketDeviceIdUpgraded struct {
	Implementation common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpgraded is a free log retrieval operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) FilterUpgraded(opts *bind.FilterOpts, implementation []common.Address) (*AftermarketDeviceIdUpgradedIterator, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _AftermarketDeviceId.contract.FilterLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return &AftermarketDeviceIdUpgradedIterator{contract: _AftermarketDeviceId.contract, event: "Upgraded", logs: logs, sub: sub}, nil
}

// WatchUpgraded is a free log subscription operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) WatchUpgraded(opts *bind.WatchOpts, sink chan<- *AftermarketDeviceIdUpgraded, implementation []common.Address) (event.Subscription, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _AftermarketDeviceId.contract.WatchLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AftermarketDeviceIdUpgraded)
				if err := _AftermarketDeviceId.contract.UnpackLog(event, "Upgraded", log); err != nil {
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
func (_AftermarketDeviceId *AftermarketDeviceIdFilterer) ParseUpgraded(log types.Log) (*AftermarketDeviceIdUpgraded, error) {
	event := new(AftermarketDeviceIdUpgraded)
	if err := _AftermarketDeviceId.contract.UnpackLog(event, "Upgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
