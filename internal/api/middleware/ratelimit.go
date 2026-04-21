package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/yenug1k/cars-api/config"
)

func NewRateLimiter(cfg *config.Config) func(*fiber.Ctx) error {
	return limiter.New(limiter.Config{
		Max:        cfg.RateLimitBurst,
		Expiration: time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			if ip := c.Get("X-Forwarded-For"); ip != "" {
				return ip
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			traceID, _ := c.Locals("requestid").(string)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":    "rate limit exceeded",
				"trace_id": traceID,
			})
		},
	})
}
