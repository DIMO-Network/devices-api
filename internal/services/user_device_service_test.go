package services

import (
	"context"
	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"os"
	"testing"
)

func Test_userDeviceService_CreateUserDevice(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := context.Background()
	deviceDefSvc := mock_services.NewMockDeviceDefinitionService(mockCtrl)
	eventSvc := mock_services.NewMockEventService(mockCtrl)

	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "devices-api").
		Logger()

	ddID := ksuid.New().String()
	styleID := ksuid.New().String()
	userID := ksuid.New().String()
	vin := "VINNNY1231231"
	can := "7"
	apInt := test.BuildIntegrationGRPC("autopi", 10, 12)
	dd := test.BuildDeviceDefinitionGRPC(ddID, "Ford", "Escaped", 2023, apInt)
	deviceDefSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), ddID).Times(1).Return(dd, nil)
	// style will have hybrid name in it and powertrain attr HEV
	deviceDefSvc.EXPECT().GetDeviceStyleByID(gomock.Any(), styleID).Times(1).Return(&ddgrpc.DeviceStyle{
		Id:                 ksuid.New().String(),
		Name:               "Super Hybrid",
		DeviceDefinitionId: ddID,
		DeviceAttributes: []*ddgrpc.DeviceTypeAttribute{
			{
				Name:  "powertrain_type",
				Value: "HEV",
			},
		},
	}, nil)

	userDeviceSvc := NewUserDeviceService(deviceDefSvc, logger, pdb.DBS, eventSvc)

	userDevice, _, err := userDeviceSvc.CreateUserDevice(ctx, ddID, styleID, "USA", userID, &vin, &can)
	require.NoError(t, err)
	assert.Equal(t, vin, userDevice.VinIdentifier.String)
	assert.Equal(t, userID, userDevice.UserID)
	assert.Equal(t, ddID, userDevice.DeviceDefinitionID)
	assert.Equal(t, styleID, userDevice.DeviceStyleID.String)
	assert.Equal(t, "USA", userDevice.CountryCode.String)
	assert.Equal(t, can, gjson.GetBytes(userDevice.Metadata.JSON, "canProtocol"))
	assert.Equal(t, "HEV", gjson.GetBytes(userDevice.Metadata.JSON, "powertrainType"))
}
