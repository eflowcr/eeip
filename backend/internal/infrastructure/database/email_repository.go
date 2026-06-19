package database

import (
	"context"
	"github.com/eprac/eeip-backend/internal/domain/models"
	"github.com/jmoiron/sqlx"
)

type EmailRepository interface {
	SaveEmail(ctx context.Context, email *models.Email) error
	GetEmailsByAccount(ctx context.Context, accountID string, limit, offset int) ([]models.Email, error)
	GetImportantEmails(ctx context.Context, limit int) ([]models.Email, error)
	GetGlobalInbox(ctx context.Context, limit int) ([]models.Email, error)
}

type emailRepository struct {
	db *sqlx.DB
}

func NewEmailRepository(db *sqlx.DB) EmailRepository {
	return &emailRepository{db: db}
}

func (r *emailRepository) SaveEmail(ctx context.Context, email *models.Email) error {
	query := `
		INSERT INTO emails (
			account_id, message_id, thread_id, sender_email, sender_name,
			recipient_emails, subject, body_text, body_html, received_at,
			category, priority, requires_action, requires_approval, is_delegable, deadline,
			sentiment, sentiment_score, dissatisfaction_score, escalation_risk_score,
			customer_risk_score, detected_tone, recommended_action, ai_confidence_score,
			classification_explanation, status, suggested_assignee
		) VALUES (
			:account_id, :message_id, :thread_id, :sender_email, :sender_name,
			:recipient_emails, :subject, :body_text, :body_html, :received_at,
			:category, :priority, :requires_action, :requires_approval, :is_delegable, :deadline,
			:sentiment, :sentiment_score, :dissatisfaction_score, :escalation_risk_score,
			:customer_risk_score, :detected_tone, :recommended_action, :ai_confidence_score,
			:classification_explanation, :status, :suggested_assignee
		) RETURNING id, created_at, updated_at
	`
	
	rows, err := r.db.NamedQueryContext(ctx, query, email)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&email.ID, &email.CreatedAt, &email.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *emailRepository) GetEmailsByAccount(ctx context.Context, accountID string, limit, offset int) ([]models.Email, error) {
	var emails []models.Email
	query := `SELECT * FROM emails WHERE account_id = $1 ORDER BY received_at DESC LIMIT $2 OFFSET $3`
	err := r.db.SelectContext(ctx, &emails, query, accountID, limit, offset)
	return emails, err
}

func (r *emailRepository) GetImportantEmails(ctx context.Context, limit int) ([]models.Email, error) {
	var emails []models.Email
	query := `SELECT * FROM emails WHERE (priority IN ('Critical', 'High') OR requires_action = true) AND category NOT IN ('Ruido', 'Informativo') ORDER BY received_at DESC LIMIT $1`
	err := r.db.SelectContext(ctx, &emails, query, limit)
	return emails, err
}

func (r *emailRepository) GetGlobalInbox(ctx context.Context, limit int) ([]models.Email, error) {
	var emails []models.Email
	query := `SELECT * FROM emails ORDER BY received_at DESC LIMIT $1`
	err := r.db.SelectContext(ctx, &emails, query, limit)
	return emails, err
}
