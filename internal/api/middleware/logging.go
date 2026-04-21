package middleware

import (
	"io"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberlog "github.com/gofiber/fiber/v2/middleware/logger"
)

func NewLogger(output io.Writer) func(*fiber.Ctx) error {
	return fiberlog.New(fiberlog.Config{
		Format:     `{"time":"${time}","method":"${method}","path":"${path}","status":${status},"latency":"${latency}","ip":"${ip}","request_id":"${locals:requestid}"}` + "\n",
		TimeFormat: time.RFC3339,
		Output:     output,
	})
}
