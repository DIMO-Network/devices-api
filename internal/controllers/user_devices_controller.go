package controllers

import (
	"bytes"
	"cmp"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/contracts"
	sig2 "github.com/DIMO-Network/devices-api/internal/contracts/signature"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/ipfs"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
	"github.com/DIMO-Network/devices-api/internal/utils"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/redis"
	pb_oracle "github.com/DIMO-Network/tesla-oracle/pkg/grpc"
	"github.com/IBM/sarama"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ericlagergren/decimal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	signer "github.com/ethereum/go-ethereum/signer/core/apitypes"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/drivers"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
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
	teslaTaskService          services.TeslaTaskService
	teslaOracle               pb_oracle.TeslaOracleClient
	cipher                    shared.Cipher
	autoPiSvc                 services.AutoPiAPIService
	autoPiIngestRegistrar     services.IngestRegistrar
	s3                        *s3.Client
	producer                  sarama.SyncProducer
	deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	redisCache                redis.CacheService
	openAI                    services.OpenAI
	NATSSvc                   *services.NATSService
	wallet                    services.SyntheticWalletInstanceService
	userDeviceSvc             services.UserDeviceService
	teslaFleetAPISvc          services.TeslaFleetAPIService
	ipfsSvc                   *ipfs.IPFS
	clickHouseConn            clickhouse.Conn
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
	teslaTaskService services.TeslaTaskService,
	teslaOracle pb_oracle.TeslaOracleClient,
	cipher shared.Cipher,
	autoPiSvc services.AutoPiAPIService,
	autoPiIngestRegistrar services.IngestRegistrar,
	deviceDefinitionRegistrar services.DeviceDefinitionRegistrar,
	producer sarama.SyncProducer,
	s3NFTClient *s3.Client,
	cache redis.CacheService,
	openAI services.OpenAI,
	natsSvc *services.NATSService,
	wallet services.SyntheticWalletInstanceService,
	userDeviceSvc services.UserDeviceService,
	teslaFleetAPISvc services.TeslaFleetAPIService,
	ipfsSvc *ipfs.IPFS,
	chConn clickhouse.Conn,
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
		teslaTaskService:          teslaTaskService,
		teslaOracle:               teslaOracle,
		cipher:                    cipher,
		autoPiSvc:                 autoPiSvc,
		autoPiIngestRegistrar:     autoPiIngestRegistrar,
		s3:                        s3NFTClient,
		producer:                  producer,
		deviceDefinitionRegistrar: deviceDefinitionRegistrar,
		redisCache:                cache,
		openAI:                    openAI,
		NATSSvc:                   natsSvc,
		wallet:                    wallet,
		userDeviceSvc:             userDeviceSvc,
		teslaFleetAPISvc:          teslaFleetAPISvc,
		ipfsSvc:                   ipfsSvc,
		clickHouseConn:            chConn,
	}
}

