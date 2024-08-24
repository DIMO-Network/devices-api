package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	"github.com/DIMO-Network/devices-api/internal/services/dex"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/sqlboiler/v4/types"

	smock "github.com/IBM/sarama/mocks"
)

type mockTestEntity struct {
	Contract    []byte
	TokenID     types.Decimal
	PrivilegeID int64
	UserAddress []byte
	ExpiresAt   time.Time
}

type mockTestArgs struct {
	contract    common.Address
	tokenID     types.Decimal
	userAddress common.Address
	expiresAt   int64
	privilegeID int64
	from        common.Address
	to          common.Address
}

type cEventsTestHelper struct {
	logger    zerolog.Logger
	pdb       db.Store
	container testcontainers.Container
	ctx       context.Context
	t         *testing.T
	require   *require.Assertions
	settings  *config.Settings
}

type eventsFactoryResp struct {
	args    mockTestArgs
	payload string
}

const AftermarketDeviceContractAddress = "0x00000000000000000000000000000000000000c1"

func marshalMockPayload(payload string) (*shared.CloudEvent[json.RawMessage], error) {
	event := new(shared.CloudEvent[json.RawMessage])
	err := json.Unmarshal([]byte(payload), event)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func TestProcessContractsEventsMessages(t *testing.T) {
	ctx := context.Background()

	s := initCEventsTestHelper(t)
	defer s.destroy()

	e := privilegeEventsPayloadFactory(1, 1, "", 0, s.settings.DIMORegistryChainID)
	factoryResp := e[0]

	c := NewContractsEventsConsumer(s.pdb, &s.logger, s.settings, nil, nil, nil, nil, nil, nil)

	event, err := marshalMockPayload(factoryResp.payload)
	s.require.NoError(err)

	err = c.processEvent(ctx, event)
	s.require.NoError(err)

	args := factoryResp.args

	nft, err := models.FindNFTPrivilege(s.ctx, s.pdb.DBS().Reader, args.contract.Bytes(), args.tokenID, args.privilegeID, args.userAddress.Bytes())
	s.require.NoError(err)

	s.require.NotNil(nft)

	actual := mockTestEntity{
		Contract:    nft.ContractAddress,
		TokenID:     nft.TokenID,
		PrivilegeID: nft.Privilege,
		UserAddress: nft.UserAddress,
		ExpiresAt:   nft.Expiry,
	}

	expected := mockTestEntity{
		Contract:    args.contract.Bytes(),
		UserAddress: args.userAddress.Bytes(),
		TokenID:     args.tokenID,
		ExpiresAt:   time.Unix(args.expiresAt, 0).UTC(),
		PrivilegeID: args.privilegeID,
	}

	s.require.Equal(expected, actual, "Event was persisted properly")
}

func TestIgnoreWrongEventNames(t *testing.T) {
	ctx := context.Background()

	s := initCEventsTestHelper(t)
	defer s.destroy()

	e := privilegeEventsPayloadFactory(2, 2, "SomeEvent", 0, s.settings.DIMORegistryChainID)
	factoryResp := e[0]

	c := NewContractsEventsConsumer(s.pdb, &s.logger, s.settings, nil, nil, nil, nil, nil, nil)

	event, err := marshalMockPayload(factoryResp.payload)
	require.NoError(t, err)

	err = c.processEvent(ctx, event)
	s.require.NoError(err)

	s.require.Nil(err)

	args := factoryResp.args

	nft, err := models.FindNFTPrivilege(s.ctx, s.pdb.DBS().Reader, args.contract.Bytes(), args.tokenID, args.privilegeID, args.userAddress.Bytes())
	s.require.EqualError(err, "sql: no rows in result set")

	s.require.Nil(nft)
}

func TestUpdatedTimestamp(t *testing.T) {
	ctx := context.Background()

	s := initCEventsTestHelper(t)
	defer s.destroy()

	e := privilegeEventsPayloadFactory(3, 3, "", 0, s.settings.DIMORegistryChainID)
	factoryResp := e[0]

	c := NewContractsEventsConsumer(s.pdb, &s.logger, s.settings, nil, nil, nil, nil, nil, nil)

	event, err := marshalMockPayload(factoryResp.payload)
	require.NoError(t, err)

	err = c.processEvent(ctx, event)
	s.require.NoError(err)

	args := factoryResp.args

	oldNft, err := models.FindNFTPrivilege(s.ctx, s.pdb.DBS().Reader, args.contract.Bytes(), args.tokenID, args.privilegeID, args.userAddress.Bytes())
	s.require.NoError(err)

	s.require.NotNil(oldNft)

	expiry := time.Now().Add(time.Hour + time.Duration(4)).UTC().Unix()
	e = privilegeEventsPayloadFactory(3, 3, "", expiry, s.settings.DIMORegistryChainID)
	factoryResp = e[0]

	event, err = marshalMockPayload(factoryResp.payload)
	require.NoError(t, err)

	err = c.processEvent(ctx, event)
	s.require.NoError(err)

	a, _ := models.NFTPrivileges().All(s.ctx, s.pdb.DBS().Reader)
	s.require.Equal(len(a), 1)

	newNft, err := models.FindNFTPrivilege(s.ctx, s.pdb.DBS().Reader, args.contract.Bytes(), args.tokenID, args.privilegeID, args.userAddress.Bytes())
	s.require.NoError(err)

	actual := mockTestEntity{
		Contract:    newNft.ContractAddress,
		TokenID:     newNft.TokenID,
		PrivilegeID: newNft.Privilege,
		UserAddress: newNft.UserAddress,
		ExpiresAt:   newNft.Expiry,
	}

	expected := mockTestEntity{
		Contract:    args.contract.Bytes(),
		UserAddress: args.userAddress.Bytes(),
		TokenID:     args.tokenID,
		PrivilegeID: args.privilegeID,
		ExpiresAt:   time.Unix(expiry, 0).UTC(),
	}

	s.require.Equal(expected, actual, "Event was updated successful")
	s.require.Equal(oldNft.CreatedAt, newNft.CreatedAt)
	s.require.NotEqual(oldNft.UpdatedAt, newNft.UpdatedAt)
}

func Test_Transfer_Event_Handled_Correctly(t *testing.T) {
	ctx := context.Background()

	s := initCEventsTestHelper(t)
	defer s.destroy()

	tokenID := int64(4)
	nullTkID := types.NewDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(tokenID), 0))
	factoryResp := transferEventsPayloadFactory(2, 3, tokenID, s.settings.DIMORegistryChainID, AftermarketDeviceContractAddress)

	cm := common.BytesToAddress([]byte{uint8(9)})
	autopiUnit := models.AftermarketDevice{
		UserID:       null.StringFrom("SomeID"),
		OwnerAddress: null.BytesFrom(cm.Bytes()),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		TokenID:      types.NewDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(tokenID), 0)),
		Beneficiary:  null.BytesFrom(common.BytesToAddress([]byte{uint8(1)}).Bytes()),
	}

	err := autopiUnit.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.require.NoError(err)

	c := NewContractsEventsConsumer(s.pdb, &s.logger, s.settings, nil, nil, nil, nil, nil, nil)
	event, err := marshalMockPayload(factoryResp.payload)
	require.NoError(t, err)

	err = c.processEvent(ctx, event)
	s.require.NoError(err)

	aUnit, err := models.AftermarketDevices(models.AftermarketDeviceWhere.TokenID.EQ(nullTkID)).One(s.ctx, s.pdb.DBS().Reader)
	s.require.NoError(err)

	newOner := common.BytesToAddress([]byte{uint8(3)})
	s.require.Equal(aUnit.OwnerAddress, null.BytesFrom(newOner.Bytes()))
	s.require.Equal(null.String{}, aUnit.UserID)
	s.require.Equal(null.Bytes{Bytes: []byte{}}, aUnit.Beneficiary)
}

