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

// AftermarketDeviceInfos is an auto generated low-level Go binding around an user-defined struct.
type AftermarketDeviceInfos struct {
	Addr          common.Address
	AttrInfoPairs []AttributeInfoPair
}

// AftermarketDeviceOwnerPair is an auto generated low-level Go binding around an user-defined struct.
type AftermarketDeviceOwnerPair struct {
	AftermarketDeviceNodeId *big.Int
	Owner                   common.Address
}

// AttributeInfoPair is an auto generated low-level Go binding around an user-defined struct.
type AttributeInfoPair struct {
	Attribute string
	Info      string
}

// DevAdminIdManufacturerName is an auto generated low-level Go binding around an user-defined struct.
type DevAdminIdManufacturerName struct {
	TokenId *big.Int
	Name    string
}

// RegistryMetaData contains all meta data concerning the Registry contract.
var RegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"UintUtils__InsufficientHexLength\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"moduleAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes4[]\",\"name\":\"selectors\",\"type\":\"bytes4[]\"}],\"name\":\"ModuleAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"moduleAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes4[]\",\"name\":\"selectors\",\"type\":\"bytes4[]\"}],\"name\":\"ModuleRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldImplementation\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes4[]\",\"name\":\"oldSelectors\",\"type\":\"bytes4[]\"},{\"indexed\":false,\"internalType\":\"bytes4[]\",\"name\":\"newSelectors\",\"type\":\"bytes4[]\"}],\"name\":\"ModuleUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"},{\"internalType\":\"bytes4[]\",\"name\":\"selectors\",\"type\":\"bytes4[]\"}],\"name\":\"addModule\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"},{\"internalType\":\"bytes4[]\",\"name\":\"selectors\",\"type\":\"bytes4[]\"}],\"name\":\"removeModule\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"oldImplementation\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes4[]\",\"name\":\"oldSelectors\",\"type\":\"bytes4[]\"},{\"internalType\":\"bytes4[]\",\"name\":\"newSelectors\",\"type\":\"bytes4[]\"}],\"name\":\"updateModule\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"AftermarketDeviceTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"}],\"name\":\"AftermarketDeviceUnclaimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"vehicleNode\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"AftermarketDeviceUnpaired\",\"type\":\"event\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"}],\"internalType\":\"structDevAdmin.IdManufacturerName[]\",\"name\":\"idManufacturerNames\",\"type\":\"tuple[]\"}],\"name\":\"renameManufacturers\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferAftermarketDeviceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"aftermarketDeviceNodes\",\"type\":\"uint256[]\"}],\"name\":\"unclaimAftermarketDeviceNode\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"aftermarketDeviceNodes\",\"type\":\"uint256[]\"}],\"name\":\"unpairAftermarketDeviceByDeviceNode\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"vehicleNodes\",\"type\":\"uint256[]\"}],\"name\":\"unpairAftermarketDeviceByVehicleNode\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes[]\",\"name\":\"data\",\"type\":\"bytes[]\"}],\"name\":\"multiDelegateCall\",\"outputs\":[{\"internalType\":\"bytes[]\",\"name\":\"results\",\"type\":\"bytes[]\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes[]\",\"name\":\"data\",\"type\":\"bytes[]\"}],\"name\":\"multiStaticCall\",\"outputs\":[{\"internalType\":\"bytes[]\",\"name\":\"results\",\"type\":\"bytes[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAdMintCost\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"adMintCost\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_adMintCost\",\"type\":\"uint256\"}],\"name\":\"setAdMintCost\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_dimoToken\",\"type\":\"address\"}],\"name\":\"setDimoToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_foundation\",\"type\":\"address\"}],\"name\":\"setFoundationAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_license\",\"type\":\"address\"}],\"name\":\"setLicense\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"}],\"name\":\"AftermarketDeviceAttributeAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"info\",\"type\":\"string\"}],\"name\":\"AftermarketDeviceAttributeSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"AftermarketDeviceClaimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"proxy\",\"type\":\"address\"}],\"name\":\"AftermarketDeviceIdProxySet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"aftermarketDeviceAddress\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"AftermarketDeviceNodeMinted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"vehicleNode\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"AftermarketDevicePaired\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"vehicleNode\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"AftermarketDeviceUnpaired\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"}],\"name\":\"addAftermarketDeviceAttribute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNodeId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"internalType\":\"structAftermarketDeviceOwnerPair[]\",\"name\":\"adOwnerPair\",\"type\":\"tuple[]\"}],\"name\":\"claimAftermarketDeviceBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"ownerSig\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"aftermarketDeviceSig\",\"type\":\"bytes\"}],\"name\":\"claimAftermarketDeviceSign\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getAftermarketDeviceIdByAddress\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeId\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"manufacturerNode\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"info\",\"type\":\"string\"}],\"internalType\":\"structAttributeInfoPair[]\",\"name\":\"attrInfoPairs\",\"type\":\"tuple[]\"}],\"internalType\":\"structAftermarketDeviceInfos[]\",\"name\":\"adInfos\",\"type\":\"tuple[]\"}],\"name\":\"mintAftermarketDeviceByManufacturerBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"vehicleNode\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"aftermarketDeviceSig\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"vehicleOwnerSig\",\"type\":\"bytes\"}],\"name\":\"pairAftermarketDeviceSign\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"vehicleNode\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"pairAftermarketDeviceSign\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"setAftermarketDeviceIdProxyAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"info\",\"type\":\"string\"}],\"internalType\":\"structAttributeInfoPair[]\",\"name\":\"attrInfo\",\"type\":\"tuple[]\"}],\"name\":\"setAftermarketDeviceInfo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"vehicleNode\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"unpairAftermarketDeviceSign\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"}],\"name\":\"verifyAftermarketDeviceTransfer\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"controller\",\"type\":\"address\"}],\"name\":\"ControllerSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"}],\"name\":\"ManufacturerAttributeAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"info\",\"type\":\"string\"}],\"name\":\"ManufacturerAttributeSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"proxy\",\"type\":\"address\"}],\"name\":\"ManufacturerIdProxySet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"ManufacturerNodeMinted\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"}],\"name\":\"addManufacturerAttribute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"}],\"name\":\"getManufacturerIdByName\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeId\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getManufacturerNameById\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"isAllowedToOwnManufacturerNode\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"_isAllowed\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"isController\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"_isController\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"isManufacturerMinted\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"_isManufacturerMinted\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"info\",\"type\":\"string\"}],\"internalType\":\"structAttributeInfoPair[]\",\"name\":\"attrInfoPairList\",\"type\":\"tuple[]\"}],\"name\":\"mintManufacturer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string[]\",\"name\":\"names\",\"type\":\"string[]\"}],\"name\":\"mintManufacturerBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_controller\",\"type\":\"address\"}],\"name\":\"setController\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"setManufacturerIdProxyAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"info\",\"type\":\"string\"}],\"internalType\":\"structAttributeInfoPair[]\",\"name\":\"attrInfoList\",\"type\":\"tuple[]\"}],\"name\":\"setManufacturerInfo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"setManufacturerMinted\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"}],\"name\":\"VehicleAttributeAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"info\",\"type\":\"string\"}],\"name\":\"VehicleAttributeSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"proxy\",\"type\":\"address\"}],\"name\":\"VehicleIdProxySet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"VehicleNodeMinted\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"}],\"name\":\"addVehicleAttribute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"manufacturerNode\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"info\",\"type\":\"string\"}],\"internalType\":\"structAttributeInfoPair[]\",\"name\":\"attrInfo\",\"type\":\"tuple[]\"}],\"name\":\"mintVehicle\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"manufacturerNode\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"info\",\"type\":\"string\"}],\"internalType\":\"structAttributeInfoPair[]\",\"name\":\"attrInfo\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"mintVehicleSign\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"setVehicleIdProxyAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"info\",\"type\":\"string\"}],\"internalType\":\"structAttributeInfoPair[]\",\"name\":\"attrInfo\",\"type\":\"tuple[]\"}],\"name\":\"setVehicleInfo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"idProxyAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"}],\"name\":\"getInfo\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"info\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"idProxyAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getParentNode\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"parentNode\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"idProxyAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"sourceNode\",\"type\":\"uint256\"}],\"name\":\"getLink\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"targetNode\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// RegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use RegistryMetaData.ABI instead.
var RegistryABI = RegistryMetaData.ABI

// Registry is an auto generated Go binding around an Ethereum contract.
type Registry struct {
	RegistryCaller     // Read-only binding to the contract
	RegistryTransactor // Write-only binding to the contract
	RegistryFilterer   // Log filterer for contract events
}

// RegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type RegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type RegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type RegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type RegistrySession struct {
	Contract     *Registry         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type RegistryCallerSession struct {
	Contract *RegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// RegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type RegistryTransactorSession struct {
	Contract     *RegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// RegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type RegistryRaw struct {
	Contract *Registry // Generic contract binding to access the raw methods on
}

// RegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type RegistryCallerRaw struct {
	Contract *RegistryCaller // Generic read-only contract binding to access the raw methods on
}

// RegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type RegistryTransactorRaw struct {
	Contract *RegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewRegistry creates a new instance of Registry, bound to a specific deployed contract.
func NewRegistry(address common.Address, backend bind.ContractBackend) (*Registry, error) {
	contract, err := bindRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Registry{RegistryCaller: RegistryCaller{contract: contract}, RegistryTransactor: RegistryTransactor{contract: contract}, RegistryFilterer: RegistryFilterer{contract: contract}}, nil
}

// NewRegistryCaller creates a new read-only instance of Registry, bound to a specific deployed contract.
func NewRegistryCaller(address common.Address, caller bind.ContractCaller) (*RegistryCaller, error) {
	contract, err := bindRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &RegistryCaller{contract: contract}, nil
}

// NewRegistryTransactor creates a new write-only instance of Registry, bound to a specific deployed contract.
func NewRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*RegistryTransactor, error) {
	contract, err := bindRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &RegistryTransactor{contract: contract}, nil
}

// NewRegistryFilterer creates a new log filterer instance of Registry, bound to a specific deployed contract.
func NewRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*RegistryFilterer, error) {
	contract, err := bindRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &RegistryFilterer{contract: contract}, nil
}

// bindRegistry binds a generic wrapper to an already deployed contract.
func bindRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := RegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Registry *RegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Registry.Contract.RegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Registry *RegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Registry.Contract.RegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Registry *RegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Registry.Contract.RegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Registry *RegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Registry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Registry *RegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Registry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Registry *RegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Registry.Contract.contract.Transact(opts, method, params...)
}

// GetAdMintCost is a free data retrieval call binding the contract method 0x46946743.
//
// Solidity: function getAdMintCost() view returns(uint256 adMintCost)
func (_Registry *RegistryCaller) GetAdMintCost(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "getAdMintCost")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAdMintCost is a free data retrieval call binding the contract method 0x46946743.
//
// Solidity: function getAdMintCost() view returns(uint256 adMintCost)
func (_Registry *RegistrySession) GetAdMintCost() (*big.Int, error) {
	return _Registry.Contract.GetAdMintCost(&_Registry.CallOpts)
}

// GetAdMintCost is a free data retrieval call binding the contract method 0x46946743.
//
// Solidity: function getAdMintCost() view returns(uint256 adMintCost)
func (_Registry *RegistryCallerSession) GetAdMintCost() (*big.Int, error) {
	return _Registry.Contract.GetAdMintCost(&_Registry.CallOpts)
}

// GetAftermarketDeviceIdByAddress is a free data retrieval call binding the contract method 0x9796cf22.
//
// Solidity: function getAftermarketDeviceIdByAddress(address addr) view returns(uint256 nodeId)
func (_Registry *RegistryCaller) GetAftermarketDeviceIdByAddress(opts *bind.CallOpts, addr common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "getAftermarketDeviceIdByAddress", addr)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAftermarketDeviceIdByAddress is a free data retrieval call binding the contract method 0x9796cf22.
//
// Solidity: function getAftermarketDeviceIdByAddress(address addr) view returns(uint256 nodeId)
func (_Registry *RegistrySession) GetAftermarketDeviceIdByAddress(addr common.Address) (*big.Int, error) {
	return _Registry.Contract.GetAftermarketDeviceIdByAddress(&_Registry.CallOpts, addr)
}

// GetAftermarketDeviceIdByAddress is a free data retrieval call binding the contract method 0x9796cf22.
//
// Solidity: function getAftermarketDeviceIdByAddress(address addr) view returns(uint256 nodeId)
func (_Registry *RegistryCallerSession) GetAftermarketDeviceIdByAddress(addr common.Address) (*big.Int, error) {
	return _Registry.Contract.GetAftermarketDeviceIdByAddress(&_Registry.CallOpts, addr)
}

// GetInfo is a free data retrieval call binding the contract method 0xdce2f860.
//
// Solidity: function getInfo(address idProxyAddress, uint256 tokenId, string attribute) view returns(string info)
func (_Registry *RegistryCaller) GetInfo(opts *bind.CallOpts, idProxyAddress common.Address, tokenId *big.Int, attribute string) (string, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "getInfo", idProxyAddress, tokenId, attribute)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetInfo is a free data retrieval call binding the contract method 0xdce2f860.
//
// Solidity: function getInfo(address idProxyAddress, uint256 tokenId, string attribute) view returns(string info)
func (_Registry *RegistrySession) GetInfo(idProxyAddress common.Address, tokenId *big.Int, attribute string) (string, error) {
	return _Registry.Contract.GetInfo(&_Registry.CallOpts, idProxyAddress, tokenId, attribute)
}

// GetInfo is a free data retrieval call binding the contract method 0xdce2f860.
//
// Solidity: function getInfo(address idProxyAddress, uint256 tokenId, string attribute) view returns(string info)
func (_Registry *RegistryCallerSession) GetInfo(idProxyAddress common.Address, tokenId *big.Int, attribute string) (string, error) {
	return _Registry.Contract.GetInfo(&_Registry.CallOpts, idProxyAddress, tokenId, attribute)
}

// GetLink is a free data retrieval call binding the contract method 0x112e62a2.
//
// Solidity: function getLink(address idProxyAddress, uint256 sourceNode) view returns(uint256 targetNode)
func (_Registry *RegistryCaller) GetLink(opts *bind.CallOpts, idProxyAddress common.Address, sourceNode *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "getLink", idProxyAddress, sourceNode)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetLink is a free data retrieval call binding the contract method 0x112e62a2.
//
// Solidity: function getLink(address idProxyAddress, uint256 sourceNode) view returns(uint256 targetNode)
func (_Registry *RegistrySession) GetLink(idProxyAddress common.Address, sourceNode *big.Int) (*big.Int, error) {
	return _Registry.Contract.GetLink(&_Registry.CallOpts, idProxyAddress, sourceNode)
}

// GetLink is a free data retrieval call binding the contract method 0x112e62a2.
//
// Solidity: function getLink(address idProxyAddress, uint256 sourceNode) view returns(uint256 targetNode)
func (_Registry *RegistryCallerSession) GetLink(idProxyAddress common.Address, sourceNode *big.Int) (*big.Int, error) {
	return _Registry.Contract.GetLink(&_Registry.CallOpts, idProxyAddress, sourceNode)
}

// GetManufacturerIdByName is a free data retrieval call binding the contract method 0xce55aab0.
//
// Solidity: function getManufacturerIdByName(string name) view returns(uint256 nodeId)
func (_Registry *RegistryCaller) GetManufacturerIdByName(opts *bind.CallOpts, name string) (*big.Int, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "getManufacturerIdByName", name)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetManufacturerIdByName is a free data retrieval call binding the contract method 0xce55aab0.
//
// Solidity: function getManufacturerIdByName(string name) view returns(uint256 nodeId)
func (_Registry *RegistrySession) GetManufacturerIdByName(name string) (*big.Int, error) {
	return _Registry.Contract.GetManufacturerIdByName(&_Registry.CallOpts, name)
}

// GetManufacturerIdByName is a free data retrieval call binding the contract method 0xce55aab0.
//
// Solidity: function getManufacturerIdByName(string name) view returns(uint256 nodeId)
func (_Registry *RegistryCallerSession) GetManufacturerIdByName(name string) (*big.Int, error) {
	return _Registry.Contract.GetManufacturerIdByName(&_Registry.CallOpts, name)
}

