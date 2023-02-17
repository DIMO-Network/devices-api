package controllers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/autopi"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/db"
	"github.com/Shopify/sarama"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	signer "github.com/ethereum/go-ethereum/signer/core/apitypes"
)

type UserDevicesController struct {
	Settings                  *config.Settings
	DBS                       func() *db.ReaderWriter
	DeviceDefSvc              services.DeviceDefinitionService
	DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
	log                       *zerolog.Logger
	eventService              services.EventService
	smartcarClient            services.SmartcarClient
	smartcarTaskSvc           services.SmartcarTaskService
	teslaService              services.TeslaService
	teslaTaskService          services.TeslaTaskService
	cipher                    shared.Cipher
	autoPiSvc                 services.AutoPiAPIService
	nhtsaService              services.INHTSAService
	autoPiIngestRegistrar     services.IngestRegistrar
	autoPiTaskService         services.AutoPiTaskService
	drivlyTaskService         services.DrivlyTaskService
	s3                        *s3.Client
	producer                  sarama.SyncProducer
	deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	autoPiIntegration         *autopi.Integration
}

// PrivilegedDevices contains all devices for which a privilege has been shared
type PrivilegedDevices struct {
	Devices []PrivilegedAccessDevice `json:"devices"`
}

// PrivilegedAccessDevice device details for which a privilege has been shared
type PrivilegedAccessDevice struct {
	TokenID      *big.Int       `json:"tokenId"`
	OwnerAddress common.Address `json:"ownerAddress"`
	Device       Device         `json:"type"`
	Privileges   []Privilege    `json:"privileges"`
}

// Privilege ID associated with privilege and expiration time
type Privilege struct {
	ID        int64     `json:"id"`
	UpdatedAt time.Time `json:"updatedAt"`
	ExpiresAt time.Time `json:"expiry"`
}

// Device vehicle make, model, year
type Device struct {
	Make  string `json:"make"`
	Model string `json:"model"`
	Year  int    `json:"year"`
}

// NewUserDevicesController constructor
func NewUserDevicesController(
	settings *config.Settings,
	dbs func() *db.ReaderWriter,
	logger *zerolog.Logger,
	ddSvc services.DeviceDefinitionService,
	ddIntSvc services.DeviceDefinitionIntegrationService,
	eventService services.EventService,
	smartcarClient services.SmartcarClient,
	smartcarTaskSvc services.SmartcarTaskService,
	teslaService services.TeslaService,
	teslaTaskService services.TeslaTaskService,
	cipher shared.Cipher,
	autoPiSvc services.AutoPiAPIService,
	nhtsaService services.INHTSAService,
	autoPiIngestRegistrar services.IngestRegistrar,
	deviceDefinitionRegistrar services.DeviceDefinitionRegistrar,
	autoPiTaskService services.AutoPiTaskService,
	producer sarama.SyncProducer,
	s3NFTClient *s3.Client,
	drivlyTaskService services.DrivlyTaskService,
	autoPi *autopi.Integration,
) UserDevicesController {
	return UserDevicesController{
		Settings:                  settings,
		DBS:                       dbs,
		log:                       logger,
		DeviceDefSvc:              ddSvc,
		DeviceDefIntSvc:           ddIntSvc,
		eventService:              eventService,
		smartcarClient:            smartcarClient,
		smartcarTaskSvc:           smartcarTaskSvc,
		teslaService:              teslaService,
		teslaTaskService:          teslaTaskService,
		cipher:                    cipher,
		autoPiSvc:                 autoPiSvc,
		nhtsaService:              nhtsaService,
		autoPiIngestRegistrar:     autoPiIngestRegistrar,
		autoPiTaskService:         autoPiTaskService,
		s3:                        s3NFTClient,
		producer:                  producer,
		drivlyTaskService:         drivlyTaskService,
		deviceDefinitionRegistrar: deviceDefinitionRegistrar,
		autoPiIntegration:         autoPi,
	}
}

func (udc *UserDevicesController) dbDevicesToDisplay(ctx context.Context, devices []*models.UserDevice) ([]UserDeviceFull, error) {
	apiDevices := []UserDeviceFull{}

	if len(devices) == 0 {
		return apiDevices, nil
	}

	ddIDs := make([]string, len(devices))
	for i, d := range devices {
		ddIDs[i] = d.DeviceDefinitionID
	}

	deviceDefinitionResponse, err := udc.DeviceDefSvc.GetDeviceDefinitionsByIDs(ctx, ddIDs)
	if err != nil {
		return nil, helpers.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+ddIDs[0])
	}

	filterDeviceDefinition := func(id string, items []*ddgrpc.GetDeviceDefinitionItemResponse) (*ddgrpc.GetDeviceDefinitionItemResponse, error) {
		for _, dd := range items {
			if id == dd.DeviceDefinitionId {
				return dd, nil
			}
		}
		return nil, errors.New("no device definition")
	}

	integrations, err := udc.DeviceDefSvc.GetIntegrations(ctx)
	if err != nil {
		return nil, helpers.GrpcErrorToFiber(err, "failed to get integrations")
	}

	for _, d := range devices {
		deviceDefinition, err := filterDeviceDefinition(d.DeviceDefinitionID, deviceDefinitionResponse)
		if err != nil {
			return nil, err
		}

		dd, err := NewDeviceDefinitionFromGRPC(deviceDefinition)
		if err != nil {
			return nil, err
		}

		filteredIntegrations := []services.DeviceCompatibility{}
		if d.CountryCode.Valid {
			if countryRecord := constants.FindCountry(d.CountryCode.String); countryRecord != nil {
				for _, integration := range dd.CompatibleIntegrations {
					if integration.Region == countryRecord.Region {
						integration.Country = d.CountryCode.String // Faking it until the UI updates for regions.
						filteredIntegrations = append(filteredIntegrations, integration)
					}
				}
			}
		}

		dd.CompatibleIntegrations = filteredIntegrations

		var md services.UserDeviceMetadata
		if d.Metadata.Valid {
			if err := d.Metadata.Unmarshal(&md); err != nil {
				return nil, opaqueInternalError
			}
		}

		var nft *NFTData
		pu := []PrivilegeUser{}

		if vnft := d.R.VehicleNFT; vnft != nil {
			nftStatus := vnft.R.MintRequest
			nft = &NFTData{
				Status: nftStatus.Status,
			}
			if nftStatus.Hash.Valid {
				hash := hexutil.Encode(nftStatus.Hash.Bytes)
				nft.TxHash = &hash
			}
			if !vnft.TokenID.IsZero() {
				nft.TokenID = vnft.TokenID.Int(nil)
				nft.TokenURI = fmt.Sprintf("%s/v1/nfts/%s", udc.Settings.DeploymentBaseURL, nft.TokenID)
			}
			if vnft.OwnerAddress.Valid {
				addr := common.BytesToAddress(vnft.OwnerAddress.Bytes)
				nft.OwnerAddress = &addr
			}

			// NFT Privileges
			udp, err := models.NFTPrivileges(
				models.NFTPrivilegeWhere.TokenID.EQ(types.Decimal(d.R.VehicleNFT.TokenID)),
				models.NFTPrivilegeWhere.Expiry.GT(time.Now()),
				models.NFTPrivilegeWhere.ContractAddress.EQ(common.FromHex(udc.Settings.VehicleNFTAddress)),
			).All(ctx, udc.DBS().Reader)
			if err != nil {
				return nil, err
			}

			privByAddr := make(map[string][]Privilege)
			for _, v := range udp {
				ua := common.BytesToAddress(v.UserAddress).Hex()
				privByAddr[ua] = append(privByAddr[ua], Privilege{
					ID:        v.Privilege,
					ExpiresAt: v.Expiry,
					UpdatedAt: v.UpdatedAt,
				})
			}

			for k, v := range privByAddr {
				pu = append(pu, PrivilegeUser{
					Address:    k,
					Privileges: v,
				})
			}
		}

		udf := UserDeviceFull{
			ID:               d.ID,
			VIN:              d.VinIdentifier.Ptr(),
			VINConfirmed:     d.VinConfirmed,
			Name:             d.Name.Ptr(),
			CustomImageURL:   d.CustomImageURL.Ptr(),
			CountryCode:      d.CountryCode.Ptr(),
			DeviceDefinition: dd,
			Integrations:     NewUserDeviceIntegrationStatusesFromDatabase(d.R.UserDeviceAPIIntegrations, integrations),
			Metadata:         md,
			NFT:              nft,
			OptedInAt:        d.OptedInAt.Ptr(),
			PrivilegeUsers:   pu,
		}

		apiDevices = append(apiDevices, udf)
	}

	return apiDevices, nil
}

