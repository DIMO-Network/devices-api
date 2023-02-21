package main

import (
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/controllers"
	"github.com/DIMO-Network/shared/middleware/privilegetoken"
	pr "github.com/DIMO-Network/shared/middleware/privilegetoken"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

type httpTestTemplate struct {
	description  string // description of the test case
	route        string // route path to test
	expectedCode int    // expected HTTP status code
}

type testHelper struct {
	app                 *fiber.App
	t                   *testing.T
	assert              *assert.Assertions
	logger              zerolog.Logger
	privilegeMiddleware pr.IVerifyPrivilegeToken
}

type CustomClaims struct {
	ContractAddress common.Address `json:"contract_address"`
	TokenID         string         `json:"token_id"`
	PrivilegeIDs    []int64        `json:"privilege_ids"`
}

type Token struct {
	jwt.RegisteredClaims
	CustomClaims
}

func initTestHelper(t *testing.T) testHelper {
	assert := assert.New(t)
	app := fiber.New() // This can be moved into a new var to avoid re-initializing TBD.
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return testHelper{
		assert: assert,
		t:      t,
		app:    app,
		logger: logger,
		privilegeMiddleware: pr.New(pr.Config{
			Log: &logger,
		}),
	}
}

func (t testHelper) signToken(p jwt.MapClaims) *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, p)
}

func TestSuccessOnValidSinglePrivilege(t *testing.T) {
	th := initTestHelper(t)

	test := httpTestTemplate{

		description:  "Test success response when token contains at only allowed privilege on endpoint",
		route:        fmt.Sprintf("/v1/test/%d", controllers.Commands),
		expectedCode: fiber.StatusOK,
	}

	vehicleAddr := common.BytesToAddress([]byte{uint8(1)})

	th.app.Use(func(c *fiber.Ctx) error {
		token := th.signToken((jwt.MapClaims{
			"token_id":         "2",
			"contract_address": vehicleAddr,
			"privilege_ids":    []int64{controllers.Commands},
		}))

		c.Locals("user", token)
		return c.Next()
	})

	th.app.Get("/v1/test/:tokenID", th.privilegeMiddleware.OneOf(vehicleAddr, []int64{controllers.Commands}), func(c *fiber.Ctx) error {
		return c.SendString("Ok")
	})

	req := httptest.NewRequest("GET", test.route, nil)

	resp, _ := th.app.Test(req, 1)

	// Verify, if the status code is as expected
	th.assert.Equalf(test.expectedCode, resp.StatusCode, test.description)
}

func TestSuccessOnValidTokenPrivilegeOnMany(t *testing.T) {
	th := initTestHelper(t)

	test := httpTestTemplate{
		description:  "Test success response when token contains at least 1 of allowed privileges on endpoint",
		route:        fmt.Sprintf("/v1/test/%d", controllers.Commands),
		expectedCode: fiber.StatusOK,
	}

	vehicleAddr := common.BytesToAddress([]byte{uint8(1)})

	th.app.Use(func(c *fiber.Ctx) error {
		token := th.signToken((jwt.MapClaims{
			"token_id":         "2",
			"contract_address": vehicleAddr,
			"privilege_ids":    []int64{controllers.Commands},
		}))

		c.Locals("user", token)
		return c.Next()
	})

	th.app.Get("/v1/test/:tokenID", th.privilegeMiddleware.OneOf(vehicleAddr, []int64{controllers.Commands, controllers.AllTimeLocation}), func(c *fiber.Ctx) error {

		return c.SendString("Ok")
	})

	req := httptest.NewRequest("GET", test.route, nil)

	resp, _ := th.app.Test(req, 1)

	// Verify, if the status code is as expected
	th.assert.Equalf(test.expectedCode, resp.StatusCode, test.description)
}