func (udc *UserDevicesController) dbDevicesToDisplay(ctx context.Context, devices []*models.UserDevice) ([]UserDeviceFull, error) {
	apiDevices := []UserDeviceFull{}

	if len(devices) == 0 {
		return apiDevices, nil
	}

	deviceDefinitionResponse := make([]*ddgrpc.GetDeviceDefinitionItemResponse, len(devices))
	for i, userDevice := range devices {
		definitionID := userDevice.DefinitionID

		def, err := udc.DeviceDefSvc.GetDeviceDefinitionBySlug(ctx, definitionID)
		if err != nil {
			udc.log.Err(err).Str("userDeviceId", userDevice.ID).
				Str("definitionId", userDevice.DefinitionID).
				Msg("failed to resolve device definition for vehicle.")
			return nil, shared.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+definitionID)
		}
		deviceDefinitionResponse[i] = def
	}

	filterDeviceDefinition := func(id string, items []*ddgrpc.GetDeviceDefinitionItemResponse) (*ddgrpc.GetDeviceDefinitionItemResponse, error) {
		for _, dd := range items {
			if id == dd.Id {
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
		deviceDefinition, err := filterDeviceDefinition(d.DefinitionID, deviceDefinitionResponse)
		if err != nil {
			return nil, fmt.Errorf("user device %s has unknown definition %s", d.ID, d.DefinitionID)
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

				nft.Status = "Confirmed"

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

const sourcePrefix = "dimo/integration/"

var (
	dialect = drivers.Dialect{
		LQ: '`',
		RQ: '`',
	}
	connectionIDToIntegrationID = map[string]string{
		"0xF26421509Efe92861a587482100c6d728aBf1CD0": "2lcaMFuCO0HJIUfdq8o780Kx5n3", // ruptela
		"0x5e31bBc786D7bEd95216383787deA1ab0f1c1897": "27qftVRWQYpVDcO5DltO5Ojbjxk", // autopi
		"0xc4035Fecb1cc906130423EF05f9C20977F643722": "26A5Dk3vvvQutjSyF0Jka2DP5lg", // tesla
		"0x4c674ddE8189aEF6e3b58F5a36d7438b2b1f6Bc2": "2ULfuC8U9dOqRshZBAi0lMM1Rrx", // macaron
		"0xcd445F4c6bDAD32b68a2939b912150Fe3C88803E": "22N2xaPOq2WW2gAHBHd0Ikn4Zob", // smartcar
		"0x55BF1c27d468314Ea119CF74979E2b59F962295c": "2szgr5WqMQtK2ZFM8F8qW8WUfJa", // compass
	}
	integrationIDToConnectionID = func() map[string]string {
		// reverse of integrationId2ConnectionId
		out := make(map[string]string, len(connectionIDToIntegrationID))
		for k, v := range connectionIDToIntegrationID {
			out[v] = k
		}
		return out
	}()
)

func chSourceToIntegrationID(s string) string {
	if integrationID, ok := connectionIDToIntegrationID[s]; ok {
		return integrationID
	}
	return strings.TrimPrefix(s, sourcePrefix)
}

func integrationIDToCHSource(id string) []string {
	var sources []string
	if chSources, ok := integrationIDToConnectionID[id]; ok {
		sources = append(sources, chSources)
	}
	return append(sources, sourcePrefix+id)
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

	userAddr, err := helpers.GetJWTEthAddr(c)
	if err != nil {
		return err
	}

	devices, err := models.UserDevices(
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Or2(models.UserDeviceWhere.OwnerAddress.EQ(null.BytesFrom(userAddr.Bytes()))),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
		qm.Load(models.UserDeviceRels.MintRequest),
		qm.Load(models.UserDeviceRels.BurnRequest),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleTokenSyntheticDevice, models.SyntheticDeviceRels.MintRequest)),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleTokenSyntheticDevice, models.SyntheticDeviceRels.BurnRequest)),
		qm.OrderBy(models.UserDeviceColumns.CreatedAt+" DESC"),
	).All(c.Context(), udc.DBS().Reader)
	if err != nil {
		return helpers.ErrorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	{
		type checkKey struct {
			TokenID       uint32
			IntegrationID string
		}
		toCheck := make(map[checkKey]*models.UserDeviceAPIIntegration)
		for _, ud := range devices {
			if ud.TokenID.IsZero() {
				continue
			}
			for _, udai := range ud.R.UserDeviceAPIIntegrations {
				// TODO(elffjs): Really no point in doing this for synthetics if the job hasn't started.
				if udai.Status == models.UserDeviceAPIIntegrationStatusPending || udai.Status == models.UserDeviceAPIIntegrationStatusPendingFirstData {
					tok, _ := ud.TokenID.Uint64()
					toCheck[checkKey{uint32(tok), udai.IntegrationID}] = udai
				}
			}
		}

		if len(toCheck) != 0 {
			udc.log.Debug().Str("userId", userID).Msgf("Checking %d inactive connections.", len(toCheck))
			var innerList []qm.QueryMod

			for key, udai := range toCheck {
				clause := qm.Expr(
					qmhelper.Where("token_id", qmhelper.EQ, key.TokenID),
					qm.WhereIn("source IN ?", integrationIDToCHSource(key.IntegrationID)),
					qmhelper.Where("timestamp", qmhelper.GT, udai.UpdatedAt))
				if len(innerList) == 0 {
					innerList = append(innerList, clause)
				} else {
					innerList = append(innerList, qm.Or2(clause))
				}
			}

			// Please query optimizer. PLEASE.
			list := []qm.QueryMod{
				qm.Distinct("token_id, source"),
				qm.From("signal"),
				qm.Expr(innerList...),
			}

			q := &queries.Query{}
			queries.SetDialect(q, &dialect)
			qm.Apply(q, list...)

			query, args := queries.BuildQuery(q)

			udc.log.Debug().Str("userId", userID).Msgf("Querying for inactives. Query %q, args %q", query, args)

			rows, err := udc.clickHouseConn.Query(c.Context(), query, args...)
			if err != nil {
				return err
			}
			defer rows.Close()

			var toModify []*models.UserDeviceAPIIntegration

			for rows.Next() {
				var tokenID uint32
				var source string
				if err := rows.Scan(&tokenID, &source); err != nil {
					return err
				}
				if udai, ok := toCheck[checkKey{tokenID, chSourceToIntegrationID(source)}]; ok {
					toModify = append(toModify, udai)
				} else {
					return fmt.Errorf("signal activity query returned a token id %d not in the query", tokenID)
				}
			}

			if rows.Err() != nil {
				return fmt.Errorf("clickhouse scan error: %w", rows.Err())
			}

			if len(toModify) != 0 {
				modTime := time.Now()
				tx, err := udc.DBS().Writer.BeginTx(c.Context(), nil)
				if err != nil {
					return err
				}

				for _, udai := range toModify {
					udc.log.Info().Str("userId", userID).Str("userDeviceId", udai.UserDeviceID).Str("integrationId", udai.IntegrationID).Msg("Setting connection active.")
					udai.Status = models.UserDeviceAPIIntegrationStatusActive
					udai.UpdatedAt = modTime
					_, err := udai.Update(c.Context(), tx, boil.Whitelist(models.UserDeviceAPIIntegrationColumns.Status, models.UserDeviceAPIIntegrationColumns.UpdatedAt))
					if err != nil {
						return err
					}
				}

				if err := tx.Commit(); err != nil {
					return err
				}
			}
		}
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

	userAddr, err := helpers.GetJWTEthAddr(c)
	if err != nil {
		return err
	}

	var sharedVehicles []*models.UserDevice

	// This is N+1 hell.
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

	definitionID := reg.DefinitionID
	// if definitionID is blank, it means we need to use old DeviceDefinitionID to resolve
	if definitionID == "" {
		if reg.DeviceDefinitionID == nil {
			return fiber.NewError(fiber.StatusBadRequest, "definitionId is required")
		}
		url := fmt.Sprintf("%s%s", udc.Settings.DeviceDefinitionsGetByKSUIDEndpoint, *reg.DeviceDefinitionID)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, errors.Wrap(err, "failed to create request for get device definition").Error())
		}
		req.Header.Set("Accept", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request to %s: %v", url, err)
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, errors.Wrap(err, "failed to read body to get device definition").Error())
		}
		// use gjson to get the new id
		definitionID = gjson.GetBytes(body, "nameSlug").String()

		if definitionID == "" {
			udc.log.Error().Msgf("Failed to get device definition nameSlug from dd api response. url: %s response body: %s", url, string(body))
		}
	}

	udFull, err := udc.createUserDevice(c.Context(), definitionID, "", reg.CountryCode, userID, nil, nil, false)
	if err != nil {
		return shared.GrpcErrorToFiber(err, "")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"userDevice": udFull,
	})
}

