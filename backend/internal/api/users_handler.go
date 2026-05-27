package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"toggleflow/internal/auth"
	"toggleflow/internal/db"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *handler) ListUsers(c *fiber.Ctx) error {
	var users []db.User
	if err := h.db.NewSelect().Model(&users).OrderExpr("created_at ASC").Scan(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db error"})
	}
	return c.JSON(users)
}

type createUserRequest struct {
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	Role       db.Role `json:"role"`
	ExpiryDays int     `json:"expiry_days"`
}

const tokenCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generateWelcomeToken returns a token in XXXX-XXXX-XXXX-XXXX-XXXX format
// (20 mixed-case alphanumeric chars, 4 hyphens) and its SHA-256 hash.
// Only the hash is stored; the plain token is shown to the inviter once.
func generateWelcomeToken() (plain, hashed string, err error) {
	b := make([]byte, 20)
	if _, err = rand.Read(b); err != nil {
		return
	}
	chars := make([]byte, 20)
	for i, v := range b {
		chars[i] = tokenCharset[int(v)%len(tokenCharset)]
	}
	plain = fmt.Sprintf("%s-%s-%s-%s-%s",
		chars[0:4], chars[4:8], chars[8:12], chars[12:16], chars[16:20])
	sum := sha256.Sum256([]byte(plain))
	hashed = hex.EncodeToString(sum[:])
	return
}

// CreateUser creates an inactive user account and returns a one-time welcome token.
// The invitee must use the token to set their password before they can log in.
func (h *handler) CreateUser(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)

	var req createUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Name == "" || req.Email == "" || req.Role == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name, email and role are required"})
	}

	if db.RoleRank(req.Role) >= db.RoleRank(db.RoleAdmin) && claims.Role != db.RoleSuperuser {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "only superuser can assign admin or superuser roles"})
	}

	expiryDays := req.ExpiryDays
	if expiryDays <= 0 {
		expiryDays = 7
	}

	plainToken, hashedToken, err := generateWelcomeToken()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate token"})
	}

	expiresAt := time.Now().Add(time.Duration(expiryDays) * 24 * time.Hour)
	user := &db.User{
		UUID:                  uuid.NewString(),
		Name:                  req.Name,
		Email:                 req.Email,
		PasswordHash:          "",
		Role:                  req.Role,
		WelcomeToken:          hashedToken,
		WelcomeTokenExpiresAt: &expiresAt,
		CreatedBy:             &claims.UserID,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if _, err := h.db.NewInsert().Model(user).Exec(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create user"})
	}

	h.writeAudit(0, h.actorName(c), "user.created", user.Email, "",
		toJSON(map[string]any{"name": user.Name, "email": user.Email, "role": string(user.Role)}))

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user":          user,
		"welcome_token": plainToken,
	})
}

type updateUserRequest struct {
	Name  string  `json:"name"`
	Email string  `json:"email"`
	Role  db.Role `json:"role"`
}

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

	if db.RoleRank(target.Role) >= db.RoleRank(claims.Role) && int64(targetID) != claims.UserID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "cannot modify a user with equal or higher role"})
	}

	var req updateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	oldName, oldEmail, oldRole := target.Name, target.Email, target.Role

	if req.Name != "" {
		target.Name = req.Name
	}
	if req.Email != "" && req.Email != target.Email {
		// Check uniqueness before updating
		exists, err := h.db.NewSelect().Model((*db.User)(nil)).
			Where("email = ? AND id != ?", req.Email, targetID).
			Exists(context.Background())
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db error"})
		}
		if exists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "email already in use"})
		}
		target.Email = req.Email
	}
	if req.Role != "" {
		if claims.Role != db.RoleSuperuser {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "only superuser can change roles"})
		}
		// A superuser cannot demote themselves — only another superuser can do that
		if int64(targetID) == claims.UserID && req.Role != db.RoleSuperuser {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "superusers cannot demote themselves — another superuser must do it"})
		}
		target.Role = req.Role
	}
	target.UpdatedAt = time.Now()

	if _, err := h.db.NewUpdate().Model(&target).WherePK().Exec(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update user"})
	}

	h.writeAudit(0, h.actorName(c), "user.updated", target.Email,
		toJSON(map[string]any{"name": oldName, "email": oldEmail, "role": string(oldRole)}),
		toJSON(map[string]any{"name": target.Name, "email": target.Email, "role": string(target.Role)}))

	return c.JSON(target)
}

