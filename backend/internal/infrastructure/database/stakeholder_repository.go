package database

import (
	"log"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/eprac/eeip-backend/internal/domain/models"
)

type StakeholderRepository struct {
	db *sqlx.DB
}

func NewStakeholderRepository(db *sqlx.DB) *StakeholderRepository {
	return &StakeholderRepository{db: db}
}

func (r *StakeholderRepository) Create(s *models.Stakeholder) error {
	s.ID = uuid.New().String()
	query := `
		INSERT INTO stakeholders (id, name, email, telegram_chat_id, created_at)
		VALUES (:id, :name, :email, :telegram_chat_id, CURRENT_TIMESTAMP)
	`
	_, err := r.db.NamedExec(query, s)
	if err != nil {
		log.Printf("Error creating stakeholder: %v", err)
		return err
	}
	return nil
}

func (r *StakeholderRepository) GetAll() ([]models.Stakeholder, error) {
	var stakeholders []models.Stakeholder
	query := `SELECT id, name, email, telegram_chat_id, created_at FROM stakeholders ORDER BY created_at DESC`
	err := r.db.Select(&stakeholders, query)
	if err != nil {
		log.Printf("Error getting all stakeholders: %v", err)
		return nil, err
	}
	return stakeholders, nil
}

func (r *StakeholderRepository) Delete(id string) error {
	query := `DELETE FROM stakeholders WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
