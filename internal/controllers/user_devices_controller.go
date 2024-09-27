package controllers

import (
	"bytes"
	"cmp"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	sig2 "github.com/DIMO-Network/devices-api/internal/contracts/signature"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/ipfs"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/redis"
	"github.com/IBM/sarama"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	signer "github.com/ethereum/go-ethereum/signer/core/apitypes"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	smartcar "github.com/smartcar/go-sdk"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
)

var _ = signer.TypedData{} // Use this package so that the swag command doesn't throw a fit.

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
	autoPiIngestRegistrar     services.IngestRegistrar
	s3                        *s3.Client
	producer                  sarama.SyncProducer
	deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	redisCache                redis.CacheService
	openAI                    services.OpenAI
	usersClient               pb.UserServiceClient
	deviceDataSvc             services.DeviceDataService
	NATSSvc                   *services.NATSService
	wallet                    services.SyntheticWalletInstanceService
	userDeviceSvc             services.UserDeviceService
	teslaFleetAPISvc          services.TeslaFleetAPIService
	ipfsSvc                   *ipfs.IPFS
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
func NewUserDevicesController(settings *config.Settings,
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
	autoPiIngestRegistrar services.IngestRegistrar,
	deviceDefinitionRegistrar services.DeviceDefinitionRegistrar,
	producer sarama.SyncProducer,
	s3NFTClient *s3.Client,
	cache redis.CacheService,
	openAI services.OpenAI,
	usersClient pb.UserServiceClient,
	deviceDataSvc services.DeviceDataService,
	natsSvc *services.NATSService,
	wallet services.SyntheticWalletInstanceService,
	userDeviceSvc services.UserDeviceService,
	teslaFleetAPISvc services.TeslaFleetAPIService,
	ipfsSvc *ipfs.IPFS,
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
		autoPiIngestRegistrar:     autoPiIngestRegistrar,
		s3:                        s3NFTClient,
		producer:                  producer,
		deviceDefinitionRegistrar: deviceDefinitionRegistrar,
		redisCache:                cache,
		openAI:                    openAI,
		usersClient:               usersClient,
		deviceDataSvc:             deviceDataSvc,
		NATSSvc:                   natsSvc,
		wallet:                    wallet,
		userDeviceSvc:             userDeviceSvc,
		teslaFleetAPISvc:          teslaFleetAPISvc,
		ipfsSvc:                   ipfsSvc,
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
		return nil, shared.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+ddIDs[0])
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
		return nil, shared.GrpcErrorToFiber(err, "failed to get integrations")
	}

	for _, d := range devices {
		deviceDefinition, err := filterDeviceDefinition(d.DeviceDefinitionID, deviceDefinitionResponse)
		if err != nil {
			return nil, fmt.Errorf("user device %s has unknown device definition %s", d.ID, d.DeviceDefinitionID)
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

		var sdStat *SyntheticDeviceStatus

		var nft *VehicleNFTData
		pu := []PrivilegeUser{}

		if !d.TokenID.IsZero() || d.R.MintRequest != nil {
			nft = &VehicleNFTData{}

			if !d.TokenID.IsZero() {
				nft.TokenID = d.TokenID.Int(nil)

				nft.TokenURI = fmt.Sprintf("%s/v1/vehicle/%s", udc.Settings.DeploymentBaseURL, nft.TokenID)

				addr := common.BytesToAddress(d.OwnerAddress.Bytes)
				nft.OwnerAddress = &addr

				// NFT Privileges
				udp, err := models.NFTPrivileges(
					models.NFTPrivilegeWhere.TokenID.EQ(types.Decimal(d.TokenID)),
					models.NFTPrivilegeWhere.Expiry.GT(time.Now()),
					models.NFTPrivilegeWhere.ContractAddress.EQ(common.FromHex(udc.Settings.VehicleNFTAddress)),
					qm.OrderBy(models.NFTPrivilegeColumns.UpdatedAt+" DESC, "+models.NFTPrivilegeColumns.Privilege+" ASC"),
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

				slices.SortFunc(pu, func(a, b PrivilegeUser) int {
					return cmp.Compare(a.Address, b.Address)
				})
			}

			if mtr := d.R.MintRequest; mtr != nil {
				nft.Status = mtr.Status
				nft.FailureReason = mtr.FailureReason.Ptr()

				if mtr.Hash.Valid {
					hash := hexutil.Encode(mtr.Hash.Bytes)
					nft.TxHash = &hash
				}
			}

			if mtr := d.R.BurnRequest; mtr != nil {
				var maybeHash *string

				if mtr.Hash.Valid {
					hash := common.BytesToHash(mtr.Hash.Bytes).Hex()
					maybeHash = &hash
				}

				nft.BurnTransaction = &TransactionStatus{
					Status:        mtr.Status,
					Hash:          maybeHash,
					CreatedAt:     mtr.CreatedAt,
					UpdatedAt:     mtr.UpdatedAt,
					FailureReason: mtr.FailureReason.Ptr(),
				}
			}
		}

		if sd := d.R.VehicleTokenSyntheticDevice; sd != nil {
			ii, _ := sd.IntegrationTokenID.Uint64()
			mtr := sd.R.MintRequest
			sdStat = &SyntheticDeviceStatus{
				IntegrationID: ii,
				Status:        mtr.Status,
				FailureReason: mtr.FailureReason.Ptr(),
			}
			if mtr.Hash.Valid {
				h := hexutil.Encode(mtr.Hash.Bytes)
				sdStat.TxHash = &h
			}

			if !sd.TokenID.IsZero() {
				sdStat.TokenID = sd.TokenID.Int(nil)
				a := common.BytesToAddress(sd.WalletAddress)
				sdStat.Address = &a
			}

			if mtr := sd.R.BurnRequest; mtr != nil {
				var maybeHash *string

				if mtr.Hash.Valid {
					hash := common.BytesToHash(mtr.Hash.Bytes).Hex()
					maybeHash = &hash
				}

				sdStat.BurnTransaction = &TransactionStatus{
					Status:        mtr.Status,
					Hash:          maybeHash,
					CreatedAt:     mtr.CreatedAt,
					UpdatedAt:     mtr.UpdatedAt,
					FailureReason: mtr.FailureReason.Ptr(),
				}
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
			Integrations:     NewUserDeviceIntegrationStatusesFromDatabase(d.R.UserDeviceAPIIntegrations, integrations, sdStat),
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
	userID := helpers.GetUserID(c)
	user, err := udc.usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		return helpers.ErrorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	var query []qm.QueryMod

	if user.EthereumAddress == nil {
		query = []qm.QueryMod{
			models.UserDeviceWhere.UserID.EQ(userID),
		}
	} else {
		query = []qm.QueryMod{
			models.UserDeviceWhere.UserID.EQ(userID),
			qm.Or2(models.UserDeviceWhere.OwnerAddress.EQ(null.BytesFrom(common.HexToAddress(*user.EthereumAddress).Bytes()))),
		}
	}

	query = append(query,
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
		qm.Load(models.UserDeviceRels.MintRequest),
		qm.Load(models.UserDeviceRels.BurnRequest),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleTokenSyntheticDevice, models.SyntheticDeviceRels.MintRequest)),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleTokenSyntheticDevice, models.SyntheticDeviceRels.BurnRequest)),
		qm.OrderBy(models.UserDeviceColumns.CreatedAt+" DESC"))

	devices, err := models.UserDevices(query...).All(c.Context(), udc.DBS().Reader)
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

	user, err := udc.usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		udc.log.Err(err).Msg("Couldn't retrieve user record.")
		return opaqueInternalError
	}

	var sharedVehicles []*models.UserDevice

	if user.EthereumAddress != nil {
		// This is N+1 hell.
		userAddr := common.HexToAddress(*user.EthereumAddress)

		privs, err := models.NFTPrivileges(
			models.NFTPrivilegeWhere.ContractAddress.EQ(common.FromHex(udc.Settings.VehicleNFTAddress)),
			models.NFTPrivilegeWhere.UserAddress.EQ(userAddr.Bytes()),
			models.NFTPrivilegeWhere.Expiry.GT(time.Now()),
			qm.OrderBy(models.NFTPrivilegeColumns.UpdatedAt+" DESC, "+models.NFTPrivilegeColumns.TokenID+" DESC"),
		).All(c.Context(), udc.DBS().Reader)
		if err != nil {
			return err
		}

		var processedIDs []types.Decimal

	PrivLoop:
		for _, priv := range privs {
			for _, tok := range processedIDs {
				if tok.Cmp(priv.TokenID.Big) == 0 {
					continue PrivLoop
				}
			}

			processedIDs = append(processedIDs, priv.TokenID)

			ud, err := models.UserDevices(
				models.UserDeviceWhere.TokenID.EQ(types.NewNullDecimal(priv.TokenID.Big)),
				qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
				qm.Load(models.UserDeviceRels.MintRequest),
				qm.Load(models.UserDeviceRels.BurnRequest),
				qm.Load(qm.Rels(models.UserDeviceRels.VehicleTokenSyntheticDevice, models.SyntheticDeviceRels.MintRequest)),
				qm.Load(qm.Rels(models.UserDeviceRels.VehicleTokenSyntheticDevice, models.SyntheticDeviceRels.BurnRequest)),
			).One(c.Context(), udc.DBS().Reader)
			if err != nil {
				if err == sql.ErrNoRows {
					udc.log.Warn().Msgf("User %s has privileges on a vehicle %d of which we have no record.", userAddr, priv.TokenID)
					continue
				}
				return err
			}

			sharedVehicles = append(sharedVehicles, ud)
		}
	}

	apiSharedDevices, err := udc.dbDevicesToDisplay(c.Context(), sharedVehicles)
	if err != nil {
		return err
	}

	return c.JSON(MyDevicesResp{SharedDevices: apiSharedDevices})
}

