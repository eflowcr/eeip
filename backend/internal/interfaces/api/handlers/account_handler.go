package handlers

import (
	"net/http"

	"github.com/eprac/eeip-backend/internal/domain/models"
	"github.com/eprac/eeip-backend/internal/infrastructure/database"
	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	repo database.AccountRepository
}

func NewAccountHandler(repo database.AccountRepository) *AccountHandler {
	return &AccountHandler{repo: repo}
}

func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req models.EmailAccount
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hardcode the default user ID for MVP
	req.UserID = "00000000-0000-0000-0000-000000000001"

	if err := h.repo.CreateAccount(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

func (h *AccountHandler) GetAccounts(c *gin.Context) {
	accounts, err := h.repo.GetAccounts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch accounts", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, accounts)
}
