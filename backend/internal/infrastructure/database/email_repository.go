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
	UpdateEmailStatus(ctx context.Context, emailID string, status string) error
	GetEmailByID(ctx context.Context, emailID string) (*models.Email, error)
	UpdateEmailSummary(ctx context.Context, emailID string, summary string) error
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
			classification_explanation, status, suggested_assignee, is_replied
		) VALUES (
			:account_id, :message_id, :thread_id, :sender_email, :sender_name,
			:recipient_emails, :subject, :body_text, :body_html, :received_at,
			:category, :priority, :requires_action, :requires_approval, :is_delegable, :deadline,
			:sentiment, :sentiment_score, :dissatisfaction_score, :escalation_risk_score,
			:customer_risk_score, :detected_tone, :recommended_action, :ai_confidence_score,
			:classification_explanation, :status, :suggested_assignee, :is_replied
		) 
		ON CONFLICT (account_id, sender_email, subject, received_at) 
		DO UPDATE SET is_replied = EXCLUDED.is_replied, updated_at = NOW()
		RETURNING id, created_at, updated_at
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
	query := `SELECT e.*, COALESCE(NULLIF(a.account_name, ''), a.email_address) as monitored_account
	          FROM emails e
	          JOIN email_accounts a ON e.account_id = a.id
	          WHERE (e.priority IN ('Critical', 'High') OR e.requires_action = true)
	          AND e.category NOT IN ('Ruido', 'Informativo')
	          AND (e.status != 'Actioned' OR (e.status = 'Actioned' AND DATE(e.updated_at) = CURRENT_DATE))
	          ORDER BY e.received_at DESC LIMIT $1`
	err := r.db.SelectContext(ctx, &emails, query, limit)
	return emails, err
}

func (r *emailRepository) GetGlobalInbox(ctx context.Context, limit int) ([]models.Email, error) {
	var emails []models.Email
	query := `SELECT e.*, COALESCE(NULLIF(a.account_name, ''), a.email_address) as monitored_account
	          FROM emails e
	          JOIN email_accounts a ON e.account_id = a.id
	          ORDER BY e.received_at DESC LIMIT $1`
	err := r.db.SelectContext(ctx, &emails, query, limit)
	return emails, err
}

func (r *emailRepository) UpdateEmailStatus(ctx context.Context, emailID string, status string) error {
	query := `UPDATE emails SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, emailID)
	return err
}

func (r *emailRepository) GetEmailByID(ctx context.Context, emailID string) (*models.Email, error) {
	var email models.Email
	query := `SELECT * FROM emails WHERE id = $1`
	err := r.db.GetContext(ctx, &email, query, emailID)
	return &email, err
}

func (r *emailRepository) UpdateEmailSummary(ctx context.Context, emailID string, summary string) error {
	query := `UPDATE emails SET summary = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, summary, emailID)
	return err
}
