package rpc

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	pb "github.com/DIMO-Network/devices-api/pkg/grpc"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewTeslaRPCService(
	dbs func() *db.ReaderWriter,
	settings *config.Settings,
	cipher shared.Cipher,
	teslaAPI services.TeslaFleetAPIService,
	logger *zerolog.Logger,
) pb.TeslaServiceServer {
	return &teslaRPCServer{
		dbs:      dbs,
		logger:   logger,
		settings: settings,
		cipher:   cipher,
		teslaAPI: teslaAPI,
	}
}

// userDeviceRPCServer is the grpc server implementation for the proto services
type teslaRPCServer struct {
	pb.UnimplementedTeslaServiceServer
	dbs      func() *db.ReaderWriter
	logger   *zerolog.Logger
	settings *config.Settings
	cipher   shared.Cipher
	teslaAPI services.TeslaFleetAPIService
}

func (s *teslaRPCServer) CheckFleetTelemetryCapable(ctx context.Context, req *pb.CheckFleetTelemetryCapableRequest) (*pb.CheckFleetTelemetryCapableResponse, error) {
	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(req.UserDeviceId),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ("26A5Dk3vvvQutjSyF0Jka2DP5lg"),
		qm.Load(qm.Rels(models.UserDeviceAPIIntegrationRels.UserDevice, models.UserDeviceRels.VehicleTokenSyntheticDevice)),
	).One(ctx, s.dbs().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "No Tesla integration with that userDeviceId found.")
		}
		return nil, err
	}

	if !udai.ExternalID.Valid || !udai.AccessToken.Valid || !udai.AccessExpiresAt.Valid || udai.AccessExpiresAt.Time.Before(time.Now()) {
		return nil, status.Error(codes.FailedPrecondition, "Credentials invalid.")
	}

	if !udai.R.UserDevice.VinIdentifier.Valid || !udai.R.UserDevice.VinConfirmed {
		return nil, status.Error(codes.FailedPrecondition, "Credentials invalid.")
	}

	sd := udai.R.UserDevice

	if sd == nil || sd.TokenID.IsZero() {
		return nil, status.Error(codes.FailedPrecondition, "Synthetic device not minted.")
	}

	var meta services.UserDeviceAPIIntegrationsMetadata
	err = udai.Metadata.Unmarshal(&meta)
	if err != nil {
		return nil, err
	}

	apiVersion := 1
	if meta.TeslaAPIVersion != 0 {
		apiVersion = meta.TeslaAPIVersion
	}

	if apiVersion != 2 {
		return nil, status.Error(codes.FailedPrecondition, "Integration is not v2.")
	}

	accessToken, err := s.cipher.Decrypt(udai.AccessToken.String)
	if err != nil {
		return nil, err
	}

	teslaID, err := strconv.Atoi(udai.ExternalID.String)
	if err != nil {
		return nil, err
	}

	fleetStatus, err := s.teslaAPI.GetTelemetrySubscriptionStatus(ctx, accessToken, teslaID)
	if err != nil {
		return nil, err
	}
	if fleetStatus.Configured {
		return &pb.CheckFleetTelemetryCapableResponse{TelemetryCapable: true}, nil
	}

	vid, _ := udai.R.UserDevice.TokenID.Int64()

	err = s.teslaAPI.SubscribeForTelemetryData(ctx, accessToken, udai.R.UserDevice.VinIdentifier.String)
	if err != nil {
		s.logger.Err(err).Int64("vehicleId", vid).Int64("integrationId", 2).Msg("Failed to configure Fleet Telemetry.")
		var subErr *services.TeslaSubscriptionError
		if errors.As(err, &subErr) {
			switch subErr.Type {
			case services.KeyUnpaired, services.UnsupportedFirmware:
				return &pb.CheckFleetTelemetryCapableResponse{TelemetryCapable: true}, nil
			case services.UnsupportedVehicle:
				return &pb.CheckFleetTelemetryCapableResponse{TelemetryCapable: false}, nil
			}
		}
		return nil, err
	}

	s.logger.Info().Int64("vehicleId", vid).Int64("integrationId", 2).Msg("Successfully configured Fleet Telemetry.")

	return &pb.CheckFleetTelemetryCapableResponse{TelemetryCapable: true}, nil
}
