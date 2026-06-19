package handlers

import (
	"net/http"
	"strconv"

	"github.com/eprac/eeip-backend/internal/infrastructure/database"
	"github.com/gin-gonic/gin"
)

type EmailHandler struct {
	repo database.EmailRepository
}

func NewEmailHandler(repo database.EmailRepository) *EmailHandler {
	return &EmailHandler{repo: repo}
}

func (h *EmailHandler) GetImportantEmails(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	emails, err := h.repo.GetImportantEmails(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch emails", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, emails)
}

func (h *EmailHandler) GetGlobalInbox(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	emails, err := h.repo.GetGlobalInbox(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch emails", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, emails)
}

func (h *EmailHandler) GetEmailsByAccount(c *gin.Context) {
	accountID := c.Param("accountId")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	emails, err := h.repo.GetEmailsByAccount(c.Request.Context(), accountID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch emails", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, emails)
}
