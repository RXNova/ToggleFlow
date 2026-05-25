package api

import (
	"context"
	"time"

	"github.com/RXNova/ToggleFlow/internal/auth"
	"github.com/RXNova/ToggleFlow/internal/db"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// SetupStatus returns whether the system has been initialized.
// The frontend calls this on load to decide whether to show setup or login.
func (h *handler) SetupStatus(c *fiber.Ctx) error {
	count, err := h.db.NewSelect().Model((*db.User)(nil)).Count(context.Background())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db error"})
	}
	return c.JSON(fiber.Map{"initialized": count > 0})
}

type setupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Locale   string `json:"locale"`
}

// Setup creates the first superuser. Fails if any user already exists.
func (h *handler) Setup(c *fiber.Ctx) error {
	count, err := h.db.NewSelect().Model((*db.User)(nil)).Count(context.Background())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db error"})
	}
	if count > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "already initialized"})
	}

	var req setupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name, email and password are required"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to hash password"})
	}

	locale := req.Locale
	if locale != "en" && locale != "de" {
		locale = "en"
	}

	user := &db.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         db.RoleSuperuser,
		Locale:       locale,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if _, err := h.db.NewInsert().Model(user).Exec(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create user"})
	}

	token, err := auth.GenerateToken(user.ID, user.Role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate token"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"token": token, "user": user})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login validates credentials and returns a JWT token.
func (h *handler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	var user db.User
	err := h.db.NewSelect().Model(&user).Where("email = ?", req.Email).Scan(context.Background())
	if err != nil {
		// Return same error for wrong email and wrong password to avoid user enumeration
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	token, err := auth.GenerateToken(user.ID, user.Role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate token"})
	}

	return c.JSON(fiber.Map{"token": token, "user": user})
}

type updateProfileRequest struct {
	Name   string `json:"name"`
	Locale string `json:"locale"`
}

// UpdateProfile lets a user update their own name and locale preference.
func (h *handler) UpdateProfile(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)

	var req updateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	var user db.User
	if err := h.db.NewSelect().Model(&user).Where("id = ?", claims.UserID).Scan(context.Background()); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Locale == "en" || req.Locale == "de" {
		user.Locale = req.Locale
	}
	user.UpdatedAt = time.Now()

	if _, err := h.db.NewUpdate().Model(&user).WherePK().Exec(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update profile"})
	}

	return c.JSON(user)
}

// Me returns the currently authenticated user.
func (h *handler) Me(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)

	var user db.User
	if err := h.db.NewSelect().Model(&user).Where("id = ?", claims.UserID).Scan(context.Background()); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	return c.JSON(user)
}