var opaqueInternalError = fiber.NewError(fiber.StatusInternalServerError, "Internal error.")

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
// @Deprecated
func (udc *UserDevicesController) RegisterDeviceForUserFromSmartcar(_ *fiber.Ctx) error {
	return fiber.NewError(fiber.StatusBadRequest, "Creating Smartcar devices is no longer supported.")
}

func (udc *UserDevicesController) createUserDevice(ctx context.Context, definitionID, styleID, countryCode, userID string, vin, canProtocol *string, vinConfirmed bool) (*UserDeviceFull, error) {
	ud, dd, err := udc.userDeviceSvc.CreateUserDevice(ctx, definitionID, styleID, countryCode, userID, vin, canProtocol, vinConfirmed)
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

const (
	PowerTrainTypeKey = "powertrain_type"
)

// todo revisit this depending on what observe with below log message
func (udc *UserDevicesController) updatePowerTrain(ctx context.Context, userDevice *models.UserDevice) error {
	md := new(services.UserDeviceMetadata)
	if err := userDevice.Metadata.Unmarshal(md); err != nil {
		return err
	}
	resp, err := udc.DeviceDefSvc.GetDeviceDefinitionBySlug(ctx, userDevice.DefinitionID)
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

	dd, err := udc.DeviceDefSvc.GetDeviceDefinitionBySlug(c.Context(), userDevice.DefinitionID)
	if err != nil {
		return shared.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+userDevice.DefinitionID)
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
				ID:           udi,
				Make:         dd.Make.Name,
				Model:        dd.Model,
				Year:         int(dd.Year),
				DefinitionID: dd.Id,
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
	userDeviceID := c.Params("userDeviceID")

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(models.UserDeviceRels.MintRequest),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "No vehicle with that id found.")
	}

	mvs, dd, err := udc.checkVehicleMint(c, userDevice)
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
		DeviceDefinitionID: dd.Id,
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

	if udc.Settings.BlockMinting {
		return fiber.NewError(fiber.StatusInternalServerError, "Smartcar and Tesla device minting temporarily offline for a network upgrade.")
	}

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
	mvs, dd, err := udc.checkVehicleMint(c, userDevice)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errors.Wrapf(err, "failed to checkVehicleMint. user device id: %s", userDeviceID).Error())
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
		DeviceDefinitionID: dd.Id,
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
		var newIdents *utils.ConnectionChainIDs

		for _, udai := range udais {
			if info, ok := utils.SyntheticIntegrationKSUIDToOtherIDs[udai.IntegrationID]; ok {
				newIdents = info
				break
			}
		}

		if newIdents != nil {
			if newIdents.Name == "Smartcar" {
				return fiber.NewError(fiber.StatusBadRequest, "Smartcar mints are no longer supported.")
			}

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
				IntegrationTokenID: types.NewDecimal(decimal.New(newIdents.IntegrationNode.Int64(), 0)),
				MintRequestID:      requestID,
				WalletChildNumber:  seq.NextVal,
				WalletAddress:      addr,
			}

			if err := sd.Insert(c.Context(), tx, boil.Infer()); err != nil {
				return err
			}

			var msg registry.Message

			if udc.Settings.ConnectionsReplacedIntegrations {
				msg = &registry.MintVehicleAndSdSignV2{
					ConnectionID: newIdents.ConnectionID,
				}
			} else {
				msg = &registry.MintVehicleAndSdSign{
					IntegrationNode: newIdents.IntegrationNode,
				}
			}

			hash, err := client.Hash(msg)
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

			// register synthetic device with tesla oracle
			if newIdents.Name == "Tesla" {
				if _, err := udc.teslaOracle.RegisterNewSyntheticDevice(c.Context(), &pb_oracle.RegisterNewSyntheticDeviceRequest{
					Vin:                    userDevice.VinIdentifier.String,
					SyntheticDeviceAddress: sd.WalletAddress,
					WalletChildNum:         uint64(sd.WalletChildNumber),
				}); err != nil {
					logger.Err(err).Msg("failed to register synthetic device with tesla oracle")
				}
			}

			var maybeIntegrationNode *big.Int
			if udc.Settings.ConnectionsReplacedIntegrations {
				maybeIntegrationNode = newIdents.ConnectionID
			} else {
				maybeIntegrationNode = newIdents.IntegrationNode
			}

			if mr.SACDInput == nil || !udc.Settings.EnableSACDMint {
				return client.MintVehicleAndSdWithDeviceDefinitionSign(requestID, contracts.MintVehicleAndSdWithDdInput{
					ManufacturerNode:     mvs.ManufacturerNode,
					Owner:                mvs.Owner,
					DeviceDefinitionId:   dd.Id,
					IntegrationNode:      maybeIntegrationNode,
					VehicleOwnerSig:      sigBytes,
					SyntheticDeviceSig:   sign,
					SyntheticDeviceAddr:  common.BytesToAddress(addr),
					AttrInfoPairsVehicle: attrListsToAttrPairs(mvs.Attributes, mvs.Infos),
					AttrInfoPairsDevice:  []contracts.AttributeInfoPair{},
				})
			}

			return client.MintVehicleAndSdWithDeviceDefinitionSignAndSacd(requestID, contracts.MintVehicleAndSdWithDdInput{
				ManufacturerNode:     mvs.ManufacturerNode,
				Owner:                mvs.Owner,
				DeviceDefinitionId:   dd.Id,
				IntegrationNode:      maybeIntegrationNode,
				VehicleOwnerSig:      sigBytes,
				SyntheticDeviceSig:   sign,
				SyntheticDeviceAddr:  common.BytesToAddress(addr),
				AttrInfoPairsVehicle: attrListsToAttrPairs(mvs.Attributes, mvs.Infos),
				AttrInfoPairsDevice:  []contracts.AttributeInfoPair{},
			}, contracts.SacdInput{
				Grantee:     mr.SACDInput.Grantee,
				Permissions: mr.SACDInput.Permissions,
				Expiration:  mr.SACDInput.Expiration,
				Source:      mr.SACDInput.Source,
			})

		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// TODO(elffjs): No one should be hitting this.
	logger.Info().Msgf("Submitted metatransaction request %s", requestID)

	return client.MintVehicleWithDeviceDefinitionSign(requestID, mvs.ManufacturerNode, mvs.Owner, dd.Id, attrListsToAttrPairs(mvs.Attributes, mvs.Infos), sigBytes)
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
	// SACDInput contains user signed permission grant, including grantee, permissions, expiration and link to signed document
	SACDInput *SACDInput `json:"sacdInput,omitempty"`
}

