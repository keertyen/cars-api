package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func NewRecovery() func(*fiber.Ctx) error {
	return recover.New(recover.Config{
		EnableStackTrace: true,
	})
}
