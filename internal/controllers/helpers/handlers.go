package helpers

import (
	"encoding/json"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
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

// ErrorHandler custom handler to log recovered errors using our logger and return json instead of string
func ErrorHandler(c *fiber.Ctx, err error, logger zerolog.Logger, isProduction bool) error {
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

	return c.Status(code).JSON(fiber.Map{
		"code":    code,
		"message": err.Error(),
	})
}