func Test_Ignore_Transfer_Mint_Event(t *testing.T) {
	ctx := context.Background()

	s := initCEventsTestHelper(t)
	defer s.destroy()

	tokenID := int64(4)
	factoryResp := transferEventsPayloadFactory(0, 3, tokenID, s.settings.DIMORegistryChainID, AftermarketDeviceContractAddress)

	tkID := types.NewDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(tokenID), 0))

	cm := common.BytesToAddress([]byte{uint8(9)})
	autopiUnit := models.AftermarketDevice{
		UserID:       null.StringFrom("SomeID"),
		OwnerAddress: null.BytesFrom(cm.Bytes()),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		TokenID:      tkID,
	}

	err := autopiUnit.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.require.NoError(err)

	c := NewContractsEventsConsumer(s.pdb, &s.logger, s.settings, nil, nil, nil, nil, nil, nil)

	event, err := marshalMockPayload(factoryResp.payload)
	require.NoError(t, err)

	err = c.processEvent(ctx, event)
	s.require.NoError(err)

	aUnit, err := models.AftermarketDevices(models.AftermarketDeviceWhere.TokenID.EQ(tkID)).One(s.ctx, s.pdb.DBS().Reader)
	s.require.NoError(err)
	s.require.Equal(autopiUnit.OwnerAddress, aUnit.OwnerAddress)
	s.require.Equal(autopiUnit.UserID, aUnit.UserID)
}

