package helpers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	//d "github.com/dexidp/dex/api/v2"
)

type CustomClaims struct {
	ContractAddress common.Address `json:"contract_address"`
	TokenID         string         `json:"token_id"`
	PrivilegeIDs    []int64        `json:"privilege_ids"`
}

type Token struct {
	jwt.RegisteredClaims
	CustomClaims
}

// ErrorResponseHandler is deprecated. it doesn't log. We prefer to return an err and have the ErrorHandler in api.go handle stuff.
func ErrorResponseHandler(c *fiber.Ctx, err error, status int) error {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	return c.Status(status).JSON(fiber.Map{
		"errorMessage": msg,
	})
}

func GetUserID(c *fiber.Ctx) string {
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userID := claims["sub"].(string)
	return userID
}

type VehicleTokenClaims struct {
	VehicleTokenID string
	UserEthAddress string
	Privileges     []int64
}

type VehicleTokenClaimsResponseRaw struct {
	Sub        string
	UserID     string
	Privileges []int64
}

func GetVehicleTokenClaims(c *fiber.Ctx) (VehicleTokenClaims, error) {
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)

	jsonbody, err := json.Marshal(claims)
	if err != nil {
		return VehicleTokenClaims{}, err
	}

	p := VehicleTokenClaimsResponseRaw{}

	if err := json.Unmarshal(jsonbody, &p); err != nil {
		return VehicleTokenClaims{}, err
	}

	return VehicleTokenClaims{
		VehicleTokenID: p.Sub,
		UserEthAddress: p.UserID,
		Privileges:     p.Privileges,
	}, nil
}

func GetPrivilegeTokenClaims(c *fiber.Ctx) (Token, error) {
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)

	jsonbody, err := json.Marshal(claims)
	if err != nil {
		return Token{}, err
	}

	var t Token
	err = json.Unmarshal(jsonbody, &t)
	if err != nil {
		return Token{}, err
	}

	return t, nil
}

// CreateResponse is a generic response with an ID of the created entity
type CreateResponse struct {
	ID string `json:"id"`
}

// GrpcErrorToFiber useful anywhere calling a grpc underlying service and wanting to augment the error for fiber from grpc status codes
// meant to play nicely with the ErrorHandler in api.go that this would hand off errors to.
// msgAppend appends to error string, to eg. help if this gets logged
func GrpcErrorToFiber(err error, msgAppend string) error {
	if err == nil {
		return nil
	}
	// pull out grpc error status to then convert to fiber http equivalent
	errStatus, _ := status.FromError(err)

	switch errStatus.Code() {
	case codes.InvalidArgument:
		return fiber.NewError(fiber.StatusBadRequest, errStatus.Message()+". "+msgAppend)
	case codes.NotFound:
		return fiber.NewError(fiber.StatusNotFound, errStatus.Message()+". "+msgAppend)
	case codes.Aborted:
		return fiber.NewError(fiber.StatusConflict, errStatus.Message()+". "+msgAppend)
	case codes.Internal:
		return fiber.NewError(fiber.StatusInternalServerError, errStatus.Message()+". "+msgAppend)
	}
	return errors.Wrap(err, msgAppend)
}

func GetLogger(c *fiber.Ctx, d *zerolog.Logger) *zerolog.Logger {
	m := c.Locals("logger")
	if m == nil {
		return d
	}

	l, ok := m.(*zerolog.Logger)
	if !ok {
		return d
	}

	return l
}

type ErrorRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ErrorHandler custom handler to log recovered errors using our logger and return json instead of string
func ErrorHandler(c *fiber.Ctx, err error, logger *zerolog.Logger, isProduction bool) error {
	logger = GetLogger(c, logger)

	code := fiber.StatusInternalServerError // Default 500 statuscode

	e, fiberTypeErr := err.(*fiber.Error)
	if fiberTypeErr {
		// Override status code if fiber.Error type
		code = e.Code
	}
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	codeStr := strconv.Itoa(code)

	logger.Err(err).Str("httpStatusCode", codeStr).
		Str("httpMethod", c.Method()).
		Str("httpPath", c.Path()).
		Msg("caught an error from http request")
	// return an opaque error if we're in a higher level environment and we haven't specified an fiber type err.
	if !fiberTypeErr && isProduction {
		err = fiber.NewError(fiber.StatusInternalServerError, "Internal error")
	}

	return c.Status(code).JSON(ErrorRes{
		Code:    code,
		Message: err.Error(),
	})
}

type DeviceAttributeType string

const (
	Mpg                 DeviceAttributeType = "mpg"
	FuelTankCapacityGal DeviceAttributeType = "fuel_tank_capacity_gal"
	MpgHighway          DeviceAttributeType = "mpg_highway"
)

type DeviceDefinitionRange struct {
	FuelTankCapGal float64 `json:"fuel_tank_capacity_gal"`
	Mpg            float64 `json:"mpg"`
	MpgHwy         float64 `json:"mpg_highway"`
}

func GetActualDeviceDefinitionMetadataValues(dd *grpc.GetDeviceDefinitionItemResponse, deviceStyleID null.String) *DeviceDefinitionRange {

	var fuelTankCapGal, mpg, mpgHwy float64 = 0, 0, 0

	var metadata []*grpc.DeviceTypeAttribute

	if !deviceStyleID.IsZero() {
		for _, style := range dd.DeviceStyles {
			if style.Id == deviceStyleID.String {
				metadata = style.DeviceAttributes
				break
			}
		}
	}

	if len(metadata) == 0 && dd != nil && dd.DeviceAttributes != nil {
		metadata = dd.DeviceAttributes
	}

	for _, attr := range metadata {
		switch DeviceAttributeType(attr.Name) {
		case FuelTankCapacityGal:
			if v, err := strconv.ParseFloat(attr.Value, 32); err == nil {
				fuelTankCapGal = v
			}
		case Mpg:
			if v, err := strconv.ParseFloat(attr.Value, 32); err == nil {
				mpg = v
			}
		case MpgHighway:
			if v, err := strconv.ParseFloat(attr.Value, 32); err == nil {
				mpgHwy = v
			}
		}
	}

	return &DeviceDefinitionRange{
		FuelTankCapGal: fuelTankCapGal,
		Mpg:            mpg,
		MpgHwy:         mpgHwy,
	}
}

// ECRecoverSol used to verify message signature
// accepts message and signature hash and returns address of signer
func ECRecoverSol(hash []byte, sig []byte) (common.Address, error) {
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