func TestMiddlewareWriteClaimsToContext(t *testing.T) {
	th := initTestHelper(t)

	test := httpTestTemplate{
		description:  "Test success response when token contains at least 1 of allowed privileges on endpoint",
		route:        fmt.Sprintf("/v1/test/%d", controllers.Commands),
		expectedCode: fiber.StatusOK,
	}

	vehicleAddr := common.BytesToAddress([]byte{uint8(1)})
	cClaims := privilegetoken.CustomClaims{
		ContractAddress: vehicleAddr,
		TokenID:         "2",
		PrivilegeIDs:    []int64{controllers.Commands},
	}
	th.app.Use(func(c *fiber.Ctx) error {
		token := th.signToken((jwt.MapClaims{
			"token_id":         cClaims.TokenID,
			"contract_address": cClaims.ContractAddress,
			"privilege_ids":    cClaims.PrivilegeIDs,
		}))

		c.Locals("user", token)
		return c.Next()
	})

	th.app.Get("/v1/test/:tokenID", th.privilegeMiddleware.OneOf(vehicleAddr, []int64{controllers.Commands, controllers.AllTimeLocation}), func(c *fiber.Ctx) error {
		cl := c.Locals("tokenClaims").(privilegetoken.CustomClaims)
		th.assert.Equal(cl, cClaims)
		return c.SendString("Ok")
	})

	req := httptest.NewRequest("GET", test.route, nil)

	resp, _ := th.app.Test(req, 1)

	// Verify, if the status code is as expected
	th.assert.Equalf(test.expectedCode, resp.StatusCode, test.description)
}

func TestFailureOnInvalidPrivilegeInToken(t *testing.T) {
	th := initTestHelper(t)

	test := httpTestTemplate{
		description:  "Test unauthorized response when token does not contain at least 1 of allowed privileges on endpoint",
		route:        fmt.Sprintf("/v1/test/%d", controllers.Commands),
		expectedCode: fiber.StatusUnauthorized,
	}

	vehicleAddr := common.BytesToAddress([]byte{uint8(1)})

	th.app.Use(func(c *fiber.Ctx) error {
		token := th.signToken((jwt.MapClaims{
			"token_id":         "2",
			"contract_address": vehicleAddr,
			"privilege_ids":    []int64{controllers.Commands},
		}))

		c.Locals("user", token)
		return c.Next()
	})

	th.app.Get("/v1/test/:tokenID", th.privilegeMiddleware.OneOf(vehicleAddr, []int64{controllers.AllTimeLocation}), func(c *fiber.Ctx) error {
		return c.SendString("Ok")
	})

	req := httptest.NewRequest("GET", test.route, nil)

	resp, _ := th.app.Test(req, 1)

	// Verify, if the status code is as expected
	th.assert.Equalf(test.expectedCode, resp.StatusCode, test.description)
}

func TestFailureOnInvalidContractAddress(t *testing.T) {
	th := initTestHelper(t)

	test := httpTestTemplate{
		description:  "Test unauthorized response when token does not contain at least 1 of allowed privileges on endpoint",
		route:        fmt.Sprintf("/v1/test/%d", controllers.Commands),
		expectedCode: fiber.StatusUnauthorized,
	}

	vehicleAddr := common.BytesToAddress([]byte{uint8(1)})

	th.app.Use(func(c *fiber.Ctx) error {
		token := th.signToken((jwt.MapClaims{
			"token_id":         "2",
			"contract_address": common.BytesToAddress([]byte{uint8(2)}),
			"privilege_ids":    []int64{controllers.Commands},
		}))

		c.Locals("user", token)
		return c.Next()
	})

	th.app.Get("/v1/test/:tokenID", th.privilegeMiddleware.OneOf(vehicleAddr, []int64{controllers.AllTimeLocation}), func(c *fiber.Ctx) error {
		return c.SendString("Ok")
	})

	req := httptest.NewRequest("GET", test.route, nil)

	resp, _ := th.app.Test(req, 1)

	// Verify, if the status code is as expected
	th.assert.Equalf(test.expectedCode, resp.StatusCode, test.description)
}
