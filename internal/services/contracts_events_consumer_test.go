package services

import (
	"context"

	"errors"
	"fmt"
	"os"
	"strconv"
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

const contractEventPayload = "{\"data\":{\"contract\":\"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266\",\"transactionHash\":\"0x29d1aa4f5eb409bf7d334a7f50fcba50264fbefe00c991cc278f444eb64fdfe5\",\"blockCompleted\":true,\"eventSignature\":\"0x61a24679288162b799d80b2bb2b8b0fcdd5c5f53ac19e9246cc190b60196c359\",\"eventName\":\"PrivilegeSet\",\"arguments\":{\"0\":\"1\",\"1\":\"1\",\"2\":\"2\",\"3\":\"0x70997970C51812dc3A010C7d01b50e0d17dc79C8\",\"4\":\"1668722877\",\"tokenId\":\"1\",\"version\":\"1\",\"privId\":\"2\",\"user\":\"0x70997970C51812dc3A010C7d01b50e0d17dc79C8\",\"expires\":\"1676451061\"}},\"type\":\"zone.dimo.contract.event\"}"

// contract_address, token_id, privilege, user_address
type mockTestEntity struct {
	Contract    []byte
	TokenID     types.Decimal
	PrivilegeID int64
	UserAddress []byte
	ExpiresAt   time.Time
}

func TestProcessContractsEventsMessages(t *testing.T) {
	s := initCEventsTestHelper(t)
	defer s.destroy()

	msg := &message.Message{
		Payload: []byte(contractEventPayload),
	}
	c := NewContractsEventsConsumer(s.pdb, &s.logger)

	err := c.processMessage(msg)
	assert.NoError(t, err)

	mock, err := createMockEntities("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", "0x70997970C51812dc3A010C7d01b50e0d17dc79C8", "1", "1676451061", 2)
	assert.NoError(t, err)

	nft, err := models.FindNFTPrivilege(s.ctx, s.pdb.DBS().Reader, mock.Contract, mock.TokenID, mock.PrivilegeID, mock.UserAddress)
	assert.NoError(t, err)

	assert.NotNil(t, nft)

	actual := mockTestEntity{
		Contract:    nft.ContractAddress,
		TokenID:     nft.TokenID,
		PrivilegeID: nft.Privilege,
		UserAddress: nft.UserAddress,
		ExpiresAt:   nft.Expiry,
	}

	assert.Equal(t, mock, actual)
}

type cEventsTestHelper struct {
	logger    zerolog.Logger
	pdb       db.Store
	container testcontainers.Container
	ctx       context.Context
	t         *testing.T
}

func initCEventsTestHelper(t *testing.T) cEventsTestHelper {
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	return cEventsTestHelper{
		logger:    zerolog.New(os.Stdout).With().Timestamp().Logger(),
		pdb:       pdb,
		container: container,
		ctx:       ctx,
		t:         t,
	}
}

func (s cEventsTestHelper) destroy() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.t)
	if err := s.container.Terminate(s.ctx); err != nil {
		s.t.Fatal(err)
	}
}

func createMockEntities(contract, userAddress, tokenID, expiresAt string, privilegeID int64) (mockTestEntity, error) {
	ti, ok := new(decimal.Big).SetString(tokenID)
	if !ok {
		return mockTestEntity{}, fmt.Errorf("couldn't parse token id %q", tokenID)
	}

	tid := types.NewDecimal(ti)

	t, err := strconv.ParseInt(expiresAt, 10, 64)
	if err != nil {
		return mockTestEntity{}, errors.New("could not parse timestamp")
	}
	tm := time.Unix(t, 0)

	return mockTestEntity{
		Contract:    common.FromHex(contract),
		UserAddress: common.FromHex(userAddress),
		PrivilegeID: privilegeID,
		TokenID:     tid,
		ExpiresAt:   tm.UTC(),
	}, nil
}
