package main

import (
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/controllers"
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

func TestPrivilegeMiddleware(t *testing.T) {
	th := initTestHelper(t)

	tests := []httpTestTemplate{
		{
			description:  "Test simple 200 is returned",
			route:        fmt.Sprintf("/v1/test/%d", controllers.Commands),
			expectedCode: 200,
		},
	}

	th.app.Use(func(c *fiber.Ctx) error {
		token := th.signToken((jwt.MapClaims{
			"UserEthAddress":     common.BytesToAddress([]byte{uint8(2)}).Hex(),
			"TokenID":            "2",
			"NFTContractAddress": common.BytesToAddress([]byte{uint8(1)}).Hex(),
			"PrivilegeIDs":       []int64{controllers.Commands},
		}))

		c.Locals("user", token)
		return c.Next()
	})

	vehicleAddr := common.BytesToAddress([]byte{uint8(1)})

	th.app.Get("/v1/test/:tokenID", th.privilegeMiddleware.OneOf(vehicleAddr, []int64{controllers.Commands}), func(c *fiber.Ctx) error {
		return c.SendString("Ok")
	})

	// Iterate through test single test cases
	for _, test := range tests {
		req := httptest.NewRequest("GET", test.route, nil)

		resp, _ := th.app.Test(req, 1)

		// Verify, if the status code is as expected
		assert.Equalf(t, test.expectedCode, resp.StatusCode, test.description)
	}
}