type generateResetRequest struct {
	ExpiryDays int `json:"expiry_days"`
}

// GenerateResetLink creates a password-reset token for an existing activated user.
// Superuser only. Returns the plain token so the superuser can share it out-of-band.
func (h *handler) GenerateResetLink(c *fiber.Ctx) error {
	targetID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	var user db.User
	if err := h.db.NewSelect().Model(&user).Where("id = ?", targetID).Scan(context.Background()); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	if user.ActivatedAt == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user has not activated their account yet"})
	}

	var req generateResetRequest
	_ = c.BodyParser(&req)
	expiryDays := req.ExpiryDays
	if expiryDays <= 0 {
		expiryDays = 1
	}

	plainToken, hashedToken, err := generateWelcomeToken()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate token"})
	}

	expiresAt := time.Now().Add(time.Duration(expiryDays) * 24 * time.Hour)
	user.ResetToken = hashedToken
	user.ResetTokenExpiresAt = &expiresAt
	user.UpdatedAt = time.Now()

	if _, err := h.db.NewUpdate().Model(&user).WherePK().Exec(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate reset link"})
	}

	return c.JSON(fiber.Map{
		"user":        user,
		"reset_token": plainToken,
	})
}

type reinviteRequest struct {
	ExpiryDays int `json:"expiry_days"`
}

// ReinviteUser regenerates the welcome token for a user who hasn't activated yet.
func (h *handler) ReinviteUser(c *fiber.Ctx) error {
	targetID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	var user db.User
	if err := h.db.NewSelect().Model(&user).Where("id = ?", targetID).Scan(context.Background()); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	if user.ActivatedAt != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user has already activated their account"})
	}

	var req reinviteRequest
	_ = c.BodyParser(&req)
	expiryDays := req.ExpiryDays
	if expiryDays <= 0 {
		expiryDays = 7
	}

	plainToken, hashedToken, err := generateWelcomeToken()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate token"})
	}

	expiresAt := time.Now().Add(time.Duration(expiryDays) * 24 * time.Hour)
	user.WelcomeToken = hashedToken
	user.WelcomeTokenExpiresAt = &expiresAt
	user.UpdatedAt = time.Now()

	if _, err := h.db.NewUpdate().Model(&user).WherePK().Exec(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to regenerate invite"})
	}

	return c.JSON(fiber.Map{
		"user":          user,
		"welcome_token": plainToken,
	})
}

// GetInviteInfo returns the name and email for a pending (non-activated) invite.
// Public endpoint — used by the activate page to greet the invitee.
func (h *handler) GetInviteInfo(c *fiber.Ctx) error {
	id := c.Params("uuid")
	var user db.User
	err := h.db.NewSelect().Model(&user).
		Where("uuid = ? AND welcome_token != ''", id).
		Scan(context.Background())
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "invite not found"})
	}
	return c.JSON(fiber.Map{"name": user.Name, "email": user.Email})
}

func (h *handler) DeleteUser(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	targetID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	if int64(targetID) == claims.UserID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot delete your own account"})
	}

	ctx := context.Background()
	var target db.User
	if err := h.db.NewSelect().Model(&target).Where("id = ?", targetID).Scan(ctx); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	if _, err := h.db.NewDelete().Model((*db.User)(nil)).Where("id = ?", targetID).Exec(ctx); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete user"})
	}

	h.writeAudit(0, h.actorName(c), "user.deleted", target.Email,
		toJSON(map[string]any{"name": target.Name, "email": target.Email, "role": string(target.Role)}), "")

	return c.SendStatus(fiber.StatusNoContent)
}
