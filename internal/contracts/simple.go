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

// SimpleMetaData contains all meta data concerning the Simple contract.
var SimpleMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"hashBytes\",\"type\":\"bytes32\"}],\"name\":\"EmitBytes\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"message\",\"type\":\"string\"}],\"name\":\"EmitMessage\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"hello\",\"type\":\"string\"}],\"name\":\"SayHello\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"hashes\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"hello\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"message\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"sayhello\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_message\",\"type\":\"string\"}],\"name\":\"sendMessage\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"testData\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"transactions\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"proof\",\"type\":\"bytes32[]\"},{\"internalType\":\"bytes32\",\"name\":\"root\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"leaf\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"verify\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]",
}

// SimpleABI is the input ABI used to generate the binding from.
// Deprecated: Use SimpleMetaData.ABI instead.
var SimpleABI = SimpleMetaData.ABI

// Simple is an auto generated Go binding around an Ethereum contract.
type Simple struct {
	SimpleCaller     // Read-only binding to the contract
	SimpleTransactor // Write-only binding to the contract
	SimpleFilterer   // Log filterer for contract events
}

// SimpleCaller is an auto generated read-only Go binding around an Ethereum contract.
type SimpleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SimpleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SimpleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SimpleSession struct {
	Contract     *Simple           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SimpleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SimpleCallerSession struct {
	Contract *SimpleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// SimpleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SimpleTransactorSession struct {
	Contract     *SimpleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SimpleRaw is an auto generated low-level Go binding around an Ethereum contract.
type SimpleRaw struct {
	Contract *Simple // Generic contract binding to access the raw methods on
}

// SimpleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SimpleCallerRaw struct {
	Contract *SimpleCaller // Generic read-only contract binding to access the raw methods on
}

// SimpleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SimpleTransactorRaw struct {
	Contract *SimpleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSimple creates a new instance of Simple, bound to a specific deployed contract.
func NewSimple(address common.Address, backend bind.ContractBackend) (*Simple, error) {
	contract, err := bindSimple(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Simple{SimpleCaller: SimpleCaller{contract: contract}, SimpleTransactor: SimpleTransactor{contract: contract}, SimpleFilterer: SimpleFilterer{contract: contract}}, nil
}

// NewSimpleCaller creates a new read-only instance of Simple, bound to a specific deployed contract.
func NewSimpleCaller(address common.Address, caller bind.ContractCaller) (*SimpleCaller, error) {
	contract, err := bindSimple(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleCaller{contract: contract}, nil
}

// NewSimpleTransactor creates a new write-only instance of Simple, bound to a specific deployed contract.
func NewSimpleTransactor(address common.Address, transactor bind.ContractTransactor) (*SimpleTransactor, error) {
	contract, err := bindSimple(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleTransactor{contract: contract}, nil
}

// NewSimpleFilterer creates a new log filterer instance of Simple, bound to a specific deployed contract.
func NewSimpleFilterer(address common.Address, filterer bind.ContractFilterer) (*SimpleFilterer, error) {
	contract, err := bindSimple(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SimpleFilterer{contract: contract}, nil
}

// bindSimple binds a generic wrapper to an already deployed contract.
func bindSimple(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SimpleABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Simple *SimpleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Simple.Contract.SimpleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Simple *SimpleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Simple.Contract.SimpleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Simple *SimpleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Simple.Contract.SimpleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Simple *SimpleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Simple.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Simple *SimpleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Simple.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Simple *SimpleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Simple.Contract.contract.Transact(opts, method, params...)
}

// Hashes is a free data retrieval call binding the contract method 0x501895ae.
//
// Solidity: function hashes(uint256 ) view returns(bytes32)
func (_Simple *SimpleCaller) Hashes(opts *bind.CallOpts, arg0 *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _Simple.contract.Call(opts, &out, "hashes", arg0)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// Hashes is a free data retrieval call binding the contract method 0x501895ae.
//
// Solidity: function hashes(uint256 ) view returns(bytes32)
func (_Simple *SimpleSession) Hashes(arg0 *big.Int) ([32]byte, error) {
	return _Simple.Contract.Hashes(&_Simple.CallOpts, arg0)
}

// Hashes is a free data retrieval call binding the contract method 0x501895ae.
//
// Solidity: function hashes(uint256 ) view returns(bytes32)
func (_Simple *SimpleCallerSession) Hashes(arg0 *big.Int) ([32]byte, error) {
	return _Simple.Contract.Hashes(&_Simple.CallOpts, arg0)
}

// Hello is a free data retrieval call binding the contract method 0x19ff1d21.
//
// Solidity: function hello() view returns(string)
func (_Simple *SimpleCaller) Hello(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Simple.contract.Call(opts, &out, "hello")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Hello is a free data retrieval call binding the contract method 0x19ff1d21.
//
// Solidity: function hello() view returns(string)
func (_Simple *SimpleSession) Hello() (string, error) {
	return _Simple.Contract.Hello(&_Simple.CallOpts)
}

// Hello is a free data retrieval call binding the contract method 0x19ff1d21.
//
// Solidity: function hello() view returns(string)
func (_Simple *SimpleCallerSession) Hello() (string, error) {
	return _Simple.Contract.Hello(&_Simple.CallOpts)
}

// Message is a free data retrieval call binding the contract method 0xe21f37ce.
//
// Solidity: function message() view returns(string)
func (_Simple *SimpleCaller) Message(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Simple.contract.Call(opts, &out, "message")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Message is a free data retrieval call binding the contract method 0xe21f37ce.
//
// Solidity: function message() view returns(string)
func (_Simple *SimpleSession) Message() (string, error) {
	return _Simple.Contract.Message(&_Simple.CallOpts)
}

// Message is a free data retrieval call binding the contract method 0xe21f37ce.
//
// Solidity: function message() view returns(string)
func (_Simple *SimpleCallerSession) Message() (string, error) {
	return _Simple.Contract.Message(&_Simple.CallOpts)
}

// TestData is a free data retrieval call binding the contract method 0x016cbd51.
//
// Solidity: function testData() view returns(string)
func (_Simple *SimpleCaller) TestData(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Simple.contract.Call(opts, &out, "testData")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// TestData is a free data retrieval call binding the contract method 0x016cbd51.
//
// Solidity: function testData() view returns(string)
func (_Simple *SimpleSession) TestData() (string, error) {
	return _Simple.Contract.TestData(&_Simple.CallOpts)
}

// TestData is a free data retrieval call binding the contract method 0x016cbd51.
//
// Solidity: function testData() view returns(string)
func (_Simple *SimpleCallerSession) TestData() (string, error) {
	return _Simple.Contract.TestData(&_Simple.CallOpts)
}

// Transactions is a free data retrieval call binding the contract method 0x9ace38c2.
//
// Solidity: function transactions(uint256 ) view returns(string)
func (_Simple *SimpleCaller) Transactions(opts *bind.CallOpts, arg0 *big.Int) (string, error) {
	var out []interface{}
	err := _Simple.contract.Call(opts, &out, "transactions", arg0)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Transactions is a free data retrieval call binding the contract method 0x9ace38c2.
//
// Solidity: function transactions(uint256 ) view returns(string)
func (_Simple *SimpleSession) Transactions(arg0 *big.Int) (string, error) {
	return _Simple.Contract.Transactions(&_Simple.CallOpts, arg0)
}

// Transactions is a free data retrieval call binding the contract method 0x9ace38c2.
//
// Solidity: function transactions(uint256 ) view returns(string)
func (_Simple *SimpleCallerSession) Transactions(arg0 *big.Int) (string, error) {
	return _Simple.Contract.Transactions(&_Simple.CallOpts, arg0)
}

// Verify is a free data retrieval call binding the contract method 0x21fb335c.
//
// Solidity: function verify(bytes32[] proof, bytes32 root, bytes32 leaf, uint256 index) pure returns(bool)
func (_Simple *SimpleCaller) Verify(opts *bind.CallOpts, proof [][32]byte, root [32]byte, leaf [32]byte, index *big.Int) (bool, error) {
	var out []interface{}
	err := _Simple.contract.Call(opts, &out, "verify", proof, root, leaf, index)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Verify is a free data retrieval call binding the contract method 0x21fb335c.
//
// Solidity: function verify(bytes32[] proof, bytes32 root, bytes32 leaf, uint256 index) pure returns(bool)
func (_Simple *SimpleSession) Verify(proof [][32]byte, root [32]byte, leaf [32]byte, index *big.Int) (bool, error) {
	return _Simple.Contract.Verify(&_Simple.CallOpts, proof, root, leaf, index)
}

// Verify is a free data retrieval call binding the contract method 0x21fb335c.
//
// Solidity: function verify(bytes32[] proof, bytes32 root, bytes32 leaf, uint256 index) pure returns(bool)
func (_Simple *SimpleCallerSession) Verify(proof [][32]byte, root [32]byte, leaf [32]byte, index *big.Int) (bool, error) {
	return _Simple.Contract.Verify(&_Simple.CallOpts, proof, root, leaf, index)
}

// Sayhello is a paid mutator transaction binding the contract method 0x27a03cd7.
//
// Solidity: function sayhello() returns()
func (_Simple *SimpleTransactor) Sayhello(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Simple.contract.Transact(opts, "sayhello")
}

// Sayhello is a paid mutator transaction binding the contract method 0x27a03cd7.
//
// Solidity: function sayhello() returns()
func (_Simple *SimpleSession) Sayhello() (*types.Transaction, error) {
	return _Simple.Contract.Sayhello(&_Simple.TransactOpts)
}

// Sayhello is a paid mutator transaction binding the contract method 0x27a03cd7.
//
// Solidity: function sayhello() returns()
func (_Simple *SimpleTransactorSession) Sayhello() (*types.Transaction, error) {
	return _Simple.Contract.Sayhello(&_Simple.TransactOpts)
}

// SendMessage is a paid mutator transaction binding the contract method 0x469c8110.
//
// Solidity: function sendMessage(string _message) returns()
func (_Simple *SimpleTransactor) SendMessage(opts *bind.TransactOpts, _message string) (*types.Transaction, error) {
	return _Simple.contract.Transact(opts, "sendMessage", _message)
}

// SendMessage is a paid mutator transaction binding the contract method 0x469c8110.
//
// Solidity: function sendMessage(string _message) returns()
func (_Simple *SimpleSession) SendMessage(_message string) (*types.Transaction, error) {
	return _Simple.Contract.SendMessage(&_Simple.TransactOpts, _message)
}

// SendMessage is a paid mutator transaction binding the contract method 0x469c8110.
//
// Solidity: function sendMessage(string _message) returns()
func (_Simple *SimpleTransactorSession) SendMessage(_message string) (*types.Transaction, error) {
	return _Simple.Contract.SendMessage(&_Simple.TransactOpts, _message)
}

// SimpleEmitBytesIterator is returned from FilterEmitBytes and is used to iterate over the raw logs and unpacked data for EmitBytes events raised by the Simple contract.
type SimpleEmitBytesIterator struct {
	Event *SimpleEmitBytes // Event containing the contract specifics and raw log

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
func (it *SimpleEmitBytesIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleEmitBytes)
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
		it.Event = new(SimpleEmitBytes)
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
func (it *SimpleEmitBytesIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleEmitBytesIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleEmitBytes represents a EmitBytes event raised by the Simple contract.
type SimpleEmitBytes struct {
	HashBytes [32]byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEmitBytes is a free log retrieval operation binding the contract event 0x125a64d55651ce95412758ac71519705b3dbd75384b6eac9566b5340011fe671.
//
// Solidity: event EmitBytes(bytes32 hashBytes)
func (_Simple *SimpleFilterer) FilterEmitBytes(opts *bind.FilterOpts) (*SimpleEmitBytesIterator, error) {

	logs, sub, err := _Simple.contract.FilterLogs(opts, "EmitBytes")
	if err != nil {
		return nil, err
	}
	return &SimpleEmitBytesIterator{contract: _Simple.contract, event: "EmitBytes", logs: logs, sub: sub}, nil
}

// WatchEmitBytes is a free log subscription operation binding the contract event 0x125a64d55651ce95412758ac71519705b3dbd75384b6eac9566b5340011fe671.
//
// Solidity: event EmitBytes(bytes32 hashBytes)
func (_Simple *SimpleFilterer) WatchEmitBytes(opts *bind.WatchOpts, sink chan<- *SimpleEmitBytes) (event.Subscription, error) {

	logs, sub, err := _Simple.contract.WatchLogs(opts, "EmitBytes")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleEmitBytes)
				if err := _Simple.contract.UnpackLog(event, "EmitBytes", log); err != nil {
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

// ParseEmitBytes is a log parse operation binding the contract event 0x125a64d55651ce95412758ac71519705b3dbd75384b6eac9566b5340011fe671.
//
// Solidity: event EmitBytes(bytes32 hashBytes)
func (_Simple *SimpleFilterer) ParseEmitBytes(log types.Log) (*SimpleEmitBytes, error) {
	event := new(SimpleEmitBytes)
	if err := _Simple.contract.UnpackLog(event, "EmitBytes", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleEmitMessageIterator is returned from FilterEmitMessage and is used to iterate over the raw logs and unpacked data for EmitMessage events raised by the Simple contract.
type SimpleEmitMessageIterator struct {
	Event *SimpleEmitMessage // Event containing the contract specifics and raw log

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
func (it *SimpleEmitMessageIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleEmitMessage)
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
		it.Event = new(SimpleEmitMessage)
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
func (it *SimpleEmitMessageIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleEmitMessageIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleEmitMessage represents a EmitMessage event raised by the Simple contract.
type SimpleEmitMessage struct {
	Message string
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterEmitMessage is a free log retrieval operation binding the contract event 0x6076819939bfe4b7e4347dc8af3bfd4d018a0a9d488d7d8348a9e1ab09d7057b.
//
// Solidity: event EmitMessage(string message)
func (_Simple *SimpleFilterer) FilterEmitMessage(opts *bind.FilterOpts) (*SimpleEmitMessageIterator, error) {

	logs, sub, err := _Simple.contract.FilterLogs(opts, "EmitMessage")
	if err != nil {
		return nil, err
	}
	return &SimpleEmitMessageIterator{contract: _Simple.contract, event: "EmitMessage", logs: logs, sub: sub}, nil
}

// WatchEmitMessage is a free log subscription operation binding the contract event 0x6076819939bfe4b7e4347dc8af3bfd4d018a0a9d488d7d8348a9e1ab09d7057b.
//
// Solidity: event EmitMessage(string message)
func (_Simple *SimpleFilterer) WatchEmitMessage(opts *bind.WatchOpts, sink chan<- *SimpleEmitMessage) (event.Subscription, error) {

	logs, sub, err := _Simple.contract.WatchLogs(opts, "EmitMessage")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleEmitMessage)
				if err := _Simple.contract.UnpackLog(event, "EmitMessage", log); err != nil {
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

// ParseEmitMessage is a log parse operation binding the contract event 0x6076819939bfe4b7e4347dc8af3bfd4d018a0a9d488d7d8348a9e1ab09d7057b.
//
// Solidity: event EmitMessage(string message)
func (_Simple *SimpleFilterer) ParseEmitMessage(log types.Log) (*SimpleEmitMessage, error) {
	event := new(SimpleEmitMessage)
	if err := _Simple.contract.UnpackLog(event, "EmitMessage", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SimpleSayHelloIterator is returned from FilterSayHello and is used to iterate over the raw logs and unpacked data for SayHello events raised by the Simple contract.
type SimpleSayHelloIterator struct {
	Event *SimpleSayHello // Event containing the contract specifics and raw log

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
func (it *SimpleSayHelloIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleSayHello)
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
		it.Event = new(SimpleSayHello)
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
func (it *SimpleSayHelloIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleSayHelloIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleSayHello represents a SayHello event raised by the Simple contract.
type SimpleSayHello struct {
	Hello string
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterSayHello is a free log retrieval operation binding the contract event 0x0739fb87028724ed02772f820b543201662e37d0701527efe0b952cba64926c2.
//
// Solidity: event SayHello(string hello)
func (_Simple *SimpleFilterer) FilterSayHello(opts *bind.FilterOpts) (*SimpleSayHelloIterator, error) {

	logs, sub, err := _Simple.contract.FilterLogs(opts, "SayHello")
	if err != nil {
		return nil, err
	}
	return &SimpleSayHelloIterator{contract: _Simple.contract, event: "SayHello", logs: logs, sub: sub}, nil
}

// WatchSayHello is a free log subscription operation binding the contract event 0x0739fb87028724ed02772f820b543201662e37d0701527efe0b952cba64926c2.
//
// Solidity: event SayHello(string hello)
func (_Simple *SimpleFilterer) WatchSayHello(opts *bind.WatchOpts, sink chan<- *SimpleSayHello) (event.Subscription, error) {

	logs, sub, err := _Simple.contract.WatchLogs(opts, "SayHello")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleSayHello)
				if err := _Simple.contract.UnpackLog(event, "SayHello", log); err != nil {
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

// ParseSayHello is a log parse operation binding the contract event 0x0739fb87028724ed02772f820b543201662e37d0701527efe0b952cba64926c2.
//
// Solidity: event SayHello(string hello)
func (_Simple *SimpleFilterer) ParseSayHello(log types.Log) (*SimpleSayHello, error) {
	event := new(SimpleSayHello)
	if err := _Simple.contract.UnpackLog(event, "SayHello", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
