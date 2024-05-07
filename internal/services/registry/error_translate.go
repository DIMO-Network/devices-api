// Code generated by github.com/DIMO-Network/solidity-error-gen. DO NOT EDIT.
//
// ABI source: DimoRegistry.json
// Translation source: translation.yaml

package registry

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type DimoRegistryErrorDecoder struct {
	abi *abi.ABI
}

const DimoRegistryRawABI = "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"AdNotClaimed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"AdPaired\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"proxy\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"InvalidNode\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"VehiclePaired\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"AdNotPaired\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"attr\",\"type\":\"string\"}],\"name\":\"AttributeNotWhitelisted\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"DeviceAlreadyClaimed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"DeviceAlreadyRegistered\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidAdSignature\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidOwnerSignature\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"InvalidParentNode\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSigner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"OwnersDoNotMatch\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"VehicleNotPaired\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSdSignature\",\"type\":\"error\"}]"

func NewDimoRegistryErrorDecoder() (*DimoRegistryErrorDecoder, error) {
	a, err := abi.JSON(strings.NewReader(DimoRegistryRawABI))
	if err != nil {
		return nil, err
	}
	return &DimoRegistryErrorDecoder{abi: &a}, nil
}

func (d *DimoRegistryErrorDecoder) Decode(data []byte) (string, error) {
	if len(data) < 4 {
		return "", fmt.Errorf("length %d is too short, must have length at least 4", len(data))
	}

	selector := *(*[4]byte)(data[:4])
	argsData := data[4:]

	switch selector {
	case [4]byte{209, 30, 53, 180}:
		values, err := d.abi.Errors["AdNotPaired"].Inputs.Unpack(argsData)
		if err != nil {
			return "", err
		}
		if l := len(values); l != 1 {
			return "", fmt.Errorf("unpacked into %d args instead of the expected 1", l)
		}
		return fmt.Sprintf("Aftermarket device %[1]s is not paired.", values[0]), nil
	case [4]byte{82, 153, 186, 183}:
		values, err := d.abi.Errors["InvalidParentNode"].Inputs.Unpack(argsData)
		if err != nil {
			return "", err
		}
		if l := len(values); l != 1 {
			return "", fmt.Errorf("unpacked into %d args instead of the expected 1", l)
		}
		return fmt.Sprintf("Parent node %[1]s does not exist.", values[0]), nil
	case [4]byte{248, 233, 93, 85}:
		return "Invalid synthetic device signature.", nil
	case [4]byte{196, 106, 81, 104}:
		values, err := d.abi.Errors["VehiclePaired"].Inputs.Unpack(argsData)
		if err != nil {
			return "", err
		}
		if l := len(values); l != 1 {
			return "", fmt.Errorf("unpacked into %d args instead of the expected 1", l)
		}
		return fmt.Sprintf("Vehicle %[1]s is paired.", values[0]), nil
	case [4]byte{118, 33, 22, 174}:
		values, err := d.abi.Errors["AdPaired"].Inputs.Unpack(argsData)
		if err != nil {
			return "", err
		}
		if l := len(values); l != 1 {
			return "", fmt.Errorf("unpacked into %d args instead of the expected 1", l)
		}
		return fmt.Sprintf("Aftermarket device %[1]s is paired.", values[0]), nil
	case [4]byte{77, 236, 136, 235}:
		values, err := d.abi.Errors["DeviceAlreadyClaimed"].Inputs.Unpack(argsData)
		if err != nil {
			return "", err
		}
		if l := len(values); l != 1 {
			return "", fmt.Errorf("unpacked into %d args instead of the expected 1", l)
		}
		return fmt.Sprintf("Aftermarket device %[1]s already claimed.", values[0]), nil
	case [4]byte{219, 229, 56, 59}:
		return "Invalid aftermarket device signature.", nil
	case [4]byte{56, 168, 90, 141}:
		return "Invalid owner signature.", nil
	case [4]byte{21, 189, 170, 193}:
		values, err := d.abi.Errors["AdNotClaimed"].Inputs.Unpack(argsData)
		if err != nil {
			return "", err
		}
		if l := len(values); l != 1 {
			return "", fmt.Errorf("unpacked into %d args instead of the expected 1", l)
		}
		return fmt.Sprintf("Aftermarket device %[1]s not claimed.", values[0]), nil
	case [4]byte{28, 72, 212, 158}:
		values, err := d.abi.Errors["AttributeNotWhitelisted"].Inputs.Unpack(argsData)
		if err != nil {
			return "", err
		}
		if l := len(values); l != 1 {
			return "", fmt.Errorf("unpacked into %d args instead of the expected 1", l)
		}
		return fmt.Sprintf("Attribute %[1]s not whitelisted.", values[0]), nil
	case [4]byte{205, 118, 232, 69}:
		values, err := d.abi.Errors["DeviceAlreadyRegistered"].Inputs.Unpack(argsData)
		if err != nil {
			return "", err
		}
		if l := len(values); l != 1 {
			return "", fmt.Errorf("unpacked into %d args instead of the expected 1", l)
		}
		return fmt.Sprintf("There is already a minted device with address %[1]s.", values[0]), nil
	case [4]byte{79, 194, 128, 171}:
		return "Vehicle and aftermarket device owners are not the same.", nil
	case [4]byte{227, 202, 150, 57}:
		values, err := d.abi.Errors["InvalidNode"].Inputs.Unpack(argsData)
		if err != nil {
			return "", err
		}
		if l := len(values); l != 2 {
			return "", fmt.Errorf("unpacked into %d args instead of the expected 2", l)
		}
		return fmt.Sprintf("Token %[2]s does not exist at address %[1]s.", values[0], values[1]), nil
	case [4]byte{129, 94, 29, 100}:
		return "Signer is owner of neither the vehicle nor the device.", nil
	case [4]byte{45, 145, 252, 181}:
		values, err := d.abi.Errors["VehicleNotPaired"].Inputs.Unpack(argsData)
		if err != nil {
			return "", err
		}
		if l := len(values); l != 1 {
			return "", fmt.Errorf("unpacked into %d args instead of the expected 1", l)
		}
		return fmt.Sprintf("Vehicle %[1]s is not paired.", values[0]), nil
	default:
		return "", fmt.Errorf("unrecognized error selector %s", hexutil.Encode(selector[:]))
	}
}