// GetManufacturerNameById is a free data retrieval call binding the contract method 0x9109b30b.
//
// Solidity: function getManufacturerNameById(uint256 tokenId) view returns(string name)
func (_Registry *RegistryCaller) GetManufacturerNameById(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "getManufacturerNameById", tokenId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetManufacturerNameById is a free data retrieval call binding the contract method 0x9109b30b.
//
// Solidity: function getManufacturerNameById(uint256 tokenId) view returns(string name)
func (_Registry *RegistrySession) GetManufacturerNameById(tokenId *big.Int) (string, error) {
	return _Registry.Contract.GetManufacturerNameById(&_Registry.CallOpts, tokenId)
}

// GetManufacturerNameById is a free data retrieval call binding the contract method 0x9109b30b.
//
// Solidity: function getManufacturerNameById(uint256 tokenId) view returns(string name)
func (_Registry *RegistryCallerSession) GetManufacturerNameById(tokenId *big.Int) (string, error) {
	return _Registry.Contract.GetManufacturerNameById(&_Registry.CallOpts, tokenId)
}

// GetParentNode is a free data retrieval call binding the contract method 0x82087d24.
//
// Solidity: function getParentNode(address idProxyAddress, uint256 tokenId) view returns(uint256 parentNode)
func (_Registry *RegistryCaller) GetParentNode(opts *bind.CallOpts, idProxyAddress common.Address, tokenId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "getParentNode", idProxyAddress, tokenId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetParentNode is a free data retrieval call binding the contract method 0x82087d24.
//
// Solidity: function getParentNode(address idProxyAddress, uint256 tokenId) view returns(uint256 parentNode)
func (_Registry *RegistrySession) GetParentNode(idProxyAddress common.Address, tokenId *big.Int) (*big.Int, error) {
	return _Registry.Contract.GetParentNode(&_Registry.CallOpts, idProxyAddress, tokenId)
}

// GetParentNode is a free data retrieval call binding the contract method 0x82087d24.
//
// Solidity: function getParentNode(address idProxyAddress, uint256 tokenId) view returns(uint256 parentNode)
func (_Registry *RegistryCallerSession) GetParentNode(idProxyAddress common.Address, tokenId *big.Int) (*big.Int, error) {
	return _Registry.Contract.GetParentNode(&_Registry.CallOpts, idProxyAddress, tokenId)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Registry *RegistryCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Registry *RegistrySession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Registry.Contract.GetRoleAdmin(&_Registry.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Registry *RegistryCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Registry.Contract.GetRoleAdmin(&_Registry.CallOpts, role)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Registry *RegistryCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Registry *RegistrySession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Registry.Contract.HasRole(&_Registry.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Registry *RegistryCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Registry.Contract.HasRole(&_Registry.CallOpts, role, account)
}

// IsAllowedToOwnManufacturerNode is a free data retrieval call binding the contract method 0xd9c27c40.
//
// Solidity: function isAllowedToOwnManufacturerNode(address addr) view returns(bool _isAllowed)
func (_Registry *RegistryCaller) IsAllowedToOwnManufacturerNode(opts *bind.CallOpts, addr common.Address) (bool, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "isAllowedToOwnManufacturerNode", addr)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsAllowedToOwnManufacturerNode is a free data retrieval call binding the contract method 0xd9c27c40.
//
// Solidity: function isAllowedToOwnManufacturerNode(address addr) view returns(bool _isAllowed)
func (_Registry *RegistrySession) IsAllowedToOwnManufacturerNode(addr common.Address) (bool, error) {
	return _Registry.Contract.IsAllowedToOwnManufacturerNode(&_Registry.CallOpts, addr)
}

// IsAllowedToOwnManufacturerNode is a free data retrieval call binding the contract method 0xd9c27c40.
//
// Solidity: function isAllowedToOwnManufacturerNode(address addr) view returns(bool _isAllowed)
func (_Registry *RegistryCallerSession) IsAllowedToOwnManufacturerNode(addr common.Address) (bool, error) {
	return _Registry.Contract.IsAllowedToOwnManufacturerNode(&_Registry.CallOpts, addr)
}

// IsController is a free data retrieval call binding the contract method 0xb429afeb.
//
// Solidity: function isController(address addr) view returns(bool _isController)
func (_Registry *RegistryCaller) IsController(opts *bind.CallOpts, addr common.Address) (bool, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "isController", addr)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsController is a free data retrieval call binding the contract method 0xb429afeb.
//
// Solidity: function isController(address addr) view returns(bool _isController)
func (_Registry *RegistrySession) IsController(addr common.Address) (bool, error) {
	return _Registry.Contract.IsController(&_Registry.CallOpts, addr)
}

// IsController is a free data retrieval call binding the contract method 0xb429afeb.
//
// Solidity: function isController(address addr) view returns(bool _isController)
func (_Registry *RegistryCallerSession) IsController(addr common.Address) (bool, error) {
	return _Registry.Contract.IsController(&_Registry.CallOpts, addr)
}

// IsManufacturerMinted is a free data retrieval call binding the contract method 0x456bf169.
//
// Solidity: function isManufacturerMinted(address addr) view returns(bool _isManufacturerMinted)
func (_Registry *RegistryCaller) IsManufacturerMinted(opts *bind.CallOpts, addr common.Address) (bool, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "isManufacturerMinted", addr)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsManufacturerMinted is a free data retrieval call binding the contract method 0x456bf169.
//
// Solidity: function isManufacturerMinted(address addr) view returns(bool _isManufacturerMinted)
func (_Registry *RegistrySession) IsManufacturerMinted(addr common.Address) (bool, error) {
	return _Registry.Contract.IsManufacturerMinted(&_Registry.CallOpts, addr)
}

// IsManufacturerMinted is a free data retrieval call binding the contract method 0x456bf169.
//
// Solidity: function isManufacturerMinted(address addr) view returns(bool _isManufacturerMinted)
func (_Registry *RegistryCallerSession) IsManufacturerMinted(addr common.Address) (bool, error) {
	return _Registry.Contract.IsManufacturerMinted(&_Registry.CallOpts, addr)
}

// MultiStaticCall is a free data retrieval call binding the contract method 0x1c0c6e51.
//
// Solidity: function multiStaticCall(bytes[] data) view returns(bytes[] results)
func (_Registry *RegistryCaller) MultiStaticCall(opts *bind.CallOpts, data [][]byte) ([][]byte, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "multiStaticCall", data)

	if err != nil {
		return *new([][]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][]byte)).(*[][]byte)

	return out0, err

}

// MultiStaticCall is a free data retrieval call binding the contract method 0x1c0c6e51.
//
// Solidity: function multiStaticCall(bytes[] data) view returns(bytes[] results)
func (_Registry *RegistrySession) MultiStaticCall(data [][]byte) ([][]byte, error) {
	return _Registry.Contract.MultiStaticCall(&_Registry.CallOpts, data)
}

// MultiStaticCall is a free data retrieval call binding the contract method 0x1c0c6e51.
//
// Solidity: function multiStaticCall(bytes[] data) view returns(bytes[] results)
func (_Registry *RegistryCallerSession) MultiStaticCall(data [][]byte) ([][]byte, error) {
	return _Registry.Contract.MultiStaticCall(&_Registry.CallOpts, data)
}

// VerifyAftermarketDeviceTransfer is a free data retrieval call binding the contract method 0x198cec36.
//
// Solidity: function verifyAftermarketDeviceTransfer(uint256 aftermarketDeviceNode) view returns()
func (_Registry *RegistryCaller) VerifyAftermarketDeviceTransfer(opts *bind.CallOpts, aftermarketDeviceNode *big.Int) error {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "verifyAftermarketDeviceTransfer", aftermarketDeviceNode)

	if err != nil {
		return err
	}

	return err

}

// VerifyAftermarketDeviceTransfer is a free data retrieval call binding the contract method 0x198cec36.
//
// Solidity: function verifyAftermarketDeviceTransfer(uint256 aftermarketDeviceNode) view returns()
func (_Registry *RegistrySession) VerifyAftermarketDeviceTransfer(aftermarketDeviceNode *big.Int) error {
	return _Registry.Contract.VerifyAftermarketDeviceTransfer(&_Registry.CallOpts, aftermarketDeviceNode)
}

// VerifyAftermarketDeviceTransfer is a free data retrieval call binding the contract method 0x198cec36.
//
// Solidity: function verifyAftermarketDeviceTransfer(uint256 aftermarketDeviceNode) view returns()
func (_Registry *RegistryCallerSession) VerifyAftermarketDeviceTransfer(aftermarketDeviceNode *big.Int) error {
	return _Registry.Contract.VerifyAftermarketDeviceTransfer(&_Registry.CallOpts, aftermarketDeviceNode)
}

// AddAftermarketDeviceAttribute is a paid mutator transaction binding the contract method 0x6111afa3.
//
// Solidity: function addAftermarketDeviceAttribute(string attribute) returns()
func (_Registry *RegistryTransactor) AddAftermarketDeviceAttribute(opts *bind.TransactOpts, attribute string) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "addAftermarketDeviceAttribute", attribute)
}

// AddAftermarketDeviceAttribute is a paid mutator transaction binding the contract method 0x6111afa3.
//
// Solidity: function addAftermarketDeviceAttribute(string attribute) returns()
func (_Registry *RegistrySession) AddAftermarketDeviceAttribute(attribute string) (*types.Transaction, error) {
	return _Registry.Contract.AddAftermarketDeviceAttribute(&_Registry.TransactOpts, attribute)
}

// AddAftermarketDeviceAttribute is a paid mutator transaction binding the contract method 0x6111afa3.
//
// Solidity: function addAftermarketDeviceAttribute(string attribute) returns()
func (_Registry *RegistryTransactorSession) AddAftermarketDeviceAttribute(attribute string) (*types.Transaction, error) {
	return _Registry.Contract.AddAftermarketDeviceAttribute(&_Registry.TransactOpts, attribute)
}

// AddManufacturerAttribute is a paid mutator transaction binding the contract method 0x50300a3f.
//
// Solidity: function addManufacturerAttribute(string attribute) returns()
func (_Registry *RegistryTransactor) AddManufacturerAttribute(opts *bind.TransactOpts, attribute string) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "addManufacturerAttribute", attribute)
}

// AddManufacturerAttribute is a paid mutator transaction binding the contract method 0x50300a3f.
//
// Solidity: function addManufacturerAttribute(string attribute) returns()
func (_Registry *RegistrySession) AddManufacturerAttribute(attribute string) (*types.Transaction, error) {
	return _Registry.Contract.AddManufacturerAttribute(&_Registry.TransactOpts, attribute)
}

// AddManufacturerAttribute is a paid mutator transaction binding the contract method 0x50300a3f.
//
// Solidity: function addManufacturerAttribute(string attribute) returns()
func (_Registry *RegistryTransactorSession) AddManufacturerAttribute(attribute string) (*types.Transaction, error) {
	return _Registry.Contract.AddManufacturerAttribute(&_Registry.TransactOpts, attribute)
}

// AddModule is a paid mutator transaction binding the contract method 0x0df5b997.
//
// Solidity: function addModule(address implementation, bytes4[] selectors) returns()
func (_Registry *RegistryTransactor) AddModule(opts *bind.TransactOpts, implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "addModule", implementation, selectors)
}

// AddModule is a paid mutator transaction binding the contract method 0x0df5b997.
//
// Solidity: function addModule(address implementation, bytes4[] selectors) returns()
func (_Registry *RegistrySession) AddModule(implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _Registry.Contract.AddModule(&_Registry.TransactOpts, implementation, selectors)
}

// AddModule is a paid mutator transaction binding the contract method 0x0df5b997.
//
// Solidity: function addModule(address implementation, bytes4[] selectors) returns()
func (_Registry *RegistryTransactorSession) AddModule(implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _Registry.Contract.AddModule(&_Registry.TransactOpts, implementation, selectors)
}

// AddVehicleAttribute is a paid mutator transaction binding the contract method 0xf0d1a557.
//
// Solidity: function addVehicleAttribute(string attribute) returns()
func (_Registry *RegistryTransactor) AddVehicleAttribute(opts *bind.TransactOpts, attribute string) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "addVehicleAttribute", attribute)
}

// AddVehicleAttribute is a paid mutator transaction binding the contract method 0xf0d1a557.
//
// Solidity: function addVehicleAttribute(string attribute) returns()
func (_Registry *RegistrySession) AddVehicleAttribute(attribute string) (*types.Transaction, error) {
	return _Registry.Contract.AddVehicleAttribute(&_Registry.TransactOpts, attribute)
}

// AddVehicleAttribute is a paid mutator transaction binding the contract method 0xf0d1a557.
//
// Solidity: function addVehicleAttribute(string attribute) returns()
func (_Registry *RegistryTransactorSession) AddVehicleAttribute(attribute string) (*types.Transaction, error) {
	return _Registry.Contract.AddVehicleAttribute(&_Registry.TransactOpts, attribute)
}

// ClaimAftermarketDeviceBatch is a paid mutator transaction binding the contract method 0xab2ae229.
//
// Solidity: function claimAftermarketDeviceBatch((uint256,address)[] adOwnerPair) returns()
func (_Registry *RegistryTransactor) ClaimAftermarketDeviceBatch(opts *bind.TransactOpts, adOwnerPair []AftermarketDeviceOwnerPair) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "claimAftermarketDeviceBatch", adOwnerPair)
}

// ClaimAftermarketDeviceBatch is a paid mutator transaction binding the contract method 0xab2ae229.
//
// Solidity: function claimAftermarketDeviceBatch((uint256,address)[] adOwnerPair) returns()
func (_Registry *RegistrySession) ClaimAftermarketDeviceBatch(adOwnerPair []AftermarketDeviceOwnerPair) (*types.Transaction, error) {
	return _Registry.Contract.ClaimAftermarketDeviceBatch(&_Registry.TransactOpts, adOwnerPair)
}

// ClaimAftermarketDeviceBatch is a paid mutator transaction binding the contract method 0xab2ae229.
//
// Solidity: function claimAftermarketDeviceBatch((uint256,address)[] adOwnerPair) returns()
func (_Registry *RegistryTransactorSession) ClaimAftermarketDeviceBatch(adOwnerPair []AftermarketDeviceOwnerPair) (*types.Transaction, error) {
	return _Registry.Contract.ClaimAftermarketDeviceBatch(&_Registry.TransactOpts, adOwnerPair)
}

// ClaimAftermarketDeviceSign is a paid mutator transaction binding the contract method 0x89a841bb.
//
// Solidity: function claimAftermarketDeviceSign(uint256 aftermarketDeviceNode, address owner, bytes ownerSig, bytes aftermarketDeviceSig) returns()
func (_Registry *RegistryTransactor) ClaimAftermarketDeviceSign(opts *bind.TransactOpts, aftermarketDeviceNode *big.Int, owner common.Address, ownerSig []byte, aftermarketDeviceSig []byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "claimAftermarketDeviceSign", aftermarketDeviceNode, owner, ownerSig, aftermarketDeviceSig)
}

// ClaimAftermarketDeviceSign is a paid mutator transaction binding the contract method 0x89a841bb.
//
// Solidity: function claimAftermarketDeviceSign(uint256 aftermarketDeviceNode, address owner, bytes ownerSig, bytes aftermarketDeviceSig) returns()
func (_Registry *RegistrySession) ClaimAftermarketDeviceSign(aftermarketDeviceNode *big.Int, owner common.Address, ownerSig []byte, aftermarketDeviceSig []byte) (*types.Transaction, error) {
	return _Registry.Contract.ClaimAftermarketDeviceSign(&_Registry.TransactOpts, aftermarketDeviceNode, owner, ownerSig, aftermarketDeviceSig)
}

// ClaimAftermarketDeviceSign is a paid mutator transaction binding the contract method 0x89a841bb.
//
// Solidity: function claimAftermarketDeviceSign(uint256 aftermarketDeviceNode, address owner, bytes ownerSig, bytes aftermarketDeviceSig) returns()
func (_Registry *RegistryTransactorSession) ClaimAftermarketDeviceSign(aftermarketDeviceNode *big.Int, owner common.Address, ownerSig []byte, aftermarketDeviceSig []byte) (*types.Transaction, error) {
	return _Registry.Contract.ClaimAftermarketDeviceSign(&_Registry.TransactOpts, aftermarketDeviceNode, owner, ownerSig, aftermarketDeviceSig)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Registry *RegistryTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Registry *RegistrySession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Registry.Contract.GrantRole(&_Registry.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Registry *RegistryTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Registry.Contract.GrantRole(&_Registry.TransactOpts, role, account)
}

// Initialize is a paid mutator transaction binding the contract method 0x4cd88b76.
//
// Solidity: function initialize(string name, string version) returns()
func (_Registry *RegistryTransactor) Initialize(opts *bind.TransactOpts, name string, version string) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "initialize", name, version)
}

// Initialize is a paid mutator transaction binding the contract method 0x4cd88b76.
//
// Solidity: function initialize(string name, string version) returns()
func (_Registry *RegistrySession) Initialize(name string, version string) (*types.Transaction, error) {
	return _Registry.Contract.Initialize(&_Registry.TransactOpts, name, version)
}

// Initialize is a paid mutator transaction binding the contract method 0x4cd88b76.
//
// Solidity: function initialize(string name, string version) returns()
func (_Registry *RegistryTransactorSession) Initialize(name string, version string) (*types.Transaction, error) {
	return _Registry.Contract.Initialize(&_Registry.TransactOpts, name, version)
}

// MintAftermarketDeviceByManufacturerBatch is a paid mutator transaction binding the contract method 0x7ba79a39.
//
// Solidity: function mintAftermarketDeviceByManufacturerBatch(uint256 manufacturerNode, (address,(string,string)[])[] adInfos) returns()
func (_Registry *RegistryTransactor) MintAftermarketDeviceByManufacturerBatch(opts *bind.TransactOpts, manufacturerNode *big.Int, adInfos []AftermarketDeviceInfos) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "mintAftermarketDeviceByManufacturerBatch", manufacturerNode, adInfos)
}

// MintAftermarketDeviceByManufacturerBatch is a paid mutator transaction binding the contract method 0x7ba79a39.
//
// Solidity: function mintAftermarketDeviceByManufacturerBatch(uint256 manufacturerNode, (address,(string,string)[])[] adInfos) returns()
func (_Registry *RegistrySession) MintAftermarketDeviceByManufacturerBatch(manufacturerNode *big.Int, adInfos []AftermarketDeviceInfos) (*types.Transaction, error) {
	return _Registry.Contract.MintAftermarketDeviceByManufacturerBatch(&_Registry.TransactOpts, manufacturerNode, adInfos)
}

