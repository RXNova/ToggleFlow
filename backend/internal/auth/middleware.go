package auth

import (
	"strings"

	"toggleflow/internal/db"

	"github.com/gofiber/fiber/v2"
)

const claimsKey = "claims"

// Require is a Fiber middleware — like an Angular AuthGuard or NestJS AuthGuard.
// It validates the JWT and attaches the claims to the request context.
func Require(c *fiber.Ctx) error {
	header := c.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing token"})
	}

	claims, err := ParseToken(strings.TrimPrefix(header, "Bearer "))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
	}

	c.Locals(claimsKey, claims)
	return c.Next()
}

// RequireRole returns middleware that enforces a minimum role.
// e.g. RequireRole(db.RoleAdmin) blocks anyone below Admin.
func RequireRole(minimum db.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := GetClaims(c)
		if claims == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthenticated"})
		}
		if db.RoleRank(claims.Role) < db.RoleRank(minimum) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "insufficient permissions"})
		}
		return c.Next()
	}
}

func GetClaims(c *fiber.Ctx) *Claims {
	claims, _ := c.Locals(claimsKey).(*Claims)
	return claims
}
