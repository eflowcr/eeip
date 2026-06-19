package handlers

import (
	"net/http"
	"strconv"

	"github.com/eprac/eeip-backend/internal/application/services"
	"github.com/eprac/eeip-backend/internal/infrastructure/database"
	"github.com/gin-gonic/gin"
)

type EmailHandler struct {
	repo       database.EmailRepository
	summaryEng services.SummaryEngine
}

func NewEmailHandler(repo database.EmailRepository, summaryEng services.SummaryEngine) *EmailHandler {
	return &EmailHandler{repo: repo, summaryEng: summaryEng}
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

func (h *EmailHandler) UpdateEmailStatus(c *gin.Context) {
	emailID := c.Param("emailId")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdateEmailStatus(c.Request.Context(), emailID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update email status", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email status updated"})
}

func (h *EmailHandler) GenerateSummary(c *gin.Context) {
	emailID := c.Param("emailId")

	email, err := h.repo.GetEmailByID(c.Request.Context(), emailID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Email not found"})
		return
	}

	if email.Summary != nil && *email.Summary != "" {
		c.JSON(http.StatusOK, gin.H{"summary": *email.Summary})
		return
	}

	body := ""
	if email.BodyText != nil {
		body = *email.BodyText
	} else if email.BodyHTML != nil {
		body = *email.BodyHTML
	}

	summary, err := h.summaryEng.GenerateEmailSummary(c.Request.Context(), body)
	if err != nil {
		// Log error but return a graceful fallback if OpenAI fails
		fallback := "No se pudo generar el resumen automáticamente debido a un error del sistema."
		c.JSON(http.StatusOK, gin.H{"summary": fallback, "error": err.Error()})
		return
	}

	_ = h.repo.UpdateEmailSummary(c.Request.Context(), emailID, summary)

	c.JSON(http.StatusOK, gin.H{"summary": summary})
}
