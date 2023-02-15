package services

import (
	"context"
	"log"

	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
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
}

type cEventsTestHelper struct {
	logger    zerolog.Logger
	pdb       db.Store
	container testcontainers.Container
	ctx       context.Context
	t         *testing.T
	assert    *assert.Assertions
}

type eventsFactoryResp struct {
	args    mockTestArgs
	payload string
}

func TestProcessContractsEventsMessages(t *testing.T) {
	s := initCEventsTestHelper(t)
	defer s.destroy()

	e := eventsPayloadFactory(1, 1, "", 0)
	factoryResp := e[0]

	msg := &message.Message{
		Payload: []byte(factoryResp.payload),
	}
	c := NewContractsEventsConsumer(s.pdb, &s.logger)

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

	e := eventsPayloadFactory(2, 2, "SomeEvent", 0)
	factoryResp := e[0]
	log.Println(factoryResp)
	msg := &message.Message{
		Payload: []byte(factoryResp.payload),
	}
	c := NewContractsEventsConsumer(s.pdb, &s.logger)

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

	e := eventsPayloadFactory(3, 3, "", 0)
	factoryResp := e[0]

	c := NewContractsEventsConsumer(s.pdb, &s.logger)

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
	e = eventsPayloadFactory(3, 3, "", expiry)
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
	s.assert.NotEqual(oldNft.UpdatedAt, newNft.UpdatedAt)
}

// Utility/Helper functions
func eventsPayloadFactory(from, to int, eventName string, exp int64) []eventsFactoryResp {
	res := []eventsFactoryResp{}

	convertTokenIDToDecimal := func(t string) types.Decimal {
		ti, ok := new(decimal.Big).SetString(t)
		if !ok {
			return types.Decimal{}
		}

		return types.NewDecimal(ti)
	}

	if eventName == "" {
		eventName = "PrivilegeSet"
	}

	for i := from; i <= to; i++ {
		contractAddr := common.BytesToAddress([]byte{uint8(i)})
		userAddr := common.BytesToAddress([]byte{uint8(i + 1)})
		tokenID := convertTokenIDToDecimal(fmt.Sprint(i))
		privID := i + 1
		expiry := time.Now().Add(time.Hour + time.Duration(i)).UTC().Unix()

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
						"tokenId": "%s",
						"version": 1,
						"privId": %d,
						"user": "%s",
						"expires": %d
					}
				},
				"type": "zone.dimo.contract.event"
			}`, contractAddr.String(), eventName, tokenID, privID, userAddr.String(), expiry)

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

	return cEventsTestHelper{
		logger:    zerolog.New(os.Stdout).With().Timestamp().Logger(),
		pdb:       pdb,
		container: container,
		ctx:       ctx,
		t:         t,
		assert:    assert,
	}
}

func (s cEventsTestHelper) destroy() {
	if err := s.container.Terminate(s.ctx); err != nil {
		s.t.Fatal(err)
	}
}