// MintAftermarketDeviceByManufacturerBatch is a paid mutator transaction binding the contract method 0x7ba79a39.
//
// Solidity: function mintAftermarketDeviceByManufacturerBatch(uint256 manufacturerNode, (address,(string,string)[])[] adInfos) returns()
func (_Registry *RegistryTransactorSession) MintAftermarketDeviceByManufacturerBatch(manufacturerNode *big.Int, adInfos []AftermarketDeviceInfos) (*types.Transaction, error) {
	return _Registry.Contract.MintAftermarketDeviceByManufacturerBatch(&_Registry.TransactOpts, manufacturerNode, adInfos)
}

// MintManufacturer is a paid mutator transaction binding the contract method 0x5f36da6b.
//
// Solidity: function mintManufacturer(address owner, string name, (string,string)[] attrInfoPairList) returns()
func (_Registry *RegistryTransactor) MintManufacturer(opts *bind.TransactOpts, owner common.Address, name string, attrInfoPairList []AttributeInfoPair) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "mintManufacturer", owner, name, attrInfoPairList)
}

// MintManufacturer is a paid mutator transaction binding the contract method 0x5f36da6b.
//
// Solidity: function mintManufacturer(address owner, string name, (string,string)[] attrInfoPairList) returns()
func (_Registry *RegistrySession) MintManufacturer(owner common.Address, name string, attrInfoPairList []AttributeInfoPair) (*types.Transaction, error) {
	return _Registry.Contract.MintManufacturer(&_Registry.TransactOpts, owner, name, attrInfoPairList)
}

// MintManufacturer is a paid mutator transaction binding the contract method 0x5f36da6b.
//
// Solidity: function mintManufacturer(address owner, string name, (string,string)[] attrInfoPairList) returns()
func (_Registry *RegistryTransactorSession) MintManufacturer(owner common.Address, name string, attrInfoPairList []AttributeInfoPair) (*types.Transaction, error) {
	return _Registry.Contract.MintManufacturer(&_Registry.TransactOpts, owner, name, attrInfoPairList)
}

// MintManufacturerBatch is a paid mutator transaction binding the contract method 0x9abb3000.
//
// Solidity: function mintManufacturerBatch(address owner, string[] names) returns()
func (_Registry *RegistryTransactor) MintManufacturerBatch(opts *bind.TransactOpts, owner common.Address, names []string) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "mintManufacturerBatch", owner, names)
}

// MintManufacturerBatch is a paid mutator transaction binding the contract method 0x9abb3000.
//
// Solidity: function mintManufacturerBatch(address owner, string[] names) returns()
func (_Registry *RegistrySession) MintManufacturerBatch(owner common.Address, names []string) (*types.Transaction, error) {
	return _Registry.Contract.MintManufacturerBatch(&_Registry.TransactOpts, owner, names)
}

// MintManufacturerBatch is a paid mutator transaction binding the contract method 0x9abb3000.
//
// Solidity: function mintManufacturerBatch(address owner, string[] names) returns()
func (_Registry *RegistryTransactorSession) MintManufacturerBatch(owner common.Address, names []string) (*types.Transaction, error) {
	return _Registry.Contract.MintManufacturerBatch(&_Registry.TransactOpts, owner, names)
}

// MintVehicle is a paid mutator transaction binding the contract method 0x3da44e56.
//
// Solidity: function mintVehicle(uint256 manufacturerNode, address owner, (string,string)[] attrInfo) returns()
func (_Registry *RegistryTransactor) MintVehicle(opts *bind.TransactOpts, manufacturerNode *big.Int, owner common.Address, attrInfo []AttributeInfoPair) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "mintVehicle", manufacturerNode, owner, attrInfo)
}

// MintVehicle is a paid mutator transaction binding the contract method 0x3da44e56.
//
// Solidity: function mintVehicle(uint256 manufacturerNode, address owner, (string,string)[] attrInfo) returns()
func (_Registry *RegistrySession) MintVehicle(manufacturerNode *big.Int, owner common.Address, attrInfo []AttributeInfoPair) (*types.Transaction, error) {
	return _Registry.Contract.MintVehicle(&_Registry.TransactOpts, manufacturerNode, owner, attrInfo)
}

// MintVehicle is a paid mutator transaction binding the contract method 0x3da44e56.
//
// Solidity: function mintVehicle(uint256 manufacturerNode, address owner, (string,string)[] attrInfo) returns()
func (_Registry *RegistryTransactorSession) MintVehicle(manufacturerNode *big.Int, owner common.Address, attrInfo []AttributeInfoPair) (*types.Transaction, error) {
	return _Registry.Contract.MintVehicle(&_Registry.TransactOpts, manufacturerNode, owner, attrInfo)
}

// MintVehicleSign is a paid mutator transaction binding the contract method 0x1b1a82c8.
//
// Solidity: function mintVehicleSign(uint256 manufacturerNode, address owner, (string,string)[] attrInfo, bytes signature) returns()
func (_Registry *RegistryTransactor) MintVehicleSign(opts *bind.TransactOpts, manufacturerNode *big.Int, owner common.Address, attrInfo []AttributeInfoPair, signature []byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "mintVehicleSign", manufacturerNode, owner, attrInfo, signature)
}

// MintVehicleSign is a paid mutator transaction binding the contract method 0x1b1a82c8.
//
// Solidity: function mintVehicleSign(uint256 manufacturerNode, address owner, (string,string)[] attrInfo, bytes signature) returns()
func (_Registry *RegistrySession) MintVehicleSign(manufacturerNode *big.Int, owner common.Address, attrInfo []AttributeInfoPair, signature []byte) (*types.Transaction, error) {
	return _Registry.Contract.MintVehicleSign(&_Registry.TransactOpts, manufacturerNode, owner, attrInfo, signature)
}

// MintVehicleSign is a paid mutator transaction binding the contract method 0x1b1a82c8.
//
// Solidity: function mintVehicleSign(uint256 manufacturerNode, address owner, (string,string)[] attrInfo, bytes signature) returns()
func (_Registry *RegistryTransactorSession) MintVehicleSign(manufacturerNode *big.Int, owner common.Address, attrInfo []AttributeInfoPair, signature []byte) (*types.Transaction, error) {
	return _Registry.Contract.MintVehicleSign(&_Registry.TransactOpts, manufacturerNode, owner, attrInfo, signature)
}

// MultiDelegateCall is a paid mutator transaction binding the contract method 0x415c2d96.
//
// Solidity: function multiDelegateCall(bytes[] data) returns(bytes[] results)
func (_Registry *RegistryTransactor) MultiDelegateCall(opts *bind.TransactOpts, data [][]byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "multiDelegateCall", data)
}

// MultiDelegateCall is a paid mutator transaction binding the contract method 0x415c2d96.
//
// Solidity: function multiDelegateCall(bytes[] data) returns(bytes[] results)
func (_Registry *RegistrySession) MultiDelegateCall(data [][]byte) (*types.Transaction, error) {
	return _Registry.Contract.MultiDelegateCall(&_Registry.TransactOpts, data)
}

// MultiDelegateCall is a paid mutator transaction binding the contract method 0x415c2d96.
//
// Solidity: function multiDelegateCall(bytes[] data) returns(bytes[] results)
func (_Registry *RegistryTransactorSession) MultiDelegateCall(data [][]byte) (*types.Transaction, error) {
	return _Registry.Contract.MultiDelegateCall(&_Registry.TransactOpts, data)
}

// PairAftermarketDeviceSign is a paid mutator transaction binding the contract method 0xb50df2f7.
//
// Solidity: function pairAftermarketDeviceSign(uint256 aftermarketDeviceNode, uint256 vehicleNode, bytes aftermarketDeviceSig, bytes vehicleOwnerSig) returns()
func (_Registry *RegistryTransactor) PairAftermarketDeviceSign(opts *bind.TransactOpts, aftermarketDeviceNode *big.Int, vehicleNode *big.Int, aftermarketDeviceSig []byte, vehicleOwnerSig []byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "pairAftermarketDeviceSign", aftermarketDeviceNode, vehicleNode, aftermarketDeviceSig, vehicleOwnerSig)
}

// PairAftermarketDeviceSign is a paid mutator transaction binding the contract method 0xb50df2f7.
//
// Solidity: function pairAftermarketDeviceSign(uint256 aftermarketDeviceNode, uint256 vehicleNode, bytes aftermarketDeviceSig, bytes vehicleOwnerSig) returns()
func (_Registry *RegistrySession) PairAftermarketDeviceSign(aftermarketDeviceNode *big.Int, vehicleNode *big.Int, aftermarketDeviceSig []byte, vehicleOwnerSig []byte) (*types.Transaction, error) {
	return _Registry.Contract.PairAftermarketDeviceSign(&_Registry.TransactOpts, aftermarketDeviceNode, vehicleNode, aftermarketDeviceSig, vehicleOwnerSig)
}

// PairAftermarketDeviceSign is a paid mutator transaction binding the contract method 0xb50df2f7.
//
// Solidity: function pairAftermarketDeviceSign(uint256 aftermarketDeviceNode, uint256 vehicleNode, bytes aftermarketDeviceSig, bytes vehicleOwnerSig) returns()
func (_Registry *RegistryTransactorSession) PairAftermarketDeviceSign(aftermarketDeviceNode *big.Int, vehicleNode *big.Int, aftermarketDeviceSig []byte, vehicleOwnerSig []byte) (*types.Transaction, error) {
	return _Registry.Contract.PairAftermarketDeviceSign(&_Registry.TransactOpts, aftermarketDeviceNode, vehicleNode, aftermarketDeviceSig, vehicleOwnerSig)
}

// PairAftermarketDeviceSign0 is a paid mutator transaction binding the contract method 0xcfe642dd.
//
// Solidity: function pairAftermarketDeviceSign(uint256 aftermarketDeviceNode, uint256 vehicleNode, bytes signature) returns()
func (_Registry *RegistryTransactor) PairAftermarketDeviceSign0(opts *bind.TransactOpts, aftermarketDeviceNode *big.Int, vehicleNode *big.Int, signature []byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "pairAftermarketDeviceSign0", aftermarketDeviceNode, vehicleNode, signature)
}

// PairAftermarketDeviceSign0 is a paid mutator transaction binding the contract method 0xcfe642dd.
//
// Solidity: function pairAftermarketDeviceSign(uint256 aftermarketDeviceNode, uint256 vehicleNode, bytes signature) returns()
func (_Registry *RegistrySession) PairAftermarketDeviceSign0(aftermarketDeviceNode *big.Int, vehicleNode *big.Int, signature []byte) (*types.Transaction, error) {
	return _Registry.Contract.PairAftermarketDeviceSign0(&_Registry.TransactOpts, aftermarketDeviceNode, vehicleNode, signature)
}

// PairAftermarketDeviceSign0 is a paid mutator transaction binding the contract method 0xcfe642dd.
//
// Solidity: function pairAftermarketDeviceSign(uint256 aftermarketDeviceNode, uint256 vehicleNode, bytes signature) returns()
func (_Registry *RegistryTransactorSession) PairAftermarketDeviceSign0(aftermarketDeviceNode *big.Int, vehicleNode *big.Int, signature []byte) (*types.Transaction, error) {
	return _Registry.Contract.PairAftermarketDeviceSign0(&_Registry.TransactOpts, aftermarketDeviceNode, vehicleNode, signature)
}

// RemoveModule is a paid mutator transaction binding the contract method 0x9748a762.
//
// Solidity: function removeModule(address implementation, bytes4[] selectors) returns()
func (_Registry *RegistryTransactor) RemoveModule(opts *bind.TransactOpts, implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "removeModule", implementation, selectors)
}

// RemoveModule is a paid mutator transaction binding the contract method 0x9748a762.
//
// Solidity: function removeModule(address implementation, bytes4[] selectors) returns()
func (_Registry *RegistrySession) RemoveModule(implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _Registry.Contract.RemoveModule(&_Registry.TransactOpts, implementation, selectors)
}

// RemoveModule is a paid mutator transaction binding the contract method 0x9748a762.
//
// Solidity: function removeModule(address implementation, bytes4[] selectors) returns()
func (_Registry *RegistryTransactorSession) RemoveModule(implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _Registry.Contract.RemoveModule(&_Registry.TransactOpts, implementation, selectors)
}

// RenameManufacturers is a paid mutator transaction binding the contract method 0xf73a8f04.
//
// Solidity: function renameManufacturers((uint256,string)[] idManufacturerNames) returns()
func (_Registry *RegistryTransactor) RenameManufacturers(opts *bind.TransactOpts, idManufacturerNames []DevAdminIdManufacturerName) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "renameManufacturers", idManufacturerNames)
}

// RenameManufacturers is a paid mutator transaction binding the contract method 0xf73a8f04.
//
// Solidity: function renameManufacturers((uint256,string)[] idManufacturerNames) returns()
func (_Registry *RegistrySession) RenameManufacturers(idManufacturerNames []DevAdminIdManufacturerName) (*types.Transaction, error) {
	return _Registry.Contract.RenameManufacturers(&_Registry.TransactOpts, idManufacturerNames)
}

// RenameManufacturers is a paid mutator transaction binding the contract method 0xf73a8f04.
//
// Solidity: function renameManufacturers((uint256,string)[] idManufacturerNames) returns()
func (_Registry *RegistryTransactorSession) RenameManufacturers(idManufacturerNames []DevAdminIdManufacturerName) (*types.Transaction, error) {
	return _Registry.Contract.RenameManufacturers(&_Registry.TransactOpts, idManufacturerNames)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x8bb9c5bf.
//
// Solidity: function renounceRole(bytes32 role) returns()
func (_Registry *RegistryTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "renounceRole", role)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x8bb9c5bf.
//
// Solidity: function renounceRole(bytes32 role) returns()
func (_Registry *RegistrySession) RenounceRole(role [32]byte) (*types.Transaction, error) {
	return _Registry.Contract.RenounceRole(&_Registry.TransactOpts, role)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x8bb9c5bf.
//
// Solidity: function renounceRole(bytes32 role) returns()
func (_Registry *RegistryTransactorSession) RenounceRole(role [32]byte) (*types.Transaction, error) {
	return _Registry.Contract.RenounceRole(&_Registry.TransactOpts, role)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Registry *RegistryTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Registry *RegistrySession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Registry.Contract.RevokeRole(&_Registry.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Registry *RegistryTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Registry.Contract.RevokeRole(&_Registry.TransactOpts, role, account)
}

// SetAdMintCost is a paid mutator transaction binding the contract method 0x2390baa8.
//
// Solidity: function setAdMintCost(uint256 _adMintCost) returns()
func (_Registry *RegistryTransactor) SetAdMintCost(opts *bind.TransactOpts, _adMintCost *big.Int) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setAdMintCost", _adMintCost)
}

// SetAdMintCost is a paid mutator transaction binding the contract method 0x2390baa8.
//
// Solidity: function setAdMintCost(uint256 _adMintCost) returns()
func (_Registry *RegistrySession) SetAdMintCost(_adMintCost *big.Int) (*types.Transaction, error) {
	return _Registry.Contract.SetAdMintCost(&_Registry.TransactOpts, _adMintCost)
}

// SetAdMintCost is a paid mutator transaction binding the contract method 0x2390baa8.
//
// Solidity: function setAdMintCost(uint256 _adMintCost) returns()
func (_Registry *RegistryTransactorSession) SetAdMintCost(_adMintCost *big.Int) (*types.Transaction, error) {
	return _Registry.Contract.SetAdMintCost(&_Registry.TransactOpts, _adMintCost)
}

// SetAftermarketDeviceIdProxyAddress is a paid mutator transaction binding the contract method 0x4d49d82a.
//
// Solidity: function setAftermarketDeviceIdProxyAddress(address addr) returns()
func (_Registry *RegistryTransactor) SetAftermarketDeviceIdProxyAddress(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setAftermarketDeviceIdProxyAddress", addr)
}

// SetAftermarketDeviceIdProxyAddress is a paid mutator transaction binding the contract method 0x4d49d82a.
//
// Solidity: function setAftermarketDeviceIdProxyAddress(address addr) returns()
func (_Registry *RegistrySession) SetAftermarketDeviceIdProxyAddress(addr common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetAftermarketDeviceIdProxyAddress(&_Registry.TransactOpts, addr)
}

// SetAftermarketDeviceIdProxyAddress is a paid mutator transaction binding the contract method 0x4d49d82a.
//
// Solidity: function setAftermarketDeviceIdProxyAddress(address addr) returns()
func (_Registry *RegistryTransactorSession) SetAftermarketDeviceIdProxyAddress(addr common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetAftermarketDeviceIdProxyAddress(&_Registry.TransactOpts, addr)
}

// SetAftermarketDeviceInfo is a paid mutator transaction binding the contract method 0x4d13b709.
//
// Solidity: function setAftermarketDeviceInfo(uint256 tokenId, (string,string)[] attrInfo) returns()
func (_Registry *RegistryTransactor) SetAftermarketDeviceInfo(opts *bind.TransactOpts, tokenId *big.Int, attrInfo []AttributeInfoPair) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setAftermarketDeviceInfo", tokenId, attrInfo)
}

// SetAftermarketDeviceInfo is a paid mutator transaction binding the contract method 0x4d13b709.
//
// Solidity: function setAftermarketDeviceInfo(uint256 tokenId, (string,string)[] attrInfo) returns()
func (_Registry *RegistrySession) SetAftermarketDeviceInfo(tokenId *big.Int, attrInfo []AttributeInfoPair) (*types.Transaction, error) {
	return _Registry.Contract.SetAftermarketDeviceInfo(&_Registry.TransactOpts, tokenId, attrInfo)
}

// SetAftermarketDeviceInfo is a paid mutator transaction binding the contract method 0x4d13b709.
//
// Solidity: function setAftermarketDeviceInfo(uint256 tokenId, (string,string)[] attrInfo) returns()
func (_Registry *RegistryTransactorSession) SetAftermarketDeviceInfo(tokenId *big.Int, attrInfo []AttributeInfoPair) (*types.Transaction, error) {
	return _Registry.Contract.SetAftermarketDeviceInfo(&_Registry.TransactOpts, tokenId, attrInfo)
}

// SetController is a paid mutator transaction binding the contract method 0x92eefe9b.
//
// Solidity: function setController(address _controller) returns()
func (_Registry *RegistryTransactor) SetController(opts *bind.TransactOpts, _controller common.Address) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setController", _controller)
}

