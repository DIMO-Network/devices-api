package services

import (
	"context"
	"encoding/json"
	"math/big"

	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
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
	assert    *assert.Assertions
	settings  *config.Settings
}

type eventsFactoryResp struct {
	args    mockTestArgs
	payload string
}

const AftermarketDeviceContractAddress = "0x00000000000000000000000000000000000000c1"

func TestProcessContractsEventsMessages(t *testing.T) {
	s := initCEventsTestHelper(t)
	defer s.destroy()

	e := privilegeEventsPayloadFactory(1, 1, "", 0, s.settings.DIMORegistryChainID)
	factoryResp := e[0]

	msg := &message.Message{
		Payload: []byte(factoryResp.payload),
	}

	c := NewContractsEventsConsumer(s.pdb, &s.logger, s.settings, nil)

	err := c.processMessage(msg)
	s.assert.NoError(err)

	args := factoryResp.args

	nft, err := models.FindNFTPrivilege(s.ctx, s.pdb.DBS().Reader, args.contract.Bytes(), args.tokenID, args.privilegeID, args.userAddress.Bytes())
	s.assert.NoError(err)

	s.assert.NotNil(nft)

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

	s.assert.Equal(expected, actual, "Event was persisted properly")
}

func TestIgnoreWrongEventNames(t *testing.T) {
	s := initCEventsTestHelper(t)
	defer s.destroy()

	e := privilegeEventsPayloadFactory(2, 2, "SomeEvent", 0, s.settings.DIMORegistryChainID)
	factoryResp := e[0]

	msg := &message.Message{
		Payload: []byte(factoryResp.payload),
	}
	c := NewContractsEventsConsumer(s.pdb, &s.logger, s.settings, nil)

	err := c.processMessage(msg)
	s.assert.NoError(err)

	s.assert.Nil(err)

	args := factoryResp.args

	nft, err := models.FindNFTPrivilege(s.ctx, s.pdb.DBS().Reader, args.contract.Bytes(), args.tokenID, args.privilegeID, args.userAddress.Bytes())
	s.assert.EqualError(err, "sql: no rows in result set")

	s.assert.Nil(nft)
}

func TestUpdatedTimestamp(t *testing.T) {
	s := initCEventsTestHelper(t)
	defer s.destroy()

	e := privilegeEventsPayloadFactory(3, 3, "", 0, s.settings.DIMORegistryChainID)
	factoryResp := e[0]

	c := NewContractsEventsConsumer(s.pdb, &s.logger, s.settings, nil)

	msg := &message.Message{
		Payload: []byte(factoryResp.payload),
	}
	err := c.processMessage(msg)
	s.assert.NoError(err)

	args := factoryResp.args

	oldNft, err := models.FindNFTPrivilege(s.ctx, s.pdb.DBS().Reader, args.contract.Bytes(), args.tokenID, args.privilegeID, args.userAddress.Bytes())
	s.assert.NoError(err)

	s.assert.NotNil(oldNft)

	expiry := time.Now().Add(time.Hour + time.Duration(4)).UTC().Unix()
	e = privilegeEventsPayloadFactory(3, 3, "", expiry, s.settings.DIMORegistryChainID)
	factoryResp = e[0]

	msg = &message.Message{
		Payload: []byte(factoryResp.payload),
	}
	err = c.processMessage(msg)
	s.assert.NoError(err)

	a, _ := models.NFTPrivileges().All(s.ctx, s.pdb.DBS().Reader)
	s.assert.Equal(len(a), 1)

	newNft, err := models.FindNFTPrivilege(s.ctx, s.pdb.DBS().Reader, args.contract.Bytes(), args.tokenID, args.privilegeID, args.userAddress.Bytes())
	s.assert.NoError(err)

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
		ExpiresAt:   time.Unix(expiry, 0).UTC(),
		PrivilegeID: args.privilegeID,
	}

	s.assert.Equal(expected, actual, "Event was updated successful")
	s.assert.Equal(oldNft.CreatedAt, newNft.CreatedAt)
	s.assert.NotEqual(oldNft.UpdatedAt, newNft.UpdatedAt)
}

