package database

import (
	"context"
	"fmt"
	"time"

	"github.com/eprac/eeip-backend/internal/domain/models"
	"github.com/jmoiron/sqlx"
)

type EmailRepository interface {
	SaveEmail(ctx context.Context, email *models.Email) error
	EmailExists(ctx context.Context, accountID, senderEmail, subject string, receivedAt time.Time) (bool, error)
	GetEmailsByAccount(ctx context.Context, accountID string, userID string, limit, offset int) ([]models.Email, error)
	GetImportantEmails(ctx context.Context, userID string, limit int) ([]models.Email, error)
	GetGlobalInbox(ctx context.Context, userID string, limit int) ([]models.Email, error)
	UpdateEmailStatus(ctx context.Context, emailID string, status string) error
	GetEmailByID(ctx context.Context, emailID string) (*models.Email, error)
	UpdateEmailSummary(ctx context.Context, emailID string, summary string) error
	GetAlertEmails(ctx context.Context, since time.Time) ([]models.Email, error)
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

func (r *emailRepository) EmailExists(ctx context.Context, accountID, senderEmail, subject string, receivedAt time.Time) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(
		SELECT 1 FROM emails 
		WHERE account_id = $1 AND sender_email = $2 AND subject = $3 AND received_at = $4
	)`
	err := r.db.QueryRowContext(ctx, query, accountID, senderEmail, subject, receivedAt).Scan(&exists)
	return exists, err
}

func (r *emailRepository) GetEmailsByAccount(ctx context.Context, accountID string, userID string, limit, offset int) ([]models.Email, error) {
	var emails []models.Email
	query := `SELECT e.* FROM emails e JOIN email_accounts a ON e.account_id = a.id WHERE e.account_id = $1`
	args := []interface{}{accountID}
	
	if userID != "" {
		query += ` AND a.user_id = $2`
		args = append(args, userID)
	}
	
	query += ` ORDER BY e.received_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)
	
	err := r.db.SelectContext(ctx, &emails, query, args...)
	return emails, err
}

func (r *emailRepository) GetImportantEmails(ctx context.Context, userID string, limit int) ([]models.Email, error) {
	var emails []models.Email
	query := `SELECT e.*, COALESCE(NULLIF(a.account_name, ''), a.email_address) as monitored_account
	          FROM emails e
	          JOIN email_accounts a ON e.account_id = a.id
	          WHERE (
	              e.priority IN ('Critical', 'High', 'Crítico', 'Alto') 
	              OR e.requires_action = true 
	              OR e.sentiment IN ('Insatisfecho', 'Molesto', 'Frustrado', 'Preocupado', 'Critico', 'Escalado', 'Amenazante', 'Agresivo/violento', 'Peligro')
	              OR e.customer_risk_score > 50
	              OR e.escalation_risk_score > 50
	          )
	          AND e.category NOT IN ('Ruido', 'Informativo')
	          AND (e.status != 'Actioned' OR (e.status = 'Actioned' AND DATE(e.updated_at) = CURRENT_DATE))`
	
	args := []interface{}{}
	if userID != "" {
		query += ` AND a.user_id = $1`
		args = append(args, userID)
	}
	
	query += ` ORDER BY e.received_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit)
	
	err := r.db.SelectContext(ctx, &emails, query, args...)
	return emails, err
}

func (r *emailRepository) GetGlobalInbox(ctx context.Context, userID string, limit int) ([]models.Email, error) {
	var emails []models.Email
	query := `SELECT e.*, COALESCE(NULLIF(a.account_name, ''), a.email_address) as monitored_account
	          FROM emails e
	          JOIN email_accounts a ON e.account_id = a.id`
	
	args := []interface{}{}
	if userID != "" {
		query += ` WHERE a.user_id = $1`
		args = append(args, userID)
	}
	
	query += ` ORDER BY e.received_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit)
	
	err := r.db.SelectContext(ctx, &emails, query, args...)
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

func (r *emailRepository) GetAlertEmails(ctx context.Context, since time.Time) ([]models.Email, error) {
	var emails []models.Email
	query := `
		SELECT * FROM emails 
		WHERE received_at >= $1 
		AND status != 'Responded'
		AND (
			priority IN ('Crítico', 'Critical', 'Urgente', 'Urgent') 
			OR sentiment IN ('Peligro', 'Amenazante', 'Agresivo/violento')
		)
		ORDER BY received_at DESC
	`
	err := r.db.SelectContext(ctx, &emails, query, since)
	return emails, err
}