// SetController is a paid mutator transaction binding the contract method 0x92eefe9b.
//
// Solidity: function setController(address _controller) returns()
func (_Registry *RegistrySession) SetController(_controller common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetController(&_Registry.TransactOpts, _controller)
}

// SetController is a paid mutator transaction binding the contract method 0x92eefe9b.
//
// Solidity: function setController(address _controller) returns()
func (_Registry *RegistryTransactorSession) SetController(_controller common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetController(&_Registry.TransactOpts, _controller)
}

// SetDimoToken is a paid mutator transaction binding the contract method 0x5b6c1979.
//
// Solidity: function setDimoToken(address _dimoToken) returns()
func (_Registry *RegistryTransactor) SetDimoToken(opts *bind.TransactOpts, _dimoToken common.Address) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setDimoToken", _dimoToken)
}

// SetDimoToken is a paid mutator transaction binding the contract method 0x5b6c1979.
//
// Solidity: function setDimoToken(address _dimoToken) returns()
func (_Registry *RegistrySession) SetDimoToken(_dimoToken common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetDimoToken(&_Registry.TransactOpts, _dimoToken)
}

// SetDimoToken is a paid mutator transaction binding the contract method 0x5b6c1979.
//
// Solidity: function setDimoToken(address _dimoToken) returns()
func (_Registry *RegistryTransactorSession) SetDimoToken(_dimoToken common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetDimoToken(&_Registry.TransactOpts, _dimoToken)
}

// SetFoundationAddress is a paid mutator transaction binding the contract method 0xf41377ca.
//
// Solidity: function setFoundationAddress(address _foundation) returns()
func (_Registry *RegistryTransactor) SetFoundationAddress(opts *bind.TransactOpts, _foundation common.Address) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setFoundationAddress", _foundation)
}

// SetFoundationAddress is a paid mutator transaction binding the contract method 0xf41377ca.
//
// Solidity: function setFoundationAddress(address _foundation) returns()
func (_Registry *RegistrySession) SetFoundationAddress(_foundation common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetFoundationAddress(&_Registry.TransactOpts, _foundation)
}

// SetFoundationAddress is a paid mutator transaction binding the contract method 0xf41377ca.
//
// Solidity: function setFoundationAddress(address _foundation) returns()
func (_Registry *RegistryTransactorSession) SetFoundationAddress(_foundation common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetFoundationAddress(&_Registry.TransactOpts, _foundation)
}

// SetLicense is a paid mutator transaction binding the contract method 0x0fd21c17.
//
// Solidity: function setLicense(address _license) returns()
func (_Registry *RegistryTransactor) SetLicense(opts *bind.TransactOpts, _license common.Address) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setLicense", _license)
}

// SetLicense is a paid mutator transaction binding the contract method 0x0fd21c17.
//
// Solidity: function setLicense(address _license) returns()
func (_Registry *RegistrySession) SetLicense(_license common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetLicense(&_Registry.TransactOpts, _license)
}

// SetLicense is a paid mutator transaction binding the contract method 0x0fd21c17.
//
// Solidity: function setLicense(address _license) returns()
func (_Registry *RegistryTransactorSession) SetLicense(_license common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetLicense(&_Registry.TransactOpts, _license)
}

// SetManufacturerIdProxyAddress is a paid mutator transaction binding the contract method 0xd159f49a.
//
// Solidity: function setManufacturerIdProxyAddress(address addr) returns()
func (_Registry *RegistryTransactor) SetManufacturerIdProxyAddress(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setManufacturerIdProxyAddress", addr)
}

// SetManufacturerIdProxyAddress is a paid mutator transaction binding the contract method 0xd159f49a.
//
// Solidity: function setManufacturerIdProxyAddress(address addr) returns()
func (_Registry *RegistrySession) SetManufacturerIdProxyAddress(addr common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetManufacturerIdProxyAddress(&_Registry.TransactOpts, addr)
}

// SetManufacturerIdProxyAddress is a paid mutator transaction binding the contract method 0xd159f49a.
//
// Solidity: function setManufacturerIdProxyAddress(address addr) returns()
func (_Registry *RegistryTransactorSession) SetManufacturerIdProxyAddress(addr common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetManufacturerIdProxyAddress(&_Registry.TransactOpts, addr)
}

// SetManufacturerInfo is a paid mutator transaction binding the contract method 0x63545ffa.
//
// Solidity: function setManufacturerInfo(uint256 tokenId, (string,string)[] attrInfoList) returns()
func (_Registry *RegistryTransactor) SetManufacturerInfo(opts *bind.TransactOpts, tokenId *big.Int, attrInfoList []AttributeInfoPair) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setManufacturerInfo", tokenId, attrInfoList)
}

// SetManufacturerInfo is a paid mutator transaction binding the contract method 0x63545ffa.
//
// Solidity: function setManufacturerInfo(uint256 tokenId, (string,string)[] attrInfoList) returns()
func (_Registry *RegistrySession) SetManufacturerInfo(tokenId *big.Int, attrInfoList []AttributeInfoPair) (*types.Transaction, error) {
	return _Registry.Contract.SetManufacturerInfo(&_Registry.TransactOpts, tokenId, attrInfoList)
}

// SetManufacturerInfo is a paid mutator transaction binding the contract method 0x63545ffa.
//
// Solidity: function setManufacturerInfo(uint256 tokenId, (string,string)[] attrInfoList) returns()
func (_Registry *RegistryTransactorSession) SetManufacturerInfo(tokenId *big.Int, attrInfoList []AttributeInfoPair) (*types.Transaction, error) {
	return _Registry.Contract.SetManufacturerInfo(&_Registry.TransactOpts, tokenId, attrInfoList)
}

// SetManufacturerMinted is a paid mutator transaction binding the contract method 0xa84ac57f.
//
// Solidity: function setManufacturerMinted(address addr) returns()
func (_Registry *RegistryTransactor) SetManufacturerMinted(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setManufacturerMinted", addr)
}

// SetManufacturerMinted is a paid mutator transaction binding the contract method 0xa84ac57f.
//
// Solidity: function setManufacturerMinted(address addr) returns()
func (_Registry *RegistrySession) SetManufacturerMinted(addr common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetManufacturerMinted(&_Registry.TransactOpts, addr)
}

// SetManufacturerMinted is a paid mutator transaction binding the contract method 0xa84ac57f.
//
// Solidity: function setManufacturerMinted(address addr) returns()
func (_Registry *RegistryTransactorSession) SetManufacturerMinted(addr common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetManufacturerMinted(&_Registry.TransactOpts, addr)
}

// SetVehicleIdProxyAddress is a paid mutator transaction binding the contract method 0x9bfae6da.
//
// Solidity: function setVehicleIdProxyAddress(address addr) returns()
func (_Registry *RegistryTransactor) SetVehicleIdProxyAddress(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setVehicleIdProxyAddress", addr)
}

// SetVehicleIdProxyAddress is a paid mutator transaction binding the contract method 0x9bfae6da.
//
// Solidity: function setVehicleIdProxyAddress(address addr) returns()
func (_Registry *RegistrySession) SetVehicleIdProxyAddress(addr common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetVehicleIdProxyAddress(&_Registry.TransactOpts, addr)
}

// SetVehicleIdProxyAddress is a paid mutator transaction binding the contract method 0x9bfae6da.
//
// Solidity: function setVehicleIdProxyAddress(address addr) returns()
func (_Registry *RegistryTransactorSession) SetVehicleIdProxyAddress(addr common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetVehicleIdProxyAddress(&_Registry.TransactOpts, addr)
}

// SetVehicleInfo is a paid mutator transaction binding the contract method 0xd9c3ae61.
//
// Solidity: function setVehicleInfo(uint256 tokenId, (string,string)[] attrInfo) returns()
func (_Registry *RegistryTransactor) SetVehicleInfo(opts *bind.TransactOpts, tokenId *big.Int, attrInfo []AttributeInfoPair) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setVehicleInfo", tokenId, attrInfo)
}

// SetVehicleInfo is a paid mutator transaction binding the contract method 0xd9c3ae61.
//
// Solidity: function setVehicleInfo(uint256 tokenId, (string,string)[] attrInfo) returns()
func (_Registry *RegistrySession) SetVehicleInfo(tokenId *big.Int, attrInfo []AttributeInfoPair) (*types.Transaction, error) {
	return _Registry.Contract.SetVehicleInfo(&_Registry.TransactOpts, tokenId, attrInfo)
}

// SetVehicleInfo is a paid mutator transaction binding the contract method 0xd9c3ae61.
//
// Solidity: function setVehicleInfo(uint256 tokenId, (string,string)[] attrInfo) returns()
func (_Registry *RegistryTransactorSession) SetVehicleInfo(tokenId *big.Int, attrInfo []AttributeInfoPair) (*types.Transaction, error) {
	return _Registry.Contract.SetVehicleInfo(&_Registry.TransactOpts, tokenId, attrInfo)
}

// TransferAftermarketDeviceOwnership is a paid mutator transaction binding the contract method 0xff96b761.
//
// Solidity: function transferAftermarketDeviceOwnership(uint256 aftermarketDeviceNode, address newOwner) returns()
func (_Registry *RegistryTransactor) TransferAftermarketDeviceOwnership(opts *bind.TransactOpts, aftermarketDeviceNode *big.Int, newOwner common.Address) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "transferAftermarketDeviceOwnership", aftermarketDeviceNode, newOwner)
}

// TransferAftermarketDeviceOwnership is a paid mutator transaction binding the contract method 0xff96b761.
//
// Solidity: function transferAftermarketDeviceOwnership(uint256 aftermarketDeviceNode, address newOwner) returns()
func (_Registry *RegistrySession) TransferAftermarketDeviceOwnership(aftermarketDeviceNode *big.Int, newOwner common.Address) (*types.Transaction, error) {
	return _Registry.Contract.TransferAftermarketDeviceOwnership(&_Registry.TransactOpts, aftermarketDeviceNode, newOwner)
}

// TransferAftermarketDeviceOwnership is a paid mutator transaction binding the contract method 0xff96b761.
//
// Solidity: function transferAftermarketDeviceOwnership(uint256 aftermarketDeviceNode, address newOwner) returns()
func (_Registry *RegistryTransactorSession) TransferAftermarketDeviceOwnership(aftermarketDeviceNode *big.Int, newOwner common.Address) (*types.Transaction, error) {
	return _Registry.Contract.TransferAftermarketDeviceOwnership(&_Registry.TransactOpts, aftermarketDeviceNode, newOwner)
}

// UnclaimAftermarketDeviceNode is a paid mutator transaction binding the contract method 0x5c129493.
//
// Solidity: function unclaimAftermarketDeviceNode(uint256[] aftermarketDeviceNodes) returns()
func (_Registry *RegistryTransactor) UnclaimAftermarketDeviceNode(opts *bind.TransactOpts, aftermarketDeviceNodes []*big.Int) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "unclaimAftermarketDeviceNode", aftermarketDeviceNodes)
}

// UnclaimAftermarketDeviceNode is a paid mutator transaction binding the contract method 0x5c129493.
//
// Solidity: function unclaimAftermarketDeviceNode(uint256[] aftermarketDeviceNodes) returns()
func (_Registry *RegistrySession) UnclaimAftermarketDeviceNode(aftermarketDeviceNodes []*big.Int) (*types.Transaction, error) {
	return _Registry.Contract.UnclaimAftermarketDeviceNode(&_Registry.TransactOpts, aftermarketDeviceNodes)
}

// UnclaimAftermarketDeviceNode is a paid mutator transaction binding the contract method 0x5c129493.
//
// Solidity: function unclaimAftermarketDeviceNode(uint256[] aftermarketDeviceNodes) returns()
func (_Registry *RegistryTransactorSession) UnclaimAftermarketDeviceNode(aftermarketDeviceNodes []*big.Int) (*types.Transaction, error) {
	return _Registry.Contract.UnclaimAftermarketDeviceNode(&_Registry.TransactOpts, aftermarketDeviceNodes)
}

// UnpairAftermarketDeviceByDeviceNode is a paid mutator transaction binding the contract method 0x71193956.
//
// Solidity: function unpairAftermarketDeviceByDeviceNode(uint256[] aftermarketDeviceNodes) returns()
func (_Registry *RegistryTransactor) UnpairAftermarketDeviceByDeviceNode(opts *bind.TransactOpts, aftermarketDeviceNodes []*big.Int) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "unpairAftermarketDeviceByDeviceNode", aftermarketDeviceNodes)
}

// UnpairAftermarketDeviceByDeviceNode is a paid mutator transaction binding the contract method 0x71193956.
//
// Solidity: function unpairAftermarketDeviceByDeviceNode(uint256[] aftermarketDeviceNodes) returns()
func (_Registry *RegistrySession) UnpairAftermarketDeviceByDeviceNode(aftermarketDeviceNodes []*big.Int) (*types.Transaction, error) {
	return _Registry.Contract.UnpairAftermarketDeviceByDeviceNode(&_Registry.TransactOpts, aftermarketDeviceNodes)
}

// UnpairAftermarketDeviceByDeviceNode is a paid mutator transaction binding the contract method 0x71193956.
//
// Solidity: function unpairAftermarketDeviceByDeviceNode(uint256[] aftermarketDeviceNodes) returns()
func (_Registry *RegistryTransactorSession) UnpairAftermarketDeviceByDeviceNode(aftermarketDeviceNodes []*big.Int) (*types.Transaction, error) {
	return _Registry.Contract.UnpairAftermarketDeviceByDeviceNode(&_Registry.TransactOpts, aftermarketDeviceNodes)
}

// UnpairAftermarketDeviceByVehicleNode is a paid mutator transaction binding the contract method 0x8c2ee9bb.
//
// Solidity: function unpairAftermarketDeviceByVehicleNode(uint256[] vehicleNodes) returns()
func (_Registry *RegistryTransactor) UnpairAftermarketDeviceByVehicleNode(opts *bind.TransactOpts, vehicleNodes []*big.Int) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "unpairAftermarketDeviceByVehicleNode", vehicleNodes)
}

// UnpairAftermarketDeviceByVehicleNode is a paid mutator transaction binding the contract method 0x8c2ee9bb.
//
// Solidity: function unpairAftermarketDeviceByVehicleNode(uint256[] vehicleNodes) returns()
func (_Registry *RegistrySession) UnpairAftermarketDeviceByVehicleNode(vehicleNodes []*big.Int) (*types.Transaction, error) {
	return _Registry.Contract.UnpairAftermarketDeviceByVehicleNode(&_Registry.TransactOpts, vehicleNodes)
}

// UnpairAftermarketDeviceByVehicleNode is a paid mutator transaction binding the contract method 0x8c2ee9bb.
//
// Solidity: function unpairAftermarketDeviceByVehicleNode(uint256[] vehicleNodes) returns()
func (_Registry *RegistryTransactorSession) UnpairAftermarketDeviceByVehicleNode(vehicleNodes []*big.Int) (*types.Transaction, error) {
	return _Registry.Contract.UnpairAftermarketDeviceByVehicleNode(&_Registry.TransactOpts, vehicleNodes)
}

// UnpairAftermarketDeviceSign is a paid mutator transaction binding the contract method 0x3f65997a.
//
// Solidity: function unpairAftermarketDeviceSign(uint256 aftermarketDeviceNode, uint256 vehicleNode, bytes signature) returns()
func (_Registry *RegistryTransactor) UnpairAftermarketDeviceSign(opts *bind.TransactOpts, aftermarketDeviceNode *big.Int, vehicleNode *big.Int, signature []byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "unpairAftermarketDeviceSign", aftermarketDeviceNode, vehicleNode, signature)
}

// UnpairAftermarketDeviceSign is a paid mutator transaction binding the contract method 0x3f65997a.
//
// Solidity: function unpairAftermarketDeviceSign(uint256 aftermarketDeviceNode, uint256 vehicleNode, bytes signature) returns()
func (_Registry *RegistrySession) UnpairAftermarketDeviceSign(aftermarketDeviceNode *big.Int, vehicleNode *big.Int, signature []byte) (*types.Transaction, error) {
	return _Registry.Contract.UnpairAftermarketDeviceSign(&_Registry.TransactOpts, aftermarketDeviceNode, vehicleNode, signature)
}