func Test_Ignore_Transfer_Claims_Event(t *testing.T) {
	ctx := context.Background()

	s := initCEventsTestHelper(t)
	defer s.destroy()

	tokenID := int64(4)
	factoryResp := transferEventsPayloadFactory(1, 3, tokenID, s.settings.DIMORegistryChainID, AftermarketDeviceContractAddress)

	tkID := types.NewDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(tokenID), 0))

	autopiUnit := models.AftermarketDevice{
		UserID:    null.StringFrom("SomeID"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		TokenID:   tkID,
	}

	err := autopiUnit.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.require.NoError(err)

	c := NewContractsEventsConsumer(s.pdb, &s.logger, s.settings, nil, nil, nil, nil, nil, nil)
	event, err := marshalMockPayload(factoryResp.payload)
	require.NoError(t, err)

	err = c.processEvent(ctx, event)
	s.require.NoError(err)

	aUnit, err := models.AftermarketDevices(models.AftermarketDeviceWhere.TokenID.EQ(tkID)).One(s.ctx, s.pdb.DBS().Reader)
	s.require.NoError(err)
	s.require.Equal(autopiUnit.OwnerAddress, aUnit.OwnerAddress)
	s.require.Equal(autopiUnit.UserID, aUnit.UserID)
}

func Test_Ignore_Transfer_Wrong_Contract(t *testing.T) {
	ctx := context.Background()

	s := initCEventsTestHelper(t)
	defer s.destroy()

	tokenID := int64(4)
	factoryResp := transferEventsPayloadFactory(1, 3, tokenID, s.settings.DIMORegistryChainID, "0x00000000000000000000000000000000000000c3")

	cm := common.BytesToAddress([]byte{uint8(1)})
	autopiUnit := models.AftermarketDevice{
		UserID:       null.StringFrom("SomeID"),
		OwnerAddress: null.BytesFrom(cm.Bytes()),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		TokenID:      types.NewDecimal(decimal.New(tokenID, 0)),
	}

	err := autopiUnit.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.require.NoError(err)

	c := NewContractsEventsConsumer(s.pdb, &s.logger, s.settings, nil, nil, nil, nil, nil, nil)

	event, err := marshalMockPayload(factoryResp.payload)
	require.NoError(t, err)

	err = c.processEvent(ctx, event)
	s.require.NoError(err)

	s.require.NoError(autopiUnit.Reload(s.ctx, s.pdb.DBS().Reader))
	s.require.Equal(autopiUnit.OwnerAddress, null.BytesFrom(cm.Bytes()))
}

func Test_Ignore_Transfer_Unit_Not_Found(t *testing.T) {
	ctx := context.Background()

	s := initCEventsTestHelper(t)
	defer s.destroy()

	tokenID := int64(4)
	factoryResp := transferEventsPayloadFactory(1, 3, 5, s.settings.DIMORegistryChainID, AftermarketDeviceContractAddress)

	cm := common.BytesToAddress([]byte{uint8(9)})
	autopiUnit := models.AftermarketDevice{
		UserID:       null.StringFrom("SomeID"),
		OwnerAddress: null.BytesFrom(cm.Bytes()),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		TokenID:      types.NewDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(tokenID), 0)),
	}

	err := autopiUnit.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.require.NoError(err)

	c := NewContractsEventsConsumer(s.pdb, &s.logger, s.settings, nil, nil, nil, nil, nil, nil)
	event, err := marshalMockPayload(factoryResp.payload)
	s.require.NoError(err)

	err = c.processEvent(ctx, event)
	s.require.EqualError(err, "record not found as this might be a newly minted device")
}

type beneficiaryCase struct {
	Name                      string
	Address                   common.Address
	Event                     ev
	AutopiUnitTable           models.AftermarketDevice
	ExpectedBeneficiaryResult null.Bytes
}
type ev struct {
	IdProxyAddress common.Address //nolint
	NodeId         *big.Int       //nolint
	Beneficiary    common.Address
}

