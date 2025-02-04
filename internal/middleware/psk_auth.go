package middleware

import "github.com/gofiber/fiber/v2"

type PSKAuthMiddleware struct {
	preSharedKey string
}

func NewPSKAuthMiddleware(preSharedKey string) *PSKAuthMiddleware {
	return &PSKAuthMiddleware{preSharedKey: preSharedKey}
}

// Middleware func for PSK authentication to be used with fiber
func (p *PSKAuthMiddleware) Middleware(c *fiber.Ctx) error {
	// Get the Authorization header
	authHeader := c.Get("Authorization")

	// Check if the header exists and matches the PSK
	if authHeader != "PSK "+p.preSharedKey {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: Invalid PSK. " + authHeader,
		})
	}

	// Proceed to the next handler
	return c.Next()
}
