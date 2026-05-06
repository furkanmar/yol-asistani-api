package waypoint

import (
	"errors"

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
	tripID := c.Params("id")
	wps, err := h.svc.List(c.Context(), tripID)
	if err != nil {
		return c.Status(500).JSON(util.Err("failed to list waypoints"))
	}
	return c.JSON(util.OK(wps))
}

func (h *Handler) Create(c *fiber.Ctx) error {
	tripID := c.Params("id")

	var req CreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(util.Err("invalid request body"))
	}

	wp, err := h.svc.Create(c.Context(), tripID, req)
	if err != nil {
		return c.Status(500).JSON(util.Err(err.Error()))
	}
	return c.Status(201).JSON(util.OK(wp))
}

func (h *Handler) Update(c *fiber.Ctx) error {
	tripID := c.Params("id")
	wpID := c.Params("wid")

	var req UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(util.Err("invalid request body"))
	}

	wp, err := h.svc.Update(c.Context(), wpID, tripID, req)
	if errors.Is(err, ErrNotFound) {
		return c.Status(404).JSON(util.Err("waypoint not found"))
	}
	if err != nil {
		return c.Status(500).JSON(util.Err("failed to update waypoint"))
	}
	return c.JSON(util.OK(wp))
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	tripID := c.Params("id")
	wpID := c.Params("wid")

	if err := h.svc.Delete(c.Context(), wpID, tripID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(404).JSON(util.Err("waypoint not found"))
		}
		return c.Status(500).JSON(util.Err("failed to delete waypoint"))
	}
	return c.JSON(util.OK("deleted"))
}

func (h *Handler) Reorder(c *fiber.Ctx) error {
	tripID := c.Params("id")

	var items []ReorderItem
	if err := c.BodyParser(&items); err != nil {
		return c.Status(400).JSON(util.Err("invalid request body"))
	}

	if err := h.svc.Reorder(c.Context(), tripID, items); err != nil {
		return c.Status(500).JSON(util.Err(err.Error()))
	}
	return c.JSON(util.OK("reordered"))
}