// GetUserDevices godoc
// @Description gets all devices associated with current user - pulled from token
// @Tags        user-devices
// @Produce     json
// @Success     200 {object} controllers.MyDevicesResp
// @Security    BearerAuth
// @Router      /user/devices/me [get]
func (udc *UserDevicesController) GetUserDevices(c *fiber.Ctx) error {
	// todo grpc call out to grpc service endpoint in the deviceDefinitionsService udc.deviceDefSvc.GetDeviceDefinitionsByIDs(c.Context(), []string{ "todo"} )

	userID := helpers.GetUserID(c)
	devices, err := models.UserDevices(
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleNFT, models.VehicleNFTRels.MintRequest)),
		qm.OrderBy(models.UserDeviceColumns.CreatedAt),
	).All(c.Context(), udc.DBS().Reader)
	if err != nil {
		return helpers.ErrorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	apiMyDevices, err := udc.dbDevicesToDisplay(c.Context(), devices)
	if err != nil {
		return err
	}

	return c.JSON(MyDevicesResp{UserDevices: apiMyDevices})
}

// GetSharedDevices godoc
// @Description gets all devices shared with current user - pulled from token
// @Tags        user-devices
// @Produce     json
// @Success     200 {object} controllers.MyDevicesResp
// @Security    BearerAuth
// @Router      /user/devices/shared [get]
func (udc *UserDevicesController) GetSharedDevices(c *fiber.Ctx) error {
	// todo grpc call out to grpc service endpoint in the deviceDefinitionsService udc.deviceDefSvc.GetDeviceDefinitionsByIDs(c.Context(), []string{ "todo"} )

	userID := helpers.GetUserID(c)

	// TODO(elffjs): Really shouldn't be dialing so much.
	conn, err := grpc.Dial(udc.Settings.UsersAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		udc.log.Err(err).Msg("Failed to create users API client.")
		return opaqueInternalError
	}
	defer conn.Close()

	usersClient := pb.NewUserServiceClient(conn)

	user, err := usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		udc.log.Err(err).Msg("Couldn't retrieve user record.")
		return opaqueInternalError
	}

	var sharedDev []*models.UserDevice

	if user.EthereumAddress != nil {
		// This is N+1 hell.
		userAddr := common.HexToAddress(*user.EthereumAddress)

		privs, err := models.NFTPrivileges(
			models.NFTPrivilegeWhere.ContractAddress.EQ(common.FromHex(udc.Settings.VehicleNFTAddress)),
			models.NFTPrivilegeWhere.UserAddress.EQ(userAddr.Bytes()),
			models.NFTPrivilegeWhere.Expiry.GT(time.Now()),
		).All(c.Context(), udc.DBS().Reader)
		if err != nil {
			return err
		}

		var toks []types.Decimal

	PrivLoop:
		for _, priv := range privs {
			for _, tok := range toks {
				if tok.Cmp(priv.TokenID.Big) == 0 {
					continue PrivLoop
				}
			}

			toks = append(toks, priv.TokenID)

			nft, err := models.VehicleNFTS(
				models.VehicleNFTWhere.TokenID.EQ(types.NewNullDecimal(priv.TokenID.Big)),
				qm.Load(models.VehicleNFTRels.UserDevice),
			).One(c.Context(), udc.DBS().Reader)
			if err != nil {
				if err == sql.ErrNoRows {
					continue
				}
				return err
			}

			ud, err := models.UserDevices(
				models.UserDeviceWhere.ID.EQ(nft.UserDeviceID.String),
				qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
				// Would we get this backreference for free?
				qm.Load(qm.Rels(models.UserDeviceRels.VehicleNFT, models.VehicleNFTRels.MintRequest)),
			).One(c.Context(), udc.DBS().Reader)
			if err != nil {
				return err
			}

			sharedDev = append(sharedDev, ud)
		}
	}

	apiSharedDevices, err := udc.dbDevicesToDisplay(c.Context(), sharedDev)
	if err != nil {
		return err
	}

	return c.JSON(MyDevicesResp{SharedDevices: apiSharedDevices})
}

func NewUserDeviceIntegrationStatusesFromDatabase(udis []*models.UserDeviceAPIIntegration, integrations []*ddgrpc.Integration) []UserDeviceIntegrationStatus {
	out := make([]UserDeviceIntegrationStatus, len(udis))

	for i, udi := range udis {
		// TODO(elffjs): Remove this translation when the frontend is ready for "AuthenticationFailure".
		status := udi.Status

		out[i] = UserDeviceIntegrationStatus{
			IntegrationID: udi.IntegrationID,
			Status:        status,
			ExternalID:    udi.ExternalID.Ptr(),
			CreatedAt:     udi.CreatedAt,
			UpdatedAt:     udi.UpdatedAt,
			Metadata:      udi.Metadata,
		}

		for _, integration := range integrations {
			if integration.Id == udi.IntegrationID {
				out[i].IntegrationVendor = integration.Vendor
				break
			}
		}
	}

	return out
}

const UserDeviceCreationEventType = "com.dimo.zone.device.create"

type UserDeviceEvent struct {
	Timestamp time.Time                      `json:"timestamp"`
	UserID    string                         `json:"userId"`
	Device    services.UserDeviceEventDevice `json:"device"`
}

// RegisterDeviceForUser godoc
// @Description adds a device to a user. can add with only device_definition_id or with MMY, which will create a device_definition on the fly
// @Tags        user-devices
// @Produce     json
// @Accept      json
// @Param       user_device body controllers.RegisterUserDevice true "add device to user. either MMY or id are required"
// @Security    ApiKeyAuth
// @Success     201 {object} controllers.RegisterUserDeviceResponse
// @Security    BearerAuth
// @Router      /user/devices [post]
func (udc *UserDevicesController) RegisterDeviceForUser(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	reg := &RegisterUserDevice{}
	if err := c.BodyParser(reg); err != nil {
		// Return status 400 and error message.
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := reg.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	udFull, err := udc.createUserDevice(c.Context(), *reg.DeviceDefinitionID, reg.CountryCode, userID, nil)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"userDevice": udFull,
	})
}

var opaqueInternalError = fiber.NewError(fiber.StatusInternalServerError, "Internal error.")

// RegisterDeviceForUserFromVIN godoc
// @Description adds a device to a user by decoding a VIN. If cannot decode returns 424 or 500 if error
// @Tags        user-devices
// @Produce     json
// @Accept      json
// @Param       user_device body controllers.RegisterUserDeviceVIN true "add device to user. VIN is required and so is country"
// @Security    ApiKeyAuth
// @Failure		400 "validation failure"
// @Failure		424 "unable to decode VIN"
// @Failure		500 "server error, dependency error"
// @Success     201 {object} controllers.RegisterUserDeviceResponse
// @Security    BearerAuth
// @Router      /user/devices/fromvin [post]
func (udc *UserDevicesController) RegisterDeviceForUserFromVIN(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	reg := &RegisterUserDeviceVIN{}
	if err := c.BodyParser(reg); err != nil {
		// Return status 400 and error message.
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := reg.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	// decode VIN with grpc call
	decodeVIN, err := udc.DeviceDefSvc.DecodeVIN(c.Context(), reg.VIN)
	if err != nil {
		return errors.Wrapf(err, "could not decode vin %s for country %s", reg.VIN, reg.CountryCode)
	}
	if len(decodeVIN.DeviceDefinitionId) == 0 {
		udc.log.Warn().Str("vin", reg.VIN).Str("user_id", userID).
			Msg("unable to decode vin for customer request to create vehicle")
		return fiber.NewError(fiber.StatusFailedDependency, "unable to decode vin")
	}
	// attach device def to user
	udFull, err := udc.createUserDevice(c.Context(), decodeVIN.DeviceDefinitionId, reg.CountryCode, userID, &reg.VIN)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"userDevice": udFull,
	})
}