func NewUserDeviceIntegrationStatusesFromDatabase(udis []*models.UserDeviceAPIIntegration, integrations []*ddgrpc.Integration, sdStat *SyntheticDeviceStatus) []UserDeviceIntegrationStatus {
	out := make([]UserDeviceIntegrationStatus, len(udis))

	for i, udi := range udis {
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

				if sdStat != nil && integration.TokenId == sdStat.IntegrationID {
					out[i].Mint = sdStat
				}
				break
			}
		}
	}

	return out
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

	udFull, err := udc.createUserDevice(c.Context(), *reg.DeviceDefinitionID, "", reg.CountryCode, userID, nil, nil)
	if err != nil {
		return shared.GrpcErrorToFiber(err, "")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"userDevice": udFull,
	})
}

var opaqueInternalError = fiber.NewError(fiber.StatusInternalServerError, "Internal error.")

// RegisterDeviceForUserFromVIN godoc
// @Description adds a device to a user by decoding a VIN. If cannot decode returns 424 or 500 if error. Can optionally include the can bus protocol.
// @Tags        user-devices
// @Produce     json
// @Accept      json
// @Param       user_device body controllers.RegisterUserDeviceVIN true "add device to user. VIN is required and so is country"
// @Security    ApiKeyAuth
// @Failure		400 "validation failure"
// @Failure		424 "unable to decode VIN"
// @Failure		500 "server error, dependency error"
// @Success     201 {object} controllers.UserDeviceFull
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
	country := constants.FindCountry(strings.ToUpper(reg.CountryCode))
	if country == nil {
		return fiber.NewError(fiber.StatusBadRequest, "unsupported or invalid country: "+reg.CountryCode)
	}
	vin := strings.ToUpper(reg.VIN)

	integration, err := udc.DeviceDefIntSvc.GetAutoPiIntegration(c.Context())
	if err != nil {
		udc.log.Err(err).Msg("failed to get autopi integration")
		return err
	}
	localLog := udc.log.With().Str("userId", userID).Str("integrationId", integration.Id).
		Str("countryCode", country.Alpha3).Str("vin", vin).Str("handler", "RegisterDeviceForUserFromVIN").
		Logger()

	deviceDefinitionID := ""

	// check if VIN already exists
	existingUD, err := models.UserDevices(models.UserDeviceWhere.VinIdentifier.EQ(null.StringFrom(vin)),
		models.UserDeviceWhere.VinConfirmed.EQ(true)).One(c.Context(), udc.DBS().Reader)
	if err != nil && !errors.Is(sql.ErrNoRows, err) {
		return err
	}
	var udFull *UserDeviceFull
	if existingUD != nil {
		if existingUD.UserID != userID {
			return fiber.NewError(fiber.StatusConflict, "VIN already exists for a different user: "+reg.VIN)
		}
		deviceDefinitionID = existingUD.DeviceDefinitionID
		// shortcut process, just use the already registered UD
		dd, err := udc.DeviceDefSvc.GetDeviceDefinitionByID(c.Context(), deviceDefinitionID)
		if err != nil {
			return err
		}
		udFull, err = builUserDeviceFull(existingUD, dd, reg.CountryCode)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	} else {
		// decode VIN with grpc call
		decodeVIN, err := udc.DeviceDefSvc.DecodeVIN(c.Context(), vin, "", 0, reg.CountryCode)
		if err != nil {
			localLog.Err(err).Msg("unable to decode vin for customer request to create vehicle")
			return shared.GrpcErrorToFiber(err, "unable to decode vin: "+vin)
		}
		if len(decodeVIN.DeviceDefinitionId) == 0 {
			localLog.Warn().Msg("unable to decode vin for customer request to create vehicle")
			return fiber.NewError(fiber.StatusFailedDependency, "unable to decode vin")
		}
		deviceDefinitionID = decodeVIN.DeviceDefinitionId

		udFull, err = udc.createUserDevice(c.Context(), deviceDefinitionID, decodeVIN.DeviceStyleId, reg.CountryCode, userID, &vin, &reg.CANProtocol)
		if err != nil {
			localLog.Err(err).Send()
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	}

	// create device_integration record in definitions just in case. If we got the VIN normally means Mobile App able to decode.
	_, err = udc.DeviceDefIntSvc.CreateDeviceDefinitionIntegration(c.Context(), integration.Id, deviceDefinitionID, country.Region)
	if err != nil {
		localLog.Warn().Err(err).Msgf("failed to add device_integration for autopi and dd_id: %s", deviceDefinitionID)
	}

	// request valuation
	if udc.Settings.IsProduction() {
		tokenID := int64(0)
		if udFull.NFT != nil {
			tokenID = udFull.NFT.TokenID.Int64()
		}
		udc.requestValuation(vin, udFull.ID, tokenID)
		udc.requestInstantOffer(udFull.ID, tokenID)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"userDevice": udFull,
	})
}

