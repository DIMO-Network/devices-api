package services

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/suite"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/test"
)

const mockTeslaFleetBaseURL = "https://fleet-mock-api.dimo.zone"

type TeslaFleetAPIServiceTestSuite struct {
	suite.Suite
	ctx      context.Context
	SUT      TeslaFleetAPIService
	settings *config.Settings
}

func (t *TeslaFleetAPIServiceTestSuite) SetupSuite() {
	t.ctx = context.Background()
	logger := test.Logger()
	t.settings = &config.Settings{TeslaFleetURL: mockTeslaFleetBaseURL, TeslaTelemetryCACertificate: "Ca-Cert", TeslaTelemetryPort: 443, TeslaTelemetryHostName: "tel.dimo.com"}

	var err error
	t.SUT, err = NewTeslaFleetAPIService(t.settings, logger)
	t.Require().NoError(err)
}

func TestTeslaFleetAPIServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TeslaFleetAPIServiceTestSuite))
}

func (t *TeslaFleetAPIServiceTestSuite) TestSubscribeForTelemetryData() {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	token := "someToken"
	vin := "RandomVin"

	baseURL := mockTeslaFleetBaseURL
	u := fmt.Sprintf("%s/api/1/vehicles/fleet_telemetry_config", baseURL)

	respBody := TeslaResponseWrapper[SubscribeForTelemetryDataResponse]{
		SubscribeForTelemetryDataResponse{
			UpdatedVehicles: 1,
			SkippedVehicles: SkippedVehicles{},
		},
	}

	jsonResp, err := httpmock.NewJsonResponder(http.StatusOK, respBody)
	t.Require().NoError(err)
	httpmock.RegisterResponder(http.MethodPost, u, jsonResp)

	err = t.SUT.SubscribeForTelemetryData(t.ctx, token, vin)

	t.Require().NoError(err)
}

func (t *TeslaFleetAPIServiceTestSuite) TestSubscribeForTelemetryData_Errror_Cases() {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	vin := "RandomVin"
	tests := []struct {
		response      TeslaResponseWrapper[SubscribeForTelemetryDataResponse]
		expectedError string
	}{
		{
			response: TeslaResponseWrapper[SubscribeForTelemetryDataResponse]{
				SubscribeForTelemetryDataResponse{
					UpdatedVehicles: 0,
					SkippedVehicles: SkippedVehicles{
						MissingKey:          []string{vin},
						UnsupportedHardware: nil,
						UnsupportedFirmware: nil,
					},
				},
			},
			expectedError: "virtual key not added to vehicle",
		},
		{
			response: TeslaResponseWrapper[SubscribeForTelemetryDataResponse]{
				SubscribeForTelemetryDataResponse{
					UpdatedVehicles: 0,
					SkippedVehicles: SkippedVehicles{
						MissingKey:          nil,
						UnsupportedHardware: []string{vin},
						UnsupportedFirmware: nil,
					},
				},
			},
			expectedError: "vehicle hardware not supported",
		},
		{
			response: TeslaResponseWrapper[SubscribeForTelemetryDataResponse]{
				SubscribeForTelemetryDataResponse{
					UpdatedVehicles: 0,
					SkippedVehicles: SkippedVehicles{
						MissingKey:          nil,
						UnsupportedHardware: nil,
						UnsupportedFirmware: []string{vin},
					},
				},
			},
			expectedError: "vehicle firmware not supported",
		},
	}

	for _, tst := range tests {
		token := "someToken"

		baseURL := mockTeslaFleetBaseURL
		u := fmt.Sprintf("%s/api/1/vehicles/fleet_telemetry_config", baseURL)

		responder, err := httpmock.NewJsonResponder(http.StatusOK, tst.response)
		t.Require().NoError(err)
		httpmock.RegisterResponder(http.MethodPost, u, responder)

		err = t.SUT.SubscribeForTelemetryData(t.ctx, token, vin)

		t.EqualError(err, tst.expectedError)
	}
}