// RegisterDeviceForUserFromSmartcar godoc
// @Description adds a device to a user by decoding VIN from Smartcar. If cannot decode returns 424 or 500 if error
// @Tags        user-devices
// @Produce     json
// @Accept      json
// @Param       user_device body controllers.RegisterUserDeviceSmartcar true "add device to user. all fields required"
// @Security    ApiKeyAuth
// @Failure		400 "validation failure"
// @Failure		424 "unable to decode VIN"
// @Failure		500 "server error, dependency error"
// @Success     201 {object} controllers.RegisterUserDeviceSmartcar
// @Security    BearerAuth
// @Router      /user/devices/fromsmartcar [post]
func (udc *UserDevicesController) RegisterDeviceForUserFromSmartcar(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	reg := &RegisterUserDeviceSmartcar{}
	if err := c.BodyParser(reg); err != nil {
		// Return status 400 and error message.
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := reg.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	localLog := udc.log.With().Str("user_id", userID).Logger()

	// call SC api with stuff and get VIN
	token, err := udc.smartcarClient.ExchangeCode(c.Context(), reg.Code, reg.RedirectURI)
	if err != nil {
		localLog.Err(err).Msg("Failed to exchange authorization code with Smartcar.")
		// This may not be the user's fault, but 400 for now.
		return fiber.NewError(fiber.StatusBadRequest, "Failed to exchange authorization code with Smartcar.")
	}
	externalID, err := udc.smartcarClient.GetExternalID(c.Context(), token.Access)
	if err != nil {
		localLog.Err(err).Msg("Failed to retrieve vehicle ID from Smartcar.")
		return smartcarCallErr
	}
	vin, err := udc.smartcarClient.GetVIN(c.Context(), token.Access, externalID)
	if err != nil {
		localLog.Err(err).Msg("Failed to retrieve VIN from Smartcar.")
		return smartcarCallErr
	}

	// decode VIN with grpc call
	decodeVIN, err := udc.DeviceDefSvc.DecodeVIN(c.Context(), vin)
	if err != nil {
		return errors.Wrapf(err, "could not decode vin %s for country %s", vin, reg.CountryCode)
	}
	if len(decodeVIN.DeviceDefinitionId) == 0 {
		localLog.Warn().Str("vin", vin).
			Msg("unable to decode vin for customer request to create vehicle")
		return fiber.NewError(fiber.StatusFailedDependency, "unable to decode vin")
	}
	// attach device def to user
	udFull, err := udc.createUserDevice(c.Context(), decodeVIN.DeviceDefinitionId, reg.CountryCode, userID, &vin)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"userDevice": udFull,
	})
}

func (udc *UserDevicesController) createUserDevice(ctx context.Context, deviceDefID, countryCode, userID string, vin *string) (*UserDeviceFull, error) {
	// attach device def to user
	dd, err2 := udc.DeviceDefSvc.GetDeviceDefinitionByID(ctx, deviceDefID)
	if err2 != nil {
		return nil, helpers.GrpcErrorToFiber(err2, fmt.Sprintf("error querying for device definition id: %s ", deviceDefID))
	}

	tx, err := udc.DBS().Writer.DB.BeginTx(ctx, nil)
	defer tx.Rollback() //nolint
	if err != nil {
		return nil, err
	}

	userDeviceID := ksuid.New().String()
	// register device for the user
	ud := models.UserDevice{
		ID:                 userDeviceID,
		UserID:             userID,
		DeviceDefinitionID: dd.DeviceDefinitionId,
		CountryCode:        null.StringFrom(countryCode),
		VinIdentifier:      null.StringFromPtr(vin),
	}
	err = ud.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "could not create user device for def_id: "+dd.DeviceDefinitionId)
	}

	err = tx.Commit() // commmit the transaction
	if err != nil {
		return nil, errors.Wrapf(err, "error commiting transaction to create geofence")
	}

	// todo call devide definitions to check and pull image for this device in case don't have one
	err = udc.eventService.Emit(&services.Event{
		Type:    UserDeviceCreationEventType,
		Subject: userID,
		Source:  "devices-api",
		Data: UserDeviceEvent{
			Timestamp: time.Now(),
			UserID:    userID,
			Device: services.UserDeviceEventDevice{
				ID:    userDeviceID,
				Make:  dd.Make.Name,
				Model: dd.Type.Model,
				Year:  int(dd.Type.Year), // Odd.
			},
		},
	})
	if err != nil {
		udc.log.Err(err).Msg("Failed emitting device creation event")
	}

	ddNice, err := NewDeviceDefinitionFromGRPC(dd)
	if err != nil {
		return nil, err
	}

	// Baby the frontend.
	for i := range ddNice.CompatibleIntegrations {
		ddNice.CompatibleIntegrations[i].Country = countryCode
	}

	return &UserDeviceFull{
		ID:               ud.ID,
		VIN:              ud.VinIdentifier.Ptr(),
		VINConfirmed:     ud.VinConfirmed,
		Name:             ud.Name.Ptr(),
		CustomImageURL:   ud.CustomImageURL.Ptr(),
		DeviceDefinition: ddNice,
		CountryCode:      ud.CountryCode.Ptr(),
		Integrations:     nil, // userDevice just created, there would never be any integrations setup
	}, nil
}

// DeviceOptIn godoc
// @Description Opts the device into data-sharing, and hence rewards.
// @Tags        user-devices
// @Produce     json
// @Param       userDeviceID path string                   true "user device id"
// @Success     204
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/commands/opt-in [post]
func (udc *UserDevicesController) DeviceOptIn(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)

	logger := udc.log.With().Str("routeName", c.Route().Name).Str("userId", userID).Str("userDeviceId", udi).Logger()

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.UserID.EQ(userID),
		models.UserDeviceWhere.ID.EQ(udi),
	).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Device not found.")
		}
		logger.Err(err).Msg("Database error searching for device.")
		return err
	}

	if userDevice.OptedInAt.Valid {
		logger.Info().Time("previousTime", userDevice.OptedInAt.Time).Msg("Already opted in to data-sharing.")
		return c.SendStatus(fiber.StatusNoContent)
	}

	userDevice.OptedInAt = null.TimeFrom(time.Now())

	_, err = userDevice.Update(c.Context(), udc.DBS().Writer, boil.Whitelist(models.UserDeviceColumns.OptedInAt))
	if err != nil {
		return err
	}

	logger.Info().Msg("Opted into data-sharing.")

	return nil
}

func validVINChar(r rune) bool {
	return 'A' <= r && r <= 'Z' || '0' <= r && r <= '9'
}