func Test_Transfer_Event_Handled_Correctly(t *testing.T) {
	s := initCEventsTestHelper(t)
	defer s.destroy()

	tokenID := int64(4)
	nullTkID := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(tokenID), 0))
	factoryResp := transferEventsPayloadFactory(2, 3, tokenID, s.settings.DIMORegistryChainID, AftermarketDeviceContractAddress)

	msg := &message.Message{
		Payload: []byte(factoryResp.payload),
	}

	cm := common.BytesToAddress([]byte{uint8(9)})
	autopiUnit := models.AftermarketDevice{
		UserID:       null.StringFrom("SomeID"),
		OwnerAddress: null.BytesFrom(cm.Bytes()),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		TokenID:      types.NewNullDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(tokenID), 0)),
		Beneficiary:  null.BytesFrom(common.BytesToAddress([]byte{uint8(1)}).Bytes()),
	}

	err := autopiUnit.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.assert.NoError(err)

	c := NewContractsEventsConsumer(s.pdb, &s.logger, s.settings, nil)

	err = c.processMessage(msg)
	s.assert.NoError(err)

	aUnit, err := models.AftermarketDevices(models.AftermarketDeviceWhere.TokenID.EQ(nullTkID)).One(s.ctx, s.pdb.DBS().Reader)
	s.assert.NoError(err)

	newOner := common.BytesToAddress([]byte{uint8(3)})
	s.assert.Equal(aUnit.OwnerAddress, null.BytesFrom(newOner.Bytes()))
	s.assert.Equal(null.String{}, aUnit.UserID)
	s.assert.Equal(null.Bytes{Bytes: []byte{}}, aUnit.Beneficiary)
}

func Test_Ignore_Transfer_Mint_Event(t *testing.T) {
	s := initCEventsTestHelper(t)
	defer s.destroy()

	tokenID := int64(4)
	factoryResp := transferEventsPayloadFactory(0, 3, tokenID, s.settings.DIMORegistryChainID, AftermarketDeviceContractAddress)

	msg := &message.Message{
		Payload: []byte(factoryResp.payload),
	}

	tkID := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(tokenID), 0))

	cm := common.BytesToAddress([]byte{uint8(9)})
	autopiUnit := models.AftermarketDevice{
		UserID:       null.StringFrom("SomeID"),
		OwnerAddress: null.BytesFrom(cm.Bytes()),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		TokenID:      tkID,
	}

	err := autopiUnit.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.assert.NoError(err)

	c := NewContractsEventsConsumer(s.pdb, &s.logger, s.settings, nil)

	err = c.processMessage(msg)
	s.assert.NoError(err)

	aUnit, err := models.AftermarketDevices(models.AftermarketDeviceWhere.TokenID.EQ(tkID)).One(s.ctx, s.pdb.DBS().Reader)
	s.assert.NoError(err)
	s.assert.Equal(autopiUnit.OwnerAddress, aUnit.OwnerAddress)
	s.assert.Equal(autopiUnit.UserID, aUnit.UserID)
}

func Test_Ignore_Transfer_Claims_Event(t *testing.T) {
	s := initCEventsTestHelper(t)
	defer s.destroy()

	tokenID := int64(4)
	factoryResp := transferEventsPayloadFactory(1, 3, tokenID, s.settings.DIMORegistryChainID, AftermarketDeviceContractAddress)

	msg := &message.Message{
		Payload: []byte(factoryResp.payload),
	}

	tkID := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(tokenID), 0))

	autopiUnit := models.AftermarketDevice{
		UserID:    null.StringFrom("SomeID"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		TokenID:   tkID,
	}

	err := autopiUnit.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.assert.NoError(err)

	c := NewContractsEventsConsumer(s.pdb, &s.logger, s.settings, nil)

	err = c.processMessage(msg)
	s.assert.NoError(err)

	aUnit, err := models.AftermarketDevices(models.AftermarketDeviceWhere.TokenID.EQ(tkID)).One(s.ctx, s.pdb.DBS().Reader)
	s.assert.NoError(err)
	s.assert.Equal(autopiUnit.OwnerAddress, aUnit.OwnerAddress)
	s.assert.Equal(autopiUnit.UserID, aUnit.UserID)
}

