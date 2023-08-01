package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/Shopify/sarama"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// TODO(elffjs): Setting?
const maxFenceTiles = 12

type GeofencesController struct {
	Settings     *config.Settings
	DBS          func() *db.ReaderWriter
	log          *zerolog.Logger
	producer     sarama.SyncProducer
	deviceDefSvc services.DeviceDefinitionService
}

// NewGeofencesController constructor
func NewGeofencesController(settings *config.Settings, dbs func() *db.ReaderWriter, logger *zerolog.Logger, producer sarama.SyncProducer, deviceDefSvc services.DeviceDefinitionService) GeofencesController {
	return GeofencesController{
		Settings:     settings,
		DBS:          dbs,
		log:          logger,
		producer:     producer,
		deviceDefSvc: deviceDefSvc,
	}
}

const PrivacyFenceEventType = "zone.dimo.device.privacyfence.update"

// Create godoc
// @Description adds a new geofence to the user's account, optionally attached to specific user_devices
// @Tags        geofence
// @Produce     json
// @Accept      json
// @Param       geofence body     controllers.CreateGeofence true "add geofence to user."
// @Success     201      {object} helpers.CreateResponse
// @Security    ApiKeyAuth
// @Security    BearerAuth
// @Router      /user/geofences [post]
func (g *GeofencesController) Create(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	create := CreateGeofence{}
	if err := c.BodyParser(&create); err != nil {
		// Return status 400 and error message.
		return helpers.ErrorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	if err := create.Validate(); err != nil {
		return helpers.ErrorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	tx, err := g.DBS().Writer.DB.BeginTx(c.Context(), nil)
	defer tx.Rollback() //nolint
	if err != nil {
		return err
	}

	// check if already exists
	exists, err := models.Geofences(models.GeofenceWhere.UserID.EQ(userID), models.GeofenceWhere.Name.EQ(create.Name)).Exists(c.Context(), tx)
	if err != nil {
		return err
	}
	if exists {
		return helpers.ErrorResponseHandler(c, errors.New("Geofence with that name already exists for this user"), fiber.StatusBadRequest)
	}

	// Check that the user has access to the devices in the request.
	if len(create.UserDeviceIDs) > 0 {
		allUserDevices, err := models.UserDevices(models.UserDeviceWhere.UserID.EQ(userID)).All(c.Context(), tx)
		if err != nil {
			return fmt.Errorf("failed to look up user's devices: %w", err)
		}

		allUserDeviceIDs := shared.NewStringSet()
		for _, userDevice := range allUserDevices {
			allUserDeviceIDs.Add(userDevice.ID)
		}

		for _, userDeviceID := range create.UserDeviceIDs {
			if !allUserDeviceIDs.Contains(userDeviceID) {
				return helpers.ErrorResponseHandler(c, fmt.Errorf("user does not have a device with id %s", userDeviceID), fiber.StatusBadRequest)
			}
		}
	}

	geofence := models.Geofence{
		ID:        ksuid.New().String(),
		UserID:    userID,
		Name:      create.Name,
		Type:      create.Type,
		H3Indexes: create.H3Indexes,
	}
	err = geofence.Insert(c.Context(), tx, boil.Infer())
	if err != nil {
		return errors.Wrap(err, "error inserting geofence")
	}
	for _, uID := range create.UserDeviceIDs {
		geoToUser := models.UserDeviceToGeofence{
			UserDeviceID: uID,
			GeofenceID:   geofence.ID,
		}
		err = geoToUser.Upsert(c.Context(), tx, true, []string{"user_device_id", "geofence_id"}, boil.Infer(), boil.Infer())
		if err != nil {
			return errors.Wrapf(err, "error upserting user_device_to_geofence")
		}
	}

	if create.Type == models.GeofenceTypePrivacyFence {
		for _, userDeviceID := range create.UserDeviceIDs {
			if err := g.EmitPrivacyFenceUpdates(c.Context(), tx, userDeviceID); err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrapf(err, "error commiting transaction to create geofence")
	}

	return c.Status(fiber.StatusCreated).JSON(helpers.CreateResponse{ID: geofence.ID})
}

type FenceData struct {
	H3Indexes []string `json:"h3Indexes"`
}

func (g *GeofencesController) EmitPrivacyFenceUpdates(ctx context.Context, db boil.ContextExecutor, userDeviceID string) error {
	rels, err := models.UserDeviceToGeofences(
		models.UserDeviceToGeofenceWhere.UserDeviceID.EQ(userDeviceID),
		qm.Load(models.UserDeviceToGeofenceRels.Geofence),
	).All(ctx, db)
	if err != nil {
		return err
	}

	indexes := shared.NewStringSet()

	for _, rel := range rels {
		if rel.R.Geofence.Type != models.GeofenceTypePrivacyFence {
			continue
		}
		for _, index := range rel.R.Geofence.H3Indexes {
			indexes.Add(index)
		}
	}

	// Delete the device's entry from the table if there are no indexes left.
	var value sarama.Encoder

	if indexes.Len() > 0 {
		ce := shared.CloudEvent[FenceData]{
			ID:          ksuid.New().String(),
			Source:      "devices-api",
			SpecVersion: "1.0",
			Subject:     userDeviceID,
			Time:        time.Now(),
			Type:        PrivacyFenceEventType,
			Data: FenceData{
				H3Indexes: indexes.Slice(),
			},
		}
		b, err := json.Marshal(ce)
		if err != nil {
			return err
		}

		value = sarama.ByteEncoder(b)
	}
	msg := &sarama.ProducerMessage{
		Topic: g.Settings.PrivacyFenceTopic,
		Key:   sarama.StringEncoder(userDeviceID),
		Value: value,
	}
	if _, _, err := g.producer.SendMessage(msg); err != nil {
		return err
	}

	return nil
}

// GetAll godoc
// @Description gets all geofences for the current user
// @Tags        geofence
// @Produce     json
// @Success     200 {object} []controllers.GetGeofence
// @Security    ApiKeyAuth
// @Security    BearerAuth
// @Router      /user/geofences [get]
func (g *GeofencesController) GetAll(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	//could not find LoadUserDevices method for eager loading
	items, err := models.Geofences(models.GeofenceWhere.UserID.EQ(userID),
		qm.Load(models.GeofenceRels.UserDeviceToGeofences),
		qm.Load(qm.Rels(models.GeofenceRels.UserDeviceToGeofences, models.UserDeviceToGeofenceRels.UserDevice)),
	).All(c.Context(), g.DBS().Reader)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return c.JSON(fiber.Map{
			"geofences": []GetGeofence{},
		})
	}

	// pull out list of udtg. device def ids
	var ddIds []string
	for _, item := range items {
		for _, udtg := range item.R.UserDeviceToGeofences {
			if !services.Contains(ddIds, udtg.R.UserDevice.DeviceDefinitionID) {
				ddIds = append(ddIds, udtg.R.UserDevice.DeviceDefinitionID)
			}
		}
	}
	// log in odd case ddIds is empty
	if len(ddIds) == 0 {
		log.Warn().Str("userId", userID).Str("httpPath", c.Path()).Str("geofenceItemsLen", fmt.Sprint(len(items))).
			Msg("unexpected case: device definition IDs was empty from geofences with values")
		return c.JSON(fiber.Map{
			"geofences": []GetGeofence{},
		})
	}
	dds, err := g.deviceDefSvc.GetDeviceDefinitionsByIDs(c.Context(), ddIds)
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "failed to pull device definitions")
	}

	fences := make([]GetGeofence, len(items))
	for i, item := range items {
		f := GetGeofence{
			ID:        item.ID,
			Name:      item.Name,
			Type:      item.Type,
			H3Indexes: item.H3Indexes,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
		for _, udtg := range item.R.UserDeviceToGeofences {
			var deviceDef *ddgrpc.GetDeviceDefinitionItemResponse
			for _, dd := range dds {
				if dd.DeviceDefinitionId == udtg.R.UserDevice.DeviceDefinitionID {
					deviceDef = dd
				}
			}
			f.UserDevices = append(f.UserDevices, GeoFenceUserDevice{
				UserDeviceID: udtg.UserDeviceID,
				Name:         udtg.R.UserDevice.Name.Ptr(),
				MMY:          deviceDef.Name,
			})
		}
		fences[i] = f
	}

	return c.JSON(fiber.Map{
		"geofences": fences,
	})
}

// Update godoc
// @Description updates an existing geofence for the current user
// @Tags        geofence
// @Produce     json
// @Accept      json
// @Param       geofenceID path string                     true "geofence id"
// @Param       geofence   body controllers.CreateGeofence true "add geofence to user."
// @Success     204
// @Security    ApiKeyAuth
// @Security    BearerAuth
// @Router      /user/geofences/{geofenceID} [put]
func (g *GeofencesController) Update(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	id := c.Params("geofenceID")
	update := CreateGeofence{}
	if err := c.BodyParser(&update); err != nil {
		return helpers.ErrorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	if err := update.Validate(); err != nil {
		return helpers.ErrorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	tx, err := g.DBS().Writer.DB.BeginTx(c.Context(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	// Return status 400 and error message.
	geofence, err := models.Geofences(models.GeofenceWhere.UserID.EQ(userID), models.GeofenceWhere.ID.EQ(id),
		qm.Load(models.GeofenceRels.UserDeviceToGeofences)).One(c.Context(), tx)
	if err != nil {
		return err
	}

	affectedDeviceIDs := shared.NewStringSet()
	for _, rel := range geofence.R.UserDeviceToGeofences {
		affectedDeviceIDs.Add(rel.UserDeviceID)
	}

	geofence.Name = update.Name
	geofence.Type = update.Type
	geofence.H3Indexes = update.H3Indexes
	_, err = geofence.Update(c.Context(), tx, boil.Whitelist(
		models.GeofenceColumns.Name,
		models.GeofenceColumns.Type,
		models.GeofenceColumns.H3Indexes,
		models.GeofenceColumns.UpdatedAt))
	if err != nil {
		return errors.Wrap(err, "error updating geofence")
	}
	for _, uID := range update.UserDeviceIDs {
		affectedDeviceIDs.Add(uID)
		geoToUser := models.UserDeviceToGeofence{
			UserDeviceID: uID,
			GeofenceID:   geofence.ID,
		}
		err = geoToUser.Upsert(c.Context(), tx, true,
			[]string{models.UserDeviceToGeofenceColumns.UserDeviceID, models.UserDeviceToGeofenceColumns.GeofenceID}, boil.Infer(), boil.Infer())
		if err != nil {
			return errors.Wrapf(err, "error upserting user_device_to_geofence")
		}
	}

	for _, uID := range affectedDeviceIDs.Slice() {
		if err := g.EmitPrivacyFenceUpdates(c.Context(), tx, uID); err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrapf(err, "error commiting transaction to create geofence")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Delete godoc
// @Description hard deletes a geofence from db
// @Tags        geofence
// @Param       geofenceID path string true "geofence id"
// @Success     204
// @Security    ApiKeyAuth
// @Security    BearerAuth
// @Router      /user/geofences/{geofenceID} [delete]
func (g *GeofencesController) Delete(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	id := c.Params("geofenceID")

	tx, err := g.DBS().Writer.DB.BeginTx(c.Context(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	geo, err := models.Geofences(
		models.GeofenceWhere.UserID.EQ(userID),
		models.GeofenceWhere.ID.EQ(id),
		qm.Load(models.GeofenceRels.UserDeviceToGeofences),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return helpers.ErrorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return helpers.ErrorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	for _, rel := range geo.R.UserDeviceToGeofences {
		if err := g.EmitPrivacyFenceUpdates(c.Context(), tx, rel.UserDeviceID); err != nil {
			return err
		}
	}

	if _, err := geo.Delete(c.Context(), tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

type GetGeofence struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Type        string               `json:"type"`
	H3Indexes   []string             `json:"h3Indexes"`
	UserDevices []GeoFenceUserDevice `json:"userDevices"`
	CreatedAt   time.Time            `json:"createdAt"`
	UpdatedAt   time.Time            `json:"updatedAt"`
}

type GeoFenceUserDevice struct {
	UserDeviceID string  `json:"userDeviceId"`
	Name         *string `json:"name"`
	MMY          string  `json:"mmy"`
}

type CreateGeofence struct {
	// required: true
	Name string `json:"name"`
	// one of following: "PrivacyFence", "TriggerEntry", "TriggerExit"
	// required: true
	Type string `json:"type"`
	// required: false
	H3Indexes []string `json:"h3Indexes"`
	// Optionally link the geofence with a list of user device ID
	UserDeviceIDs []string `json:"userDeviceIds"`
}

func (g *CreateGeofence) Validate() error {
	return validation.ValidateStruct(g,
		validation.Field(&g.Name, validation.Required),
		validation.Field(&g.Type, validation.Required, validation.In("PrivacyFence", "TriggerEntry", "TriggerExit")),
		validation.Field(&g.H3Indexes, validation.Length(0, maxFenceTiles)),
	)
}