// UpdateVIN godoc
// @Description updates the VIN on the user device record
// @Tags        user-devices
// @Produce     json
// @Accept      json
// @Param       vin          body controllers.UpdateVINReq true "VIN"
// @Param       userDeviceID path string                   true "user id"
// @Success     204
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/vin [patch]
func (udc *UserDevicesController) UpdateVIN(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)

	logger := udc.log.With().Str("route", c.Route().Name).Str("userId", userID).Str("userDeviceId", udi).Logger()

	var req UpdateVINReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Could not parse request body.")
	}

	req.VIN = strings.TrimSpace(strings.ToUpper(req.VIN))
	if len(req.VIN) != 17 {
		return fiber.NewError(fiber.StatusBadRequest, "VIN is not 17 characters long.")
	}

	for _, r := range req.VIN {
		if !validVINChar(r) {
			return fiber.NewError(fiber.StatusBadRequest, "VIN contains a non-alphanumeric character.")
		}
	}

	// If signed, we should be able to set the VIN to validated.
	if req.Signature != "" {
		vinByte := []byte(req.VIN)
		sig := common.FromHex(req.Signature)
		if len(sig) != 65 {
			logger.Error().Str("rawSignature", req.Signature).Msg("Signature was not 65 bytes.")
			return fiber.NewError(fiber.StatusBadRequest, "Signature is not 65 bytes long.")
		}

		hash := crypto.Keccak256Hash(vinByte)

		recAddr, err := recoverAddress2(hash.Bytes(), sig)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Couldn't recover signer address.")
		}

		found, err := models.AutopiUnits(
			models.AutopiUnitWhere.EthereumAddress.EQ(null.BytesFrom(recAddr.Bytes())),
		).Exists(c.Context(), udc.DBS().Reader)
		if err != nil {
			return err
		}
		if !found {
			return fiber.NewError(fiber.StatusBadRequest, "Signature does not match any known AutoPi.")
		}
	}

	// Don't want phantom reads.
	tx, err := udc.DBS().GetWriterConn().BeginTx(c.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return opaqueInternalError
	}
	defer tx.Rollback() //nolint

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
		models.UserDeviceWhere.UserID.EQ(userID),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Vehicle not found.")
		}
		return err
	}

	if userDevice.VinConfirmed {
		switch {
		case req.Signature == "":
			return fiber.NewError(fiber.StatusConflict, "Vehicle already has a confirmed VIN.")
		case req.VIN != userDevice.VinIdentifier.String:
			return fiber.NewError(fiber.StatusConflict, "Submitted VIN does not match confirmed VIN.")
		default:
			return c.SendStatus(fiber.StatusNoContent)
		}
	}

	if req.Signature != "" {
		existing, err := models.UserDevices(
			models.UserDeviceWhere.VinIdentifier.EQ(null.StringFrom(req.VIN)),
			models.UserDeviceWhere.VinConfirmed.EQ(true),
		).Exists(c.Context(), tx)
		if err != nil {
			return err
		}
		if existing {
			return fiber.NewError(fiber.StatusConflict, "VIN already in use by another vehicle.")
		}
		userDevice.VinConfirmed = true
	}

	userDevice.VinIdentifier = null.StringFrom(req.VIN)

	if _, err := userDevice.Update(c.Context(), tx, boil.Infer()); err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	// TODO: Genericize this for more countries.
	if userDevice.CountryCode.Valid && userDevice.CountryCode.String == "USA" {
		if err := udc.updateUSAPowertrain(c.Context(), userDevice); err != nil {
			logger.Err(err).Msg("Failed to update American powertrain type.")
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (udc *UserDevicesController) updateUSAPowertrain(ctx context.Context, userDevice *models.UserDevice) error {
	// todo grpc pull vin decoder via grpc from device definitions
	resp, err := udc.nhtsaService.DecodeVIN(userDevice.VinIdentifier.String)
	if err != nil {
		return err
	}

	dt, err := resp.DriveType()
	if err != nil {
		return err
	}

	md := new(services.UserDeviceMetadata)
	if err := userDevice.Metadata.Unmarshal(md); err != nil {
		return err
	}

	md.PowertrainType = &dt
	if err := userDevice.Metadata.Marshal(md); err != nil {
		return err
	}
	if _, err := userDevice.Update(ctx, udc.DBS().Writer, boil.Infer()); err != nil {
		return err
	}

	return nil
}

// UpdateName godoc
// @Description updates the Name on the user device record
// @Tags        user-devices
// @Produce     json
// @Accept      json
// @Param       name           body controllers.UpdateNameReq true "Name"
// @Param       user_device_id path string                    true "user id"
// @Success     204
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/name [patch]
func (udc *UserDevicesController) UpdateName(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)

	userDevice, err := models.UserDevices(models.UserDeviceWhere.ID.EQ(udi), models.UserDeviceWhere.UserID.EQ(userID)).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return err
	}
	name := &UpdateNameReq{}
	if err := c.BodyParser(name); err != nil {
		// Return status 400 and error message.
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if name.Name == nil {
		return fiber.NewError(fiber.StatusBadRequest, "name cannot be empty")
	}
	*name.Name = strings.TrimSpace(*name.Name)

	if err := name.validate(); err != nil {
		if name.Name != nil {
			udc.log.Warn().Err(err).Str("userDeviceId", udi).Str("userId", userID).Str("name", *name.Name).Msg("Proposed device name is invalid.")
		}
		return fiber.NewError(fiber.StatusBadRequest, "Name field is limited to 16 alphanumeric characters.")
	}

	userDevice.Name = null.StringFromPtr(name.Name)
	_, err = userDevice.Update(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateCountryCode godoc
// @Description updates the CountryCode on the user device record
// @Tags        user-devices
// @Produce     json
// @Accept      json
// @Param       name body controllers.UpdateCountryCodeReq true "Country code"
// @Success     204
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/country_code [patch]
func (udc *UserDevicesController) UpdateCountryCode(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)
	userDevice, err := models.UserDevices(models.UserDeviceWhere.ID.EQ(udi), models.UserDeviceWhere.UserID.EQ(userID)).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return helpers.ErrorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return err
	}
	countryCode := &UpdateCountryCodeReq{}
	if err := c.BodyParser(countryCode); err != nil {
		// Return status 400 and error message.
		return helpers.ErrorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	userDevice.CountryCode = null.StringFromPtr(countryCode.CountryCode)
	_, err = userDevice.Update(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return helpers.ErrorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateImage godoc
// @Description updates the ImageUrl on the user device record
// @Tags        user-devices
// @Produce     json
// @Accept      json
// @Param       name body controllers.UpdateImageURLReq true "Image URL"
// @Success     204
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/image [patch]
func (udc *UserDevicesController) UpdateImage(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)

	userDevice, err := models.UserDevices(models.UserDeviceWhere.ID.EQ(udi), models.UserDeviceWhere.UserID.EQ(userID)).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return helpers.ErrorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return err
	}
	req := &UpdateImageURLReq{}
	if err := c.BodyParser(req); err != nil {
		// Return status 400 and error message.
		return helpers.ErrorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	userDevice.CustomImageURL = null.StringFromPtr(req.ImageURL)
	_, err = userDevice.Update(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return helpers.ErrorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

type DeviceValuation struct {
	// Contains a list of valuation sets, one for each vendor
	ValuationSets []ValuationSet `json:"valuationSets"`
}
type ValuationSet struct {
	// The source of the valuation (eg. "drivly" or "blackbook")
	Vendor string `json:"vendor"`
	// The time the valuation was pulled or in the case of blackbook, this may be the event time of the device odometer which was used for the valuation
	Updated string `json:"updated,omitempty"`
	// The mileage used for the valuation
	Mileage int `json:"mileage,omitempty"`
	// This will be the zip code used (if any) for the valuation request regardless if the vendor uses it
	ZipCode string `json:"zipCode,omitempty"`
	// Useful when Drivly returns multiple vendors and we've selected one (eg. "drivly:blackbook")
	TradeInSource string `json:"tradeInSource,omitempty"`
	// tradeIn is equal to tradeInAverage when available
	TradeIn int `json:"tradeIn,omitempty"`
	// tradeInClean, tradeInAverage, and tradeInRough my not always be available
	TradeInClean   int `json:"tradeInClean,omitempty"`
	TradeInAverage int `json:"tradeInAverage,omitempty"`
	TradeInRough   int `json:"tradeInRough,omitempty"`
	// Useful when Drivly returns multiple vendors and we've selected one (eg. "drivly:blackbook")
	RetailSource string `json:"retailSource,omitempty"`
	// retail is equal to retailAverage when available
	Retail int `json:"retail,omitempty"`
	// retailClean, retailAverage, and retailRough my not always be available
	RetailClean   int    `json:"retailClean,omitempty"`
	RetailAverage int    `json:"retailAverage,omitempty"`
	RetailRough   int    `json:"retailRough,omitempty"`
	OdometerUnit  string `json:"odometerUnit"`
	Odometer      int    `json:"odometer"`
	// UserDisplayPrice the top level value to show to users in mobile app
	UserDisplayPrice int `json:"userDisplayPrice"`
}

// GetValuations godoc
// @Description gets valuations for a particular user device. Includes only price valuations, not offers. only gets the latest valuation.
// @Tags        user-devices
// @Produce     json
// @Param       userDeviceID path string true "user device id"
// @Success     200 {object} controllers.DeviceValuation
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/valuations [get]
func (udc *UserDevicesController) GetValuations(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)

	// Ensure user is owner of user device
	userDeviceExists, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
		models.UserDeviceWhere.UserID.EQ(userID),
	).Exists(c.Context(), udc.DBS().Reader)
	if err != nil {
		return err
	}
	if !userDeviceExists {
		return c.SendStatus(fiber.StatusForbidden)
	}

	logger := udc.log.With().Str("route", c.Route().Path).Str("userId", userID).Str("userDeviceId", udi).Logger()

	dVal := DeviceValuation{
		ValuationSets: []ValuationSet{},
	}

	// Drivly data
	valuationData, err := models.ExternalVinData(
		models.ExternalVinDatumWhere.UserDeviceID.EQ(null.StringFrom(udi)),
		qm.Where("pricing_metadata is not null or vincario_metadata is not null"),
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).One(c.Context(), udc.DBS().Reader)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if valuationData != nil {
		if valuationData.PricingMetadata.Valid {
			drivlyVal := ValuationSet{
				Vendor:        "drivly",
				TradeInSource: "drivly",
				RetailSource:  "drivly",
				Updated:       valuationData.UpdatedAt.Format(time.RFC3339),
			}
			drivlyJSON := valuationData.PricingMetadata.JSON
			requestJSON := valuationData.RequestMetadata.JSON
			drivlyMileage := gjson.GetBytes(drivlyJSON, "mileage")
			if drivlyMileage.Exists() {
				drivlyVal.Mileage = int(drivlyMileage.Int())
				drivlyVal.Odometer = int(drivlyMileage.Int())
				drivlyVal.OdometerUnit = "miles"
			} else {
				requestMileage := gjson.GetBytes(requestJSON, "mileage")
				if requestMileage.Exists() {
					drivlyVal.Mileage = int(requestMileage.Int())
				}
			}
			requestZipCode := gjson.GetBytes(requestJSON, "zipCode")
			if requestZipCode.Exists() {
				drivlyVal.ZipCode = requestZipCode.String()
			}
			// Drivly Trade-In
			drivlyVal.TradeIn = extractDrivlyValuation(drivlyJSON, "trade")
			drivlyVal.TradeInAverage = drivlyVal.TradeIn
			// Drivly Retail
			drivlyVal.Retail = extractDrivlyValuation(drivlyJSON, "retail")
			drivlyVal.RetailAverage = drivlyVal.Retail

			// often drivly saves valuations with 0 for value, if this is case do not consider it
			if drivlyVal.Retail > 0 || drivlyVal.TradeIn > 0 {
				// set the price to display to users
				drivlyVal.UserDisplayPrice = (drivlyVal.Retail + drivlyVal.TradeIn) / 2
				dVal.ValuationSets = append(dVal.ValuationSets, drivlyVal)
			} else {
				logger.Warn().Msg("did not find a drivly trade-in or retail value, or json in unexpected format")
			}
		} else if valuationData.VincarioMetadata.Valid {
			vincarioVal := ValuationSet{
				Vendor:        "vincario",
				TradeInSource: "vincario",
				RetailSource:  "vincario",
				Updated:       valuationData.UpdatedAt.Format(time.RFC3339),
			}
			valJSON := valuationData.VincarioMetadata.JSON
			requestJSON := valuationData.RequestMetadata.JSON
			odometerMarket := gjson.GetBytes(valJSON, "market_odometer.odometer_avg")
			if odometerMarket.Exists() {
				vincarioVal.Mileage = int(odometerMarket.Int())
				vincarioVal.Odometer = int(odometerMarket.Int())
				vincarioVal.OdometerUnit = gjson.GetBytes(valJSON, "market_odometer.odometer_unit").String()
			}
			// todo this needs to be implemented in the load_valuations script
			requestPostalCode := gjson.GetBytes(requestJSON, "postalCode")
			if requestPostalCode.Exists() {
				vincarioVal.ZipCode = requestPostalCode.String()
			}
			// vincario Trade-In - just using the price below mkt mean
			vincarioVal.TradeIn = int(gjson.GetBytes(valJSON, "market_price.price_below").Int())
			vincarioVal.TradeInAverage = vincarioVal.TradeIn
			// vincario Retail - just using the price above mkt mean
			vincarioVal.Retail = int(gjson.GetBytes(valJSON, "market_price.price_above").Int())
			vincarioVal.RetailAverage = vincarioVal.Retail

			vincarioVal.UserDisplayPrice = int(gjson.GetBytes(valJSON, "market_price.price_avg").Int())

			// often drivly saves valuations with 0 for value, if this is case do not consider it
			if vincarioVal.Retail > 0 || vincarioVal.TradeIn > 0 {
				dVal.ValuationSets = append(dVal.ValuationSets, vincarioVal)
			} else {
				logger.Warn().Msg("did not find a market value from vincario, or valJSON in unexpected format")
			}

		}
	}

	return c.JSON(dVal)
}

// extractDrivlyValuation pulls out the price from the drivly json, based on the passed in key, eg. trade or retail. calculates average if no root property found
func extractDrivlyValuation(drivlyJSON []byte, key string) int {
	if gjson.GetBytes(drivlyJSON, key).Exists() && !gjson.GetBytes(drivlyJSON, key).IsObject() {
		v := gjson.GetBytes(drivlyJSON, key).String()
		vf, _ := strconv.ParseFloat(v, 64)
		return int(vf)
	}
	// get all values
	pricings := map[string]int{}
	if gjson.GetBytes(drivlyJSON, key+".blackBook.totalAvg").Exists() {
		values := gjson.GetManyBytes(drivlyJSON, key+".blackBook.totalRough", key+".blackBook.totalAvg", key+".blackBook.totalClean")
		pricings["blackbook"] = int(values[1].Int())
	}
	if gjson.GetBytes(drivlyJSON, key+".kelley.good").Exists() {
		pricings["kbb"] = int(gjson.GetBytes(drivlyJSON, key+".kelley.good").Int())
	}
	if gjson.GetBytes(drivlyJSON, key+".edmunds.average").Exists() {
		values := gjson.GetManyBytes(drivlyJSON, key+".edmunds.rough", key+".edmunds.average", key+".edmunds.clean")
		pricings["edmunds"] = int(values[1].Int())
	}
	if len(pricings) > 1 {
		sum := 0
		for _, v := range pricings {
			sum += v
		}
		return sum / len(pricings)
	}

	return 0
}

type DeviceOffer struct {
	// Contains a list of offer sets, one for each source
	OfferSets []OfferSet `json:"offerSets"`
}
type OfferSet struct {
	// The source of the offers (eg. "drivly")
	Source string `json:"source"`
	// The time the offers were pulled
	Updated string `json:"updated,omitempty"`
	// The mileage used for the offers
	Mileage int `json:"mileage,omitempty"`
	// This will be the zip code used (if any) for the offers request regardless if the source uses it
	ZipCode string `json:"zipCode,omitempty"`
	// Contains a list of offers from the source
	Offers []Offer `json:"offers"`
}
type Offer struct {
	// The vendor of the offer (eg. "carmax", "carvana", etc.)
	Vendor string `json:"vendor"`
	// The offer price from the vendor
	Price int `json:"price,omitempty"`
	// The offer URL from the vendor
	URL string `json:"url,omitempty"`
	// An error from the vendor (eg. when the VIN is invalid)
	Error string `json:"error,omitempty"`
	// The grade of the offer from the vendor (eg. "RETAIL")
	Grade string `json:"grade,omitempty"`
	// The reason the offer was declined from the vendor
	DeclineReason string `json:"declineReason,omitempty"`
}

// GetOffers godoc
// @Description gets offers for a particular user device
// @Tags        user-devices
// @Produce     json
// @Success     200 {object} controllers.DeviceOffer
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/offers [get]
func (udc *UserDevicesController) GetOffers(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)

	// Ensure user is owner of user device
	userDeviceExists, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
		models.UserDeviceWhere.UserID.EQ(userID),
	).Exists(c.Context(), udc.DBS().Reader)
	if err != nil {
		return err
	}
	if !userDeviceExists {
		return c.SendStatus(fiber.StatusForbidden)
	}

	dOffer := DeviceOffer{
		OfferSets: []OfferSet{},
	}

	// Drivly data
	drivlyVinData, err := models.ExternalVinData(
		models.ExternalVinDatumWhere.UserDeviceID.EQ(null.StringFrom(udi)),
		models.ExternalVinDatumWhere.OfferMetadata.IsNotNull(), // offer_metadata is sourced from drivly
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).One(c.Context(), udc.DBS().Reader)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if drivlyVinData != nil {
		drivlyOffers := OfferSet{}
		drivlyOffers.Source = "drivly"
		drivlyJSON := drivlyVinData.OfferMetadata.JSON
		requestJSON := drivlyVinData.RequestMetadata.JSON
		drivlyOffers.Updated = drivlyVinData.UpdatedAt.Format(time.RFC3339)
		requestMileage := gjson.GetBytes(requestJSON, "mileage")
		if requestMileage.Exists() {
			drivlyOffers.Mileage = int(requestMileage.Int())
		}
		requestZipCode := gjson.GetBytes(requestJSON, "zipCode")
		if requestZipCode.Exists() {
			drivlyOffers.ZipCode = requestZipCode.String()
		}
		// Drivly Offers
		gjson.GetBytes(drivlyJSON, `@keys.#(%"*Price")#`).ForEach(func(key, value gjson.Result) bool {
			offer := Offer{}
			offer.Vendor = strings.TrimSuffix(value.String(), "Price") // eg. vroom, carvana, or carmax
			gjson.GetBytes(drivlyJSON, `@keys.#(%"`+offer.Vendor+`*")#`).ForEach(func(key, value gjson.Result) bool {
				prop := strings.TrimPrefix(value.String(), offer.Vendor)
				if prop == "Url" {
					prop = "URL"
				}
				if !reflect.ValueOf(&offer).Elem().FieldByName(prop).CanSet() {
					return true
				}
				val := gjson.GetBytes(drivlyJSON, value.String())
				switch val.Type {
				case gjson.Null: // ignore null values
					return true
				case gjson.Number: // for "Price"
					reflect.ValueOf(&offer).Elem().FieldByName(prop).Set(reflect.ValueOf(int(val.Int())))
				case gjson.JSON: // for "Error"
					if prop == "Error" {
						val = gjson.GetBytes(drivlyJSON, value.String()+".error.title")
						if val.Exists() {
							offer.Error = val.String()
							// reflect.ValueOf(&offer).Elem().FieldByName(prop).Set(reflect.ValueOf(val.String()))
						}
					}
				default: // for everything else
					reflect.ValueOf(&offer).Elem().FieldByName(prop).Set(reflect.ValueOf(val.String()))
				}
				return true
			})
			drivlyOffers.Offers = append(drivlyOffers.Offers, offer)
			return true
		})
		dOffer.OfferSets = append(dOffer.OfferSets, drivlyOffers)
	}

	return c.JSON(dOffer)

}

type DeviceRange struct {
	// Contains a list of range sets, one for each range basis (may be empty)
	RangeSets []RangeSet `json:"rangeSets"`
}

type RangeSet struct {
	// The time the data was collected
	Updated string `json:"updated"`
	// The basis for the range calculation (eg. "MPG" or "MPG Highway")
	RangeBasis string `json:"rangeBasis"`
	// The estimated range distance
	RangeDistance int `json:"rangeDistance"`
	// The unit used for the rangeDistance (eg. "miles" or "kilometers")
	RangeUnit string `json:"rangeUnit"`
}

// GetRange godoc
// @Description gets the estimated range for a particular user device
// @Tags        user-devices
// @Produce     json
// @Success     200 {object} controllers.DeviceRange
// @Security    BearerAuth
// @Param       userDeviceID path string true "user device id"
// @Router      /user/devices/{userDeviceID}/range [get]
func (udc *UserDevicesController) GetRange(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)

	// Ensure user is owner of user device
	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(models.UserDeviceRels.UserDeviceData),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		return err
	}

	dds, err := udc.DeviceDefSvc.GetDeviceDefinitionsByIDs(c.Context(), []string{userDevice.DeviceDefinitionID})
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+userDevice.DeviceDefinitionID)
	}

	deviceRange := DeviceRange{
		RangeSets: []RangeSet{},
	}
	udd := userDevice.R.UserDeviceData
	if len(dds) > 0 && dds[0].DeviceAttributes != nil && len(udd) > 0 {
		var fuelTankCapGal, mpg, mpgHwy float64
		for _, attr := range dds[0].DeviceAttributes {
			switch attr.Name {
			case "fuel_tank_capacity_gal":
				if v, err := strconv.ParseFloat(attr.Value, 32); err == nil {
					fuelTankCapGal = v
				}
			case "mpg":
				if v, err := strconv.ParseFloat(attr.Value, 32); err == nil {
					mpg = v
				}
			case "mpg_highway":
				if v, err := strconv.ParseFloat(attr.Value, 32); err == nil {
					mpgHwy = v
				}
			}
		}
		sortByJSONFieldMostRecent(udd, "fuelPercentRemaining")
		fuelPercentRemaining := gjson.GetBytes(udd[0].Data.JSON, "fuelPercentRemaining")
		dataUpdatedOn := gjson.GetBytes(udd[0].Data.JSON, "timestamp").Time()
		if fuelPercentRemaining.Exists() && fuelTankCapGal > 0 && mpg > 0 {
			fuelTankAtGal := fuelTankCapGal * fuelPercentRemaining.Float()
			rangeSet := RangeSet{
				Updated:       dataUpdatedOn.Format(time.RFC3339),
				RangeBasis:    "MPG",
				RangeDistance: int(mpg * fuelTankAtGal),
				RangeUnit:     "miles",
			}
			deviceRange.RangeSets = append(deviceRange.RangeSets, rangeSet)
			if mpgHwy > 0 {
				rangeSet.RangeBasis = "MPG Highway"
				rangeSet.RangeDistance = int(mpgHwy * fuelTankAtGal)
				deviceRange.RangeSets = append(deviceRange.RangeSets, rangeSet)
			}
		}
		sortByJSONFieldMostRecent(udd, "range")
		reportedRange := gjson.GetBytes(udd[0].Data.JSON, "range")
		dataUpdatedOn = gjson.GetBytes(udd[0].Data.JSON, "timestamp").Time()
		if reportedRange.Exists() {
			reportedRangeMiles := int(reportedRange.Float() / services.MilesToKmFactor)
			rangeSet := RangeSet{
				Updated:       dataUpdatedOn.Format(time.RFC3339),
				RangeBasis:    "Vehicle Reported",
				RangeDistance: reportedRangeMiles,
				RangeUnit:     "miles",
			}
			deviceRange.RangeSets = append(deviceRange.RangeSets, rangeSet)
		}
	}

	return c.JSON(deviceRange)

}