func Test_Ignore_Transfer_Wrong_Contract(t *testing.T) {
	s := initCEventsTestHelper(t)
	defer s.destroy()

	tokenID := int64(4)
	factoryResp := transferEventsPayloadFactory(1, 3, tokenID, s.settings.DIMORegistryChainID, "0x00000000000000000000000000000000000000c3")

	msg := &message.Message{
		Payload: []byte(factoryResp.payload),
	}

	cm := common.BytesToAddress([]byte{uint8(1)})
	autopiUnit := models.AftermarketDevice{
		UserID:       null.StringFrom("SomeID"),
		OwnerAddress: null.BytesFrom(cm.Bytes()),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		TokenID:      types.NewNullDecimal(decimal.New(tokenID, 0)),
	}

	err := autopiUnit.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.assert.NoError(err)

	c := NewContractsEventsConsumer(s.pdb, &s.logger, s.settings, nil)

	err = c.processMessage(msg)
	s.assert.NoError(err)

	s.assert.NoError(autopiUnit.Reload(s.ctx, s.pdb.DBS().Reader))
	s.assert.Equal(autopiUnit.OwnerAddress, null.BytesFrom(cm.Bytes()))
}

func Test_Ignore_Transfer_Unit_Not_Found(t *testing.T) {
	s := initCEventsTestHelper(t)
	defer s.destroy()

	tokenID := int64(4)
	factoryResp := transferEventsPayloadFactory(1, 3, 5, s.settings.DIMORegistryChainID, AftermarketDeviceContractAddress)

	msg := &message.Message{
		Payload: []byte(factoryResp.payload),
	}

	cm := common.BytesToAddress([]byte{uint8(9)})
	autopiUnit := models.AftermarketDevice{
		UserID:       null.StringFrom("SomeID"),
		OwnerAddress: null.BytesFrom(cm.Bytes()),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		TokenID:      types.NewNullDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(tokenID), 0)),
	}

	err := autopiUnit.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.assert.NoError(err)

	c := NewContractsEventsConsumer(s.pdb, &s.logger, s.settings, nil)

	err = c.processMessage(msg)
	s.assert.EqualError(err, "record not found as this might be a newly minted device")
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
				TokenID:      types.NewNullDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(2), 0)),
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
				TokenID:      types.NewNullDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(1), 0)),
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
				TokenID:      types.NewNullDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(3), 0)),
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
				TokenID:      types.NewNullDecimal(new(decimal.Big).SetBigMantScale(big.NewInt(3), 0)),
				Beneficiary:  null.BytesFrom(common.BigToAddress(big.NewInt(2)).Bytes()),
			},
			ExpectedBeneficiaryResult: null.Bytes{Bytes: []byte{}},
		},
	}

	for _, c := range cases {
		err := c.AutopiUnitTable.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
		s.assert.NoError(err)

		consumer := NewContractsEventsConsumer(s.pdb, &s.logger, s.settings, nil)

		b, err := json.Marshal(c.Event)
		s.assert.NoError(err)

		abi, err := contracts.RegistryMetaData.GetAbi()
		s.assert.NoError(err)

		ce := shared.CloudEvent[ContractEventData]{
			Source: fmt.Sprintf("chain/%d", s.settings.DIMORegistryChainID),
			Type:   "zone.dimo.contract.event",
			Data: ContractEventData{
				Contract:       c.Address,
				EventName:      "BeneficiarySet",
				EventSignature: abi.Events["BeneficiarySet"].ID,
				Arguments:      b,
			},
		}

		b, err = json.Marshal(ce)
		s.assert.NoError(err)

		err = consumer.processMessage(&message.Message{Payload: b})
		s.assert.NoError(err)

		err = c.AutopiUnitTable.Reload(s.ctx, s.pdb.DBS().Reader)
		s.assert.NoError(err)

		s.assert.Equal(c.ExpectedBeneficiaryResult, c.AutopiUnitTable.Beneficiary)

		test.TruncateTables(s.pdb.DBS().Writer.DB, t)
	}
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