func (udc *UserDevicesController) requestValuation(vin string, userDeviceID string, tokenID int64) {
	message := services.ValuationDecodeCommand{
		VIN:          vin,
		UserDeviceID: userDeviceID,
		TokenID:      tokenID,
	}
	messageBytes, err := json.Marshal(message)

	if err != nil {
		udc.log.Err(err).Msg("Failed to marshal message.")
		return
	}

	pubAck, err := udc.NATSSvc.JetStream.Publish(udc.NATSSvc.ValuationSubject, messageBytes)
	if err != nil {
		udc.log.Err(err).Msg("failed to publish to NATS")
		return
	}

	udc.log.Info().Str("user_device_id", userDeviceID).Msgf("published valuation request to NATS with Ack: %+v", pubAck)
}

func (udc *UserDevicesController) requestInstantOffer(userDeviceID string, tokenID int64) {
	message := services.OfferRequest{
		UserDeviceID: userDeviceID,
		TokenID:      tokenID,
	}
	messageBytes, err := json.Marshal(message)

	if err != nil {
		udc.log.Err(err).Msg("Failed to marshal message.")
		return
	}

	pubAck, err := udc.NATSSvc.JetStream.Publish(udc.NATSSvc.OfferSubject, messageBytes)
	if err != nil {
		udc.log.Err(err).Msg("failed to publish to NATS")
		return
	}

	udc.log.Info().Str("user_device_id", userDeviceID).Msgf("published instant offer request to NATS with Ack: %+v", pubAck)
}