// DeleteUserDevice godoc
// @Description delete the user device record (hard delete)
// @Tags        user-devices
// @Param       userDeviceID path string true "device id"
// @Success     204
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID} [delete]
func (udc *UserDevicesController) DeleteUserDevice(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(qm.Rels(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationRels.AutopiUnit)),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Device not found.")
		}
		return err
	}

	dd, err := udc.DeviceDefSvc.GetDeviceDefinitionByID(c.Context(), userDevice.DeviceDefinitionID)
	if err != nil {
		return helpers.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+userDevice.DeviceDefinitionID)
	}

	for _, apiInteg := range userDevice.R.UserDeviceAPIIntegrations {
		if unit := apiInteg.R.AutopiUnit; unit != nil && !unit.VehicleTokenID.IsZero() {
			return fiber.NewError(fiber.StatusConflict, fmt.Sprintf("Cannot delete vehicle before unpairing AutoPi %s on-chain.", unit.AutopiUnitID))
		}
	}

	for _, apiInteg := range userDevice.R.UserDeviceAPIIntegrations {
		err := udc.deleteDeviceIntegration(c.Context(), userID, udi, apiInteg.IntegrationID, dd)
		if err != nil {
			return err
		}
	}

	_, err = userDevice.Delete(c.Context(), udc.DBS().Writer)
	if err != nil {
		return err
	}

	err = udc.eventService.Emit(&services.Event{
		Type:    "com.dimo.zone.device.delete",
		Subject: userID,
		Source:  "devices-api",
		Data: UserDeviceEvent{
			Timestamp: time.Now(),
			UserID:    userID,
			Device: services.UserDeviceEventDevice{
				ID:    udi,
				Make:  dd.Make.Name,
				Model: dd.Type.Model,
				Year:  int(dd.Type.Year),
			},
		},
	})
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetMintDevice godoc
// @Description Returns the data the user must sign in order to mint this device.
// @Tags        user-devices
// @Param       userDeviceID path     string true "user device ID"
// @Success     200          {object} signer.TypedData
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/commands/mint [get]
func (udc *UserDevicesController) GetMintDevice(c *fiber.Ctx) error {
	userDeviceID := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		models.UserDeviceWhere.UserID.EQ(userID),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "No device with that ID found.")
	}

	dd, err := udc.DeviceDefSvc.GetDeviceDefinitionByID(c.Context(), userDevice.DeviceDefinitionID)
	if err != nil {
		return helpers.GrpcErrorToFiber(err, fmt.Sprintf("error querying for device definition id: %s ", userDevice.DeviceDefinitionID))
	}

	if dd.Make.TokenId == 0 {
		return fiber.NewError(fiber.StatusConflict, "Device make not yet minted.")
	}
	makeTokenID := big.NewInt(int64(dd.Make.TokenId))

	conn, err := grpc.Dial(udc.Settings.UsersAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		udc.log.Err(err).Msg("Failed to create users API client.")
		return opaqueInternalError
	}
	defer conn.Close()

	usersClient := pb.NewUserServiceClient(conn)

	user, err := usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		udc.log.Err(err).Msg("Couldn't retrieve user record.")
		return opaqueInternalError
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusBadRequest, "User does not have an Ethereum address on file.")
	}

	client := registry.Client{
		Producer:     udc.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(udc.Settings.DIMORegistryChainID),
			Address: common.HexToAddress(udc.Settings.DIMORegistryAddr),
			Name:    "DIMO",
			Version: "1",
		},
	}

	deviceMake := dd.Make.Name
	deviceModel := dd.Type.Model
	deviceYear := strconv.Itoa(int(dd.Type.Year))

	mvs := registry.MintVehicleSign{
		ManufacturerNode: makeTokenID,
		Owner:            common.HexToAddress(*user.EthereumAddress),
		Attributes:       []string{"Make", "Model", "Year"},
		Infos:            []string{deviceMake, deviceModel, deviceYear},
	}

	return c.JSON(client.GetPayload(&mvs))
}