// UnpairAftermarketDeviceSign is a paid mutator transaction binding the contract method 0x3f65997a.
//
// Solidity: function unpairAftermarketDeviceSign(uint256 aftermarketDeviceNode, uint256 vehicleNode, bytes signature) returns()
func (_Registry *RegistryTransactorSession) UnpairAftermarketDeviceSign(aftermarketDeviceNode *big.Int, vehicleNode *big.Int, signature []byte) (*types.Transaction, error) {
	return _Registry.Contract.UnpairAftermarketDeviceSign(&_Registry.TransactOpts, aftermarketDeviceNode, vehicleNode, signature)
}

// UpdateModule is a paid mutator transaction binding the contract method 0x06d1d2a1.
//
// Solidity: function updateModule(address oldImplementation, address newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors) returns()
func (_Registry *RegistryTransactor) UpdateModule(opts *bind.TransactOpts, oldImplementation common.Address, newImplementation common.Address, oldSelectors [][4]byte, newSelectors [][4]byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "updateModule", oldImplementation, newImplementation, oldSelectors, newSelectors)
}

// UpdateModule is a paid mutator transaction binding the contract method 0x06d1d2a1.
//
// Solidity: function updateModule(address oldImplementation, address newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors) returns()
func (_Registry *RegistrySession) UpdateModule(oldImplementation common.Address, newImplementation common.Address, oldSelectors [][4]byte, newSelectors [][4]byte) (*types.Transaction, error) {
	return _Registry.Contract.UpdateModule(&_Registry.TransactOpts, oldImplementation, newImplementation, oldSelectors, newSelectors)
}

