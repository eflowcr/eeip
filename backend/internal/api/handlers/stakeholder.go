package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/eflowcr/eeip/internal/domain/models"
	"github.com/eflowcr/eeip/internal/infrastructure/database"
)

type StakeholderHandler struct {
	repo *database.StakeholderRepository
}

func NewStakeholderHandler(repo *database.StakeholderRepository) *StakeholderHandler {
	return &StakeholderHandler{repo: repo}
}

func (h *StakeholderHandler) CreateStakeholder(c *gin.Context) {
	var stakeholder models.Stakeholder
	if err := c.ShouldBindJSON(&stakeholder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if err := h.repo.Create(&stakeholder); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create stakeholder"})
		return
	}

	c.JSON(http.StatusCreated, stakeholder)
}

func (h *StakeholderHandler) GetStakeholders(c *gin.Context) {
	stakeholders, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stakeholders"})
		return
	}
	if stakeholders == nil {
		stakeholders = []models.Stakeholder{}
	}
	c.JSON(http.StatusOK, stakeholders)
}

func (h *StakeholderHandler) DeleteStakeholder(c *gin.Context) {
	id := c.Param("id")
	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete stakeholder"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