// SACDInput user signed permission grant
type SACDInput struct {
	// Grantee is the Ethereum address permissions are being granted to.
	Grantee common.Address `json:"grantee" swaggertype:"string" example:"0xAb5801a7D398351b8bE11C439e05C5b3259aec9B"`
	// Permissions are a numerical representation of what permissions are being given to the grantee.
	Permissions *big.Int `json:"permissions" swaggertype:"number" example:"262140"`
	// Expiration permissions granted are valid until this time.
	Expiration *big.Int `json:"expiration" swaggertype:"number" example:"2933125200"`
	// Source external link to signed permission document.
	Source string `json:"source" example:"ipfs://QmWfVnjhbJqAtGCp926jq13kDiszdM8LP15Z2ij5bY4eZD"`
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
	// deprecated
	DeviceDefinitionID *string `json:"deviceDefinitionId"`
	CountryCode        string  `json:"countryCode"`
	// DefinitionID new slug id
	DefinitionID string `json:"definitionId"`
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
	CANProtocol    string `json:"canProtocol"`
	PreApprovedPSK string `json:"preApprovedPSK"`
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
	// CountryCode optional. Is set on the user device record
	CountryCode string `json:"countryCode"`
	// CANProtocol optional. Numeric style made up protocol. 6 = CAN11_500, 7 = CAN29_500, 66/77 are some UDS thing etc
	CANProtocol string `json:"canProtocol"`
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
		// todo add DefinitionId as validated after mobile app updates
		validation.Field(&reg.CountryCode, validation.Required, validation.Length(3, 3)),
	)
}

