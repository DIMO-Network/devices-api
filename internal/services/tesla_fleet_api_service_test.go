package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/suite"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/test"
)

const mockTeslaFleetBaeURL = "https://fleet-mock-api.%s.tesla.com"

type TeslaFleetAPIServiceTestSuite struct {
	suite.Suite
	ctx      context.Context
	SUT      TeslaFleetAPIService
	settings *config.Settings
}

func (t *TeslaFleetAPIServiceTestSuite) SetupSuite() {
	t.ctx = context.Background()
	logger := test.Logger()
	t.settings = &config.Settings{TeslaFleetURL: mockTeslaFleetBaeURL, TeslaTelemetryCACertificate: "Ca-Cert", TeslaTelemetryPort: 443, TeslaTelemetryHostName: "tel.dimo.com"}

	t.SUT = NewTeslaFleetAPIService(t.settings, logger)
}

func TestTeslaFleetAPIServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TeslaFleetAPIServiceTestSuite))
}

func (t *TeslaFleetAPIServiceTestSuite) TestSubscribeForTelemetryData() {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	token := "someToken"
	region := "mockRegion"
	vin := "RandomVin"

	baseURL := fmt.Sprintf(mockTeslaFleetBaeURL, region)
	u := fmt.Sprintf("%s/api/1/vehicles/fleet_telemetry_config", baseURL)

	exp := time.Now().AddDate(0, 0, 364).Unix()
	expected := SubscribeForTelemetryDataRequest{
		Vins: []string{vin},
		Config: TelemetryConfigRequest{
			HostName:            t.settings.TeslaTelemetryHostName,
			PublicCACertificate: t.settings.TeslaTelemetryCACertificate,
			Expiration:          exp,
			Port:                t.settings.TeslaTelemetryPort,
			Fields:              fields,
			AlertTypes:          []string{"service"},
		},
	}

	httpmock.RegisterResponder(http.MethodPost, u, func(req *http.Request) (*http.Response, error) {
		r := SubscribeForTelemetryDataRequest{}
		if err := json.NewDecoder(req.Body).Decode(&r); err != nil {
			return httpmock.NewStringResponse(400, ""), nil
		}

		t.Require().Equal(expected, r)

		resp, err := httpmock.NewJsonResponse(200, "")
		if err != nil {
			return httpmock.NewStringResponse(500, ""), nil
		}
		return resp, nil
	})

	err := t.SUT.SubscribeForTelemetryData(t.ctx, token, region, vin)

	t.Require().NoError(err)
}