// UpdateModule is a paid mutator transaction binding the contract method 0x06d1d2a1.
//
// Solidity: function updateModule(address oldImplementation, address newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors) returns()
func (_Registry *RegistryTransactorSession) UpdateModule(oldImplementation common.Address, newImplementation common.Address, oldSelectors [][4]byte, newSelectors [][4]byte) (*types.Transaction, error) {
	return _Registry.Contract.UpdateModule(&_Registry.TransactOpts, oldImplementation, newImplementation, oldSelectors, newSelectors)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_Registry *RegistryTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _Registry.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_Registry *RegistrySession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Registry.Contract.Fallback(&_Registry.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_Registry *RegistryTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Registry.Contract.Fallback(&_Registry.TransactOpts, calldata)
}

// RegistryAftermarketDeviceAttributeAddedIterator is returned from FilterAftermarketDeviceAttributeAdded and is used to iterate over the raw logs and unpacked data for AftermarketDeviceAttributeAdded events raised by the Registry contract.
type RegistryAftermarketDeviceAttributeAddedIterator struct {
	Event *RegistryAftermarketDeviceAttributeAdded // Event containing the contract specifics and raw log

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
func (it *RegistryAftermarketDeviceAttributeAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryAftermarketDeviceAttributeAdded)
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
		it.Event = new(RegistryAftermarketDeviceAttributeAdded)
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
func (it *RegistryAftermarketDeviceAttributeAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryAftermarketDeviceAttributeAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryAftermarketDeviceAttributeAdded represents a AftermarketDeviceAttributeAdded event raised by the Registry contract.
type RegistryAftermarketDeviceAttributeAdded struct {
	Attribute string
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterAftermarketDeviceAttributeAdded is a free log retrieval operation binding the contract event 0x3ef2473cbfb66e153539befafe6ba95e95c6cc0659ebc0d7e8a56f014de7eb5f.
//
// Solidity: event AftermarketDeviceAttributeAdded(string attribute)
func (_Registry *RegistryFilterer) FilterAftermarketDeviceAttributeAdded(opts *bind.FilterOpts) (*RegistryAftermarketDeviceAttributeAddedIterator, error) {

	logs, sub, err := _Registry.contract.FilterLogs(opts, "AftermarketDeviceAttributeAdded")
	if err != nil {
		return nil, err
	}
	return &RegistryAftermarketDeviceAttributeAddedIterator{contract: _Registry.contract, event: "AftermarketDeviceAttributeAdded", logs: logs, sub: sub}, nil
}

// WatchAftermarketDeviceAttributeAdded is a free log subscription operation binding the contract event 0x3ef2473cbfb66e153539befafe6ba95e95c6cc0659ebc0d7e8a56f014de7eb5f.
//
// Solidity: event AftermarketDeviceAttributeAdded(string attribute)
func (_Registry *RegistryFilterer) WatchAftermarketDeviceAttributeAdded(opts *bind.WatchOpts, sink chan<- *RegistryAftermarketDeviceAttributeAdded) (event.Subscription, error) {

	logs, sub, err := _Registry.contract.WatchLogs(opts, "AftermarketDeviceAttributeAdded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryAftermarketDeviceAttributeAdded)
				if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceAttributeAdded", log); err != nil {
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

// ParseAftermarketDeviceAttributeAdded is a log parse operation binding the contract event 0x3ef2473cbfb66e153539befafe6ba95e95c6cc0659ebc0d7e8a56f014de7eb5f.
//
// Solidity: event AftermarketDeviceAttributeAdded(string attribute)
func (_Registry *RegistryFilterer) ParseAftermarketDeviceAttributeAdded(log types.Log) (*RegistryAftermarketDeviceAttributeAdded, error) {
	event := new(RegistryAftermarketDeviceAttributeAdded)
	if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceAttributeAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryAftermarketDeviceAttributeSetIterator is returned from FilterAftermarketDeviceAttributeSet and is used to iterate over the raw logs and unpacked data for AftermarketDeviceAttributeSet events raised by the Registry contract.
type RegistryAftermarketDeviceAttributeSetIterator struct {
	Event *RegistryAftermarketDeviceAttributeSet // Event containing the contract specifics and raw log

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
func (it *RegistryAftermarketDeviceAttributeSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryAftermarketDeviceAttributeSet)
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
		it.Event = new(RegistryAftermarketDeviceAttributeSet)
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
func (it *RegistryAftermarketDeviceAttributeSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryAftermarketDeviceAttributeSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryAftermarketDeviceAttributeSet represents a AftermarketDeviceAttributeSet event raised by the Registry contract.
type RegistryAftermarketDeviceAttributeSet struct {
	TokenId   *big.Int
	Attribute string
	Info      string
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterAftermarketDeviceAttributeSet is a free log retrieval operation binding the contract event 0x977fe0ddf8485988af0b93d70bf5977b48236e9969cdb9b1f55977fbab7cd417.
//
// Solidity: event AftermarketDeviceAttributeSet(uint256 tokenId, string attribute, string info)
func (_Registry *RegistryFilterer) FilterAftermarketDeviceAttributeSet(opts *bind.FilterOpts) (*RegistryAftermarketDeviceAttributeSetIterator, error) {

	logs, sub, err := _Registry.contract.FilterLogs(opts, "AftermarketDeviceAttributeSet")
	if err != nil {
		return nil, err
	}
	return &RegistryAftermarketDeviceAttributeSetIterator{contract: _Registry.contract, event: "AftermarketDeviceAttributeSet", logs: logs, sub: sub}, nil
}

// WatchAftermarketDeviceAttributeSet is a free log subscription operation binding the contract event 0x977fe0ddf8485988af0b93d70bf5977b48236e9969cdb9b1f55977fbab7cd417.
//
// Solidity: event AftermarketDeviceAttributeSet(uint256 tokenId, string attribute, string info)
func (_Registry *RegistryFilterer) WatchAftermarketDeviceAttributeSet(opts *bind.WatchOpts, sink chan<- *RegistryAftermarketDeviceAttributeSet) (event.Subscription, error) {

	logs, sub, err := _Registry.contract.WatchLogs(opts, "AftermarketDeviceAttributeSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryAftermarketDeviceAttributeSet)
				if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceAttributeSet", log); err != nil {
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

// ParseAftermarketDeviceAttributeSet is a log parse operation binding the contract event 0x977fe0ddf8485988af0b93d70bf5977b48236e9969cdb9b1f55977fbab7cd417.
//
// Solidity: event AftermarketDeviceAttributeSet(uint256 tokenId, string attribute, string info)
func (_Registry *RegistryFilterer) ParseAftermarketDeviceAttributeSet(log types.Log) (*RegistryAftermarketDeviceAttributeSet, error) {
	event := new(RegistryAftermarketDeviceAttributeSet)
	if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceAttributeSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryAftermarketDeviceClaimedIterator is returned from FilterAftermarketDeviceClaimed and is used to iterate over the raw logs and unpacked data for AftermarketDeviceClaimed events raised by the Registry contract.
type RegistryAftermarketDeviceClaimedIterator struct {
	Event *RegistryAftermarketDeviceClaimed // Event containing the contract specifics and raw log

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
func (it *RegistryAftermarketDeviceClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryAftermarketDeviceClaimed)
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
		it.Event = new(RegistryAftermarketDeviceClaimed)
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
func (it *RegistryAftermarketDeviceClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryAftermarketDeviceClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryAftermarketDeviceClaimed represents a AftermarketDeviceClaimed event raised by the Registry contract.
type RegistryAftermarketDeviceClaimed struct {
	AftermarketDeviceNode *big.Int
	Owner                 common.Address
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterAftermarketDeviceClaimed is a free log retrieval operation binding the contract event 0x8468d811e5090d3b1a07e28af524e66c128f624e16b07638f419012c779f76ec.
//
// Solidity: event AftermarketDeviceClaimed(uint256 aftermarketDeviceNode, address indexed owner)
func (_Registry *RegistryFilterer) FilterAftermarketDeviceClaimed(opts *bind.FilterOpts, owner []common.Address) (*RegistryAftermarketDeviceClaimedIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "AftermarketDeviceClaimed", ownerRule)
	if err != nil {
		return nil, err
	}
	return &RegistryAftermarketDeviceClaimedIterator{contract: _Registry.contract, event: "AftermarketDeviceClaimed", logs: logs, sub: sub}, nil
}

// WatchAftermarketDeviceClaimed is a free log subscription operation binding the contract event 0x8468d811e5090d3b1a07e28af524e66c128f624e16b07638f419012c779f76ec.
//
// Solidity: event AftermarketDeviceClaimed(uint256 aftermarketDeviceNode, address indexed owner)
func (_Registry *RegistryFilterer) WatchAftermarketDeviceClaimed(opts *bind.WatchOpts, sink chan<- *RegistryAftermarketDeviceClaimed, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "AftermarketDeviceClaimed", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryAftermarketDeviceClaimed)
				if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceClaimed", log); err != nil {
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

// ParseAftermarketDeviceClaimed is a log parse operation binding the contract event 0x8468d811e5090d3b1a07e28af524e66c128f624e16b07638f419012c779f76ec.
//
// Solidity: event AftermarketDeviceClaimed(uint256 aftermarketDeviceNode, address indexed owner)
func (_Registry *RegistryFilterer) ParseAftermarketDeviceClaimed(log types.Log) (*RegistryAftermarketDeviceClaimed, error) {
	event := new(RegistryAftermarketDeviceClaimed)
	if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceClaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryAftermarketDeviceIdProxySetIterator is returned from FilterAftermarketDeviceIdProxySet and is used to iterate over the raw logs and unpacked data for AftermarketDeviceIdProxySet events raised by the Registry contract.
type RegistryAftermarketDeviceIdProxySetIterator struct {
	Event *RegistryAftermarketDeviceIdProxySet // Event containing the contract specifics and raw log

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
func (it *RegistryAftermarketDeviceIdProxySetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryAftermarketDeviceIdProxySet)
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
		it.Event = new(RegistryAftermarketDeviceIdProxySet)
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
func (it *RegistryAftermarketDeviceIdProxySetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryAftermarketDeviceIdProxySetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryAftermarketDeviceIdProxySet represents a AftermarketDeviceIdProxySet event raised by the Registry contract.
type RegistryAftermarketDeviceIdProxySet struct {
	Proxy common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterAftermarketDeviceIdProxySet is a free log retrieval operation binding the contract event 0xe2daa727eb82f2761802221c26f72d54501ca8abd6da081e50fedaaab21f4036.
//
// Solidity: event AftermarketDeviceIdProxySet(address indexed proxy)
func (_Registry *RegistryFilterer) FilterAftermarketDeviceIdProxySet(opts *bind.FilterOpts, proxy []common.Address) (*RegistryAftermarketDeviceIdProxySetIterator, error) {

	var proxyRule []interface{}
	for _, proxyItem := range proxy {
		proxyRule = append(proxyRule, proxyItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "AftermarketDeviceIdProxySet", proxyRule)
	if err != nil {
		return nil, err
	}
	return &RegistryAftermarketDeviceIdProxySetIterator{contract: _Registry.contract, event: "AftermarketDeviceIdProxySet", logs: logs, sub: sub}, nil
}

// WatchAftermarketDeviceIdProxySet is a free log subscription operation binding the contract event 0xe2daa727eb82f2761802221c26f72d54501ca8abd6da081e50fedaaab21f4036.
//
// Solidity: event AftermarketDeviceIdProxySet(address indexed proxy)
func (_Registry *RegistryFilterer) WatchAftermarketDeviceIdProxySet(opts *bind.WatchOpts, sink chan<- *RegistryAftermarketDeviceIdProxySet, proxy []common.Address) (event.Subscription, error) {

	var proxyRule []interface{}
	for _, proxyItem := range proxy {
		proxyRule = append(proxyRule, proxyItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "AftermarketDeviceIdProxySet", proxyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryAftermarketDeviceIdProxySet)
				if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceIdProxySet", log); err != nil {
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

// ParseAftermarketDeviceIdProxySet is a log parse operation binding the contract event 0xe2daa727eb82f2761802221c26f72d54501ca8abd6da081e50fedaaab21f4036.
//
// Solidity: event AftermarketDeviceIdProxySet(address indexed proxy)
func (_Registry *RegistryFilterer) ParseAftermarketDeviceIdProxySet(log types.Log) (*RegistryAftermarketDeviceIdProxySet, error) {
	event := new(RegistryAftermarketDeviceIdProxySet)
	if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceIdProxySet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryAftermarketDeviceNodeMintedIterator is returned from FilterAftermarketDeviceNodeMinted and is used to iterate over the raw logs and unpacked data for AftermarketDeviceNodeMinted events raised by the Registry contract.
type RegistryAftermarketDeviceNodeMintedIterator struct {
	Event *RegistryAftermarketDeviceNodeMinted // Event containing the contract specifics and raw log

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
func (it *RegistryAftermarketDeviceNodeMintedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryAftermarketDeviceNodeMinted)
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
		it.Event = new(RegistryAftermarketDeviceNodeMinted)
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
func (it *RegistryAftermarketDeviceNodeMintedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryAftermarketDeviceNodeMintedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryAftermarketDeviceNodeMinted represents a AftermarketDeviceNodeMinted event raised by the Registry contract.
type RegistryAftermarketDeviceNodeMinted struct {
	TokenId                  *big.Int
	AftermarketDeviceAddress common.Address
	Owner                    common.Address
	Raw                      types.Log // Blockchain specific contextual infos
}

// FilterAftermarketDeviceNodeMinted is a free log retrieval operation binding the contract event 0x476b4c986061e7e4f957a5753096dea5ff957cca8bee688c4085a85c7590dfe3.
//
// Solidity: event AftermarketDeviceNodeMinted(uint256 tokenId, address indexed aftermarketDeviceAddress, address indexed owner)
func (_Registry *RegistryFilterer) FilterAftermarketDeviceNodeMinted(opts *bind.FilterOpts, aftermarketDeviceAddress []common.Address, owner []common.Address) (*RegistryAftermarketDeviceNodeMintedIterator, error) {

	var aftermarketDeviceAddressRule []interface{}
	for _, aftermarketDeviceAddressItem := range aftermarketDeviceAddress {
		aftermarketDeviceAddressRule = append(aftermarketDeviceAddressRule, aftermarketDeviceAddressItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "AftermarketDeviceNodeMinted", aftermarketDeviceAddressRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &RegistryAftermarketDeviceNodeMintedIterator{contract: _Registry.contract, event: "AftermarketDeviceNodeMinted", logs: logs, sub: sub}, nil
}

// WatchAftermarketDeviceNodeMinted is a free log subscription operation binding the contract event 0x476b4c986061e7e4f957a5753096dea5ff957cca8bee688c4085a85c7590dfe3.
//
// Solidity: event AftermarketDeviceNodeMinted(uint256 tokenId, address indexed aftermarketDeviceAddress, address indexed owner)
func (_Registry *RegistryFilterer) WatchAftermarketDeviceNodeMinted(opts *bind.WatchOpts, sink chan<- *RegistryAftermarketDeviceNodeMinted, aftermarketDeviceAddress []common.Address, owner []common.Address) (event.Subscription, error) {

	var aftermarketDeviceAddressRule []interface{}
	for _, aftermarketDeviceAddressItem := range aftermarketDeviceAddress {
		aftermarketDeviceAddressRule = append(aftermarketDeviceAddressRule, aftermarketDeviceAddressItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "AftermarketDeviceNodeMinted", aftermarketDeviceAddressRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryAftermarketDeviceNodeMinted)
				if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceNodeMinted", log); err != nil {
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

// ParseAftermarketDeviceNodeMinted is a log parse operation binding the contract event 0x476b4c986061e7e4f957a5753096dea5ff957cca8bee688c4085a85c7590dfe3.
//
// Solidity: event AftermarketDeviceNodeMinted(uint256 tokenId, address indexed aftermarketDeviceAddress, address indexed owner)
func (_Registry *RegistryFilterer) ParseAftermarketDeviceNodeMinted(log types.Log) (*RegistryAftermarketDeviceNodeMinted, error) {
	event := new(RegistryAftermarketDeviceNodeMinted)
	if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceNodeMinted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryAftermarketDevicePairedIterator is returned from FilterAftermarketDevicePaired and is used to iterate over the raw logs and unpacked data for AftermarketDevicePaired events raised by the Registry contract.
type RegistryAftermarketDevicePairedIterator struct {
	Event *RegistryAftermarketDevicePaired // Event containing the contract specifics and raw log

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
func (it *RegistryAftermarketDevicePairedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryAftermarketDevicePaired)
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
		it.Event = new(RegistryAftermarketDevicePaired)
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
func (it *RegistryAftermarketDevicePairedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryAftermarketDevicePairedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryAftermarketDevicePaired represents a AftermarketDevicePaired event raised by the Registry contract.
type RegistryAftermarketDevicePaired struct {
	AftermarketDeviceNode *big.Int
	VehicleNode           *big.Int
	Owner                 common.Address
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterAftermarketDevicePaired is a free log retrieval operation binding the contract event 0x89ec132808bbf01af00b90fd34e04fd6cfb8dba2813ca5446a415500b83c7938.
//
// Solidity: event AftermarketDevicePaired(uint256 aftermarketDeviceNode, uint256 vehicleNode, address indexed owner)
func (_Registry *RegistryFilterer) FilterAftermarketDevicePaired(opts *bind.FilterOpts, owner []common.Address) (*RegistryAftermarketDevicePairedIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "AftermarketDevicePaired", ownerRule)
	if err != nil {
		return nil, err
	}
	return &RegistryAftermarketDevicePairedIterator{contract: _Registry.contract, event: "AftermarketDevicePaired", logs: logs, sub: sub}, nil
}

// WatchAftermarketDevicePaired is a free log subscription operation binding the contract event 0x89ec132808bbf01af00b90fd34e04fd6cfb8dba2813ca5446a415500b83c7938.
//
// Solidity: event AftermarketDevicePaired(uint256 aftermarketDeviceNode, uint256 vehicleNode, address indexed owner)
func (_Registry *RegistryFilterer) WatchAftermarketDevicePaired(opts *bind.WatchOpts, sink chan<- *RegistryAftermarketDevicePaired, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "AftermarketDevicePaired", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryAftermarketDevicePaired)
				if err := _Registry.contract.UnpackLog(event, "AftermarketDevicePaired", log); err != nil {
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

// ParseAftermarketDevicePaired is a log parse operation binding the contract event 0x89ec132808bbf01af00b90fd34e04fd6cfb8dba2813ca5446a415500b83c7938.
//
// Solidity: event AftermarketDevicePaired(uint256 aftermarketDeviceNode, uint256 vehicleNode, address indexed owner)
func (_Registry *RegistryFilterer) ParseAftermarketDevicePaired(log types.Log) (*RegistryAftermarketDevicePaired, error) {
	event := new(RegistryAftermarketDevicePaired)
	if err := _Registry.contract.UnpackLog(event, "AftermarketDevicePaired", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryAftermarketDeviceTransferredIterator is returned from FilterAftermarketDeviceTransferred and is used to iterate over the raw logs and unpacked data for AftermarketDeviceTransferred events raised by the Registry contract.
type RegistryAftermarketDeviceTransferredIterator struct {
	Event *RegistryAftermarketDeviceTransferred // Event containing the contract specifics and raw log

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
func (it *RegistryAftermarketDeviceTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryAftermarketDeviceTransferred)
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
		it.Event = new(RegistryAftermarketDeviceTransferred)
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
func (it *RegistryAftermarketDeviceTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryAftermarketDeviceTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryAftermarketDeviceTransferred represents a AftermarketDeviceTransferred event raised by the Registry contract.
type RegistryAftermarketDeviceTransferred struct {
	AftermarketDeviceNode *big.Int
	OldOwner              common.Address
	NewOwner              common.Address
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterAftermarketDeviceTransferred is a free log retrieval operation binding the contract event 0x1d2e88640b58e7fc67878851d97e2cfae3bc7eb7db3226dec94b1c499d631637.
//
// Solidity: event AftermarketDeviceTransferred(uint256 indexed aftermarketDeviceNode, address indexed oldOwner, address indexed newOwner)
func (_Registry *RegistryFilterer) FilterAftermarketDeviceTransferred(opts *bind.FilterOpts, aftermarketDeviceNode []*big.Int, oldOwner []common.Address, newOwner []common.Address) (*RegistryAftermarketDeviceTransferredIterator, error) {

	var aftermarketDeviceNodeRule []interface{}
	for _, aftermarketDeviceNodeItem := range aftermarketDeviceNode {
		aftermarketDeviceNodeRule = append(aftermarketDeviceNodeRule, aftermarketDeviceNodeItem)
	}
	var oldOwnerRule []interface{}
	for _, oldOwnerItem := range oldOwner {
		oldOwnerRule = append(oldOwnerRule, oldOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "AftermarketDeviceTransferred", aftermarketDeviceNodeRule, oldOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &RegistryAftermarketDeviceTransferredIterator{contract: _Registry.contract, event: "AftermarketDeviceTransferred", logs: logs, sub: sub}, nil
}

// WatchAftermarketDeviceTransferred is a free log subscription operation binding the contract event 0x1d2e88640b58e7fc67878851d97e2cfae3bc7eb7db3226dec94b1c499d631637.
//
// Solidity: event AftermarketDeviceTransferred(uint256 indexed aftermarketDeviceNode, address indexed oldOwner, address indexed newOwner)
func (_Registry *RegistryFilterer) WatchAftermarketDeviceTransferred(opts *bind.WatchOpts, sink chan<- *RegistryAftermarketDeviceTransferred, aftermarketDeviceNode []*big.Int, oldOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var aftermarketDeviceNodeRule []interface{}
	for _, aftermarketDeviceNodeItem := range aftermarketDeviceNode {
		aftermarketDeviceNodeRule = append(aftermarketDeviceNodeRule, aftermarketDeviceNodeItem)
	}
	var oldOwnerRule []interface{}
	for _, oldOwnerItem := range oldOwner {
		oldOwnerRule = append(oldOwnerRule, oldOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "AftermarketDeviceTransferred", aftermarketDeviceNodeRule, oldOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryAftermarketDeviceTransferred)
				if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceTransferred", log); err != nil {
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

// ParseAftermarketDeviceTransferred is a log parse operation binding the contract event 0x1d2e88640b58e7fc67878851d97e2cfae3bc7eb7db3226dec94b1c499d631637.
//
// Solidity: event AftermarketDeviceTransferred(uint256 indexed aftermarketDeviceNode, address indexed oldOwner, address indexed newOwner)
func (_Registry *RegistryFilterer) ParseAftermarketDeviceTransferred(log types.Log) (*RegistryAftermarketDeviceTransferred, error) {
	event := new(RegistryAftermarketDeviceTransferred)
	if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryAftermarketDeviceUnclaimedIterator is returned from FilterAftermarketDeviceUnclaimed and is used to iterate over the raw logs and unpacked data for AftermarketDeviceUnclaimed events raised by the Registry contract.
type RegistryAftermarketDeviceUnclaimedIterator struct {
	Event *RegistryAftermarketDeviceUnclaimed // Event containing the contract specifics and raw log

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
func (it *RegistryAftermarketDeviceUnclaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryAftermarketDeviceUnclaimed)
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
		it.Event = new(RegistryAftermarketDeviceUnclaimed)
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
func (it *RegistryAftermarketDeviceUnclaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryAftermarketDeviceUnclaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryAftermarketDeviceUnclaimed represents a AftermarketDeviceUnclaimed event raised by the Registry contract.
type RegistryAftermarketDeviceUnclaimed struct {
	AftermarketDeviceNode *big.Int
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterAftermarketDeviceUnclaimed is a free log retrieval operation binding the contract event 0x9811dbcee5e7af2d698bf7ed98b64e8e28ab2e5e5e5fa4a03aace3c4f9b48e84.
//
// Solidity: event AftermarketDeviceUnclaimed(uint256 indexed aftermarketDeviceNode)
func (_Registry *RegistryFilterer) FilterAftermarketDeviceUnclaimed(opts *bind.FilterOpts, aftermarketDeviceNode []*big.Int) (*RegistryAftermarketDeviceUnclaimedIterator, error) {

	var aftermarketDeviceNodeRule []interface{}
	for _, aftermarketDeviceNodeItem := range aftermarketDeviceNode {
		aftermarketDeviceNodeRule = append(aftermarketDeviceNodeRule, aftermarketDeviceNodeItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "AftermarketDeviceUnclaimed", aftermarketDeviceNodeRule)
	if err != nil {
		return nil, err
	}
	return &RegistryAftermarketDeviceUnclaimedIterator{contract: _Registry.contract, event: "AftermarketDeviceUnclaimed", logs: logs, sub: sub}, nil
}

// WatchAftermarketDeviceUnclaimed is a free log subscription operation binding the contract event 0x9811dbcee5e7af2d698bf7ed98b64e8e28ab2e5e5e5fa4a03aace3c4f9b48e84.
//
// Solidity: event AftermarketDeviceUnclaimed(uint256 indexed aftermarketDeviceNode)
func (_Registry *RegistryFilterer) WatchAftermarketDeviceUnclaimed(opts *bind.WatchOpts, sink chan<- *RegistryAftermarketDeviceUnclaimed, aftermarketDeviceNode []*big.Int) (event.Subscription, error) {

	var aftermarketDeviceNodeRule []interface{}
	for _, aftermarketDeviceNodeItem := range aftermarketDeviceNode {
		aftermarketDeviceNodeRule = append(aftermarketDeviceNodeRule, aftermarketDeviceNodeItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "AftermarketDeviceUnclaimed", aftermarketDeviceNodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryAftermarketDeviceUnclaimed)
				if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceUnclaimed", log); err != nil {
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

// ParseAftermarketDeviceUnclaimed is a log parse operation binding the contract event 0x9811dbcee5e7af2d698bf7ed98b64e8e28ab2e5e5e5fa4a03aace3c4f9b48e84.
//
// Solidity: event AftermarketDeviceUnclaimed(uint256 indexed aftermarketDeviceNode)
func (_Registry *RegistryFilterer) ParseAftermarketDeviceUnclaimed(log types.Log) (*RegistryAftermarketDeviceUnclaimed, error) {
	event := new(RegistryAftermarketDeviceUnclaimed)
	if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceUnclaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryAftermarketDeviceUnpairedIterator is returned from FilterAftermarketDeviceUnpaired and is used to iterate over the raw logs and unpacked data for AftermarketDeviceUnpaired events raised by the Registry contract.
type RegistryAftermarketDeviceUnpairedIterator struct {
	Event *RegistryAftermarketDeviceUnpaired // Event containing the contract specifics and raw log

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
func (it *RegistryAftermarketDeviceUnpairedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryAftermarketDeviceUnpaired)
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
		it.Event = new(RegistryAftermarketDeviceUnpaired)
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
func (it *RegistryAftermarketDeviceUnpairedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryAftermarketDeviceUnpairedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryAftermarketDeviceUnpaired represents a AftermarketDeviceUnpaired event raised by the Registry contract.
type RegistryAftermarketDeviceUnpaired struct {
	AftermarketDeviceNode *big.Int
	VehicleNode           *big.Int
	Owner                 common.Address
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterAftermarketDeviceUnpaired is a free log retrieval operation binding the contract event 0xd9135724aa6cdaa5b3ea73e3e0d74cb1a3a6d3cddcb9d58583f05f17bac82a8e.
//
// Solidity: event AftermarketDeviceUnpaired(uint256 indexed aftermarketDeviceNode, uint256 indexed vehicleNode, address indexed owner)
func (_Registry *RegistryFilterer) FilterAftermarketDeviceUnpaired(opts *bind.FilterOpts, aftermarketDeviceNode []*big.Int, vehicleNode []*big.Int, owner []common.Address) (*RegistryAftermarketDeviceUnpairedIterator, error) {

	var aftermarketDeviceNodeRule []interface{}
	for _, aftermarketDeviceNodeItem := range aftermarketDeviceNode {
		aftermarketDeviceNodeRule = append(aftermarketDeviceNodeRule, aftermarketDeviceNodeItem)
	}
	var vehicleNodeRule []interface{}
	for _, vehicleNodeItem := range vehicleNode {
		vehicleNodeRule = append(vehicleNodeRule, vehicleNodeItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "AftermarketDeviceUnpaired", aftermarketDeviceNodeRule, vehicleNodeRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &RegistryAftermarketDeviceUnpairedIterator{contract: _Registry.contract, event: "AftermarketDeviceUnpaired", logs: logs, sub: sub}, nil
}

// WatchAftermarketDeviceUnpaired is a free log subscription operation binding the contract event 0xd9135724aa6cdaa5b3ea73e3e0d74cb1a3a6d3cddcb9d58583f05f17bac82a8e.
//
// Solidity: event AftermarketDeviceUnpaired(uint256 indexed aftermarketDeviceNode, uint256 indexed vehicleNode, address indexed owner)
func (_Registry *RegistryFilterer) WatchAftermarketDeviceUnpaired(opts *bind.WatchOpts, sink chan<- *RegistryAftermarketDeviceUnpaired, aftermarketDeviceNode []*big.Int, vehicleNode []*big.Int, owner []common.Address) (event.Subscription, error) {

	var aftermarketDeviceNodeRule []interface{}
	for _, aftermarketDeviceNodeItem := range aftermarketDeviceNode {
		aftermarketDeviceNodeRule = append(aftermarketDeviceNodeRule, aftermarketDeviceNodeItem)
	}
	var vehicleNodeRule []interface{}
	for _, vehicleNodeItem := range vehicleNode {
		vehicleNodeRule = append(vehicleNodeRule, vehicleNodeItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "AftermarketDeviceUnpaired", aftermarketDeviceNodeRule, vehicleNodeRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryAftermarketDeviceUnpaired)
				if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceUnpaired", log); err != nil {
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

// ParseAftermarketDeviceUnpaired is a log parse operation binding the contract event 0xd9135724aa6cdaa5b3ea73e3e0d74cb1a3a6d3cddcb9d58583f05f17bac82a8e.
//
// Solidity: event AftermarketDeviceUnpaired(uint256 indexed aftermarketDeviceNode, uint256 indexed vehicleNode, address indexed owner)
func (_Registry *RegistryFilterer) ParseAftermarketDeviceUnpaired(log types.Log) (*RegistryAftermarketDeviceUnpaired, error) {
	event := new(RegistryAftermarketDeviceUnpaired)
	if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceUnpaired", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryAftermarketDeviceUnpaired0Iterator is returned from FilterAftermarketDeviceUnpaired0 and is used to iterate over the raw logs and unpacked data for AftermarketDeviceUnpaired0 events raised by the Registry contract.
type RegistryAftermarketDeviceUnpaired0Iterator struct {
	Event *RegistryAftermarketDeviceUnpaired0 // Event containing the contract specifics and raw log

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
func (it *RegistryAftermarketDeviceUnpaired0Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryAftermarketDeviceUnpaired0)
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
		it.Event = new(RegistryAftermarketDeviceUnpaired0)
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
func (it *RegistryAftermarketDeviceUnpaired0Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryAftermarketDeviceUnpaired0Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryAftermarketDeviceUnpaired0 represents a AftermarketDeviceUnpaired0 event raised by the Registry contract.
type RegistryAftermarketDeviceUnpaired0 struct {
	AftermarketDeviceNode *big.Int
	VehicleNode           *big.Int
	Owner                 common.Address
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterAftermarketDeviceUnpaired0 is a free log retrieval operation binding the contract event 0xd9135724aa6cdaa5b3ea73e3e0d74cb1a3a6d3cddcb9d58583f05f17bac82a8e.
//
// Solidity: event AftermarketDeviceUnpaired(uint256 aftermarketDeviceNode, uint256 vehicleNode, address indexed owner)
func (_Registry *RegistryFilterer) FilterAftermarketDeviceUnpaired0(opts *bind.FilterOpts, owner []common.Address) (*RegistryAftermarketDeviceUnpaired0Iterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "AftermarketDeviceUnpaired0", ownerRule)
	if err != nil {
		return nil, err
	}
	return &RegistryAftermarketDeviceUnpaired0Iterator{contract: _Registry.contract, event: "AftermarketDeviceUnpaired0", logs: logs, sub: sub}, nil
}

// WatchAftermarketDeviceUnpaired0 is a free log subscription operation binding the contract event 0xd9135724aa6cdaa5b3ea73e3e0d74cb1a3a6d3cddcb9d58583f05f17bac82a8e.
//
// Solidity: event AftermarketDeviceUnpaired(uint256 aftermarketDeviceNode, uint256 vehicleNode, address indexed owner)
func (_Registry *RegistryFilterer) WatchAftermarketDeviceUnpaired0(opts *bind.WatchOpts, sink chan<- *RegistryAftermarketDeviceUnpaired0, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "AftermarketDeviceUnpaired0", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryAftermarketDeviceUnpaired0)
				if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceUnpaired0", log); err != nil {
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

// ParseAftermarketDeviceUnpaired0 is a log parse operation binding the contract event 0xd9135724aa6cdaa5b3ea73e3e0d74cb1a3a6d3cddcb9d58583f05f17bac82a8e.
//
// Solidity: event AftermarketDeviceUnpaired(uint256 aftermarketDeviceNode, uint256 vehicleNode, address indexed owner)
func (_Registry *RegistryFilterer) ParseAftermarketDeviceUnpaired0(log types.Log) (*RegistryAftermarketDeviceUnpaired0, error) {
	event := new(RegistryAftermarketDeviceUnpaired0)
	if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceUnpaired0", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryControllerSetIterator is returned from FilterControllerSet and is used to iterate over the raw logs and unpacked data for ControllerSet events raised by the Registry contract.
type RegistryControllerSetIterator struct {
	Event *RegistryControllerSet // Event containing the contract specifics and raw log

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
func (it *RegistryControllerSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryControllerSet)
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
		it.Event = new(RegistryControllerSet)
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
func (it *RegistryControllerSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryControllerSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryControllerSet represents a ControllerSet event raised by the Registry contract.
type RegistryControllerSet struct {
	Controller common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterControllerSet is a free log retrieval operation binding the contract event 0x79f74fd5964b6943d8a1865abfb7f668c92fa3f32c0a2e3195da7d0946703ad7.
//
// Solidity: event ControllerSet(address indexed controller)
func (_Registry *RegistryFilterer) FilterControllerSet(opts *bind.FilterOpts, controller []common.Address) (*RegistryControllerSetIterator, error) {

	var controllerRule []interface{}
	for _, controllerItem := range controller {
		controllerRule = append(controllerRule, controllerItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "ControllerSet", controllerRule)
	if err != nil {
		return nil, err
	}
	return &RegistryControllerSetIterator{contract: _Registry.contract, event: "ControllerSet", logs: logs, sub: sub}, nil
}

// WatchControllerSet is a free log subscription operation binding the contract event 0x79f74fd5964b6943d8a1865abfb7f668c92fa3f32c0a2e3195da7d0946703ad7.
//
// Solidity: event ControllerSet(address indexed controller)
func (_Registry *RegistryFilterer) WatchControllerSet(opts *bind.WatchOpts, sink chan<- *RegistryControllerSet, controller []common.Address) (event.Subscription, error) {

	var controllerRule []interface{}
	for _, controllerItem := range controller {
		controllerRule = append(controllerRule, controllerItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "ControllerSet", controllerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryControllerSet)
				if err := _Registry.contract.UnpackLog(event, "ControllerSet", log); err != nil {
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

// ParseControllerSet is a log parse operation binding the contract event 0x79f74fd5964b6943d8a1865abfb7f668c92fa3f32c0a2e3195da7d0946703ad7.
//
// Solidity: event ControllerSet(address indexed controller)
func (_Registry *RegistryFilterer) ParseControllerSet(log types.Log) (*RegistryControllerSet, error) {
	event := new(RegistryControllerSet)
	if err := _Registry.contract.UnpackLog(event, "ControllerSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryManufacturerAttributeAddedIterator is returned from FilterManufacturerAttributeAdded and is used to iterate over the raw logs and unpacked data for ManufacturerAttributeAdded events raised by the Registry contract.
type RegistryManufacturerAttributeAddedIterator struct {
	Event *RegistryManufacturerAttributeAdded // Event containing the contract specifics and raw log

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
func (it *RegistryManufacturerAttributeAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryManufacturerAttributeAdded)
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
		it.Event = new(RegistryManufacturerAttributeAdded)
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
func (it *RegistryManufacturerAttributeAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryManufacturerAttributeAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryManufacturerAttributeAdded represents a ManufacturerAttributeAdded event raised by the Registry contract.
type RegistryManufacturerAttributeAdded struct {
	Attribute string
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterManufacturerAttributeAdded is a free log retrieval operation binding the contract event 0x47ff34ba477617ab4dc2aefe5ea26ba19a207b052ec44d59b86c2ff3e7fd53b3.
//
// Solidity: event ManufacturerAttributeAdded(string attribute)
func (_Registry *RegistryFilterer) FilterManufacturerAttributeAdded(opts *bind.FilterOpts) (*RegistryManufacturerAttributeAddedIterator, error) {

	logs, sub, err := _Registry.contract.FilterLogs(opts, "ManufacturerAttributeAdded")
	if err != nil {
		return nil, err
	}
	return &RegistryManufacturerAttributeAddedIterator{contract: _Registry.contract, event: "ManufacturerAttributeAdded", logs: logs, sub: sub}, nil
}

// WatchManufacturerAttributeAdded is a free log subscription operation binding the contract event 0x47ff34ba477617ab4dc2aefe5ea26ba19a207b052ec44d59b86c2ff3e7fd53b3.
//
// Solidity: event ManufacturerAttributeAdded(string attribute)
func (_Registry *RegistryFilterer) WatchManufacturerAttributeAdded(opts *bind.WatchOpts, sink chan<- *RegistryManufacturerAttributeAdded) (event.Subscription, error) {

	logs, sub, err := _Registry.contract.WatchLogs(opts, "ManufacturerAttributeAdded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryManufacturerAttributeAdded)
				if err := _Registry.contract.UnpackLog(event, "ManufacturerAttributeAdded", log); err != nil {
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

// ParseManufacturerAttributeAdded is a log parse operation binding the contract event 0x47ff34ba477617ab4dc2aefe5ea26ba19a207b052ec44d59b86c2ff3e7fd53b3.
//
// Solidity: event ManufacturerAttributeAdded(string attribute)
func (_Registry *RegistryFilterer) ParseManufacturerAttributeAdded(log types.Log) (*RegistryManufacturerAttributeAdded, error) {
	event := new(RegistryManufacturerAttributeAdded)
	if err := _Registry.contract.UnpackLog(event, "ManufacturerAttributeAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryManufacturerAttributeSetIterator is returned from FilterManufacturerAttributeSet and is used to iterate over the raw logs and unpacked data for ManufacturerAttributeSet events raised by the Registry contract.
type RegistryManufacturerAttributeSetIterator struct {
	Event *RegistryManufacturerAttributeSet // Event containing the contract specifics and raw log

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
func (it *RegistryManufacturerAttributeSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryManufacturerAttributeSet)
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
		it.Event = new(RegistryManufacturerAttributeSet)
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
func (it *RegistryManufacturerAttributeSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryManufacturerAttributeSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryManufacturerAttributeSet represents a ManufacturerAttributeSet event raised by the Registry contract.
type RegistryManufacturerAttributeSet struct {
	TokenId   *big.Int
	Attribute string
	Info      string
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterManufacturerAttributeSet is a free log retrieval operation binding the contract event 0xb81a4ce1a42b79dd48c79a2c5a0b170cebf3c78b5ecb25df31066eb9b656a929.
//
// Solidity: event ManufacturerAttributeSet(uint256 tokenId, string attribute, string info)
func (_Registry *RegistryFilterer) FilterManufacturerAttributeSet(opts *bind.FilterOpts) (*RegistryManufacturerAttributeSetIterator, error) {

	logs, sub, err := _Registry.contract.FilterLogs(opts, "ManufacturerAttributeSet")
	if err != nil {
		return nil, err
	}
	return &RegistryManufacturerAttributeSetIterator{contract: _Registry.contract, event: "ManufacturerAttributeSet", logs: logs, sub: sub}, nil
}

// WatchManufacturerAttributeSet is a free log subscription operation binding the contract event 0xb81a4ce1a42b79dd48c79a2c5a0b170cebf3c78b5ecb25df31066eb9b656a929.
//
// Solidity: event ManufacturerAttributeSet(uint256 tokenId, string attribute, string info)
func (_Registry *RegistryFilterer) WatchManufacturerAttributeSet(opts *bind.WatchOpts, sink chan<- *RegistryManufacturerAttributeSet) (event.Subscription, error) {

	logs, sub, err := _Registry.contract.WatchLogs(opts, "ManufacturerAttributeSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryManufacturerAttributeSet)
				if err := _Registry.contract.UnpackLog(event, "ManufacturerAttributeSet", log); err != nil {
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

// ParseManufacturerAttributeSet is a log parse operation binding the contract event 0xb81a4ce1a42b79dd48c79a2c5a0b170cebf3c78b5ecb25df31066eb9b656a929.
//
// Solidity: event ManufacturerAttributeSet(uint256 tokenId, string attribute, string info)
func (_Registry *RegistryFilterer) ParseManufacturerAttributeSet(log types.Log) (*RegistryManufacturerAttributeSet, error) {
	event := new(RegistryManufacturerAttributeSet)
	if err := _Registry.contract.UnpackLog(event, "ManufacturerAttributeSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryManufacturerIdProxySetIterator is returned from FilterManufacturerIdProxySet and is used to iterate over the raw logs and unpacked data for ManufacturerIdProxySet events raised by the Registry contract.
type RegistryManufacturerIdProxySetIterator struct {
	Event *RegistryManufacturerIdProxySet // Event containing the contract specifics and raw log

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
func (it *RegistryManufacturerIdProxySetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryManufacturerIdProxySet)
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
		it.Event = new(RegistryManufacturerIdProxySet)
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
func (it *RegistryManufacturerIdProxySetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryManufacturerIdProxySetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryManufacturerIdProxySet represents a ManufacturerIdProxySet event raised by the Registry contract.
type RegistryManufacturerIdProxySet struct {
	Proxy common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterManufacturerIdProxySet is a free log retrieval operation binding the contract event 0xf9bca5f2d5444f9fbb6e6d0fb2b4c2cda766bd110a1326420b883ffc6978f5e2.
//
// Solidity: event ManufacturerIdProxySet(address indexed proxy)
func (_Registry *RegistryFilterer) FilterManufacturerIdProxySet(opts *bind.FilterOpts, proxy []common.Address) (*RegistryManufacturerIdProxySetIterator, error) {

	var proxyRule []interface{}
	for _, proxyItem := range proxy {
		proxyRule = append(proxyRule, proxyItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "ManufacturerIdProxySet", proxyRule)
	if err != nil {
		return nil, err
	}
	return &RegistryManufacturerIdProxySetIterator{contract: _Registry.contract, event: "ManufacturerIdProxySet", logs: logs, sub: sub}, nil
}

// WatchManufacturerIdProxySet is a free log subscription operation binding the contract event 0xf9bca5f2d5444f9fbb6e6d0fb2b4c2cda766bd110a1326420b883ffc6978f5e2.
//
// Solidity: event ManufacturerIdProxySet(address indexed proxy)
func (_Registry *RegistryFilterer) WatchManufacturerIdProxySet(opts *bind.WatchOpts, sink chan<- *RegistryManufacturerIdProxySet, proxy []common.Address) (event.Subscription, error) {

	var proxyRule []interface{}
	for _, proxyItem := range proxy {
		proxyRule = append(proxyRule, proxyItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "ManufacturerIdProxySet", proxyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryManufacturerIdProxySet)
				if err := _Registry.contract.UnpackLog(event, "ManufacturerIdProxySet", log); err != nil {
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

// ParseManufacturerIdProxySet is a log parse operation binding the contract event 0xf9bca5f2d5444f9fbb6e6d0fb2b4c2cda766bd110a1326420b883ffc6978f5e2.
//
// Solidity: event ManufacturerIdProxySet(address indexed proxy)
func (_Registry *RegistryFilterer) ParseManufacturerIdProxySet(log types.Log) (*RegistryManufacturerIdProxySet, error) {
	event := new(RegistryManufacturerIdProxySet)
	if err := _Registry.contract.UnpackLog(event, "ManufacturerIdProxySet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryManufacturerNodeMintedIterator is returned from FilterManufacturerNodeMinted and is used to iterate over the raw logs and unpacked data for ManufacturerNodeMinted events raised by the Registry contract.
type RegistryManufacturerNodeMintedIterator struct {
	Event *RegistryManufacturerNodeMinted // Event containing the contract specifics and raw log

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
func (it *RegistryManufacturerNodeMintedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryManufacturerNodeMinted)
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
		it.Event = new(RegistryManufacturerNodeMinted)
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
func (it *RegistryManufacturerNodeMintedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryManufacturerNodeMintedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryManufacturerNodeMinted represents a ManufacturerNodeMinted event raised by the Registry contract.
type RegistryManufacturerNodeMinted struct {
	TokenId *big.Int
	Owner   common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterManufacturerNodeMinted is a free log retrieval operation binding the contract event 0x8f9181f0193eb4d214d8c50ec79854ead4cd18fef4b2d6e1243ff591592ec420.
//
// Solidity: event ManufacturerNodeMinted(uint256 tokenId, address indexed owner)
func (_Registry *RegistryFilterer) FilterManufacturerNodeMinted(opts *bind.FilterOpts, owner []common.Address) (*RegistryManufacturerNodeMintedIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "ManufacturerNodeMinted", ownerRule)
	if err != nil {
		return nil, err
	}
	return &RegistryManufacturerNodeMintedIterator{contract: _Registry.contract, event: "ManufacturerNodeMinted", logs: logs, sub: sub}, nil
}

// WatchManufacturerNodeMinted is a free log subscription operation binding the contract event 0x8f9181f0193eb4d214d8c50ec79854ead4cd18fef4b2d6e1243ff591592ec420.
//
// Solidity: event ManufacturerNodeMinted(uint256 tokenId, address indexed owner)
func (_Registry *RegistryFilterer) WatchManufacturerNodeMinted(opts *bind.WatchOpts, sink chan<- *RegistryManufacturerNodeMinted, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "ManufacturerNodeMinted", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryManufacturerNodeMinted)
				if err := _Registry.contract.UnpackLog(event, "ManufacturerNodeMinted", log); err != nil {
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

// ParseManufacturerNodeMinted is a log parse operation binding the contract event 0x8f9181f0193eb4d214d8c50ec79854ead4cd18fef4b2d6e1243ff591592ec420.
//
// Solidity: event ManufacturerNodeMinted(uint256 tokenId, address indexed owner)
func (_Registry *RegistryFilterer) ParseManufacturerNodeMinted(log types.Log) (*RegistryManufacturerNodeMinted, error) {
	event := new(RegistryManufacturerNodeMinted)
	if err := _Registry.contract.UnpackLog(event, "ManufacturerNodeMinted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryModuleAddedIterator is returned from FilterModuleAdded and is used to iterate over the raw logs and unpacked data for ModuleAdded events raised by the Registry contract.
type RegistryModuleAddedIterator struct {
	Event *RegistryModuleAdded // Event containing the contract specifics and raw log

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
func (it *RegistryModuleAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryModuleAdded)
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
		it.Event = new(RegistryModuleAdded)
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
func (it *RegistryModuleAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryModuleAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryModuleAdded represents a ModuleAdded event raised by the Registry contract.
type RegistryModuleAdded struct {
	ModuleAddr common.Address
	Selectors  [][4]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterModuleAdded is a free log retrieval operation binding the contract event 0x02d0c334c706cd2f08faf7bc03674fc7f3970dd8921776c655069cde33b7fb29.
//
// Solidity: event ModuleAdded(address indexed moduleAddr, bytes4[] selectors)
func (_Registry *RegistryFilterer) FilterModuleAdded(opts *bind.FilterOpts, moduleAddr []common.Address) (*RegistryModuleAddedIterator, error) {

	var moduleAddrRule []interface{}
	for _, moduleAddrItem := range moduleAddr {
		moduleAddrRule = append(moduleAddrRule, moduleAddrItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "ModuleAdded", moduleAddrRule)
	if err != nil {
		return nil, err
	}
	return &RegistryModuleAddedIterator{contract: _Registry.contract, event: "ModuleAdded", logs: logs, sub: sub}, nil
}

// WatchModuleAdded is a free log subscription operation binding the contract event 0x02d0c334c706cd2f08faf7bc03674fc7f3970dd8921776c655069cde33b7fb29.
//
// Solidity: event ModuleAdded(address indexed moduleAddr, bytes4[] selectors)
func (_Registry *RegistryFilterer) WatchModuleAdded(opts *bind.WatchOpts, sink chan<- *RegistryModuleAdded, moduleAddr []common.Address) (event.Subscription, error) {

	var moduleAddrRule []interface{}
	for _, moduleAddrItem := range moduleAddr {
		moduleAddrRule = append(moduleAddrRule, moduleAddrItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "ModuleAdded", moduleAddrRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryModuleAdded)
				if err := _Registry.contract.UnpackLog(event, "ModuleAdded", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseModuleAdded(log types.Log) (*RegistryModuleAdded, error) {
	event := new(RegistryModuleAdded)
	if err := _Registry.contract.UnpackLog(event, "ModuleAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryModuleRemovedIterator is returned from FilterModuleRemoved and is used to iterate over the raw logs and unpacked data for ModuleRemoved events raised by the Registry contract.
type RegistryModuleRemovedIterator struct {
	Event *RegistryModuleRemoved // Event containing the contract specifics and raw log

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
func (it *RegistryModuleRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryModuleRemoved)
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
		it.Event = new(RegistryModuleRemoved)
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
func (it *RegistryModuleRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryModuleRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryModuleRemoved represents a ModuleRemoved event raised by the Registry contract.
type RegistryModuleRemoved struct {
	ModuleAddr common.Address
	Selectors  [][4]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterModuleRemoved is a free log retrieval operation binding the contract event 0x7c3eb4f9083f75cbed2bd3f703e24b4bbcb77d345d3c50945f3abf3e967755cb.
//
// Solidity: event ModuleRemoved(address indexed moduleAddr, bytes4[] selectors)
func (_Registry *RegistryFilterer) FilterModuleRemoved(opts *bind.FilterOpts, moduleAddr []common.Address) (*RegistryModuleRemovedIterator, error) {

	var moduleAddrRule []interface{}
	for _, moduleAddrItem := range moduleAddr {
		moduleAddrRule = append(moduleAddrRule, moduleAddrItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "ModuleRemoved", moduleAddrRule)
	if err != nil {
		return nil, err
	}
	return &RegistryModuleRemovedIterator{contract: _Registry.contract, event: "ModuleRemoved", logs: logs, sub: sub}, nil
}

// WatchModuleRemoved is a free log subscription operation binding the contract event 0x7c3eb4f9083f75cbed2bd3f703e24b4bbcb77d345d3c50945f3abf3e967755cb.
//
// Solidity: event ModuleRemoved(address indexed moduleAddr, bytes4[] selectors)
func (_Registry *RegistryFilterer) WatchModuleRemoved(opts *bind.WatchOpts, sink chan<- *RegistryModuleRemoved, moduleAddr []common.Address) (event.Subscription, error) {

	var moduleAddrRule []interface{}
	for _, moduleAddrItem := range moduleAddr {
		moduleAddrRule = append(moduleAddrRule, moduleAddrItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "ModuleRemoved", moduleAddrRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryModuleRemoved)
				if err := _Registry.contract.UnpackLog(event, "ModuleRemoved", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseModuleRemoved(log types.Log) (*RegistryModuleRemoved, error) {
	event := new(RegistryModuleRemoved)
	if err := _Registry.contract.UnpackLog(event, "ModuleRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryModuleUpdatedIterator is returned from FilterModuleUpdated and is used to iterate over the raw logs and unpacked data for ModuleUpdated events raised by the Registry contract.
type RegistryModuleUpdatedIterator struct {
	Event *RegistryModuleUpdated // Event containing the contract specifics and raw log

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
func (it *RegistryModuleUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryModuleUpdated)
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
		it.Event = new(RegistryModuleUpdated)
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
func (it *RegistryModuleUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryModuleUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryModuleUpdated represents a ModuleUpdated event raised by the Registry contract.
type RegistryModuleUpdated struct {
	OldImplementation common.Address
	NewImplementation common.Address
	OldSelectors      [][4]byte
	NewSelectors      [][4]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterModuleUpdated is a free log retrieval operation binding the contract event 0xa062c2c046aa14dc9284b13bde77061cb034f0aa820f20057af6b164651eaa08.
//
// Solidity: event ModuleUpdated(address indexed oldImplementation, address indexed newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors)
func (_Registry *RegistryFilterer) FilterModuleUpdated(opts *bind.FilterOpts, oldImplementation []common.Address, newImplementation []common.Address) (*RegistryModuleUpdatedIterator, error) {

	var oldImplementationRule []interface{}
	for _, oldImplementationItem := range oldImplementation {
		oldImplementationRule = append(oldImplementationRule, oldImplementationItem)
	}
	var newImplementationRule []interface{}
	for _, newImplementationItem := range newImplementation {
		newImplementationRule = append(newImplementationRule, newImplementationItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "ModuleUpdated", oldImplementationRule, newImplementationRule)
	if err != nil {
		return nil, err
	}
	return &RegistryModuleUpdatedIterator{contract: _Registry.contract, event: "ModuleUpdated", logs: logs, sub: sub}, nil
}

// WatchModuleUpdated is a free log subscription operation binding the contract event 0xa062c2c046aa14dc9284b13bde77061cb034f0aa820f20057af6b164651eaa08.
//
// Solidity: event ModuleUpdated(address indexed oldImplementation, address indexed newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors)
func (_Registry *RegistryFilterer) WatchModuleUpdated(opts *bind.WatchOpts, sink chan<- *RegistryModuleUpdated, oldImplementation []common.Address, newImplementation []common.Address) (event.Subscription, error) {

	var oldImplementationRule []interface{}
	for _, oldImplementationItem := range oldImplementation {
		oldImplementationRule = append(oldImplementationRule, oldImplementationItem)
	}
	var newImplementationRule []interface{}
	for _, newImplementationItem := range newImplementation {
		newImplementationRule = append(newImplementationRule, newImplementationItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "ModuleUpdated", oldImplementationRule, newImplementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryModuleUpdated)
				if err := _Registry.contract.UnpackLog(event, "ModuleUpdated", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseModuleUpdated(log types.Log) (*RegistryModuleUpdated, error) {
	event := new(RegistryModuleUpdated)
	if err := _Registry.contract.UnpackLog(event, "ModuleUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the Registry contract.
type RegistryRoleAdminChangedIterator struct {
	Event *RegistryRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *RegistryRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryRoleAdminChanged)
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
		it.Event = new(RegistryRoleAdminChanged)
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
func (it *RegistryRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryRoleAdminChanged represents a RoleAdminChanged event raised by the Registry contract.
type RegistryRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Registry *RegistryFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*RegistryRoleAdminChangedIterator, error) {

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

	logs, sub, err := _Registry.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &RegistryRoleAdminChangedIterator{contract: _Registry.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Registry *RegistryFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *RegistryRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

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

	logs, sub, err := _Registry.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryRoleAdminChanged)
				if err := _Registry.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseRoleAdminChanged(log types.Log) (*RegistryRoleAdminChanged, error) {
	event := new(RegistryRoleAdminChanged)
	if err := _Registry.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the Registry contract.
type RegistryRoleGrantedIterator struct {
	Event *RegistryRoleGranted // Event containing the contract specifics and raw log

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
func (it *RegistryRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryRoleGranted)
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
		it.Event = new(RegistryRoleGranted)
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
func (it *RegistryRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryRoleGranted represents a RoleGranted event raised by the Registry contract.
type RegistryRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Registry *RegistryFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*RegistryRoleGrantedIterator, error) {

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

	logs, sub, err := _Registry.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &RegistryRoleGrantedIterator{contract: _Registry.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Registry *RegistryFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *RegistryRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _Registry.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryRoleGranted)
				if err := _Registry.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseRoleGranted(log types.Log) (*RegistryRoleGranted, error) {
	event := new(RegistryRoleGranted)
	if err := _Registry.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the Registry contract.
type RegistryRoleRevokedIterator struct {
	Event *RegistryRoleRevoked // Event containing the contract specifics and raw log

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
func (it *RegistryRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryRoleRevoked)
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
		it.Event = new(RegistryRoleRevoked)
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
func (it *RegistryRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryRoleRevoked represents a RoleRevoked event raised by the Registry contract.
type RegistryRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Registry *RegistryFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*RegistryRoleRevokedIterator, error) {

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

	logs, sub, err := _Registry.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &RegistryRoleRevokedIterator{contract: _Registry.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Registry *RegistryFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *RegistryRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _Registry.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryRoleRevoked)
				if err := _Registry.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseRoleRevoked(log types.Log) (*RegistryRoleRevoked, error) {
	event := new(RegistryRoleRevoked)
	if err := _Registry.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryVehicleAttributeAddedIterator is returned from FilterVehicleAttributeAdded and is used to iterate over the raw logs and unpacked data for VehicleAttributeAdded events raised by the Registry contract.
type RegistryVehicleAttributeAddedIterator struct {
	Event *RegistryVehicleAttributeAdded // Event containing the contract specifics and raw log

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
func (it *RegistryVehicleAttributeAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryVehicleAttributeAdded)
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
		it.Event = new(RegistryVehicleAttributeAdded)
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
func (it *RegistryVehicleAttributeAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryVehicleAttributeAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryVehicleAttributeAdded represents a VehicleAttributeAdded event raised by the Registry contract.
type RegistryVehicleAttributeAdded struct {
	Attribute string
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterVehicleAttributeAdded is a free log retrieval operation binding the contract event 0x2b7d41dc33ffd58029f53ebfc3232e4f343480b078458bc17c527583e0172c1a.
//
// Solidity: event VehicleAttributeAdded(string attribute)
func (_Registry *RegistryFilterer) FilterVehicleAttributeAdded(opts *bind.FilterOpts) (*RegistryVehicleAttributeAddedIterator, error) {

	logs, sub, err := _Registry.contract.FilterLogs(opts, "VehicleAttributeAdded")
	if err != nil {
		return nil, err
	}
	return &RegistryVehicleAttributeAddedIterator{contract: _Registry.contract, event: "VehicleAttributeAdded", logs: logs, sub: sub}, nil
}

// WatchVehicleAttributeAdded is a free log subscription operation binding the contract event 0x2b7d41dc33ffd58029f53ebfc3232e4f343480b078458bc17c527583e0172c1a.
//
// Solidity: event VehicleAttributeAdded(string attribute)
func (_Registry *RegistryFilterer) WatchVehicleAttributeAdded(opts *bind.WatchOpts, sink chan<- *RegistryVehicleAttributeAdded) (event.Subscription, error) {

	logs, sub, err := _Registry.contract.WatchLogs(opts, "VehicleAttributeAdded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryVehicleAttributeAdded)
				if err := _Registry.contract.UnpackLog(event, "VehicleAttributeAdded", log); err != nil {
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

// ParseVehicleAttributeAdded is a log parse operation binding the contract event 0x2b7d41dc33ffd58029f53ebfc3232e4f343480b078458bc17c527583e0172c1a.
//
// Solidity: event VehicleAttributeAdded(string attribute)
func (_Registry *RegistryFilterer) ParseVehicleAttributeAdded(log types.Log) (*RegistryVehicleAttributeAdded, error) {
	event := new(RegistryVehicleAttributeAdded)
	if err := _Registry.contract.UnpackLog(event, "VehicleAttributeAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryVehicleAttributeSetIterator is returned from FilterVehicleAttributeSet and is used to iterate over the raw logs and unpacked data for VehicleAttributeSet events raised by the Registry contract.
type RegistryVehicleAttributeSetIterator struct {
	Event *RegistryVehicleAttributeSet // Event containing the contract specifics and raw log

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
func (it *RegistryVehicleAttributeSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryVehicleAttributeSet)
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
		it.Event = new(RegistryVehicleAttributeSet)
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
func (it *RegistryVehicleAttributeSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryVehicleAttributeSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryVehicleAttributeSet represents a VehicleAttributeSet event raised by the Registry contract.
type RegistryVehicleAttributeSet struct {
	TokenId   *big.Int
	Attribute string
	Info      string
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterVehicleAttributeSet is a free log retrieval operation binding the contract event 0x3a259e5d4c53f11c343582a8291a82a8cc0b36ec211d5ab48c2f29ebb068e5fb.
//
// Solidity: event VehicleAttributeSet(uint256 tokenId, string attribute, string info)
func (_Registry *RegistryFilterer) FilterVehicleAttributeSet(opts *bind.FilterOpts) (*RegistryVehicleAttributeSetIterator, error) {

	logs, sub, err := _Registry.contract.FilterLogs(opts, "VehicleAttributeSet")
	if err != nil {
		return nil, err
	}
	return &RegistryVehicleAttributeSetIterator{contract: _Registry.contract, event: "VehicleAttributeSet", logs: logs, sub: sub}, nil
}

// WatchVehicleAttributeSet is a free log subscription operation binding the contract event 0x3a259e5d4c53f11c343582a8291a82a8cc0b36ec211d5ab48c2f29ebb068e5fb.
//
// Solidity: event VehicleAttributeSet(uint256 tokenId, string attribute, string info)
func (_Registry *RegistryFilterer) WatchVehicleAttributeSet(opts *bind.WatchOpts, sink chan<- *RegistryVehicleAttributeSet) (event.Subscription, error) {

	logs, sub, err := _Registry.contract.WatchLogs(opts, "VehicleAttributeSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryVehicleAttributeSet)
				if err := _Registry.contract.UnpackLog(event, "VehicleAttributeSet", log); err != nil {
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

// ParseVehicleAttributeSet is a log parse operation binding the contract event 0x3a259e5d4c53f11c343582a8291a82a8cc0b36ec211d5ab48c2f29ebb068e5fb.
//
// Solidity: event VehicleAttributeSet(uint256 tokenId, string attribute, string info)
func (_Registry *RegistryFilterer) ParseVehicleAttributeSet(log types.Log) (*RegistryVehicleAttributeSet, error) {
	event := new(RegistryVehicleAttributeSet)
	if err := _Registry.contract.UnpackLog(event, "VehicleAttributeSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryVehicleIdProxySetIterator is returned from FilterVehicleIdProxySet and is used to iterate over the raw logs and unpacked data for VehicleIdProxySet events raised by the Registry contract.
type RegistryVehicleIdProxySetIterator struct {
	Event *RegistryVehicleIdProxySet // Event containing the contract specifics and raw log

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
func (it *RegistryVehicleIdProxySetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryVehicleIdProxySet)
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
		it.Event = new(RegistryVehicleIdProxySet)
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
func (it *RegistryVehicleIdProxySetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryVehicleIdProxySetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryVehicleIdProxySet represents a VehicleIdProxySet event raised by the Registry contract.
type RegistryVehicleIdProxySet struct {
	Proxy common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterVehicleIdProxySet is a free log retrieval operation binding the contract event 0x3e7484c4e57f7d92e9f02eba6cd805d89112e48db8c21aeb8485fcf0020e479d.
//
// Solidity: event VehicleIdProxySet(address indexed proxy)
func (_Registry *RegistryFilterer) FilterVehicleIdProxySet(opts *bind.FilterOpts, proxy []common.Address) (*RegistryVehicleIdProxySetIterator, error) {

	var proxyRule []interface{}
	for _, proxyItem := range proxy {
		proxyRule = append(proxyRule, proxyItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "VehicleIdProxySet", proxyRule)
	if err != nil {
		return nil, err
	}
	return &RegistryVehicleIdProxySetIterator{contract: _Registry.contract, event: "VehicleIdProxySet", logs: logs, sub: sub}, nil
}

// WatchVehicleIdProxySet is a free log subscription operation binding the contract event 0x3e7484c4e57f7d92e9f02eba6cd805d89112e48db8c21aeb8485fcf0020e479d.
//
// Solidity: event VehicleIdProxySet(address indexed proxy)
func (_Registry *RegistryFilterer) WatchVehicleIdProxySet(opts *bind.WatchOpts, sink chan<- *RegistryVehicleIdProxySet, proxy []common.Address) (event.Subscription, error) {

	var proxyRule []interface{}
	for _, proxyItem := range proxy {
		proxyRule = append(proxyRule, proxyItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "VehicleIdProxySet", proxyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryVehicleIdProxySet)
				if err := _Registry.contract.UnpackLog(event, "VehicleIdProxySet", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseVehicleIdProxySet(log types.Log) (*RegistryVehicleIdProxySet, error) {
	event := new(RegistryVehicleIdProxySet)
	if err := _Registry.contract.UnpackLog(event, "VehicleIdProxySet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryVehicleNodeMintedIterator is returned from FilterVehicleNodeMinted and is used to iterate over the raw logs and unpacked data for VehicleNodeMinted events raised by the Registry contract.
type RegistryVehicleNodeMintedIterator struct {
	Event *RegistryVehicleNodeMinted // Event containing the contract specifics and raw log

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
func (it *RegistryVehicleNodeMintedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryVehicleNodeMinted)
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
		it.Event = new(RegistryVehicleNodeMinted)
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
func (it *RegistryVehicleNodeMintedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryVehicleNodeMintedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryVehicleNodeMinted represents a VehicleNodeMinted event raised by the Registry contract.
type RegistryVehicleNodeMinted struct {
	TokenId *big.Int
	Owner   common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterVehicleNodeMinted is a free log retrieval operation binding the contract event 0x09ec7fe5281be92443463e1061ce315afc1142b6c31c98a90b711012a54cc32f.
//
// Solidity: event VehicleNodeMinted(uint256 tokenId, address owner)
func (_Registry *RegistryFilterer) FilterVehicleNodeMinted(opts *bind.FilterOpts) (*RegistryVehicleNodeMintedIterator, error) {

	logs, sub, err := _Registry.contract.FilterLogs(opts, "VehicleNodeMinted")
	if err != nil {
		return nil, err
	}
	return &RegistryVehicleNodeMintedIterator{contract: _Registry.contract, event: "VehicleNodeMinted", logs: logs, sub: sub}, nil
}

// WatchVehicleNodeMinted is a free log subscription operation binding the contract event 0x09ec7fe5281be92443463e1061ce315afc1142b6c31c98a90b711012a54cc32f.
//
// Solidity: event VehicleNodeMinted(uint256 tokenId, address owner)
func (_Registry *RegistryFilterer) WatchVehicleNodeMinted(opts *bind.WatchOpts, sink chan<- *RegistryVehicleNodeMinted) (event.Subscription, error) {

	logs, sub, err := _Registry.contract.WatchLogs(opts, "VehicleNodeMinted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryVehicleNodeMinted)
				if err := _Registry.contract.UnpackLog(event, "VehicleNodeMinted", log); err != nil {
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

// ParseVehicleNodeMinted is a log parse operation binding the contract event 0x09ec7fe5281be92443463e1061ce315afc1142b6c31c98a90b711012a54cc32f.
//
// Solidity: event VehicleNodeMinted(uint256 tokenId, address owner)
func (_Registry *RegistryFilterer) ParseVehicleNodeMinted(log types.Log) (*RegistryVehicleNodeMinted, error) {
	event := new(RegistryVehicleNodeMinted)
	if err := _Registry.contract.UnpackLog(event, "VehicleNodeMinted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