func TestSetBeneficiary(t *testing.T) {
	ctx := context.Background()

	s := initCEventsTestHelper(t)
	defer s.destroy()

	cases := []beneficiaryCase{
		{
			Name:    "Ignore other contracts",
			Address: common.BigToAddress(big.NewInt(2)),
			Event: ev{
				IdProxyAddress: common.BigToAddress(big.NewInt(2)),
				NodeId:         big.NewInt(2),
				Beneficiary:    common.BigToAddress(big.NewInt(2)),
			},
			AutopiUnitTable: models.AftermarketDevice{
				OwnerAddress: null.BytesFrom(common.BigToAddress(big.NewInt(2)).Bytes()),
				TokenID:      types.NewDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(2), 0)),
			},
			ExpectedBeneficiaryResult: null.Bytes{Bytes: []byte{}},
		},
		{
			Name:    "Go from null to explicitly set beneficiary",
			Address: common.HexToAddress(s.settings.DIMORegistryAddr),
			Event: ev{
				IdProxyAddress: common.HexToAddress(s.settings.AftermarketDeviceContractAddress),
				NodeId:         big.NewInt(1),
				Beneficiary:    common.BigToAddress(big.NewInt(1)),
			},
			AutopiUnitTable: models.AftermarketDevice{
				OwnerAddress: null.BytesFrom(common.BigToAddress(big.NewInt(1)).Bytes()),
				TokenID:      types.NewDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(1), 0)),
			},
			ExpectedBeneficiaryResult: null.BytesFrom(common.BigToAddress(big.NewInt(1)).Bytes()),
		},
		{
			Name:    "Go from one explicitly set beneficiary to another",
			Address: common.HexToAddress(s.settings.DIMORegistryAddr),
			Event: ev{
				IdProxyAddress: common.HexToAddress(s.settings.AftermarketDeviceContractAddress),
				NodeId:         big.NewInt(3),
				Beneficiary:    common.BigToAddress(big.NewInt(3)),
			},
			AutopiUnitTable: models.AftermarketDevice{
				OwnerAddress: null.BytesFrom(common.BigToAddress(big.NewInt(1)).Bytes()),
				TokenID:      types.NewDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(3), 0)),
				Beneficiary:  null.BytesFrom(common.BigToAddress(big.NewInt(2)).Bytes()),
			},
			ExpectedBeneficiaryResult: null.BytesFrom(common.BigToAddress(big.NewInt(3)).Bytes()),
		},
		{
			Name:    "Go from beneficiary to explicitly cleared beneficiary",
			Address: common.HexToAddress(s.settings.DIMORegistryAddr),
			Event: ev{
				IdProxyAddress: common.HexToAddress(s.settings.AftermarketDeviceContractAddress),
				NodeId:         big.NewInt(3),
				Beneficiary:    common.BigToAddress(big.NewInt(0)),
			},
			AutopiUnitTable: models.AftermarketDevice{
				OwnerAddress: null.BytesFrom(common.BigToAddress(big.NewInt(1)).Bytes()),
				TokenID:      types.NewDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(3), 0)),
				Beneficiary:  null.BytesFrom(common.BigToAddress(big.NewInt(2)).Bytes()),
			},
			ExpectedBeneficiaryResult: null.Bytes{Bytes: []byte{}},
		},
	}

	for _, c := range cases {
		err := c.AutopiUnitTable.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
		s.require.NoError(err)

		abi, err := contracts.RegistryMetaData.GetAbi()
		s.require.NoError(err)

		payload := fmt.Sprintf(`{
		"data": {
			"contract": "%s",
			"eventSignature": "%s",
			"eventName": "BeneficiarySet",
			"arguments": {
				"nodeID": %d,
				"beneficiary": "%s",
				"idProxyAddress": "%s"
			}
		},
		"type": "zone.dimo.contract.event",	
		"source": "chain/%d"
		}`, c.Address.Hex(), abi.Events["BeneficiarySet"].ID, c.Event.NodeId, c.Event.Beneficiary.Hex(), c.Event.IdProxyAddress.Hex(), s.settings.DIMORegistryChainID)

		consumer := NewContractsEventsConsumer(s.pdb, &s.logger, s.settings, nil, nil, nil, nil, nil, nil)

		event, err := marshalMockPayload(payload)
		require.NoError(t, err)

		err = consumer.processEvent(ctx, event)
		s.require.NoError(err)

		err = c.AutopiUnitTable.Reload(s.ctx, s.pdb.DBS().Reader)
		s.require.NoError(err)

		s.require.Equal(c.ExpectedBeneficiaryResult, c.AutopiUnitTable.Beneficiary)

		test.TruncateTables(s.pdb.DBS().Writer.DB, t)
	}
}

