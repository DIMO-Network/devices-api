package registry

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/Shopify/sarama"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	signer "github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/segmentio/ksuid"
)

type Client struct {
	Producer     sarama.SyncProducer
	RequestTopic string
	Contract     Contract
}

type Contract struct {
	ChainID *big.Int
	Address common.Address
	Name    string
	Version string
}

type requestData struct {
	ID   string `json:"id"`
	To   string `json:"to"`
	Data string `json:"data"`
}

// MintVehicleSign(uint256 manufacturerNode,address owner,string[] attributes,string[] infos)
type MintVehicleSign struct {
	ManufacturerNode *big.Int
	Owner            common.Address
	Attributes       []string
	Infos            []string
}

func anySlice[A any](v []A) []any {
	n := len(v)
	out := make([]any, n)

	for i := 0; i < n; i++ {
		out[i] = v[i]
	}

	return out
}

func (m *MintVehicleSign) Name() string {
	return "MintVehicleSign"
}

func (m *MintVehicleSign) Type() []signer.Type {
	return []signer.Type{
		{Name: "manufacturerNode", Type: "uint256"},
		{Name: "owner", Type: "address"},
		{Name: "attributes", Type: "string[]"},
		{Name: "infos", Type: "string[]"},
	}
}

func (m *MintVehicleSign) Message() signer.TypedDataMessage {
	return signer.TypedDataMessage{
		"manufacturerNode": hexutil.EncodeBig(m.ManufacturerNode),
		"owner":            m.Owner.Hex(),
		"attributes":       anySlice(m.Attributes),
		"infos":            anySlice(m.Infos),
	}
}

// ClaimAftermarketDeviceSign(uint256 aftermarketDeviceNode,address owner)
type ClaimAftermarketDeviceSign struct {
	AftermarketDeviceNode *big.Int
	Owner                 common.Address
}

func (m *ClaimAftermarketDeviceSign) Name() string {
	return "ClaimAftermarketDeviceSign"
}

func (m *ClaimAftermarketDeviceSign) Type() []signer.Type {
	return []signer.Type{
		{Name: "aftermarketDeviceNode", Type: "uint256"},
		{Name: "owner", Type: "address"},
	}
}

func (m *ClaimAftermarketDeviceSign) Message() signer.TypedDataMessage {
	return signer.TypedDataMessage{
		"aftermarketDeviceNode": hexutil.EncodeBig(m.AftermarketDeviceNode),
		"owner":                 m.Owner.Hex(),
	}
}

// PairAftermarketDeviceSign(uint256 aftermarketDeviceNode,uint256 vehicleNode)
type PairAftermarketDeviceSign struct {
	AftermarketDeviceNode *big.Int
	VehicleNode           *big.Int
}

func (m *PairAftermarketDeviceSign) Name() string {
	return "PairAftermarketDeviceSign"
}

func (m *PairAftermarketDeviceSign) Type() []signer.Type {
	return []signer.Type{
		{Name: "aftermarketDeviceNode", Type: "uint256"},
		{Name: "vehicleNode", Type: "uint256"},
	}
}

func (m *PairAftermarketDeviceSign) Message() signer.TypedDataMessage {
	return signer.TypedDataMessage{
		"aftermarketDeviceNode": hexutil.EncodeBig(m.AftermarketDeviceNode),
		"vehicleNode":           hexutil.EncodeBig(m.VehicleNode),
	}
}

// UnPairAftermarketDeviceSign(uint256 aftermarketDeviceNode,uint256 vehicleNode)
// Looks exactly like the pairing message outside of the name.
type UnPairAftermarketDeviceSign struct {
	AftermarketDeviceNode *big.Int
	VehicleNode           *big.Int
}

func (m *UnPairAftermarketDeviceSign) Name() string {
	return "UnPairAftermarketDeviceSign"
}

func (m *UnPairAftermarketDeviceSign) Type() []signer.Type {
	return []signer.Type{
		{Name: "aftermarketDeviceNode", Type: "uint256"},
		{Name: "vehicleNode", Type: "uint256"},
	}
}

func (m *UnPairAftermarketDeviceSign) Message() signer.TypedDataMessage {
	return signer.TypedDataMessage{
		"aftermarketDeviceNode": hexutil.EncodeBig(m.AftermarketDeviceNode),
		"vehicleNode":           hexutil.EncodeBig(m.VehicleNode),
	}
}