// TODO(elffjs): Do not keep these functions in this file!
func computeTypedDataHash(td *signer.TypedData) (hash common.Hash, err error) {
	domainSep, err := td.HashStruct("EIP712Domain", td.Domain.Map())
	if err != nil {
		return
	}
	msgHash, err := td.HashStruct(td.PrimaryType, td.Message)
	if err != nil {
		return
	}

	payload := []byte{0x19, 0x01}
	payload = append(payload, domainSep...)
	payload = append(payload, msgHash...)

	hash = crypto.Keccak256Hash(payload)
	return
}

func recoverAddress2(hash []byte, sig []byte) (common.Address, error) {
	if len(sig) != 65 {
		return common.Address{}, fmt.Errorf("signature has invalid length %d", len(sig))
	}

	fixedSig := make([]byte, len(sig))
	copy(fixedSig, sig)
	fixedSig[64] -= 27

	uncPubKey, err := crypto.Ecrecover(hash, fixedSig)
	if err != nil {
		return common.Address{}, err
	}

	pubKey, err := crypto.UnmarshalPubkey(uncPubKey)
	if err != nil {
		return common.Address{}, err
	}

	return crypto.PubkeyToAddress(*pubKey), nil
}

func recoverAddress(td *signer.TypedData, signature []byte) (addr common.Address, err error) {
	hash, err := computeTypedDataHash(td)
	if err != nil {
		return
	}
	signature[64] -= 27
	rawPub, err := crypto.Ecrecover(hash[:], signature)
	if err != nil {
		return
	}

	pub, err := crypto.UnmarshalPubkey(rawPub)
	if err != nil {
		return
	}
	addr = crypto.PubkeyToAddress(*pub)
	return
}

