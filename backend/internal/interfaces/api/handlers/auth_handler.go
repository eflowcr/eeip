package handlers

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/emersion/go-imap/client"
	"github.com/eprac/eeip-backend/internal/application/services"
	"github.com/eprac/eeip-backend/internal/domain/models"
	"github.com/eprac/eeip-backend/internal/infrastructure/auth"
	"github.com/eprac/eeip-backend/internal/infrastructure/database"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type AuthHandler struct {
	userRepo     database.UserRepository
	accountRepo  database.AccountRepository
	tokenManager auth.TokenManager
	collector    services.EmailCollector
}

func NewAuthHandler(userRepo database.UserRepository, accountRepo database.AccountRepository, tokenManager auth.TokenManager, collector services.EmailCollector) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, accountRepo: accountRepo, tokenManager: tokenManager, collector: collector}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email        string `json:"email" binding:"required"`
	Password     string `json:"password" binding:"required"`
	Role         string `json:"role" binding:"required"` // Admin, Auditor, Normal
	AccountName  string `json:"account_name" binding:"required"`
	IMAPHost     string `json:"imap_host" binding:"required"`
	IMAPPort     int    `json:"imap_port" binding:"required"`
	IMAPUser     string `json:"imap_user" binding:"required"`
	IMAPPassword string `json:"imap_password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	tokenString, err := h.tokenManager.GenerateToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user":  user,
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Validate IMAP connection before proceeding
	var cIMAP *client.Client
	var err error
	if req.IMAPPort == 993 {
		tlsConfig := &tls.Config{InsecureSkipVerify: true}
		cIMAP, err = client.DialTLS(fmt.Sprintf("%s:%d", req.IMAPHost, req.IMAPPort), tlsConfig)
	} else {
		cIMAP, err = client.Dial(fmt.Sprintf("%s:%d", req.IMAPHost, req.IMAPPort))
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to connect to email server", "details": err.Error()})
		return
	}
	defer cIMAP.Logout()

	if err := cIMAP.Login(req.IMAPUser, req.IMAPPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email credentials", "details": err.Error()})
		return
	}

	// 2. Hash EEIP password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// 3. Create User
	user := &database.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         req.Role,
	}

	if err := h.userRepo.CreateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// 4. Create Account and link to User
	account := &models.EmailAccount{
		UserID:       user.ID,
		EmailAddress: req.Email,
		AccountName:  req.AccountName,
		IMAPHost:     req.IMAPHost,
		IMAPPort:     req.IMAPPort,
		IMAPUser:     req.IMAPUser,
		IMAPPassword: req.IMAPPassword,
		IsActive:     true,
	}

	if err := h.accountRepo.CreateAccount(c.Request.Context(), account); err != nil {
		// Log error but don't fail user creation, or delete user if strict
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User created but failed to save email account details"})
		return
	}

	// 5. Trigger initial sync
	go func() {
		h.collector.CollectEmails(context.Background(), account)
	}()

	c.JSON(http.StatusCreated, user)
}