type Message interface {
	Name() string
	Type() []signer.Type
	Message() signer.TypedDataMessage
}

// mintVehicleSign(uint256 manufacturerNode, address owner,	string[] calldata attributes, string[] calldata infos, bytes calldata signature)
func (c *Client) MintVehicleSign(requestID string, manufacturerNode *big.Int, owner common.Address, attrInfo []AttributeInfoPair, signature []byte) error {
	abi, err := RegistryMetaData.GetAbi()
	if err != nil {
		return err
	}

	data, err := abi.Pack("mintVehicleSign", manufacturerNode, owner, attrInfo, signature)
	if err != nil {
		return err
	}

	return c.sendRequest(requestID, data)
}

// claimAftermarketDeviceSign(uint256 aftermarketDeviceNode, address owner,	bytes calldata ownerSig, bytes calldata aftermarketDeviceSig)
func (c *Client) ClaimAftermarketDeviceSign(requestID string, aftermarketDeviceNode *big.Int, owner common.Address, ownerSig []byte, aftermarketDeviceSig []byte) error {
	abi, err := RegistryMetaData.GetAbi()
	if err != nil {
		return err
	}

	data, err := abi.Pack("claimAftermarketDeviceSign", aftermarketDeviceNode, owner, ownerSig, aftermarketDeviceSig)
	if err != nil {
		return err
	}

	return c.sendRequest(requestID, data)
}

// function pairAftermarketDeviceSign(uint256 aftermarketDeviceNode, uint256 vehicleNode, bytes calldata signature)
func (c *Client) PairAftermarketDeviceSign(requestID string, aftermarketDeviceNode *big.Int, vehicleNode *big.Int, signature []byte) error {
	abi, err := RegistryMetaData.GetAbi()
	if err != nil {
		return err
	}

	data, err := abi.Pack("pairAftermarketDeviceSign", aftermarketDeviceNode, vehicleNode, signature)
	if err != nil {
		return err
	}

	return c.sendRequest(requestID, data)
}

// function unpairAftermarketDeviceSign(uint256 aftermarketDeviceNode, uint256 vehicleNode, bytes calldata signature)
func (c *Client) UnPairAftermarketDeviceSign(requestID string, aftermarketDeviceNode *big.Int, vehicleNode *big.Int, signature []byte) error {
	abi, err := RegistryMetaData.GetAbi()
	if err != nil {
		return err
	}

	data, err := abi.Pack("unpairAftermarketDeviceSign", aftermarketDeviceNode, vehicleNode, signature)
	if err != nil {
		return err
	}

	return c.sendRequest(requestID, data)
}

func (c *Client) sendRequest(requestID string, data []byte) error {
	event := shared.CloudEvent[requestData]{
		ID:          ksuid.New().String(),
		Source:      "devices-api",
		SpecVersion: "1.0",
		Subject:     requestID,
		Time:        time.Now(),
		Type:        "zone.dimo.transaction.request",
		Data: requestData{
			ID:   requestID,
			To:   hexutil.Encode(c.Contract.Address[:]),
			Data: hexutil.Encode(data),
		},
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, _, err = c.Producer.SendMessage(
		&sarama.ProducerMessage{
			Topic: c.RequestTopic,
			Key:   sarama.StringEncoder(requestID),
			Value: sarama.ByteEncoder(eventBytes),
		},
	)

	return err
}

func (c *Client) GetPayload(msg Message) *signer.TypedData {
	return &signer.TypedData{
		Types: signer.Types{
			"EIP712Domain": []signer.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			msg.Name(): msg.Type(),
		},
		PrimaryType: msg.Name(),
		Domain: signer.TypedDataDomain{
			Name:              c.Contract.Name,
			Version:           c.Contract.Version,
			ChainId:           (*math.HexOrDecimal256)(c.Contract.ChainID),
			VerifyingContract: c.Contract.Address.Hex(),
		},
		Message: msg.Message(),
	}
}

func (c *Client) Hash(msg Message) (common.Hash, error) {
	td := c.GetPayload(msg)
	domHash, err := td.HashStruct("EIP712Domain", td.Domain.Map())
	if err != nil {
		return common.Hash{}, err
	}

	msgHash, err := td.HashStruct(td.PrimaryType, td.Message)
	if err != nil {
		return common.Hash{}, err
	}

	payload := []byte{0x19, 0x01}
	payload = append(payload, domHash...)
	payload = append(payload, msgHash...)

	return crypto.Keccak256Hash(payload), nil
}