// UpdateNFTImage godoc
// @Description Updates a user's NFT image.
// @Tags        user-devices
// @Param       userDeviceId path string                   true "user device id"
// @Param       nftIamges body controllers.NFTImageData true "base64-encoded NFT image data"
// @Success     204
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceId}/commands/update-nft-image [post]
func (udc *UserDevicesController) UpdateNFTImage(c *fiber.Ctx) error {
	userDeviceID := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(models.UserDeviceRels.VehicleNFT),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "No device with that ID found.")
	}

	if userDevice.R.VehicleNFT == nil || userDevice.R.VehicleNFT.TokenID.IsZero() {
		return fiber.NewError(fiber.StatusBadRequest, "Vehicle not minted.")
	}

	mr := new(MintRequest)
	if err := c.BodyParser(mr); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request body.")
	}

	// This may not be there, but if it is we should delete it.
	imageData := strings.TrimPrefix(mr.ImageData, "data:image/png;base64,")

	image, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Primary image not properly base64-encoded.")
	}

	if len(image) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Empty image field.")
	}

	_, err = udc.s3.PutObject(c.Context(), &s3.PutObjectInput{
		Bucket: &udc.Settings.NFTS3Bucket,
		Key:    aws.String(userDevice.R.VehicleNFT.MintRequestID + ".png"),
		Body:   bytes.NewReader(image),
	})
	if err != nil {
		udc.log.Err(err).Msg("Failed to save image to S3.")
		return opaqueInternalError
	}

	// This may not be there, but if it is we should delete it.
	imageDataTransp := strings.TrimPrefix(mr.ImageDataTransparent, "data:image/png;base64,")

	// Should be okay if empty or not provided.
	imageTransp, err := base64.StdEncoding.DecodeString(imageDataTransp)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Transparent image not properly base64-encoded.")
	}

	if len(imageTransp) != 0 {
		_, err = udc.s3.PutObject(c.Context(), &s3.PutObjectInput{
			Bucket: &udc.Settings.NFTS3Bucket,
			Key:    aws.String(userDevice.R.VehicleNFT.MintRequestID + "_transparent.png"),
			Body:   bytes.NewReader(imageTransp),
		})
		if err != nil {
			udc.log.Err(err).Msg("Failed to save transparent image to S3.")
			return opaqueInternalError
		}
	}

	return err
}

// PostMintDevice godoc
// @Description Sends a mint device request to the blockchain
// @Tags        user-devices
// @Param       userDeviceID path string                  true "user device ID"
// @Param       mintRequest  body controllers.MintRequest true "Signature and NFT data"
// @Success     200
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/commands/mint [post]
func (udc *UserDevicesController) PostMintDevice(c *fiber.Ctx) error {
	userDeviceID := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)

	logger := udc.log.With().
		Str("userId", userID).
		Str("userDeviceId", userDeviceID).
		Str("route", c.Route().Name).
		Logger()

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		models.UserDeviceWhere.UserID.EQ(userID),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "No device with that ID found.")
	}

	if !userDevice.VinConfirmed {
		return fiber.NewError(fiber.StatusConflict, "VIN not confirmed.")
	}

	dd, err2 := udc.DeviceDefSvc.GetDeviceDefinitionByID(c.Context(), userDevice.DeviceDefinitionID)
	if err2 != nil {
		return helpers.GrpcErrorToFiber(err2, fmt.Sprintf("error querying for device definition id: %s ", userDevice.DeviceDefinitionID))
	}

	if dd.Make.TokenId == 0 {
		return fiber.NewError(fiber.StatusConflict, "Device make not yet minted.")
	}

	makeTokenID := big.NewInt(int64(dd.Make.TokenId))

	conn, err := grpc.Dial(udc.Settings.UsersAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		udc.log.Err(err).Msg("Failed to create users API client.")
		return opaqueInternalError
	}
	defer conn.Close()

	usersClient := pb.NewUserServiceClient(conn)

	user, err := usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		udc.log.Err(err).Msg("Couldn't retrieve user record.")
		return opaqueInternalError
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusBadRequest, "User does not have an Ethereum address on file.")
	}

	client := registry.Client{
		Producer:     udc.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(udc.Settings.DIMORegistryChainID),
			Address: common.HexToAddress(udc.Settings.DIMORegistryAddr),
			Name:    "DIMO",
			Version: "1",
		},
	}

	deviceMake := dd.Make.Name
	deviceModel := dd.Type.Model
	deviceYear := strconv.Itoa(int(dd.Type.Year))

	mvs := registry.MintVehicleSign{
		ManufacturerNode: makeTokenID,
		Owner:            common.HexToAddress(*user.EthereumAddress),
		Attributes:       []string{"Make", "Model", "Year"},
		Infos:            []string{deviceMake, deviceModel, deviceYear},
	}

	mr := new(MintRequest)
	if err := c.BodyParser(mr); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request body.")
	}

	// This may not be there, but if it is we should delete it.
	imageData := strings.TrimPrefix(mr.ImageData, "data:image/png;base64,")

	image, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Primary image not properly base64-encoded.")
	}

	if len(image) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Empty image field.")
	}

	requestID := ksuid.New().String()

	logger.Info().
		Interface("httpRequestBody", mr).
		Interface("client", client).
		Interface("mintVehicleSign", mvs).
		Interface("typedData", client.GetPayload(&mvs)).
		Msg("Got request.")

	_, err = udc.s3.PutObject(c.Context(), &s3.PutObjectInput{
		Bucket: &udc.Settings.NFTS3Bucket,
		Key:    aws.String(requestID + ".png"),
		Body:   bytes.NewReader(image),
	})
	if err != nil {
		logger.Err(err).Msg("Failed to save image to S3.")
		return opaqueInternalError
	}

	// This may not be there, but if it is we should delete it.
	imageDataTransp := strings.TrimPrefix(mr.ImageDataTransparent, "data:image/png;base64,")

	imageTransp, err := base64.StdEncoding.DecodeString(imageDataTransp)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Transparent image not properly base64-encoded.")
	}

	if len(imageTransp) != 0 {
		_, err = udc.s3.PutObject(c.Context(), &s3.PutObjectInput{
			Bucket: &udc.Settings.NFTS3Bucket,
			Key:    aws.String(requestID + "_transparent.png"),
			Body:   bytes.NewReader(imageTransp),
		})
		if err != nil {
			logger.Err(err).Msg("Failed to save transparent image to S3.")
			return opaqueInternalError
		}
	}

	hash, err := client.Hash(&mvs)
	if err != nil {
		return opaqueInternalError
	}

	sigBytes := common.FromHex(mr.Signature)

	if len(sigBytes) != 65 {
		logger.Error().Str("rawSignature", mr.Signature).Msg("Signature was not 65 bytes.")
		return fiber.NewError(fiber.StatusBadRequest, "Signature must be 65 bytes.")
	}

	sigBytesYellowPaper := make([]byte, len(sigBytes))
	copy(sigBytesYellowPaper, sigBytes)
	sigBytesYellowPaper[64] -= 27

	recUncPubKey, err := crypto.Ecrecover(hash[:], sigBytesYellowPaper)
	if err != nil {
		return err
	}

	recPubKey, err := crypto.UnmarshalPubkey(recUncPubKey)
	if err != nil {
		return err
	}

	recAddr := crypto.PubkeyToAddress(*recPubKey)
	realAddr := common.HexToAddress(*user.EthereumAddress)

	if recAddr != realAddr {
		return fiber.NewError(fiber.StatusBadRequest, "Signature incorrect.")
	}

	mtr := models.MetaTransactionRequest{
		ID:     requestID,
		Status: "Unsubmitted",
	}
	err = mtr.Insert(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return err
	}

	nft := models.VehicleNFT{
		MintRequestID: requestID,
		UserDeviceID:  null.StringFrom(userDevice.ID),
		Vin:           userDevice.VinIdentifier.String,
	}

	err = nft.Insert(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return err
	}

	udc.log.Info().Str("userDeviceId", userDevice.ID).Str("requestId", requestID).Msg("Submitted metatransaction request.")

	return client.MintVehicleSign(requestID, makeTokenID, realAddr, []registry.AttributeInfoPair{
		{Attribute: "Make", Info: deviceMake},
		{Attribute: "Model", Info: deviceModel},
		{Attribute: "Year", Info: deviceYear},
	}, sigBytes)
}

