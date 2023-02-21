package main

import (
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/controllers"
	"github.com/DIMO-Network/shared/middleware/privilegetoken"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog"
)

type httpTestTemplate struct {
	description  string // description of the test case
	route        string // route path to test
	expectedCode int    // expected HTTP status code
}

// type testHelper struct {
// 	app                 *fiber.App
// 	t                   *testing.T
// 	assert              *assert.Assertions
// 	logger              zerolog.Logger
// 	privilegeMiddleware pr.IVerifyPrivilegeToken
// }

type CustomClaims struct {
	ContractAddress common.Address `json:"contract_address"`
	TokenID         string         `json:"token_id"`
	PrivilegeIDs    []int64        `json:"privilege_ids"`
}

type Token struct {
	jwt.RegisteredClaims
	CustomClaims
}

func TestPrivilegeMiddleware(t *testing.T) {
	// th := initTestHelper(t)

	// tests := []httpTestTemplate{
	// 	{
	// 		description:  "Test simple 200 is returned",
	// 		route:        fmt.Sprintf("/v1/test/%d", controllers.Commands),
	// 		expectedCode: 200,
	// 	},
	// }

	app := fiber.New()

	vehicleAddr := "0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF"

	pre := func(c *fiber.Ctx) error {
		tok := &jwt.Token{
			Claims: jwt.MapClaims{
				// "userEthAddress":   common.BytesToAddress([]byte{uint8(2)}).Hex(),
				"token_id":         "2",
				"contract_address": vehicleAddr,
				"privilege_ids":    []int64{controllers.Commands},
			},
		}

		c.Locals("user", tok)

		return c.Next()
	}

	app.Use(pre)

	logger := zerolog.Nop()

	mw := privilegetoken.New(privilegetoken.Config{Log: &logger})

	app.Get("/v1/test/:tokenID", mw.OneOf(common.HexToAddress(vehicleAddr), []int64{controllers.Commands}), func(c *fiber.Ctx) error {
		return c.SendString("Ok")
	})

	req := httptest.NewRequest("GET", "/v1/test/2", nil)

	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, but got %d", resp.StatusCode)
	}

	fmt.Println(resp)
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(out))
}