func TestVehicleTransfer(t *testing.T) {
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer container.Terminate(ctx) //nolint

	logger := zerolog.Nop()
	settings := &config.Settings{DIMORegistryChainID: 1, VehicleNFTAddress: "0x881d40237659c251811cec9c364ef91dc08d300c"}

	mtr := models.MetaTransactionRequest{ID: "xdd"}
	_ = mtr.Insert(ctx, pdb.DBS().Writer, boil.Infer())

	ud := models.UserDevice{
		ID:                 ksuid.New().String(),
		MintRequestID:      null.StringFrom(mtr.ID),
		OwnerAddress:       null.BytesFrom(common.FromHex("0xdafea492d9c6733ae3d56b7ed1adb60692c98bc5")),
		TokenID:            types.NewNullDecimal(decimal.New(5, 0)),
		DeviceDefinitionID: ksuid.New().String(),
	}
	_ = ud.Insert(ctx, pdb.DBS().Writer, boil.Infer())

	consumer := NewContractsEventsConsumer(pdb, &logger, settings, nil, nil, nil, nil, nil, nil)
	event, err := marshalMockPayload(`
	{
		"type": "zone.dimo.contract.event",
		"source": "chain/1",
		"data": {
			"contract": "0x881d40237659c251811cec9c364ef91dc08d300c",
			"eventName": "Transfer",
			"arguments": {
				"from": "0xdafea492d9c6733ae3d56b7ed1adb60692c98bc5",
				"to": "0x4675c7e5baafbffbca748158becba61ef3b0a263",
				"tokenId": 5
			}
		}
	}
	`)
	require.NoError(t, err)

	err = consumer.processEvent(ctx, event)
	if err != nil {
		t.Errorf("failed to process event: %v", err)
	}

	_ = ud.Reload(ctx, pdb.DBS().Reader)
	if !ud.OwnerAddress.Valid {
		t.Fatal("token owner became null")
	}

	if common.BytesToAddress(ud.OwnerAddress.Bytes) != common.HexToAddress("0x4675c7e5baafbffbca748158becba61ef3b0a263") {
		t.Errorf("expected owner to become %s, but was %s", common.HexToAddress("0x4675c7e5baafbffbca748158becba61ef3b0a263"), common.BytesToAddress(ud.OwnerAddress.Bytes))
	}
}

func Test_NFTPrivileges_Cleared_On_Vehicle_Transfer(t *testing.T) {
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer container.Terminate(ctx) //nolint

	require := require.New(t)
	logger := zerolog.Nop()
	settings := &config.Settings{DIMORegistryChainID: 1, VehicleNFTAddress: "0x881d40237659c251811cec9c364ef91dc08d300c"}

	mtr := models.MetaTransactionRequest{ID: "xdd"}
	_ = mtr.Insert(ctx, pdb.DBS().Writer, boil.Infer())

	tkID := types.NewNullDecimal(decimal.New(5, 0))
	ownerAddress := null.BytesFrom(common.FromHex("0xdafea492d9c6733ae3d56b7ed1adb60692c98bc5"))

	nftPriv := models.NFTPrivilege{
		TokenID:         types.Decimal(tkID),
		ContractAddress: common.BytesToAddress([]byte{uint8(1)}).Bytes(),
		Privilege:       1,
		UserAddress:     ownerAddress.Bytes,
		Expiry:          time.Now(),
	}
	_ = nftPriv.Insert(ctx, pdb.DBS().Writer, boil.Infer())

	ud := models.UserDevice{
		ID:                 ksuid.New().String(),
		MintRequestID:      null.StringFrom(mtr.ID),
		OwnerAddress:       null.BytesFrom(common.FromHex("0xdafea492d9c6733ae3d56b7ed1adb60692c98bc5")),
		TokenID:            types.NewNullDecimal(decimal.New(5, 0)),
		DeviceDefinitionID: ksuid.New().String(),
	}
	_ = ud.Insert(ctx, pdb.DBS().Writer, boil.Infer())

	consumer := NewContractsEventsConsumer(pdb, &logger, settings, nil, nil, nil, nil, nil, nil)
	event, err := marshalMockPayload(`
	{
		"type": "zone.dimo.contract.event",
		"source": "chain/1",
		"data": {
			"contract": "0x881d40237659c251811cec9c364ef91dc08d300c",
			"eventName": "Transfer",
			"arguments": {
				"from": "0xdafea492d9c6733ae3d56b7ed1adb60692c98bc5",
				"to": "0x4675c7e5baafbffbca748158becba61ef3b0a263",
				"tokenId": 5
			}
		}
	}
	`)
	require.NoError(err)

	err = consumer.processEvent(ctx, event)
	if err != nil {
		t.Errorf("failed to process event: %v", err)
	}

	_ = ud.Reload(ctx, pdb.DBS().Reader)
	if !ud.OwnerAddress.Valid {
		t.Fatal("token owner became null")
	}

	if common.BytesToAddress(ud.OwnerAddress.Bytes) != common.HexToAddress("0x4675c7e5baafbffbca748158becba61ef3b0a263") {
		t.Errorf("expected owner to become %s, but was %s", common.HexToAddress("0x4675c7e5baafbffbca748158becba61ef3b0a263"), common.BytesToAddress(ud.OwnerAddress.Bytes))
	}

	nftPrivileges, err := models.NFTPrivileges().All(ctx, pdb.DBS().Reader)
	require.NoError(err)
	require.Equal(0, len(nftPrivileges))
}