// RegisterDeviceForUserFromSmartcar godoc
// @Description adds a device to a user by decoding VIN from Smartcar. If cannot decode returns 424 or 500 if error.
// @Description If the user device already exists from a different integration, for the same user, this will return a 200 with the full user device object
// @Tags        user-devices
// @Produce     json
// @Accept      json
// @Param       user_device body controllers.RegisterUserDeviceSmartcar true "add device to user. all fields required"
// @Security    ApiKeyAuth
// @Failure		400 "validation failure"
// @Failure		424 "unable to decode VIN"
// @Failure		409 "VIN already exists either for different a user"
// @Failure		500 "server error, dependency error"
// @Success     201 {object} controllers.UserDeviceFull
// @Success     200 {object} controllers.UserDeviceFull
// @Security    BearerAuth
// @Router      /user/devices/fromsmartcar [post]
func (udc *UserDevicesController) RegisterDeviceForUserFromSmartcar(c *fiber.Ctx) error {
	const smartCarIntegrationID = "22N2xaPOq2WW2gAHBHd0Ikn4Zob"
	userID := helpers.GetUserID(c)
	localLog := udc.log.With().Str("userId", userID).Str("integrationId", smartCarIntegrationID).
		Str("handler", "RegisterDeviceForUserFromSmartcar").Logger()

	reg := &RegisterUserDeviceSmartcar{}
	if err := c.BodyParser(reg); err != nil {
		// Return status 400 and error message.
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := reg.Validate(); err != nil {
		localLog.Error().Msgf("Smartcar device creation input invalid, code %q.", reg.Code)
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	country := constants.FindCountry(reg.CountryCode)
	if country == nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid countryCode field or country not supported: %s", reg.CountryCode))
	}
	localLog = localLog.With().Str("countryCode", reg.CountryCode).Str("region", country.Region).Logger()

	// call SC api with stuff and get VIN
	token, err := udc.smartcarClient.ExchangeCode(c.Context(), reg.Code, reg.RedirectURI)
	if err != nil {
		var scErr *services.SmartcarError
		if errors.As(err, &scErr) {
			localLog.Error().Msgf("Failed exchanging Authorization code. Status code %d, request id %s, and body `%s`.", scErr.Code, scErr.RequestID, string(scErr.Body))
		} else {
			localLog.Err(err).Msg("Failed to exchange authorization code with Smartcar.")
		}
		// This may not be the user's fault, but 400 for now.
		return fiber.NewError(fiber.StatusBadRequest, "Failed to exchange authorization code with Smartcar.")
	}

	externalID, err := udc.smartcarClient.GetExternalID(c.Context(), token.Access)
	if err != nil {
		localLog.Err(err).Msg("Failed to retrieve vehicle ID from Smartcar.")
		return fiber.NewError(fiber.StatusInternalServerError, smartcarCallErr.Error())
	}
	vin, err := udc.smartcarClient.GetVIN(c.Context(), token.Access, externalID)
	if err != nil {
		localLog.Err(err).Msg("Failed to retrieve VIN from Smartcar.")
		return fiber.NewError(fiber.StatusInternalServerError, smartcarCallErr.Error())
	}
	if len(vin) != 17 {
		localLog.Error().Msgf("invalid VIN returned from smartcar: %s", vin)
		return fiber.NewError(fiber.StatusInternalServerError, smartcarCallErr.Error())
	}
	localLog = localLog.With().Str("vin", vin).Logger()

	// duplicate vin check, only in prod. If same user has already registered this car, and are eg. trying to add autopi, client should not call this endpoint
	isSameUserConflict := false
	var existingUd *models.UserDevice
	if udc.Settings.IsProduction() {
		existingUd, err = models.UserDevices(
			models.UserDeviceWhere.VinIdentifier.EQ(null.StringFrom(vin)),
			models.UserDeviceWhere.VinConfirmed.EQ(true),
		).One(c.Context(), udc.DBS().Reader)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			localLog.Err(err).Msg("Failed to search for VIN conflicts.")
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		if existingUd != nil {
			if existingUd.UserID == userID {
				isSameUserConflict = true
			} else {
				localLog.Error().Msgf("failed to create UD from smartcar because VIN already in use. conflict vin user_id: %s", existingUd.UserID)
				return fiber.NewError(fiber.StatusConflict, fmt.Sprintf("VIN %s in use by a previously connected device", vin))
			}
		}
	}
	// persist token in redis, encrypted, for next step
	jsonToken, err := json.Marshal(token)
	if err != nil {
		return errors.Wrap(err, "failed to marshall sc token")
	}
	encToken, err := udc.cipher.Encrypt(string(jsonToken))
	if err != nil {
		return errors.Wrap(err, "failed to encrypt smartcar token json")
	}
	udc.redisCache.Set(c.Context(), buildSmartcarTokenKey(vin, userID), encToken, time.Hour*2)

	if isSameUserConflict {
		dd, err2 := udc.DeviceDefSvc.GetDeviceDefinitionByID(c.Context(), existingUd.DeviceDefinitionID)
		if err2 != nil {
			return shared.GrpcErrorToFiber(err2, fmt.Sprintf("error querying for device definition id: %s ", existingUd.DeviceDefinitionID))
		}
		udFull, err := builUserDeviceFull(existingUd, dd, reg.CountryCode)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"userDevice": udFull,
		})
	}
	// get info from smartcar, fine if fails
	info, err := udc.smartcarClient.GetInfo(c.Context(), token.Access, externalID)
	if err != nil {
		localLog.Warn().Err(err).Msg("unable to get info from smartcar")
	}
	if info == nil {
		info = &smartcar.Info{}
	}

	// decode VIN with grpc call, including any possible smartcar known info
	decodeVIN, err := udc.DeviceDefSvc.DecodeVIN(c.Context(), vin, info.Model, info.Year, reg.CountryCode)
	if err != nil {
		localLog.Err(err).Msg("unable to decode vin for customer request to create vehicle")
		return shared.GrpcErrorToFiber(err, "unable to decode vin: "+vin)
	}

	// in case err is nil but we don't get a valid decode
	if len(decodeVIN.DeviceDefinitionId) == 0 {
		localLog.Err(err).Msg("unable to decode vin for customer request to create vehicle")
		return fiber.NewError(fiber.StatusFailedDependency, "failed to decode vin")
	}
	// attach smartcar integration to device definition
	_, err = udc.DeviceDefIntSvc.CreateDeviceDefinitionIntegration(c.Context(), smartCarIntegrationID,
		decodeVIN.DeviceDefinitionId, country.Region)
	if err != nil {
		localLog.Err(err).
			Msgf("unable to CreateDeviceDefinitionIntegration for dd_id: %s", decodeVIN.DeviceDefinitionId)
		return fiber.NewError(fiber.StatusConflict, errors.Wrap(err, "unable to attach smartcar integration to device definition id").Error())
	}

	// attach device def to user
	udFull, err := udc.createUserDevice(c.Context(), decodeVIN.DeviceDefinitionId, decodeVIN.DeviceStyleId, reg.CountryCode, userID, &vin, nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if udc.Settings.IsProduction() {
		tokenID := int64(0)
		if udFull.NFT != nil {
			tokenID = udFull.NFT.TokenID.Int64()
		}
		udc.requestValuation(vin, udFull.ID, tokenID)
		udc.requestInstantOffer(udFull.ID, tokenID)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"userDevice": udFull,
	})
}

func buildSmartcarTokenKey(vin, userID string) string {
	return fmt.Sprintf("sc-temp-tok-%s-%s", vin, userID)
}

