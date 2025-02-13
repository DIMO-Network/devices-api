package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
)

type enableTelemetryCmd struct {
	logger   zerolog.Logger
	settings config.Settings
	pdb      db.Store
	cipher   shared.Cipher
}

func (*enableTelemetryCmd) Name() string { return "enable-telemetry" }
func (*enableTelemetryCmd) Synopsis() string {
	return "xdd"
}
func (*enableTelemetryCmd) Usage() string {
	return `xpp
  `
}

// nolint
func (p *enableTelemetryCmd) SetFlags(f *flag.FlagSet) {

}

func (p *enableTelemetryCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	err := enableTelemetry(&p.settings, p.pdb, &p.logger, p.cipher)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Error running Smartcar Kafka re-registration")
	}
	return subcommands.ExitSuccess
}

const teslaBasicFields = `{"ACChargingEnergyIn":{"interval_seconds":60},"ACChargingPower":{"interval_seconds":60},"BatteryLevel":{"interval_seconds":300},"ChargeLimitSoc":{"interval_seconds":3600},"ChargeState":{"interval_seconds":60},"DCChargingEnergyIn":{"interval_seconds":60},"DCChargingPower":{"interval_seconds":60},"EnergyRemaining":{"interval_seconds":60},"EstBatteryRange":{"interval_seconds":300},"Location":{"interval_seconds":20},"Odometer":{"interval_seconds":300},"OutsideTemp":{"interval_seconds":60},"Soc":{"interval_seconds":60},"TpmsPressureFl":{"interval_seconds":300},"TpmsPressureFr":{"interval_seconds":300},"TpmsPressureRl":{"interval_seconds":300},"TpmsPressureRr":{"interval_seconds":300},"VehicleSpeed":{"interval_seconds":20}}`
const teslaAdvancedFields = `{"ACChargingEnergyIn": {"interval_seconds": 60}, "ACChargingPower": {"interval_seconds": 60}, "AutomaticEmergencyBrakingOff": {"interval_seconds": 1}, "BatteryLevel": {"interval_seconds": 300}, "BlindSpotCollisionWarningChime": {"interval_seconds": 1}, "BrickVoltageMax": {"interval_seconds": 300}, "BrickVoltageMin": {"interval_seconds": 300}, "CarType": {"interval_seconds": 21600}, "ChargeAmps": {"interval_seconds": 60}, "ChargeLimitSoc": {"interval_seconds": 3600}, "ChargeState": {"interval_seconds": 60}, "ChargerVoltage": {"interval_seconds": 300}, "ChargingCableType": {"interval_seconds": 300}, "CruiseFollowDistance": {"interval_seconds": 60}, "CruiseSetSpeed": {"interval_seconds": 60}, "CurrentLimitMph": {"interval_seconds": 1}, "DCChargingEnergyIn": {"interval_seconds": 60}, "DCChargingPower": {"interval_seconds": 60}, "DetailedChargeState": {"interval_seconds": 300}, "DoorState": {"interval_seconds": 1}, "EmergencyLaneDepartureAvoidance": {"interval_seconds": 1}, "EnergyRemaining": {"interval_seconds": 60}, "EstBatteryRange": {"interval_seconds": 60}, "FastChargerPresent": {"interval_seconds": 300}, "FdWindow": {"interval_seconds": 1}, "ForwardCollisionWarning": {"interval_seconds": 1}, "FpWindow": {"interval_seconds": 1}, "GuestModeEnabled": {"interval_seconds": 3600}, "IdealBatteryRange": {"interval_seconds": 20}, "LaneDepartureAvoidance": {"interval_seconds": 1}, "Location": {"interval_seconds": 1}, "Locked": {"interval_seconds": 300}, "Odometer": {"interval_seconds": 300}, "OutsideTemp": {"interval_seconds": 60}, "Soc": {"interval_seconds": 60}, "SoftwareUpdateVersion": {"interval_seconds": 21600}, "SpeedLimitWarning": {"interval_seconds": 1}, "TpmsLastSeenPressureTimeFl": {"interval_seconds": 300}, "TpmsLastSeenPressureTimeFr": {"interval_seconds": 300}, "TpmsLastSeenPressureTimeRl": {"interval_seconds": 300}, "TpmsLastSeenPressureTimeRr": {"interval_seconds": 300}, "TpmsPressureFl": {"interval_seconds": 300}, "TpmsPressureFr": {"interval_seconds": 300}, "TpmsPressureRl": {"interval_seconds": 300}, "TpmsPressureRr": {"interval_seconds": 300}, "Trim": {"interval_seconds": 21600}, "VehicleName": {"interval_seconds": 21600}, "VehicleSpeed": {"interval_seconds": 20}, "Version": {"interval_seconds": 21600}}`
const teslaAllFields = `{"ACChargingEnergyIn":{"interval_seconds":1},"ACChargingPower":{"interval_seconds":1},"AutoSeatClimateLeft":{"interval_seconds":1},"AutoSeatClimateRight":{"interval_seconds":1},"AutomaticBlindSpotCamera":{"interval_seconds":1},"AutomaticEmergencyBrakingOff":{"interval_seconds":1},"BMSState":{"interval_seconds":1},"BatteryHeaterOn":{"interval_seconds":1},"BatteryLevel":{"interval_seconds":1},"BlindSpotCollisionWarningChime":{"interval_seconds":1},"BmsFullchargecomplete":{"interval_seconds":1},"BrakePedal":{"interval_seconds":1},"BrakePedalPos":{"interval_seconds":1},"BrickVoltageMax":{"interval_seconds":1},"BrickVoltageMin":{"interval_seconds":1},"CabinOverheatProtectionMode":{"interval_seconds":1},"CabinOverheatProtectionTemperatureLimit":{"interval_seconds":1},"CarType":{"interval_seconds":1},"CenterDisplay":{"interval_seconds":1},"ChargeAmps":{"interval_seconds":1},"ChargeCurrentRequest":{"interval_seconds":1},"ChargeCurrentRequestMax":{"interval_seconds":1},"ChargeEnableRequest":{"interval_seconds":1},"ChargeLimitSoc":{"interval_seconds":1},"ChargePort":{"interval_seconds":1},"ChargePortColdWeatherMode":{"interval_seconds":1},"ChargePortDoorOpen":{"interval_seconds":1},"ChargePortLatch":{"interval_seconds":1},"ChargeState":{"interval_seconds":1},"ChargerPhases":{"interval_seconds":1},"ChargingCableType":{"interval_seconds":1},"ClimateKeeperMode":{"interval_seconds":1},"CruiseFollowDistance":{"interval_seconds":1},"CruiseSetSpeed":{"interval_seconds":1},"CurrentLimitMph":{"interval_seconds":1},"DCChargingEnergyIn":{"interval_seconds":1},"DCChargingPower":{"interval_seconds":1},"DCDCEnable":{"interval_seconds":1},"DefrostForPreconditioning":{"interval_seconds":1},"DefrostMode":{"interval_seconds":1},"DestinationLocation":{"interval_seconds":1},"DestinationName":{"interval_seconds":1},"DetailedChargeState":{"interval_seconds":1},"DiAxleSpeedF":{"interval_seconds":1},"DiAxleSpeedR":{"interval_seconds":1},"DiAxleSpeedREL":{"interval_seconds":1},"DiAxleSpeedRER":{"interval_seconds":1},"DiHeatsinkTF":{"interval_seconds":1},"DiHeatsinkTR":{"interval_seconds":1},"DiHeatsinkTREL":{"interval_seconds":1},"DiHeatsinkTRER":{"interval_seconds":1},"DiInverterTF":{"interval_seconds":1},"DiInverterTR":{"interval_seconds":1},"DiInverterTREL":{"interval_seconds":1},"DiInverterTRER":{"interval_seconds":1},"DiMotorCurrentF":{"interval_seconds":1},"DiMotorCurrentR":{"interval_seconds":1},"DiMotorCurrentREL":{"interval_seconds":1},"DiMotorCurrentRER":{"interval_seconds":1},"DiSlaveTorqueCmd":{"interval_seconds":1},"DiStateF":{"interval_seconds":1},"DiStateR":{"interval_seconds":1},"DiStateREL":{"interval_seconds":1},"DiStateRER":{"interval_seconds":1},"DiStatorTempF":{"interval_seconds":1},"DiStatorTempR":{"interval_seconds":1},"DiStatorTempREL":{"interval_seconds":1},"DiStatorTempRER":{"interval_seconds":1},"DiTorqueActualF":{"interval_seconds":1},"DiTorqueActualR":{"interval_seconds":1},"DiTorqueActualREL":{"interval_seconds":1},"DiTorqueActualRER":{"interval_seconds":1},"DiTorquemotor":{"interval_seconds":1},"DiVBatF":{"interval_seconds":1},"DiVBatR":{"interval_seconds":1},"DiVBatREL":{"interval_seconds":1},"DiVBatRER":{"interval_seconds":1},"DoorState":{"interval_seconds":1},"DriveRail":{"interval_seconds":1},"DriverSeatBelt":{"interval_seconds":1},"DriverSeatOccupied":{"interval_seconds":1},"EfficiencyPackage":{"interval_seconds":1},"EmergencyLaneDepartureAvoidance":{"interval_seconds":1},"EnergyRemaining":{"interval_seconds":1},"EstBatteryRange":{"interval_seconds":1},"EstimatedHoursToChargeTermination":{"interval_seconds":1},"EuropeVehicle":{"interval_seconds":1},"ExpectedEnergyPercentAtTripArrival":{"interval_seconds":1},"ExteriorColor":{"interval_seconds":1},"FastChargerPresent":{"interval_seconds":1},"FastChargerType":{"interval_seconds":1},"FdWindow":{"interval_seconds":1},"ForwardCollisionWarning":{"interval_seconds":1},"FpWindow":{"interval_seconds":1},"Gear":{"interval_seconds":1},"GpsHeading":{"interval_seconds":1},"GpsState":{"interval_seconds":1},"GuestModeEnabled":{"interval_seconds":1},"GuestModeMobileAccessState":{"interval_seconds":1},"HomelinkDeviceCount":{"interval_seconds":1},"HomelinkNearby":{"interval_seconds":1},"HvacACEnabled":{"interval_seconds":1},"HvacAutoMode":{"interval_seconds":1},"HvacFanSpeed":{"interval_seconds":1},"HvacFanStatus":{"interval_seconds":1},"HvacLeftTemperatureRequest":{"interval_seconds":1},"HvacPower":{"interval_seconds":1},"HvacRightTemperatureRequest":{"interval_seconds":1},"HvacSteeringWheelHeatAuto":{"interval_seconds":1},"HvacSteeringWheelHeatLevel":{"interval_seconds":1},"Hvil":{"interval_seconds":1},"IdealBatteryRange":{"interval_seconds":1},"InsideTemp":{"interval_seconds":1},"IsolationResistance":{"interval_seconds":1},"LaneDepartureAvoidance":{"interval_seconds":1},"LateralAcceleration":{"interval_seconds":1},"LifetimeEnergyGainedRegen":{"interval_seconds":1},"LifetimeEnergyUsed":{"interval_seconds":1},"LifetimeEnergyUsedDrive":{"interval_seconds":1},"Location":{"interval_seconds":1},"Locked":{"interval_seconds":1},"LongitudinalAcceleration":{"interval_seconds":1},"MilesToArrival":{"interval_seconds":1},"MinutesToArrival":{"interval_seconds":1},"ModuleTempMax":{"interval_seconds":1},"ModuleTempMin":{"interval_seconds":1},"NotEnoughPowerToHeat":{"interval_seconds":1},"NumBrickVoltageMax":{"interval_seconds":1},"NumBrickVoltageMin":{"interval_seconds":1},"NumModuleTempMax":{"interval_seconds":1},"NumModuleTempMin":{"interval_seconds":1},"Odometer":{"interval_seconds":1},"OffroadLightbarPresent":{"interval_seconds":1},"OriginLocation":{"interval_seconds":1},"OutsideTemp":{"interval_seconds":1},"PackCurrent":{"interval_seconds":1},"PackVoltage":{"interval_seconds":1},"PairedPhoneKeyAndKeyFobQty":{"interval_seconds":1},"PassengerSeatBelt":{"interval_seconds":1},"PedalPosition":{"interval_seconds":1},"PinToDriveEnabled":{"interval_seconds":1},"PowershareHoursLeft":{"interval_seconds":1},"PowershareInstantaneousPowerKW":{"interval_seconds":1},"PowershareStatus":{"interval_seconds":1},"PowershareStopReason":{"interval_seconds":1},"PowershareType":{"interval_seconds":1},"PreconditioningEnabled":{"interval_seconds":1},"RatedRange":{"interval_seconds":1},"RdWindow":{"interval_seconds":1},"RearDisplayHvacEnabled":{"interval_seconds":1},"RearSeatHeaters":{"interval_seconds":1},"RemoteStartEnabled":{"interval_seconds":1},"RightHandDrive":{"interval_seconds":1},"RoofColor":{"interval_seconds":1},"RouteLastUpdated":{"interval_seconds":1},"RouteLine":{"interval_seconds":1},"RouteTrafficMinutesDelay":{"interval_seconds":1},"RpWindow":{"interval_seconds":1},"ScheduledChargingMode":{"interval_seconds":1},"ScheduledChargingPending":{"interval_seconds":1},"ScheduledChargingStartTime":{"interval_seconds":1},"ScheduledDepartureTime":{"interval_seconds":1},"SeatHeaterLeft":{"interval_seconds":1},"SeatHeaterRearCenter":{"interval_seconds":1},"SeatHeaterRearLeft":{"interval_seconds":1},"SeatHeaterRearRight":{"interval_seconds":1},"SeatHeaterRight":{"interval_seconds":1},"SentryMode":{"interval_seconds":1},"ServiceMode":{"interval_seconds":1},"Soc":{"interval_seconds":1},"SoftwareUpdateDownloadPercentComplete":{"interval_seconds":1},"SoftwareUpdateExpectedDurationMinutes":{"interval_seconds":1},"SoftwareUpdateInstallationPercentComplete":{"interval_seconds":1},"SoftwareUpdateScheduledStartTime":{"interval_seconds":1},"SoftwareUpdateVersion":{"interval_seconds":1},"SpeedLimitMode":{"interval_seconds":1},"SpeedLimitWarning":{"interval_seconds":1},"SuperchargerSessionTripPlanner":{"interval_seconds":1},"TimeToFullCharge":{"interval_seconds":1},"TonneauOpenPercent":{"interval_seconds":1},"TonneauPosition":{"interval_seconds":1},"TonneauTentMode":{"interval_seconds":1},"TpmsHardWarnings":{"interval_seconds":1},"TpmsLastSeenPressureTimeFl":{"interval_seconds":1},"TpmsLastSeenPressureTimeFr":{"interval_seconds":1},"TpmsLastSeenPressureTimeRl":{"interval_seconds":1},"TpmsLastSeenPressureTimeRr":{"interval_seconds":1},"TpmsPressureFl":{"interval_seconds":1},"TpmsPressureFr":{"interval_seconds":1},"TpmsPressureRl":{"interval_seconds":1},"TpmsPressureRr":{"interval_seconds":1},"TpmsSoftWarnings":{"interval_seconds":1},"Trim":{"interval_seconds":1},"ValetModeEnabled":{"interval_seconds":1},"VehicleName":{"interval_seconds":1},"VehicleSpeed":{"interval_seconds":1},"Version":{"interval_seconds":1},"WheelType":{"interval_seconds":1},"WiperHeatEnabled":{"interval_seconds":1}}`

