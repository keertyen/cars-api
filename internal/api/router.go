package api

import (
	"errors"
	"io"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	fiberlog "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/yenug1k/cars-api/config"
	"github.com/yenug1k/cars-api/internal/service"
)

func NewApp(svc service.Service, logOutput io.Writer, cfg *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		BodyLimit:    1 * 1024 * 1024, // 1 MB
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return c.Status(code).JSON(ErrorResponse{
				Error:   err.Error(),
				TraceID: traceID(c),
			})
		},
	})

	app.Use(requestid.New())

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	app.Use(fiberlog.New(fiberlog.Config{
		Format:     `{"time":"${time}","method":"${method}","path":"${path}","status":${status},"latency":"${latency}","ip":"${ip}","request_id":"${locals:requestid}"}` + "\n",
		TimeFormat: time.RFC3339,
		Output:     logOutput,
	}))

	app.Use(limiter.New(limiter.Config{
		Max:        cfg.RateLimitBurst,
		Expiration: time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			if ip := c.Get("X-Forwarded-For"); ip != "" {
				return ip
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(ErrorResponse{
				Error:   "rate limit exceeded",
				TraceID: traceID(c),
			})
		},
	}))

	app.Use(compress.New())

	h := NewHandler(svc)

	app.Get("/health", h.Health)
	app.Get("/ready", h.Health)

	v1 := app.Group("/v1")
	cars := v1.Group("/cars")
	cars.Post("/", h.CreateCar)
	cars.Get("/", h.ListCars)
	cars.Get("/:id", h.GetCar)
	cars.Put("/:id", h.UpdateCar)
	cars.Delete("/:id", h.DeleteCar)

	return app
}