func initCEventsTestHelper(t *testing.T) cEventsTestHelper {
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	assert := assert.New(t)
	settings := &config.Settings{AutoPiAPIToken: "fdff", DIMORegistryChainID: 1, AftermarketDeviceContractAddress: AftermarketDeviceContractAddress}

	return cEventsTestHelper{
		logger:    zerolog.New(os.Stdout).With().Timestamp().Logger(),
		pdb:       pdb,
		container: container,
		ctx:       ctx,
		t:         t,
		assert:    assert,
		settings:  settings,
	}
}

func (s cEventsTestHelper) destroy() {
	if err := s.container.Terminate(s.ctx); err != nil {
		s.t.Fatal(err)
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

	nft := models.VehicleNFT{MintRequestID: "xdd", OwnerAddress: null.BytesFrom(common.FromHex("0xdafea492d9c6733ae3d56b7ed1adb60692c98bc5")), TokenID: types.NewNullDecimal(decimal.New(5, 0))}
	_ = nft.Insert(ctx, pdb.DBS().Writer, boil.Infer())

	consumer := NewContractsEventsConsumer(pdb, &logger, settings, nil)
	err := consumer.processMessage(&message.Message{Payload: []byte(`
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
	`)})
	if err != nil {
		t.Errorf("failed to process event: %v", err)
	}

	_ = nft.Reload(ctx, pdb.DBS().Reader)
	if !nft.OwnerAddress.Valid {
		t.Fatal("token owner became null")
	}

	if common.BytesToAddress(nft.OwnerAddress.Bytes) != common.HexToAddress("0x4675c7e5baafbffbca748158becba61ef3b0a263") {
		t.Errorf("expected owner to become %s, but was %s", common.HexToAddress("0x4675c7e5baafbffbca748158becba61ef3b0a263"), common.BytesToAddress(nft.OwnerAddress.Bytes))
	}
}

func Test_NFTPrivileges_Cleared_On_Vehicle_Transfer(t *testing.T) {
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer container.Terminate(ctx) //nolint

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

	nft := models.VehicleNFT{MintRequestID: "xdd", OwnerAddress: ownerAddress, TokenID: tkID}
	_ = nft.Insert(ctx, pdb.DBS().Writer, boil.Infer())

	consumer := NewContractsEventsConsumer(pdb, &logger, settings, nil)
	err := consumer.processMessage(&message.Message{Payload: []byte(`
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
	`)})
	if err != nil {
		t.Errorf("failed to process event: %v", err)
	}

	_ = nft.Reload(ctx, pdb.DBS().Reader)
	if !nft.OwnerAddress.Valid {
		t.Fatal("token owner became null")
	}

	if common.BytesToAddress(nft.OwnerAddress.Bytes) != common.HexToAddress("0x4675c7e5baafbffbca748158becba61ef3b0a263") {
		t.Errorf("expected owner to become %s, but was %s", common.HexToAddress("0x4675c7e5baafbffbca748158becba61ef3b0a263"), common.BytesToAddress(nft.OwnerAddress.Bytes))
	}

	nftPrivileges, err := models.NFTPrivileges().All(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(nftPrivileges))
}
