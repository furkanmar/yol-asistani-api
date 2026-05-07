package route

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/furkan/yol-asistani-api/internal/util"
)

// Handler holds HTTP handlers for route endpoints.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GET /api/v1/routes/alternatives
// Query: origin_lat, origin_lng, dest_lat, dest_lng, alternatives
func (h *Handler) Alternatives(c *fiber.Ctx) error {
	var req AlternativesRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(400).JSON(util.Err("invalid query params: " + err.Error()))
	}

	if req.OriginLat == 0 || req.OriginLng == 0 || req.DestLat == 0 || req.DestLng == 0 {
		return c.Status(400).JSON(util.Err("origin_lat, origin_lng, dest_lat, dest_lng required"))
	}

	tier := tierFromCtx(c)

	// Enforce tier limit — over-requesting returns 403
	if req.Alternatives > 0 {
		limit := tierAlternativesLimit[tier]
		if limit == 0 {
			limit = 1
		}
		if req.Alternatives > limit {
			return c.Status(403).JSON(util.Err(
				fmt.Sprintf("tier %q allows max %d alternative(s)", tier, limit),
			))
		}
	}

	resp, err := h.svc.Alternatives(c.Context(), req, tier)
	if err != nil {
		return c.Status(502).JSON(util.Err("routing failed: " + err.Error()))
	}
	return c.JSON(util.OK(resp))
}

// GET /api/v1/routes/segment
// Query: waypoints ("lat1,lng1;lat2,lng2;..."), alternatives
func (h *Handler) Segment(c *fiber.Ctx) error {
	var req SegmentRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(400).JSON(util.Err("invalid query params: " + err.Error()))
	}

	if req.Waypoints == "" {
		return c.Status(400).JSON(util.Err("waypoints required"))
	}

	tier := tierFromCtx(c)
	resp, err := h.svc.Segment(c.Context(), req, tier)
	if err != nil {
		return c.Status(400).JSON(util.Err(err.Error()))
	}
	return c.JSON(util.OK(resp))
}

// tierFromCtx extracts subscription tier set by JWTProtected middleware.
func tierFromCtx(c *fiber.Ctx) string {
	tier, ok := c.Locals("tier").(string)
	if !ok || tier == "" {
		return "free"
	}
	return tier
}
