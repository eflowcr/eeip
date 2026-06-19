package handlers

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/emersion/go-imap/client"
	"github.com/eprac/eeip-backend/internal/application/services"
	"github.com/eprac/eeip-backend/internal/domain/models"
	"github.com/eprac/eeip-backend/internal/infrastructure/database"
	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	repo      database.AccountRepository
	collector services.EmailCollector
}

func NewAccountHandler(repo database.AccountRepository, collector services.EmailCollector) *AccountHandler {
	return &AccountHandler{repo: repo, collector: collector}
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get accounts", "details": err.Error()})
		return
	}
	
	for i := range accounts {
		accounts[i].IMAPPassword = "" // Never send password back to frontend
	}
	
	c.JSON(http.StatusOK, accounts)
}

func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	accountId := c.Param("accountId")
	var req models.EmailAccount
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = accountId

	if err := h.repo.UpdateAccount(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account updated successfully"})
}

func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	accountId := c.Param("accountId")
	if err := h.repo.DeleteAccount(c.Request.Context(), accountId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}

func (h *AccountHandler) TestConnection(c *gin.Context) {
	var req models.EmailAccount
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var cIMAP *client.Client
	var err error
	if req.IMAPPort == 993 {
		tlsConfig := &tls.Config{InsecureSkipVerify: true}
		cIMAP, err = client.DialTLS(fmt.Sprintf("%s:%d", req.IMAPHost, req.IMAPPort), tlsConfig)
	} else {
		cIMAP, err = client.Dial(fmt.Sprintf("%s:%d", req.IMAPHost, req.IMAPPort))
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to connect to IMAP server", "details": err.Error()})
		return
	}
	defer cIMAP.Logout()

	if err := cIMAP.Login(req.IMAPUser, req.IMAPPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to login to IMAP server", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Connection successful"})
}

func (h *AccountHandler) TestExistingConnection(c *gin.Context) {
	accountId := c.Param("accountId")
	
	acc, err := h.repo.GetAccountByID(c.Request.Context(), accountId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found", "details": err.Error()})
		return
	}

	var cIMAP *client.Client
	if acc.IMAPPort == 993 {
		tlsConfig := &tls.Config{InsecureSkipVerify: true}
		cIMAP, err = client.DialTLS(fmt.Sprintf("%s:%d", acc.IMAPHost, acc.IMAPPort), tlsConfig)
	} else {
		cIMAP, err = client.Dial(fmt.Sprintf("%s:%d", acc.IMAPHost, acc.IMAPPort))
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to connect to IMAP server", "details": err.Error()})
		return
	}
	defer cIMAP.Logout()

	if err := cIMAP.Login(acc.IMAPUser, acc.IMAPPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to login to IMAP server", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Connection successful"})
}

func (h *AccountHandler) SyncAccount(c *gin.Context) {
	accountId := c.Param("accountId")
	
	acc, err := h.repo.GetAccountByID(c.Request.Context(), accountId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	// For demonstration, we run synchronously to give immediate feedback.
	if err := h.collector.CollectEmails(c.Request.Context(), acc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Sync failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account synchronized successfully"})
}