type teslaFleetConfigReq struct {
	VINs   []string               `json:"vins"`
	Config teslaFleetConfigConfig `json:"config"`
}

type teslaFleetConfigConfig struct {
	Hostname string          `json:"hostname"`
	CA       string          `json:"ca"`
	Port     int             `json:"port"`
	Fields   json.RawMessage `json:"fields"`
}

type fleetConfigResp struct {
	Response struct {
		UpdatedVehicles int `json:"updated_vehicles"`
		SkippedVehicles struct {
			MissingKey          []string `json:"missing_key"`
			UnsupportedHardware []string `json:"unsupported_hardware"`
			UnsupportedFirmware []string `json:"unsupported_firmware"`
		} `json:"skipped_vehicles"`
	} `json:"response"`
}

type deleteConfigResp struct {
	Response struct {
		UpdatedVehicles int `json:"updated_vehicles"`
	} `json:"response"`
}

func enableTelemetry(settings *config.Settings, pdb db.Store, logger *zerolog.Logger, cipher shared.Cipher) error {
	configName := os.Args[2]
	userDeviceID := os.Args[3]

	var fieldsPayload []byte
	switch configName {
	case "basic":
		fieldsPayload = []byte(teslaBasicFields)
	case "advanced":
		fieldsPayload = []byte(teslaAdvancedFields)
	case "max":
		fieldsPayload = []byte(teslaAllFields)
	case "delete":
		// Handle this elsewhere.
	default:
		return fmt.Errorf("unrecognized type %q", os.Args[3])
	}

	ctx := context.Background()

	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ("26A5Dk3vvvQutjSyF0Jka2DP5lg"),
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice),
	).One(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}

	var md services.UserDeviceAPIIntegrationsMetadata
	err = udai.Metadata.Unmarshal(&md)
	if err != nil {
		logger.Warn().Str("userDeviceId", udai.UserDeviceID).Msg("Couldn't parse metadata, skipping.")
		return err
	}

	if md.TeslaAPIVersion != 2 {
		return fmt.Errorf("tesla version not %d", md.TeslaAPIVersion)
	}

	token, err := cipher.Decrypt(udai.AccessToken.String)
	if err != nil {
		return err
	}

	baseURL, err := url.ParseRequestURI(settings.TeslaFleetURL)
	if err != nil {
		return err
	}

	if configName == "delete" {
		ur := baseURL.JoinPath("api/1/vehicles", udai.R.UserDevice.VinIdentifier.String, "fleet_telemetry_config")

		req, err := http.NewRequest("DELETE", ur.String(), nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+token)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("status code %d", res.StatusCode)
		}

		respBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		var resBody deleteConfigResp

		if err := json.Unmarshal(respBytes, &resBody); err != nil {
			return fmt.Errorf("failed to unmarshal response body: %w", err)
		}

		if resBody.Response.UpdatedVehicles == 1 {
			logger.Info().Str("userDeviceId", udai.UserDeviceID).Msg("Successfully removed config.")
		} else {
			logger.Info().Str("userDeviceId", udai.UserDeviceID).Msg("Failed to remove config.")
		}

		return nil
	}

	ur := baseURL.JoinPath("api/1/vehicles/fleet_telemetry_config")

	tfcr := teslaFleetConfigReq{
		VINs: []string{udai.R.UserDevice.VinIdentifier.String},
		Config: teslaFleetConfigConfig{
			Hostname: settings.TeslaTelemetryHostName,
			CA:       settings.TeslaTelemetryCACertificate,
			Port:     settings.TeslaTelemetryPort,
			Fields:   fieldsPayload,
		},
	}

	reqBytes, err := json.Marshal(tfcr)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", ur.String(), bytes.NewBuffer(reqBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("status code %d", res.StatusCode)
	}

	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var resBody fleetConfigResp

	if err := json.Unmarshal(respBytes, &resBody); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	var xs string
	switch {
	case resBody.Response.UpdatedVehicles != 0:
		xs = fmt.Sprintf("Set config %s successfully.", configName)
	case len(resBody.Response.SkippedVehicles.MissingKey) != 0:
		xs = "Virtual key missing."
	case len(resBody.Response.SkippedVehicles.UnsupportedHardware) != 0:
		xs = "Hardware not supported, VIN is " + udai.R.UserDevice.VinIdentifier.String + ", definition is " + udai.R.UserDevice.DefinitionID
	case len(resBody.Response.SkippedVehicles.UnsupportedFirmware) != 0:
		xs = "Firmware not supported, VIN is " + udai.R.UserDevice.VinIdentifier.String + ", definition is " + udai.R.UserDevice.DefinitionID
	default:
		xs = "Something weird happened"
	}

	logger.Info().Str("userDeviceId", udai.UserDeviceID).Msg(xs)

	return nil
}
