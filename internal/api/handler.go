package api

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/yenug1k/cars-api/internal/model"
	"github.com/yenug1k/cars-api/internal/service"
	"github.com/yenug1k/cars-api/internal/store"
)

type Handler struct {
	svc      service.Service
	validate *validator.Validate
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

func (h *Handler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *Handler) CreateCar(c *fiber.Ctx) error {
	var req model.CreateCarRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid request body",
			TraceID: traceID(c),
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "validation failed",
			Details: toFieldErrors(err),
			TraceID: traceID(c),
		})
	}

	car, err := h.svc.Create(c.UserContext(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "failed to create car",
			TraceID: traceID(c),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(car)
}

func (h *Handler) GetCar(c *fiber.Ctx) error {
	id := c.Params("id")

	car, err := h.svc.Get(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "car not found",
				TraceID: traceID(c),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "failed to get car",
			TraceID: traceID(c),
		})
	}

	return c.JSON(car)
}

func (h *Handler) ListCars(c *fiber.Ctx) error {
	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		n, err := strconv.Atoi(ps)
		if err != nil || n < 1 || n > 100 {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error:   "page_size must be between 1 and 100",
				TraceID: traceID(c),
			})
		}
		pageSize = n
	}

	q := model.ListCarsQuery{
		PageSize:  pageSize,
		PageToken: c.Query("page_token"),
	}

	resp, err := h.svc.List(c.UserContext(), q)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "failed to list cars",
			TraceID: traceID(c),
		})
	}

	return c.JSON(resp)
}

func (h *Handler) UpdateCar(c *fiber.Ctx) error {
	id := c.Params("id")

	var req model.UpdateCarRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid request body",
			TraceID: traceID(c),
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "validation failed",
			Details: toFieldErrors(err),
			TraceID: traceID(c),
		})
	}

	car, err := h.svc.Update(c.UserContext(), id, &req)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "car not found",
				TraceID: traceID(c),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "failed to update car",
			TraceID: traceID(c),
		})
	}

	return c.JSON(car)
}

func (h *Handler) DeleteCar(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "car not found",
				TraceID: traceID(c),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "failed to delete car",
			TraceID: traceID(c),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func toFieldErrors(err error) map[string]string {
	out := make(map[string]string)
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			out[fe.Field()] = fieldMessage(fe)
		}
	}
	return out
}

func fieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "field is required"
	case "min":
		return fmt.Sprintf("minimum value is %s", fe.Param())
	case "max":
		return fmt.Sprintf("maximum value is %s", fe.Param())
	case "gt":
		return fmt.Sprintf("must be greater than %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	case "len":
		return fmt.Sprintf("must be exactly %s characters", fe.Param())
	case "alphanum":
		return "must contain only alphanumeric characters"
	default:
		return fmt.Sprintf("failed validation: %s", fe.Tag())
	}
}