func Test_RegistryAftermarketDeviceAddressReset(t *testing.T) {
	ctx := context.Background()

	s := initCEventsTestHelper(t)
	defer s.destroy()

	logger := zerolog.Nop()
	s.settings.DIMORegistryAddr = common.BigToAddress(big.NewInt(7)).Hex()

	tokenID := types.NewDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(1), 0))
	updatedEthAddr := common.HexToAddress("0x19995Cee27AbBe71b85A09B73D24EA26Fa9325a0")

	amd := models.AftermarketDevice{
		UserID:          null.StringFrom("SomeID"),
		EthereumAddress: common.BigToAddress(big.NewInt(1)).Bytes(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		TokenID:         tokenID,
	}

	payload :=
		fmt.Sprintf(`{
			"type": "zone.dimo.contract.event",
			"source": "chain/%d",
			"data": {
				"contract": "%s",
				"eventName": "%s",
				"chainId": %d,
				"arguments": {
				"manufacturerId": 2,
				"tokenId": 1,
				"aftermarketDeviceAddress": "%s"
				}
			}
		}`,
			s.settings.DIMORegistryChainID,
			s.settings.DIMORegistryAddr,
			AftermarketDeviceAddressReset.String(),
			s.settings.DIMORegistryChainID, updatedEthAddr)
	err := amd.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.require.NoError(err)

	consumer := NewContractsEventsConsumer(s.pdb, &logger, s.settings, nil, nil, nil, nil, nil, nil)
	event, err := marshalMockPayload(payload)
	require.NoError(t, err)

	err = consumer.processEvent(ctx, event)
	s.require.NoError(err)

	updatedAmd, err := models.AftermarketDevices(models.AftermarketDeviceWhere.TokenID.EQ(tokenID)).One(s.ctx, s.pdb.DBS().Reader)
	s.require.NoError(err)

	s.require.Equal(updatedEthAddr, common.BytesToAddress(updatedAmd.EthereumAddress))
}

func convertTokenIDToDecimal(t string) types.Decimal {
	ti, ok := new(decimal.Big).SetString(t)
	if !ok {
		return types.Decimal{}
	}

	return types.NewDecimal(ti)
}

func transferEventsPayloadFactory(from, to int, tokenID int64, dimoChainID int64, contractAddrress string) eventsFactoryResp {
	contractAddr := common.HexToAddress(contractAddrress)
	frmAddr := common.BytesToAddress([]byte{uint8(from)})
	toAddr := common.BytesToAddress([]byte{uint8(to)})
	tkID := convertTokenIDToDecimal(fmt.Sprint(tokenID))

	payload := fmt.Sprintf(`{
		"data": {
			"contract": "%s",
			"transactionHash": "0x29d1aa4f5eb409bf7d334a7f50fcba50264fbefe00c991cc278f444eb64fdfe5",
			"eventSignature": "0x61a24679288162b799d80b2bb2b8b0fcdd5c5f53ac19e9246cc190b60196c359",
			"eventName": "%s",
			"arguments": {
				"from": "%s",
				"to": "%s",
				"tokenId": %s
			}
		},
		"type": "zone.dimo.contract.event",
		"source": "chain/%d"
	}`, contractAddr.String(), "Transfer", frmAddr.String(), toAddr.String(), tkID, dimoChainID)

	return eventsFactoryResp{
		payload: payload,
		args: mockTestArgs{
			contract: contractAddr,
			tokenID:  tkID,
			from:     frmAddr,
			to:       toAddr,
		},
	}
}

// Utility/Helper functions

