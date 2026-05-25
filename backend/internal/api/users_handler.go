package api

import (
	"context"
	"time"

	"github.com/RXNova/ToggleFlow/internal/auth"
	"github.com/RXNova/ToggleFlow/internal/db"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// ListUsers returns all users. Admin and above only.
func (h *handler) ListUsers(c *fiber.Ctx) error {
	var users []db.User
	if err := h.db.NewSelect().Model(&users).OrderExpr("created_at ASC").Scan(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db error"})
	}
	return c.JSON(users)
}

type createUserRequest struct {
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Role     db.Role `json:"role"`
}

// CreateUser creates a new user. Only Admin+ can call this.
// Superuser is the only role that can create another Superuser or Admin.
func (h *handler) CreateUser(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)

	var req createUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Name == "" || req.Email == "" || req.Password == "" || req.Role == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name, email, password and role are required"})
	}

	// Only superuser can grant superuser or admin roles
	if db.RoleRank(req.Role) >= db.RoleRank(db.RoleAdmin) && claims.Role != db.RoleSuperuser {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "only superuser can assign admin or superuser roles"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to hash password"})
	}

	user := &db.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         req.Role,
		CreatedBy:    &claims.UserID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if _, err := h.db.NewInsert().Model(user).Exec(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create user"})
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

type updateUserRequest struct {
	Name  string  `json:"name"`
	Role  db.Role `json:"role"`
}

// UpdateUser updates a user's name or role. Superuser-only for role changes.
func (h *handler) UpdateUser(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	targetID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	var target db.User
	if err := h.db.NewSelect().Model(&target).Where("id = ?", targetID).Scan(context.Background()); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	// Cannot modify a user with equal or higher rank
	if db.RoleRank(target.Role) >= db.RoleRank(claims.Role) && int64(targetID) != claims.UserID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "cannot modify a user with equal or higher role"})
	}

	var req updateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	if req.Name != "" {
		target.Name = req.Name
	}
	if req.Role != "" {
		if claims.Role != db.RoleSuperuser {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "only superuser can change roles"})
		}
		target.Role = req.Role
	}
	target.UpdatedAt = time.Now()

	if _, err := h.db.NewUpdate().Model(&target).WherePK().Exec(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update user"})
	}

	return c.JSON(target)
}

// DeleteUser deletes a user. Superuser only.
func (h *handler) DeleteUser(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	targetID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	// Superuser cannot delete themselves
	if int64(targetID) == claims.UserID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot delete your own account"})
	}

	if _, err := h.db.NewDelete().Model((*db.User)(nil)).Where("id = ?", targetID).Exec(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete user"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
