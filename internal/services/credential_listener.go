package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type CredentialListener struct {
	db  func() *db.ReaderWriter
	log *zerolog.Logger
}

type TeslaCredentialsCloudEventV1V2 struct {
	CloudEventHeaders
	Data TeslaCredentialsV1V2 `json:"data"`
}

type TeslaCredentialsV1V2 struct {
	OwnerAccessToken          string    `json:"ownerAccessToken"`
	OwnerAccessTokenExpiresAt time.Time `json:"ownerAccessTokenExpiresAt"`
	AuthRefreshToken          string    `json:"authRefreshToken"`
	AccessToken               string    `json:"accessToken"`
	Expiry                    time.Time `json:"expiry"`
	RefreshToken              string    `json:"refreshToken"`
}

func NewCredentialListener(db func() *db.ReaderWriter, log *zerolog.Logger) *CredentialListener {
	return &CredentialListener{db: db, log: log}
}

func (i *CredentialListener) ProcessCredentialsMessages(messages <-chan *message.Message) {
	for msg := range messages {
		err := i.processMessage(msg)
		if err != nil {
			i.log.Err(err).Msg("error processing credential msg")
		}
	}
}

func (i *CredentialListener) processMessage(msg *message.Message) error {
	// Keep the pipeline moving no matter what.
	defer func() { msg.Ack() }()

	// Deletion messages. We're the only actor that produces these, so ignore them.
	if msg.Payload == nil {
		return nil
	}

	event := new(TeslaCredentialsCloudEventV1V2)
	if err := json.Unmarshal(msg.Payload, event); err != nil {
		return errors.Wrap(err, "error parsing device event payload")
	}

	return i.processEvent(event)
}

// Usual format of the source field in CloudEvents for anything related to an integration.
const sourcePrefix = "dimo/integration/"

func (i *CredentialListener) processEvent(event *TeslaCredentialsCloudEventV1V2) error {
	var (
		ctx          = context.Background()
		userDeviceID = event.Subject
	)

	if !strings.HasPrefix(event.Source, sourcePrefix) {
		return fmt.Errorf("unexpected event source format: %s", event.Source)
	}
	integrationID := strings.TrimPrefix(event.Source, sourcePrefix)

	var (
		accessToken, refreshToken string
		expiry                    time.Time
	)
	switch event.Type {
	case "zone.dimo.task.tesla.poll.credential":
		// Only devices-api ever sent these, so no point in reacting.
		return nil
	case "zone.dimo.task.tesla.poll.credential.v2", "zone.dimo.task.smartcar.poll.credential":
		accessToken = event.Data.AccessToken
		refreshToken = event.Data.RefreshToken
		expiry = event.Data.Expiry
	default:
		return fmt.Errorf("unexpected event type %s", event.Type)
	}

	integ, err := models.FindUserDeviceAPIIntegration(ctx, i.db().Writer, userDeviceID, integrationID)
	if err != nil {
		return fmt.Errorf("couldn't find device integration for device %s and integration %s: %w", userDeviceID, integrationID, err)
	}

	// Upon initial connection, there will be message that we sent and there's no point in updating the database.
	// TODO: Should we ignore these if they're expired?
	if !integ.AccessToken.Valid || integ.AccessToken.String != accessToken {
		i.log.Debug().Str("userDeviceId", userDeviceID).Str("integrationId", integrationID).Msgf("Saving new credentials.")
		integ.AccessToken = null.StringFrom(accessToken)
		integ.RefreshToken = null.StringFrom(refreshToken)
		integ.AccessExpiresAt = null.TimeFrom(expiry)
		if _, err := integ.Update(ctx, i.db().Writer, boil.Infer()); err != nil {
			return fmt.Errorf("failed to update integration record: %w", err)
		}
	}

	return nil
}