// Creates a specific number of payloads that the event consumer can parse and process. Can start at an arbitrary point in the array.
// @Param from - start from index.
// @Param to - end at index.
// @Param eventName - name of event we are generating for
// @Param exp - expiry of event
// @Param dIMORegistryChainID - chainId to include with the
func privilegeEventsPayloadFactory(from, to int, eventName string, exp int64, dIMORegistryChainID int64) []eventsFactoryResp {
	res := []eventsFactoryResp{}

	if eventName == "" {
		eventName = "PrivilegeSet"
	}

	for i := from; i <= to; i++ {
		contractAddr := common.BytesToAddress([]byte{uint8(i)})
		userAddr := common.BytesToAddress([]byte{uint8(i + 1)})
		tokenID := convertTokenIDToDecimal(fmt.Sprint(i))
		privID := i + 1
		expiry := time.Now().Add(time.Hour + time.Duration(i)*time.Minute).UTC().Unix()

		if exp != 0 {
			expiry = exp
		}

		payload := fmt.Sprintf(`{
				"data": {
					"contract": "%s",
					"transactionHash": "0x29d1aa4f5eb409bf7d334a7f50fcba50264fbefe00c991cc278f444eb64fdfe5",
					"eventSignature": "0x61a24679288162b799d80b2bb2b8b0fcdd5c5f53ac19e9246cc190b60196c359",
					"eventName": "%s",
					"arguments": {
						"tokenId":  %s,
						"version": 1,
						"privId": %d,
						"user": "%s",
						"expires": %d
					}
				},
				"type": "zone.dimo.contract.event",
				"source": "chain/%d"
			}`, contractAddr.String(), eventName, tokenID, privID, userAddr.String(), expiry, dIMORegistryChainID)

		res = append(res, eventsFactoryResp{
			payload: payload,
			args: mockTestArgs{
				contract:    contractAddr,
				tokenID:     tokenID,
				userAddress: userAddr,
				expiresAt:   expiry,
				privilegeID: int64(privID),
			},
		})
	}
	return res
}

func Test_VehicleNodeMintedWithDeviceDefinition_NoMtx(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer container.Terminate(ctx) //nolint

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	require := require.New(t)
	settings := &config.Settings{DIMORegistryChainID: 1, DIMORegistryAddr: "0x881d40237659c251811cec9c364ef91dc08d300c"}
	deviceDefSvc := NewMockDeviceDefinitionService(mockCtrl)

	kprod := smock.NewSyncProducer(t, nil)
	evt := NewEventService(&logger, settings, kprod)
	kprod.ExpectSendMessageAndSucceed()
	consumer := NewContractsEventsConsumer(pdb, &logger, settings, nil, nil, deviceDefSvc, evt, nil, nil)

	owner := common.HexToAddress("0xdafea492d9c6733ae3d56b7ed1adb60692c98bc5")
	ddSlug := "jeep_wrangler_2013"
	deviceDefID := ksuid.New().String()
	event, err := marshalMockPayload(fmt.Sprintf(`{
			"type": "zone.dimo.contract.event",
			"source": "chain/%d",
			"data": {
				"contract": "%s",
				"eventName": "%s",
				"chainId": %d,
				"arguments": {
				"manufacturerId": %d,
				"vehicleId": %d,
				"owner": "%s",
				"deviceDefinitionId": "%s"
				}
			}
		}`,
		settings.DIMORegistryChainID,
		settings.DIMORegistryAddr,
		VehicleNodeMintedWithDeviceDefinition.String(),
		settings.DIMORegistryChainID,
		7,           // manufacturerId
		13,          // vehicleId
		owner.Hex(), // owner
		ddSlug,      // device definition id
	))
	require.NoError(err)

	deviceDefSvc.EXPECT().GetDeviceDefinitionBySlugName(gomock.Any(), &ddgrpc.GetDeviceDefinitionBySlugNameRequest{
		Slug: ddSlug,
	}).Return(&ddgrpc.GetDeviceDefinitionItemResponse{
		DeviceDefinitionId: deviceDefID,
		Make: &ddgrpc.DeviceMake{
			TokenId: 7,
		},
	}, nil)

	err = consumer.processEvent(ctx, event)
	require.NoError(err)

	ud, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.EQ(types.NewNullDecimal(decimal.New(13, 0))),
	).One(ctx, pdb.DBS().Reader)
	require.NoError(err)

	require.Equal(deviceDefID, ud.DeviceDefinitionID)
	require.Equal(owner.Hex(), common.BytesToAddress(ud.OwnerAddress.Bytes).Hex())

	userID, err := proto.Marshal(&dex.IDTokenSubject{
		UserId: owner.Hex(),
		ConnId: "web3",
	})
	require.NoError(err)
	require.Equal(base64.RawURLEncoding.EncodeToString(userID), ud.UserID)
}

