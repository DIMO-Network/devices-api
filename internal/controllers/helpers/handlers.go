package helpers

import (
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
	//d "github.com/dexidp/dex/api/v2"
)

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

func GetUserEthAddr(c *fiber.Ctx) *common.Address {
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	ethAddr, ok := claims["ethereum_address"].(string)
	if !ok || !common.IsHexAddress(ethAddr) {
		return nil
	}
	e := common.HexToAddress(ethAddr)
	return &e
}

// CreateResponse is a generic response with an ID of the created entity
type CreateResponse struct {
	ID string `json:"id"`
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

const skipErrorLogKey = "skipErrorLog"

func SkipErrorLog(c *fiber.Ctx) {
	c.Locals(skipErrorLogKey, true)
}

// ErrorHandler custom handler to log recovered errors using our logger and return json instead of string
func ErrorHandler(c *fiber.Ctx, err error, logger *zerolog.Logger, isProduction bool) error {
	logger = GetLogger(c, logger)

	code := fiber.StatusInternalServerError // Default 500 statuscode

	var e *fiber.Error
	isFiberErr := errors.As(err, &e)
	if isFiberErr {
		// Override status code if fiber.Error type
		code = e.Code
	}
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	codeStr := strconv.Itoa(code)

	if code != fiber.StatusNotFound && c.Locals(skipErrorLogKey) != true {
		logger.Err(err).Str("httpStatusCode", codeStr).
			Str("httpMethod", c.Method()).
			Str("httpPath", c.Path()).
			Msg("caught an error from http request")
	}
	// return an opaque error if we're in a higher level environment and we haven't specified an fiber type err.
	if !isFiberErr && isProduction {
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

var zeroAddr common.Address

const sigLen = 65

// Ecrecover mimics the ecrecover opcode, returning the address that signed
// hash with signature. sig must have length 65 and the last byte, the recovery
// byte usually denoted v, must be 27 or 28.
func Ecrecover(hash, sig []byte) (common.Address, error) {
	if len(sig) != sigLen {
		return zeroAddr, fmt.Errorf("signature has invalid length %d", len(sig))
	}

	// Defensive copy: the caller shouldn't have to worry about us modifying
	// the signature. We adjust because crypto.Ecrecover demands 0 <= v <= 4.
	fixedSig := make([]byte, sigLen)
	copy(fixedSig, sig)
	fixedSig[64] -= 27

	rawPk, err := crypto.Ecrecover(hash, fixedSig)
	if err != nil {
		return zeroAddr, err
	}

	pk, err := crypto.UnmarshalPubkey(rawPk)
	if err != nil {
		return zeroAddr, err
	}

	return crypto.PubkeyToAddress(*pk), nil
}

// TODO(elffjs): This is becoming a dumping ground.

func APIError(code int, message string, args ...any) error {
	return fiber.NewError(code, fmt.Sprintf(message, args...))
}