type MintEventData struct {
	RequestID    string   `json:"requestId"`
	UserDeviceID string   `json:"userDeviceId"`
	Owner        string   `json:"owner"`
	RootNode     *big.Int `json:"rootNode"`
	Attributes   []string `json:"attributes"`
	Infos        []string `json:"infos"`
	// Signature is the EIP-712 signature of the RootNode, Attributes, and Infos fields.
	Signature string `json:"signature"`
}

// MintRequest contains the user's signature for the mint request as well as the
// NFT image.
type MintRequest struct {
	NFTImageData
	// Signature is the hex encoding of the EIP-712 signature result.
	Signature string `json:"signature" validate:"required"`
}

type NFTImageData struct {
	// ImageData contains the base64-encoded NFT PNG image.
	ImageData string `json:"imageData" validate:"required"`
	// ImageDataTransparent contains the base64-encoded NFT PNG image
	// with a transparent background, for use in the app. For compatibility
	// with older versions it is not required.
	ImageDataTransparent string `json:"imageDataTransparent" validate:"optional"`
}

type RegisterUserDevice struct {
	DeviceDefinitionID *string `json:"deviceDefinitionId"`
	CountryCode        string  `json:"countryCode"`
}

type RegisterUserDeviceResponse struct {
	UserDeviceID            string                         `json:"userDeviceId"`
	DeviceDefinitionID      string                         `json:"deviceDefinitionId"`
	IntegrationCapabilities []services.DeviceCompatibility `json:"integrationCapabilities"`
}

type RegisterUserDeviceVIN struct {
	VIN         string `json:"vin"`
	CountryCode string `json:"countryCode"`
}

type RegisterUserDeviceSmartcar struct {
	// Code refers to the auth code provided by smartcar when user logs in
	Code        string `json:"code"`
	RedirectURI string `json:"redirectURI"`
	CountryCode string `json:"countryCode"`
}

type AdminRegisterUserDevice struct {
	RegisterUserDevice
	ID          string  `json:"id"`          // KSUID from client,
	CreatedDate int64   `json:"createdDate"` // unix timestamp
	VehicleName *string `json:"vehicleName"`
	VIN         string  `json:"vin"`
	ImageURL    *string `json:"imageUrl"`
	Verified    bool    `json:"verified"`
}

type UpdateVINReq struct {
	// VIN is a vehicle identification number. At the very least, it must be
	// 17 characters in length and contain only letters and numbers.
	VIN string `json:"vin" example:"4Y1SL65848Z411439" validate:"required"`
	// Signature is the hex-encoded result of the AutoPi signing the VIN. It must
	// be present to verify the VIN.
	Signature string `json:"signature" example:"16b15f88bbd2e0a22d1d0084b8b7080f2003ea83eab1a00f80d8c18446c9c1b6224f17aa09eaf167717ca4f355bb6dc94356e037edf3adf6735a86fc3741f5231b" validate:"optional"`
}

type UpdateNameReq struct {
	Name *string `json:"name"`
}

type UpdateCountryCodeReq struct {
	CountryCode *string `json:"countryCode"`
}

type UpdateImageURLReq struct {
	ImageURL *string `json:"imageUrl"`
}

func (reg *RegisterUserDevice) Validate() error {
	return validation.ValidateStruct(reg,
		validation.Field(&reg.DeviceDefinitionID, validation.Required),
		validation.Field(&reg.CountryCode, validation.Required, validation.Length(3, 3)),
	)
}

func (reg *RegisterUserDeviceVIN) Validate() error {
	return validation.ValidateStruct(reg,
		validation.Field(&reg.VIN, validation.Required, validation.Length(17, 17)),
		validation.Field(&reg.CountryCode, validation.Required, validation.Length(3, 3)),
	)
}

func (reg *RegisterUserDeviceSmartcar) Validate() error {
	return validation.ValidateStruct(reg,
		validation.Field(&reg.Code, validation.Required),
		validation.Field(&reg.RedirectURI, validation.Required),
		validation.Field(&reg.CountryCode, validation.Required, validation.Length(3, 3)),
	)
}

func (reg *AdminRegisterUserDevice) Validate() error {
	return validation.ValidateStruct(reg,
		validation.Field(&reg.RegisterUserDevice),
		validation.Field(&reg.ID, validation.Required, validation.Length(27, 27), is.Alphanumeric),
	)
}

func (u *UpdateVINReq) validate() error {

	validateLengthAndChars := validation.ValidateStruct(u,
		// vin must be 17 characters in length, alphanumeric
		validation.Field(&u.VIN, validation.Required, validation.Match(regexp.MustCompile("^[A-Z0-9]{17}$"))),
	)
	if validateLengthAndChars != nil {
		return validateLengthAndChars
	}

	return nil
}

func (u *UpdateNameReq) validate() error {

	return validation.ValidateStruct(u,
		// name must be between 1 and 40 alphanumeric characters in length
		// NOTE: this captures characters in the latin/ chinese/ cyrillic alphabet but doesn't work as well for thai or arabic
		validation.Field(&u.Name, validation.Required, validation.Match(regexp.MustCompile(`^[\p{L}\p{N}\p{M}# ,’.@!$'":_/()+-]{1,40}$`))),
		// cannot start with space
		validation.Field(&u.Name, validation.Required, validation.Match(regexp.MustCompile(`^[^\s]`))),
		// cannot end with space
		validation.Field(&u.Name, validation.Required, validation.Match(regexp.MustCompile(`.+[^\s]$|[^\s]$`))),
	)
}

// sortByJSONFieldMostRecent Sort user device data so the latest that has the specified field is first
func sortByJSONFieldMostRecent(udd models.UserDeviceDatumSlice, field string) {
	sort.Slice(udd, func(i, j int) bool {
		fpri := gjson.GetBytes(udd[i].Data.JSON, field)
		fprj := gjson.GetBytes(udd[j].Data.JSON, field)
		if fpri.Exists() && !fprj.Exists() {
			return true
		} else if !fpri.Exists() && fprj.Exists() {
			return false
		}
		return udd[i].UpdatedAt.After(udd[j].UpdatedAt)
	})
}

// PrivilegeUser represents set of privileges I've granted to a user
type PrivilegeUser struct {
	Address    string      `json:"address"`
	Privileges []Privilege `json:"privileges"`
}

type MyDevicesResp struct {
	UserDevices   []UserDeviceFull `json:"userDevices"`
	SharedDevices []UserDeviceFull `json:"sharedDevices"`
}

// UserDeviceFull represents object user's see on frontend for listing of their devices
type UserDeviceFull struct {
	ID               string                        `json:"id"`
	VIN              *string                       `json:"vin"`
	VINConfirmed     bool                          `json:"vinConfirmed"`
	Name             *string                       `json:"name"`
	CustomImageURL   *string                       `json:"customImageUrl"`
	DeviceDefinition services.DeviceDefinition     `json:"deviceDefinition"`
	CountryCode      *string                       `json:"countryCode"`
	Integrations     []UserDeviceIntegrationStatus `json:"integrations"`
	Metadata         services.UserDeviceMetadata   `json:"metadata"`
	NFT              *NFTData                      `json:"nft,omitempty"`
	OptedInAt        *time.Time                    `json:"optedInAt"`
	PrivilegeUsers   []PrivilegeUser               `json:"privilegedUsers"`
}

type NFTData struct {
	TokenID *big.Int `json:"tokenId,omitempty" swaggertype:"number" example:"37"`
	// OwnerAddress is the Ethereum address of the NFT owner.
	OwnerAddress *common.Address `json:"ownerAddress,omitempty"`
	TokenURI     string          `json:"tokenUri,omitempty" example:"https://nft.dimo.zone/37"`
	// TxHash is the hash of the minting transaction.
	TxHash *string `json:"txHash,omitempty" example:"0x30bce3da6985897224b29a0fe064fd2b426bb85a394cc09efe823b5c83326a8e"`
	// Status is the minting status of the NFT.
	Status string `json:"status" enums:"Unstarted,Submitted,Mined,Confirmed" example:"Confirmed"`
}
