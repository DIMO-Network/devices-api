package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"strings"
)

type PSKAuthMiddleware struct {
	preSharedKey string
	logger       *zerolog.Logger
}

func NewPSKAuthMiddleware(preSharedKey string, logger *zerolog.Logger) *PSKAuthMiddleware {
	return &PSKAuthMiddleware{preSharedKey: preSharedKey, logger: logger}
}

// Middleware func for PSK authentication to be used with fiber
func (p *PSKAuthMiddleware) Middleware(c *fiber.Ctx) error {
	// Get the Authorization header
	authHeader := c.Get("Authorization")

	// Check if the header exists and matches the PSK
	expected := "PSK " + p.preSharedKey
	if strings.TrimSpace(authHeader) != expected {
		p.logger.Warn().Msgf("PSK auth header is invalid. Received Header %s, Expected Header %s", authHeader, p.preSharedKey)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: Invalid PSK. " + authHeader,
		})
	}

	// Proceed to the next handler
	return c.Next()
}
