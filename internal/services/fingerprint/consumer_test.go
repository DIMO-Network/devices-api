package fingerprint

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/shared/db"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
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
