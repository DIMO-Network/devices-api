package fingerprint

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/mock/gomock"
)

type ConsumerTestSuite struct {
	suite.Suite
	pdb       db.Store
	container testcontainers.Container
	ctx       context.Context
	mockCtrl  *gomock.Controller
	topic     string
	cons      *Consumer
}

const migrationsDirRelPath = "../../../migrations"

// SetupSuite starts container db
func (s *ConsumerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(context.Background(), s.T(), migrationsDirRelPath)
	s.mockCtrl = gomock.NewController(s.T())
	s.topic = "topic.fingerprint"

	gitSha1 := os.Getenv("GIT_SHA1")
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "devices-api").
		Str("git-sha1", gitSha1).
		Logger()

	s.cons = &Consumer{
		logger: &logger,
		DBS:    s.pdb,
	}
}

// TearDownSuite cleanup at end by terminating container
func (s *ConsumerTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
	s.mockCtrl.Finish()
}

func TestConsumerTestSuite(t *testing.T) {
	t.Skip("Isolate this test from the network before putting it in CI.")
	suite.Run(t, new(ConsumerTestSuite))
}

func (s *ConsumerTestSuite) TestSignatureValidation() {

	cases := []struct {
		Data string
	}{
		{
			Data: `{
				"data": {"rpiUptimeSecs":39,"batteryVoltage":13.49,"timestamp":1688136702634,"vin":"W1N2539531F907299","protocol":"7"},
				"id": "2RvhwjUbtoePjmXN7q9qfjLQgwP",
				"signature": "7c31e54ddcffc2a548ccaf10ed64b7e4bdd239bbaa3e5f6dba41d3e4051d930b7fbdf184724c2fb8d3b2ac8ac82662d2ed74e881dd01c09c4b2a9b4e62ede5db1b",
				"source": "aftermarket/device/fingerprint",
				"specversion": "1.0",
				"subject": "0x448cF8Fd88AD914e3585401241BC434FbEA94bbb",
				"type": "zone.dimo.aftermarket.device.fingerprint"
			}`,
		},
		{
			Data: `{
				"data": {"rpiUptimeSecs":36,"batteryVoltage":13.73,"timestamp":1688760445189,"vin":"LRBFXCSA5KD124854","protocol":"6"},
				"id": "2SG6Cu2NWOcu7LvhadPtmDGb65S",
				"signature": "5fb985f758c6224ab45630d055c7aca163329b88accfb8fd76a0dbb13b2ebcfe3c5bd8b801851f683f7a288c174a11ed8fc2631d95929c3b3cc85c75fb10ea001c",
				"source": "aftermarket/device/fingerprint",
				"specversion": "1.0",
				"subject": "0x06fF8E7A4A159EA388da7c133DC5F79727868d83",
				"type": "zone.dimo.aftermarket.device.fingerprint"
			}`,
		},
	}

	for _, c := range cases {
		var event Event
		err := json.Unmarshal([]byte(c.Data), &event)
		require.NoError(s.T(), err)
		data, err := json.Marshal(event.Data)
		require.NoError(s.T(), err)

		signature := common.FromHex(event.Signature)
		addr := common.HexToAddress(event.Subject)
		hash := crypto.Keccak256Hash(data)
		recAddr, err := helpers.Ecrecover(hash.Bytes(), signature)
		s.NoError(err)
		s.Equal(addr, recAddr)
	}
}

func (s *ConsumerTestSuite) TestInvalidSignature() {
	msg := `{
		"data": {"rpiUptimeSecs":36,"batteryVoltage":13.73,"timestamp":1688760445189,"vin":"LRBFXCSA5KD124854","protocol":"6"},
		"id": "2SG6Cu2NWOcu7LvhadPtmDGb65S",
		"signature": "5fb985f758c6224ab45630d055c7aca163329b88accfb8fd76a0dbb13b2ebcfe3c5bd8b801851f683f7a288c174a11ed8fc2631d95929c3b3cc85c75fb10ea001a",
		"source": "aftermarket/device/fingerprint",
		"specversion": "1.0",
		"subject": "0x06fF8E7A4A159EA388da7c133DC5F79727868d83",
		"type": "zone.dimo.aftermarket.device.fingerprint"
	}`

	var event Event
	err := json.Unmarshal([]byte(msg), &event)
	require.NoError(s.T(), err)
	data, err := json.Marshal(event.Data)
	require.NoError(s.T(), err)

	signature := common.FromHex(event.Signature)
	addr := common.HexToAddress(event.Subject)
	hash := crypto.Keccak256Hash(data)
	recAddr, err := helpers.Ecrecover(hash.Bytes(), signature)
	s.Error(err)
	s.Equal(err.Error(), "invalid signature recovery id")
	s.NotEqual(recAddr, addr)
}

func TestExtractProtocol(t *testing.T) {
	p7 := "7"
	tests := []struct {
		name    string
		data    []byte
		want    *string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "happy path",
			data: []byte(`{"protocol":"7"}`),
			want: &p7,
			wantErr: func(_ assert.TestingT, err error, _ ...interface{}) bool {
				return err == nil
			},
		},
		{
			name: "no protocol",
			data: []byte(`{"protocol":null}`),
			want: nil,
			wantErr: func(_ assert.TestingT, err error, _ ...interface{}) bool {
				return err == nil
			},
		},
		{
			name: "can't parse",
			data: []byte(`caca`),
			want: nil,
			wantErr: func(_ assert.TestingT, err error, _ ...interface{}) bool {
				return err != nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractProtocol(tt.data)
			if !tt.wantErr(t, err, fmt.Sprintf("ExtractProtocol(%v)", tt.data)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ExtractProtocol(%v)", tt.data)
		})
	}
}

func TestExtractProtocolMacaronType1(t *testing.T) {

	expectedProtocol := "06"

	tests := []struct {
		name    string
		data    string
		want    *string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "valid protocol data",
			data: "AW+yb2VVFVFCV6pmvwZXQkFXWjMyMDMwMEY4Njc1Ng==",
			want: &expectedProtocol,
			wantErr: func(_ assert.TestingT, err error, _ ...interface{}) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "invalid base64 encoding",
			data: "not base64 data",
			want: nil,
			wantErr: func(_ assert.TestingT, err error, _ ...interface{}) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "data too short",
			data: base64.StdEncoding.EncodeToString([]byte{0x01}), // Too short
			want: nil,
			wantErr: func(_ assert.TestingT, err error, _ ...interface{}) bool {
				return assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractProtocolMacaronType1(tt.data)
			if !tt.wantErr(t, err, fmt.Sprintf("ExtractProtocolMacaronType1(%v)", tt.data)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ExtractProtocolMacaronType1(%v)", tt.data)
		})
	}
}
