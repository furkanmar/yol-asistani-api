package auth

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

func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(util.Err("invalid request body"))
	}
	if req.Email == "" || req.Password == "" {
		return c.Status(400).JSON(util.Err("email and password required"))
	}
	if len(req.Password) < 8 {
		return c.Status(400).JSON(util.Err("password must be at least 8 characters"))
	}

	resp, err := h.svc.Register(c.Context(), req)
	if errors.Is(err, ErrEmailTaken) {
		return c.Status(409).JSON(util.Err("email already in use"))
	}
	if err != nil {
		return c.Status(500).JSON(util.Err("registration failed"))
	}

	return c.Status(201).JSON(util.OK(resp))
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(util.Err("invalid request body"))
	}

	resp, err := h.svc.Login(c.Context(), req)
	if errors.Is(err, ErrInvalidCreds) {
		return c.Status(401).JSON(util.Err("invalid email or password"))
	}
	if err != nil {
		return c.Status(500).JSON(util.Err("login failed"))
	}

	return c.JSON(util.OK(resp))
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(util.Err("invalid request body"))
	}

	resp, err := h.svc.Refresh(c.Context(), req.RefreshToken)
	if errors.Is(err, ErrInvalidToken) {
		return c.Status(401).JSON(util.Err("invalid or expired refresh token"))
	}
	if err != nil {
		return c.Status(500).JSON(util.Err("refresh failed"))
	}

	return c.JSON(util.OK(resp))
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(util.Err("invalid request body"))
	}

	h.svc.Logout(c.Context(), req.RefreshToken) // hata sessizce yut
	return c.JSON(util.OK("logged out"))
}

func (h *Handler) Me(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	user, err := h.svc.Me(c.Context(), userID)
	if err != nil {
		return c.Status(401).JSON(util.Err("user not found"))
	}
	return c.JSON(util.OK(user))
}