func (udc *UserDevicesController) createUserDevice(ctx context.Context, deviceDefID, styleID, countryCode, userID string, vin, canProtocol *string) (*UserDeviceFull, error) {
	ud, dd, err := udc.userDeviceSvc.CreateUserDevice(ctx, deviceDefID, styleID, countryCode, userID, vin, canProtocol, false)
	if err != nil {
		if errors.Is(err, services.ErrEmailUnverified) {
			return nil, fiber.NewError(fiber.StatusBadRequest,
				"Your email has not been verified. Please check your email for the DIMO verification email.")
		}
		return nil, err
	}

	return builUserDeviceFull(ud, dd, countryCode)
}

func builUserDeviceFull(ud *models.UserDevice, dd *ddgrpc.GetDeviceDefinitionItemResponse, countryCode string) (*UserDeviceFull, error) {
	ddNice, err := NewDeviceDefinitionFromGRPC(dd)
	if err != nil {
		return nil, err
	}
	// Baby the frontend.
	for i := range ddNice.CompatibleIntegrations {
		ddNice.CompatibleIntegrations[i].Country = countryCode
	}

	udMd := &services.UserDeviceMetadata{}
	_ = ud.Metadata.Unmarshal(udMd)

	return &UserDeviceFull{
		ID:               ud.ID,
		VIN:              ud.VinIdentifier.Ptr(),
		VINConfirmed:     ud.VinConfirmed,
		Name:             ud.Name.Ptr(),
		CustomImageURL:   ud.CustomImageURL.Ptr(),
		DeviceDefinition: ddNice,
		CountryCode:      ud.CountryCode.Ptr(),
		Metadata:         *udMd,
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

	logger := helpers.GetLogger(c, udc.log)

	userDevice, err := models.UserDevices(
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

	logger.Debug().Msg("Opted into data-sharing.")

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

	logger := helpers.GetLogger(c, udc.log).With().Str("route", c.Route().Name).Logger()

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

		hash := crypto.Keccak256(vinByte)

		recAddr, err := helpers.Ecrecover(hash, sig)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Couldn't recover signer address.")
		}

		found, err := models.AftermarketDevices(
			models.AftermarketDeviceWhere.EthereumAddress.EQ(recAddr.Bytes()),
		).Exists(c.Context(), udc.DBS().Reader)
		if err != nil {
			return err
		}
		if !found {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("VIN signature author %s does not match any known aftermarket device.", recAddr))
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
		if udc.Settings.IsProduction() && existing {
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

	if userDevice.CountryCode.Valid {
		if err := udc.updatePowerTrain(c.Context(), userDevice); err != nil {
			logger.Err(err).Msg("Failed to update powertrain type.")
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}

const (
	PowerTrainTypeKey = "powertrain_type"
)

// todo revisit this depending on what observe with below log message
func (udc *UserDevicesController) updatePowerTrain(ctx context.Context, userDevice *models.UserDevice) error {
	md := new(services.UserDeviceMetadata)
	if err := userDevice.Metadata.Unmarshal(md); err != nil {
		return err
	}
	resp, err := udc.DeviceDefSvc.GetDeviceDefinitionByID(ctx, userDevice.DeviceDefinitionID)
	if err != nil {
		return err
	}

	if len(resp.DeviceAttributes) > 0 {
		// Find device attribute (powertrain_type)
		for _, item := range resp.DeviceAttributes {
			if item.Name == PowerTrainTypeKey {
				powertrainType := services.ConvertPowerTrainStringToPowertrain(item.Value)
				md.PowertrainType = &powertrainType
				break
			}
		}
	}

	if err := userDevice.Metadata.Marshal(md); err != nil {
		return err
	}
	if _, err := userDevice.Update(ctx, udc.DBS().Writer, boil.Infer()); err != nil {
		return err
	}

	return nil
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

	userDevice, err := models.UserDevices(models.UserDeviceWhere.ID.EQ(udi)).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
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

// DeleteUserDevice godoc
// @Description delete the user device record (hard delete)
// @Tags        user-devices
// @Param       userDeviceID path string true "device id"
// @Success     204
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID} [delete]
func (udc *UserDevicesController) DeleteUserDevice(c *fiber.Ctx) error {
	logger := helpers.GetLogger(c, udc.log)

	udi := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)

	tx, err := udc.DBS().Writer.BeginTx(c.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
		qm.Load(models.UserDeviceRels.MintRequest),
		qm.Load(qm.Rels(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationRels.SerialAftermarketDevice)),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Device not found.")
		}
		return err
	}

	// if vehicle minted, user must delete by burning
	if !userDevice.TokenID.IsZero() {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Vehicle was minted with token id %d. Burn this NFT to delete the vehicle.", userDevice.TokenID))
	}
	if userDevice.R.MintRequest != nil && userDevice.R.MintRequest.Status != models.MetaTransactionRequestStatusFailed {
		return fiber.NewError(fiber.StatusBadRequest, "Vehicle minting in progress. Burn the resulting NFT in order to delete this vehicle.")
	}

	dd, err := udc.DeviceDefSvc.GetDeviceDefinitionByID(c.Context(), userDevice.DeviceDefinitionID)
	if err != nil {
		return shared.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+userDevice.DeviceDefinitionID)
	}
	autopiDeviceID := ""

	for _, apiInteg := range userDevice.R.UserDeviceAPIIntegrations {
		if unit := apiInteg.R.SerialAftermarketDevice; unit != nil && !unit.VehicleTokenID.IsZero() {
			return fiber.NewError(fiber.StatusConflict, fmt.Sprintf("Cannot delete vehicle before unpairing aftermarket device %s on-chain.", unit.Serial))
		}
		integr, err := udc.DeviceDefSvc.GetIntegrationByID(c.Context(), apiInteg.IntegrationID)
		if err != nil {
			return err
		}
		if integr.Vendor == constants.AutoPiVendor {
			unit, _ := udc.autoPiSvc.GetDeviceByUnitID(apiInteg.Serial.String)
			if unit != nil {
				autopiDeviceID = unit.ID
			} else {
				udc.log.Warn().Msgf("failed to find autopi device with serial: %s and user device id: %s", apiInteg.Serial.String, apiInteg.UserDeviceID)
			}
		}
	}

	for _, apiInteg := range userDevice.R.UserDeviceAPIIntegrations {
		err := udc.deleteDeviceIntegration(c.Context(), userID, udi, apiInteg.IntegrationID, dd, tx)
		if err != nil {
			return err
		}
	}

	if _, err := userDevice.Delete(c.Context(), tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	if err = udc.eventService.Emit(&shared.CloudEvent[any]{
		Type:    "com.dimo.zone.device.delete",
		Subject: userID,
		Source:  "devices-api",
		Data: services.UserDeviceEvent{
			Timestamp: time.Now(),
			UserID:    userID,
			Device: services.UserDeviceEventDevice{
				ID:    udi,
				Make:  dd.Make.Name,
				Model: dd.Type.Model,
				Year:  int(dd.Type.Year),
			},
		},
	}); err != nil {
		return err
	}

	if userDevice.VinConfirmed {
		logger.Info().Msgf("Deleted vehicle with VIN %s.", userDevice.VinIdentifier.String)
	} else {
		logger.Info().Msg("Deleted vehicle.")
	}
	udc.markAutoPiUnpaired(autopiDeviceID)

	return c.SendStatus(fiber.StatusNoContent)
}

// markAutoPiUnpaired tells the AP cloud this device has been marked as unpaired in their metadata, only if id is not empty
func (udc *UserDevicesController) markAutoPiUnpaired(autopiDeviceID string) {
	// autopi would like it if we updated the state to unpaired for these cases
	if autopiDeviceID != "" {
		err := udc.autoPiSvc.UpdateState(autopiDeviceID, "unpaired", "", "")
		if err != nil {
			udc.log.Err(err).Msgf("failed to update autopi device state with device id: %s", autopiDeviceID)
		}
	}
}

const imageURIattribute = "ImageURI"

// GetMintDevice godoc
// @Description Returns the data the user must sign in order to mint this device.
// @Tags        user-devices
// @Param       userDeviceID path     string true "user device ID"
// @Success     200          {object} signer.TypedData
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/commands/mint [get]
func (udc *UserDevicesController) GetMintDevice(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	userDeviceID := c.Params("userDeviceID")

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(models.UserDeviceRels.MintRequest),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "No vehicle with that id found.")
	}

	mvs, dd, err := udc.checkVehicleMint(c.Context(), userID, userDevice)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
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

	mvdds := registry.MintVehicleWithDeviceDefinitionSign{
		ManufacturerNode:   mvs.ManufacturerNode,
		Owner:              mvs.Owner,
		Attributes:         mvs.Attributes,
		Infos:              mvs.Infos,
		DeviceDefinitionID: dd.NameSlug,
	}

	return c.JSON(client.GetPayload(&mvdds))
}