func (reg *RegisterUserDeviceVIN) Validate() error {
	return validation.ValidateStruct(reg,
		validation.Field(&reg.VIN, validation.Required, validation.Length(13, 17)),
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

func (udc *UserDevicesController) checkVehicleMint(c *fiber.Ctx, userDevice *models.UserDevice) (*registry.MintVehicleSign, *ddgrpc.GetDeviceDefinitionItemResponse, error) {
	if !userDevice.TokenID.IsZero() {
		return nil, nil, fmt.Errorf("vehicle already minted with token id %d", userDevice.TokenID.Big)
	}

	if mr := userDevice.R.MintRequest; mr != nil && mr.Status != models.MetaTransactionRequestStatusFailed {
		return nil, nil, fmt.Errorf("vehicle minting already in process")
	}

	if !userDevice.VinConfirmed {
		return nil, nil, fmt.Errorf("VIN not confirmed")
	}

	if len(userDevice.DefinitionID) == 0 {
		return nil, nil, fmt.Errorf("vehcile definition_id not set")
	}

	dd, err := udc.DeviceDefSvc.GetDeviceDefinitionBySlug(c.Context(), userDevice.DefinitionID)
	if err != nil {
		return nil, nil, fmt.Errorf("error querying for definition by slug id: %s ", userDevice.DefinitionID)
	}

	if dd.Make.TokenId == 0 {
		return nil, nil, fmt.Errorf("vehicle make %q not yet minted", dd.Make.Name)
	}
	makeTokenID := new(big.Int).SetUint64(dd.Make.TokenId)

	userAddr, err := helpers.GetJWTEthAddr(c)
	if err != nil {
		return nil, nil, err
	}

	if dd.Id == "" {
		return nil, nil, fmt.Errorf("invalid on-chain name slug for device definition id: %s", userDevice.DefinitionID)
	}

	mvs := &registry.MintVehicleSign{
		ManufacturerNode: makeTokenID,
		Owner:            userAddr,
		Attributes:       []string{"Make", "Model", "Year"},
		Infos:            []string{dd.Make.Name, dd.Model, strconv.Itoa(int(dd.Year))},
	}

	if userDevice.IpfsImageCid.Valid {
		mvs.Attributes = append(mvs.Attributes, imageURIattribute)
		mvs.Infos = append(mvs.Infos, ipfs.URL(userDevice.IpfsImageCid.String))
	}

	return mvs, dd, nil
}

// GetCompassDeviceByVIN godoc
// @Description Temporary endpoint meant for compass-iot integration. Gets you the token id's by the VIN
// @Tags        user-devices
// @Param       vin path string                   true "VIN"
// @Success     200
// @Failure		404 "user device with VIN not found"
// @Failure		400 "invalid VIN"
// @Failure		500 "server error"
// @Security    PSK
// @Router      /compass/device-by-vin/{vin} [get]
func (udc *UserDevicesController) GetCompassDeviceByVIN(c *fiber.Ctx) error {
	vin := c.Params("vin")
	if len(vin) != 17 {
		return c.Status(fiber.StatusBadRequest).JSON(fmt.Errorf("vin should be 17 characters long"))
	}

	ud, err := models.UserDevices(
		models.UserDeviceWhere.VinIdentifier.EQ(null.StringFrom(vin)),
		qm.Load(models.UserDeviceRels.VehicleTokenSyntheticDevice),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "No device with that VIN found.")
	}
	tkID := uint64(0)
	synthID := uint64(0)
	if !ud.TokenID.IsZero() {
		tkID, _ = ud.TokenID.Uint64()

		if sd := ud.R.VehicleTokenSyntheticDevice; sd != nil {
			if !sd.TokenID.IsZero() {
				synthID, _ = sd.TokenID.Uint64()
			}
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"vin":                    ud.VinIdentifier.String,
		"userDeviceId":           ud.ID,
		"vehicleTokenId":         tkID,
		"syntheticDeviceTokenId": synthID,
		"definitionId":           ud.DefinitionID,
	})
}
