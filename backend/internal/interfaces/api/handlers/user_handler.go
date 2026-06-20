package handlers

import (
	"net/http"

	"github.com/eprac/eeip-backend/internal/infrastructure/database"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userRepo database.UserRepository
}

func NewUserHandler(userRepo database.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.userRepo.GetAllUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

func (h *UserHandler) UpdateRole(c *gin.Context) {
	userID := c.Param("userId")
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validRoles := map[string]bool{"Admin": true, "Auditor": true, "Normal": true}
	if !validRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	if err := h.userRepo.UpdateUserRole(c.Request.Context(), userID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role updated successfully"})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("userId")
	if err := h.userRepo.DeleteUser(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
