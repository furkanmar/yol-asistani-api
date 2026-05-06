package auth

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/furkan/yol-asistani-api/config"
	"github.com/furkan/yol-asistani-api/internal/util"
)

var tierRank = map[string]int{
	"free":     0,
	"pro":      1,
	"pro_plus": 2,
}

func JWTProtected(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(401).JSON(util.Err("missing or invalid authorization header"))
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(401).JSON(util.Err("invalid or expired token"))
		}

		c.Locals("userID", claims.UserID)
		c.Locals("tier", claims.SubscriptionTier)
		return c.Next()
	}
}

func RequireTier(minTier string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tier, ok := c.Locals("tier").(string)
		if !ok {
			tier = "free"
		}
		if tierRank[tier] < tierRank[minTier] {
			return c.Status(403).JSON(util.Err("subscription_required"))
		}
		return c.Next()
	}
}
