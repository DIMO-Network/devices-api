package services

import (
	"context"
	"encoding/json"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
)

type ValuationService struct {
	log                 *zerolog.Logger
	NATSSvc             *NATSService
	pdb                 db.Store
	deviceDefinitionSvc DeviceDefinitionService
}

func NewValuationService(settings *config.Settings, log *zerolog.Logger, pdb db.Store, deviceDefinitionSvc DeviceDefinitionService) *ValuationService {

	NATSSvc, err := NewNATSService(settings, log)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create NATS service")
	}

	return &ValuationService{
		log:                 log,
		NATSSvc:             NATSSvc,
		pdb:                 pdb,
		deviceDefinitionSvc: deviceDefinitionSvc,
	}
}

func (v *ValuationService) ValuationConsumer(ctx context.Context) error {
	sub, err := v.NATSSvc.JetStream.PullSubscribe(v.NATSSvc.JetStreamSubject, v.NATSSvc.DurableConsumer, nats.AckWait(v.NATSSvc.AckTimeout))

	if err != nil {
		return err
	}

	for {
		msgs, err := sub.Fetch(1, nats.MaxWait(v.NATSSvc.AckTimeout))
		if err != nil {
			if err == nats.ErrTimeout {
				continue
			}

			return err
		}

		for _, msg := range msgs {
			_, err := msg.Metadata()
			if err != nil {
				v.nak(msg, "")
				v.log.Info().Err(err).Msg("unable to parse metadata for message")
				continue
			}

			select {
			case <-ctx.Done():
				return nil
			default:

				var vin string

				if err := json.Unmarshal(msg.Data, &vin); err != nil {
					v.nak(msg, vin)
					v.log.Info().Err(err).Msg("unable to parse vin from message")
					continue
				}

				userDevice, err := models.UserDevices(models.UserDeviceWhere.VinIdentifier.EQ(null.StringFrom(vin))).One(ctx, v.pdb.DBS().Reader)

				if err != nil {
					v.nak(msg, vin)
					v.log.Info().Err(err).Msg("unable to find user device")
					continue
				}

				v.inProgress(msg)

				if userDevice.CountryCode.String == "USA" || userDevice.CountryCode.String == "CAN" || userDevice.CountryCode.String == "MEX" {
					status, err := v.deviceDefinitionSvc.PullDrivlyData(ctx, userDevice.ID, userDevice.DeviceDefinitionID, userDevice.VinIdentifier.String)
					if err != nil {
						v.log.Err(err).Str("vin", userDevice.VinIdentifier.String).Msg("error pulling drivly data")
					} else {
						v.log.Info().Msgf("Drivly   %s vin: %s, country: %s", status, userDevice.VinIdentifier.String, userDevice.CountryCode.String)
					}
				} else {
					status, err := v.deviceDefinitionSvc.PullVincarioValuation(ctx, userDevice.ID, userDevice.DeviceDefinitionID, userDevice.VinIdentifier.String)
					if err != nil {
						v.log.Err(err).Str("vin", userDevice.VinIdentifier.String).Msg("error pulling vincario data")
					} else {
						v.log.Info().Msgf("Vincario %s vin: %s, country: %s", status, userDevice.VinIdentifier.String, userDevice.CountryCode.String)
					}
				}

				if err := msg.Ack(); err != nil {
					v.log.Err(err).Msg("message ack failed")
				}
			}
		}
	}
}

func (v *ValuationService) inProgress(msg *nats.Msg) {
	if err := msg.InProgress(); err != nil {
		v.log.Err(err).Msg("message in progress failed")
	}
}

func (v *ValuationService) nak(msg *nats.Msg, params string) {
	err := msg.Nak()
	if len(params) == 0 {
		v.log.Err(err).Msg("message nak failed")
	} else {
		v.log.Err(err).Str("vin", params).Msg("message nak failed")
	}
}