func TestBurnSyntheticDevice(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()
	sdToken := 4
	vehicleID := 54
	teslaIntID := 2

	teslaIntegrationID := ksuid.New().String()

	chainID := 1
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer container.Terminate(ctx) //nolint

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	sdAddr := randomAddr(t)

	settings := &config.Settings{DIMORegistryChainID: int64(chainID), SyntheticDeviceNFTAddress: sdAddr.Hex()}
	deviceDefSvc := NewMockDeviceDefinitionService(mockCtrl)

	teslaTask := NewMockSyntheticTaskService(mockCtrl)

	mtr := models.MetaTransactionRequest{
		ID:     ksuid.New().String(),
		Status: models.MetaTransactionRequestStatusConfirmed,
	}
	err := mtr.Insert(ctx, pdb.DBS().Reader, boil.Infer())
	require.NoError(t, err)

	teslaInteg := &ddgrpc.Integration{
		Id:     teslaIntegrationID,
		Vendor: constants.TeslaVendor,
	}

	ud := models.UserDevice{
		ID:                 ksuid.New().String(),
		UserID:             "xdd",
		DeviceDefinitionID: ksuid.New().String(),
		TokenID:            types.NewNullDecimal(decimal.New(int64(vehicleID), 0)),
	}
	err = ud.Insert(ctx, pdb.DBS().Reader, boil.Infer())
	require.NoError(t, err)

	deviceDefSvc.EXPECT().GetIntegrationByTokenID(gomock.Any(), uint64(teslaIntID)).Return(teslaInteg, nil)
	deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), ud.DeviceDefinitionID).Return(&ddgrpc.GetDeviceDefinitionItemResponse{
		Make: &ddgrpc.DeviceMake{
			Name: "Tesla",
		},
		Type: &ddgrpc.DeviceType{
			Model: "Model X",
			Year:  2024,
		},
	}, nil)

	udai := models.UserDeviceAPIIntegration{
		UserDeviceID:  ud.ID,
		IntegrationID: teslaInteg.Id,
		TaskID:        null.StringFrom(ksuid.New().String()),
		Status:        models.UserDeviceAPIIntegrationStatusActive,
	}
	err = udai.Insert(ctx, pdb.DBS().Reader, boil.Infer())
	require.NoError(t, err)

	sd := models.SyntheticDevice{
		VehicleTokenID:     ud.TokenID,
		IntegrationTokenID: types.NewDecimal(decimal.New(2, 0)),
		MintRequestID:      mtr.ID,
		WalletChildNumber:  1,
		WalletAddress:      randomAddr(t).Bytes(),
		TokenID:            types.NewNullDecimal(decimal.New(int64(sdToken), 0)),
	}
	err = sd.Insert(ctx, pdb.DBS().Reader, boil.Infer())
	require.NoError(t, err)

	kprod := smock.NewSyncProducer(t, nil)
	evt := NewEventService(&logger, settings, kprod)
	kprod.ExpectSendMessageAndSucceed()
	consumer := NewContractsEventsConsumer(pdb, &logger, settings, nil, nil, deviceDefSvc, evt, nil, teslaTask)

	ownerAddr := randomAddr(t)

	ced := ContractEventData{
		ChainID:   int64(chainID),
		EventName: "Transfer",
		Contract:  sdAddr,
		Arguments: []byte(fmt.Sprintf(`{"from": "%s", "to": "%s", "tokenId": %d}`, ownerAddr, zeroAddr, sdToken)),
	}

	b, _ := json.Marshal(ced)

	teslaTask.EXPECT().StopPoll(gomock.Any())

	err = consumer.processEvent(ctx, &shared.CloudEvent[json.RawMessage]{
		Source: fmt.Sprintf("chain/%d", chainID),
		Type:   contractEventCEType,
		Data:   b,
	})
	require.NoError(t, err)

	err = sd.Reload(ctx, pdb.DBS().Reader)
	require.ErrorIs(t, err, sql.ErrNoRows)

	err = udai.Reload(ctx, pdb.DBS().Reader)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func initCEventsTestHelper(t *testing.T) cEventsTestHelper {
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	settings := &config.Settings{AutoPiAPIToken: "fdff", DIMORegistryChainID: 1, AftermarketDeviceContractAddress: AftermarketDeviceContractAddress}

	return cEventsTestHelper{
		logger:    zerolog.New(os.Stdout).With().Timestamp().Logger(),
		pdb:       pdb,
		container: container,
		ctx:       ctx,
		t:         t,
		require:   require.New(t),
		settings:  settings,
	}
}

func (s cEventsTestHelper) destroy() {
	if err := s.container.Terminate(s.ctx); err != nil {
		s.t.Fatal(err)
	}
}

func randomAddr(t *testing.T) common.Address {
	addr := make([]byte, common.AddressLength)
	_, err := rand.Read(addr)
	if err != nil {
		t.Fatalf("couldn't create a test address: %v", err)
	}
	return common.Address(addr)
}