var erc1271magicValue = [4]byte{0x16, 0x26, 0xba, 0x7e}

// PostMintDevice godoc
// @Description Sends a mint device request to the blockchain
// @Tags        user-devices
// @Param       userDeviceID path string                  true "user device ID"
// @Param       mintRequest  body controllers.VehicleMintRequest true "Signature and NFT data"
// @Success     200
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/commands/mint [post]
func (udc *UserDevicesController) PostMintDevice(c *fiber.Ctx) error {
	userDeviceID := c.Params("userDeviceID")
	userID := helpers.GetUserID(c)

	logger := helpers.GetLogger(c, udc.log)

	tx, err := udc.DBS().Writer.BeginTx(c.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(models.UserDeviceRels.MintRequest),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
	).One(c.Context(), tx)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "No device with that id found.")
		}
		return err
	}

	// This actually makes no database calls!
	mvs, dd, err := udc.checkVehicleMint(c.Context(), userID, userDevice)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	var mr VehicleMintRequest
	if err := c.BodyParser(&mr); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request body.")
	}

	// This may not be there, but if it is we should delete it.
	imageData := strings.TrimPrefix(mr.ImageData, "data:image/png;base64,")

	image, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Primary image not properly base64-encoded.")
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

	mvdds := registry.MintVehicleWithDeviceDefinitionSign{
		ManufacturerNode:   mvs.ManufacturerNode,
		Owner:              mvs.Owner,
		Attributes:         mvs.Attributes,
		Infos:              mvs.Infos,
		DeviceDefinitionID: dd.NameSlug,
	}

	logger.Info().
		Interface("httpRequestBody", mr).
		Interface("client", client).Interface("mintVehicleWithDeviceDefinitionSign", mvdds).
		Interface("typedData", client.GetPayload(&mvdds)).
		Msg("Got request.")

	hash, err := client.Hash(&mvdds)
	if err != nil {
		return opaqueInternalError
	}

	sigBytes := common.FromHex(mr.Signature)

	recAddr, origErr := helpers.Ecrecover(hash, sigBytes)
	if origErr != nil || recAddr != mvs.Owner {
		ethClient, err := ethclient.Dial(udc.Settings.MainRPCURL)
		if err != nil {
			return err
		}

		sigCon, err := sig2.NewErc1271(mvs.Owner, ethClient)
		if err != nil {
			return err
		}

		ret, err := sigCon.IsValidSignature(nil, common.BytesToHash(hash), sigBytes)
		if err != nil {
			return err
		}

		if ret != erc1271magicValue {
			return fiber.NewError(fiber.StatusBadRequest, "Could not verify ERC-1271 signature.")
		}

		return fiber.NewError(fiber.StatusInternalServerError, "You gave the right EIP-1271 signature, but we're not ready for this yet.")
	}

	requestID := ksuid.New().String()

	if len(image) == 0 {
		if !userDevice.IpfsImageCid.Valid {
			return fiber.NewError(fiber.StatusBadRequest, "No image in request body and none assigned previously.")
		}
	} else {
		if userDevice.IpfsImageCid.Valid {
			logger.Warn().Msg("Image provided in request body, but also one assigned previously.")
		}
		cid, err := udc.ipfsSvc.UploadImage(c.Context(), imageData)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Failed to upload image to IPFS.")
		}

		userDevice.IpfsImageCid = null.StringFrom(cid)
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

	mtr := models.MetaTransactionRequest{
		ID:     requestID,
		Status: models.MetaTransactionRequestStatusUnsubmitted,
	}
	if err = mtr.Insert(c.Context(), tx, boil.Infer()); err != nil {
		return err
	}

	userDevice.MintRequestID = null.StringFrom(requestID)
	if _, err = userDevice.Update(c.Context(), tx, boil.Infer()); err != nil {
		return err
	}

	if udais := userDevice.R.UserDeviceAPIIntegrations; len(udais) != 0 {
		intID := uint64(0)
		for _, udai := range udais {
			in, err := udc.DeviceDefSvc.GetIntegrationByID(c.Context(), udai.IntegrationID)
			if err != nil {
				return err
			}

			// block new kias from minting
			if (strings.ToLower(dd.Make.NameSlug) == "kia" || dd.Make.Id == "2681cSm2zmTmGHzqK3ldzoTLZIw") && in.Vendor == constants.SmartCarVendor {
				udc.log.Warn().Msgf("new kias blocked from minting with smartcar connections")
				return fiber.NewError(fiber.StatusFailedDependency, "Kia vehicles connected via smartcar cannot be manually minted for now.")
			}

			if in.Vendor == constants.TeslaVendor {
				intID = in.TokenId
				break
			} else if in.Vendor == constants.SmartCarVendor {
				intID = in.TokenId
			}
		}

		if intID != 0 {
			var seq struct {
				NextVal int `boil:"nextval"`
			}

			qry := fmt.Sprintf("SELECT nextval('%s.synthetic_devices_serial_sequence');", udc.Settings.DB.Name)
			err := queries.Raw(qry).Bind(c.Context(), tx, &seq)
			if err != nil {
				return err
			}

			childNum := seq.NextVal

			addr, err := udc.wallet.GetAddress(c.Context(), uint32(childNum))
			if err != nil {
				return err
			}

			sd := models.SyntheticDevice{
				IntegrationTokenID: types.NewDecimal(decimal.New(int64(intID), 0)),
				MintRequestID:      requestID,
				WalletChildNumber:  seq.NextVal,
				WalletAddress:      addr,
			}

			if err := sd.Insert(c.Context(), tx, boil.Infer()); err != nil {
				return err
			}

			mvss := registry.MintVehicleAndSdSign{
				IntegrationNode: new(big.Int).SetUint64(intID),
			}

			hash, err := client.Hash(&mvss)
			if err != nil {
				return err
			}

			sign, err := udc.wallet.SignHash(c.Context(), uint32(childNum), hash)
			if err != nil {
				return err
			}

			if err := tx.Commit(); err != nil {
				return err
			}

			return client.MintVehicleAndSdWithDeviceDefinitionSign(requestID, contracts.MintVehicleAndSdWithDdInput{
				ManufacturerNode:     mvs.ManufacturerNode,
				Owner:                mvs.Owner,
				DeviceDefinitionId:   dd.NameSlug,
				IntegrationNode:      new(big.Int).SetUint64(intID),
				VehicleOwnerSig:      sigBytes,
				SyntheticDeviceSig:   sign,
				SyntheticDeviceAddr:  common.BytesToAddress(addr),
				AttrInfoPairsVehicle: attrListsToAttrPairs(mvs.Attributes, mvs.Infos),
				AttrInfoPairsDevice:  []contracts.AttributeInfoPair{},
			})
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	logger.Info().Msgf("Submitted metatransaction request %s", requestID)

	return client.MintVehicleWithDeviceDefinitionSign(requestID, mvs.ManufacturerNode, mvs.Owner, dd.NameSlug, attrListsToAttrPairs(mvs.Attributes, mvs.Infos), sigBytes)
}

func attrListsToAttrPairs(attrs []string, infos []string) []contracts.AttributeInfoPair {
	out := make([]contracts.AttributeInfoPair, len(attrs))
	for i := range attrs {
		out[i] = contracts.AttributeInfoPair{
			Attribute: attrs[i],
			Info:      infos[i],
		}
	}
	return out
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

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "No device with that ID found.")
	}

	var nid NFTImageData
	if err := c.BodyParser(&nid); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request body.")
	}

	// This may not be there, but if it is we should delete it.
	imageData := strings.TrimPrefix(nid.ImageData, "data:image/png;base64,")

	image, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Primary image not properly base64-encoded.")
	}

	if len(image) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Empty image field.")
	}

	cid, err := udc.ipfsSvc.UploadImage(c.Context(), nid.ImageData)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Failed to upload image to IPFS")
	}

	userDevice.IpfsImageCid = null.StringFrom(cid)
	_, err = userDevice.Update(c.Context(), udc.DBS().Writer, boil.Whitelist(models.UserDeviceColumns.IpfsImageCid))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "failed to store IPFS CID for vehicle")
	}

	// This may not be there, but if it is we should delete it.
	imageDataTransp := strings.TrimPrefix(nid.ImageDataTransparent, "data:image/png;base64,")

	// Should be okay if empty or not provided.
	imageTransp, err := base64.StdEncoding.DecodeString(imageDataTransp)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Transparent image not properly base64-encoded.")
	}

	if len(imageTransp) != 0 {
		if userDevice.TokenID.IsZero() || !userDevice.MintRequestID.Valid {
			return fiber.NewError(fiber.StatusBadRequest, "Can't set transparent image for this vehicle.")
		}

		_, err = udc.s3.PutObject(c.Context(), &s3.PutObjectInput{
			Bucket: &udc.Settings.NFTS3Bucket,
			Key:    aws.String(userDevice.MintRequestID.String + "_transparent.png"),
			Body:   bytes.NewReader(imageTransp),
		})
		if err != nil {
			udc.log.Err(err).Msg("Failed to save transparent image to S3.")
			return opaqueInternalError
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// VehicleMintRequest contains the user's signature for the mint request as well as the
// NFT image.
type VehicleMintRequest struct {
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
	// CANProtocol is the protocol that was detected by edge-network from the autopi.
	CANProtocol string `json:"canProtocol"`
}

type RegisterUserDeviceSmartcar struct {
	// Code refers to the auth code provided by smartcar when user logs in
	Code        string `json:"code"`
	RedirectURI string `json:"redirectURI"`
	CountryCode string `json:"countryCode"`
}

type UpdateVINReq struct {
	// VIN is a vehicle identification number. At the very least, it must be
	// 17 characters in length and contain only letters and numbers.
	VIN string `json:"vin" example:"4Y1SL65848Z411439" validate:"required"`
	// Signature is the hex-encoded result of the AutoPi signing the VIN. It must
	// be present to verify the VIN.
	Signature string `json:"signature" example:"16b15f88bbd2e0a22d1d0084b8b7080f2003ea83eab1a00f80d8c18446c9c1b6224f17aa09eaf167717ca4f355bb6dc94356e037edf3adf6735a86fc3741f5231b" validate:"optional"`
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
	NFT              *VehicleNFTData               `json:"nft,omitempty"`
	OptedInAt        *time.Time                    `json:"optedInAt"`
	PrivilegeUsers   []PrivilegeUser               `json:"privilegedUsers"`
}

type VehicleNFTData struct {
	TokenID *big.Int `json:"tokenId,omitempty" swaggertype:"number" example:"37"`
	// OwnerAddress is the Ethereum address of the NFT owner.
	OwnerAddress *common.Address `json:"ownerAddress,omitempty"`
	TokenURI     string          `json:"tokenUri,omitempty" example:"https://nft.dimo.zone/37"`
	// TxHash is the hash of the minting transaction.
	TxHash *string `json:"txHash,omitempty" example:"0x30bce3da6985897224b29a0fe064fd2b426bb85a394cc09efe823b5c83326a8e"`
	// Status is the minting status of the NFT.
	Status string `json:"status,omitempty" enums:"Unstarted,Submitted,Mined,Confirmed,Failed" example:"Confirmed"`
	// FailureReason is populated if the status is "Failed" because of an on-chain revert and
	// we were able to decode the reason.
	FailureReason *string `json:"failureReason,omitempty"`
	// BurnTransaction contains the status of the vehicle burning meta-transaction, if one
	// is in flight or has failed.
	BurnTransaction *TransactionStatus `json:"burnTransaction,omitempty"`
}

type SyntheticDeviceStatus struct {
	// IntegrationID is the token id of the parent integration for this device.
	IntegrationID uint64 `json:"-"`
	// TokenID is the token id of the minted device.
	TokenID *big.Int `json:"tokenId,omitempty" swaggertype:"number" example:"15"`
	// Address is the Ethereum address of the synthetic device.
	Address *common.Address `json:"address,omitempty" swaggertype:"string" example:"0xAED7EA8035eEc47E657B34eF5D020c7005487443"`
	// TxHash is the hash of the submitted transaction.
	TxHash *string `json:"txHash,omitempty" swaggertype:"string" example:"0x30bce3da6985897224b29a0fe064fd2b426bb85a394cc09efe823b5c83326a8e"`
	// Status is the status of the minting meta-transaction.
	Status string `json:"status" enums:"Unstarted,Submitted,Mined,Confirmed,Failed" example:"Confirmed"`
	// FailureReason is populated with a human-readable error message if the status
	// is "Failed" because of an on-chain revert and we were able to decode the reason.
	FailureReason *string `json:"failureReason"`
	// BurnTransaction contains the status of the synthetic device burning meta-transaction,
	// if one is in flight or has failed.
	BurnTransaction *TransactionStatus `json:"burnTransaction,omitempty"`
}

type VINCredentialData struct {
	IssuedAt  time.Time `json:"issuedAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	Valid     bool      `json:"valid"`
	VIN       string    `json:"vin"`
}

func (udc *UserDevicesController) checkVehicleMint(ctx context.Context, userID string, userDevice *models.UserDevice) (*registry.MintVehicleSign, *ddgrpc.GetDeviceDefinitionItemResponse, error) {
	if !userDevice.TokenID.IsZero() {
		return nil, nil, fmt.Errorf("vehicle already minted with token id %d", userDevice.TokenID.Big)
	}

	if mr := userDevice.R.MintRequest; mr != nil && mr.Status != models.MetaTransactionRequestStatusFailed {
		return nil, nil, fmt.Errorf("vehicle minting already in process")
	}

	if !userDevice.VinConfirmed {
		return nil, nil, fmt.Errorf("VIN not confirmed")
	}

	dd, err := udc.DeviceDefSvc.GetDeviceDefinitionByID(ctx, userDevice.DeviceDefinitionID)
	if err != nil {
		return nil, nil, fmt.Errorf("error querying for device definition id: %s ", userDevice.DeviceDefinitionID)
	}

	if dd.Make.TokenId == 0 {
		return nil, nil, fmt.Errorf("vehicle make %q not yet minted", dd.Make.Name)
	}
	makeTokenID := new(big.Int).SetUint64(dd.Make.TokenId)

	user, err := udc.usersClient.GetUser(ctx, &pb.GetUserRequest{Id: userID})
	if err != nil {
		udc.log.Err(err).Msg("couldn't retrieve user record")
		return nil, nil, opaqueInternalError
	}

	if user.EthereumAddress == nil {
		return nil, nil, fmt.Errorf("user does not have an Ethereum address on file")
	}

	if dd.NameSlug == "" {
		return nil, nil, fmt.Errorf("invalid on-chain name slug for device definition id: %s", userDevice.DeviceDefinitionID)
	}

	mvs := &registry.MintVehicleSign{
		ManufacturerNode: makeTokenID,
		Owner:            common.HexToAddress(*user.EthereumAddress),
		Attributes:       []string{"Make", "Model", "Year"},
		Infos:            []string{dd.Make.Name, dd.Type.Model, strconv.Itoa(int(dd.Type.Year))},
	}

	if userDevice.IpfsImageCid.Valid {
		mvs.Attributes = append(mvs.Attributes, imageURIattribute)
		mvs.Infos = append(mvs.Infos, ipfs.URL(userDevice.IpfsImageCid.String))
	}

	return mvs, dd, nil
}
