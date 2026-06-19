package database

import (
	"context"
	"github.com/eprac/eeip-backend/internal/domain/models"
	"github.com/jmoiron/sqlx"
)

type AccountRepository interface {
	CreateAccount(ctx context.Context, account *models.EmailAccount) error
	GetAccounts(ctx context.Context) ([]models.EmailAccount, error)
}

type accountRepository struct {
	db *sqlx.DB
}

func NewAccountRepository(db *sqlx.DB) AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) CreateAccount(ctx context.Context, account *models.EmailAccount) error {
	query := `
		INSERT INTO email_accounts (
			user_id, email_address, imap_host, imap_port, imap_user, imap_password
		) VALUES (
			:user_id, :email_address, :imap_host, :imap_port, :imap_user, :imap_password
		) RETURNING id, created_at, updated_at
	`
	rows, err := r.db.NamedQueryContext(ctx, query, account)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)
	}
	return nil
}

func (r *accountRepository) GetAccounts(ctx context.Context) ([]models.EmailAccount, error) {
	var accounts []models.EmailAccount
	query := `SELECT * FROM email_accounts ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &accounts, query)
	return accounts, err
}
