package trip

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/furkan/yol-asistani-api/internal/util"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	trips, total, err := h.svc.List(c.Context(), userID, page, limit)
	if err != nil {
		return c.Status(500).JSON(util.Err("failed to list trips"))
	}
	if trips == nil {
		trips = []Trip{}
	}

	return c.JSON(util.OKWithMeta(trips, util.Meta{
		Page:  page,
		Limit: limit,
		Total: total,
	}))
}

func (h *Handler) Get(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	id := c.Params("id")

	trip, err := h.svc.Get(c.Context(), id, userID)
	if errors.Is(err, ErrNotFound) {
		return c.Status(404).JSON(util.Err("trip not found"))
	}
	if err != nil {
		return c.Status(500).JSON(util.Err("failed to get trip"))
	}

	return c.JSON(util.OK(trip))
}

func (h *Handler) Create(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	tier, _ := c.Locals("tier").(string)

	var req CreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(util.Err("invalid request body"))
	}

	trip, err := h.svc.Create(c.Context(), userID, tier, req)
	if errors.Is(err, ErrTripLimitReached) {
		return c.Status(403).JSON(util.Err("free tier limit: max 3 trips. Upgrade to Pro."))
	}
	if err != nil {
		return c.Status(500).JSON(util.Err(err.Error()))
	}

	return c.Status(201).JSON(util.OK(trip))
}

func (h *Handler) Update(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	id := c.Params("id")

	var req UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(util.Err("invalid request body"))
	}

	trip, err := h.svc.Update(c.Context(), id, userID, req)
	if errors.Is(err, ErrNotFound) {
		return c.Status(404).JSON(util.Err("trip not found"))
	}
	if err != nil {
		return c.Status(500).JSON(util.Err("failed to update trip"))
	}

	return c.JSON(util.OK(trip))
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	id := c.Params("id")

	if err := h.svc.Delete(c.Context(), id, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(404).JSON(util.Err("trip not found"))
		}
		return c.Status(500).JSON(util.Err("failed to delete trip"))
	}

	return c.JSON(util.OK("deleted"))
}
