package database

import (
	"context"
	"github.com/eprac/eeip-backend/internal/domain/models"
	"github.com/jmoiron/sqlx"
)

type AccountRepository interface {
	CreateAccount(ctx context.Context, account *models.EmailAccount) error
	GetAccounts(ctx context.Context) ([]models.EmailAccount, error)
	GetAccountByID(ctx context.Context, id string) (*models.EmailAccount, error)
	UpdateAccount(ctx context.Context, account *models.EmailAccount) error
	DeleteAccount(ctx context.Context, id string) error
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
			user_id, email_address, account_name, imap_host, imap_port, imap_user, imap_password
		) VALUES (
			:user_id, :email_address, :account_name, :imap_host, :imap_port, :imap_user, :imap_password
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

func (r *accountRepository) GetAccountByID(ctx context.Context, id string) (*models.EmailAccount, error) {
	var account models.EmailAccount
	query := `SELECT * FROM email_accounts WHERE id = $1`
	err := r.db.GetContext(ctx, &account, query, id)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *accountRepository) UpdateAccount(ctx context.Context, account *models.EmailAccount) error {
	query := `
		UPDATE email_accounts SET
			email_address = :email_address,
			account_name = :account_name,
			imap_host = :imap_host,
			imap_port = :imap_port,
			imap_user = :imap_user,
			imap_password = CASE WHEN :imap_password != '' THEN :imap_password ELSE imap_password END,
			updated_at = NOW()
		WHERE id = :id
	`
	_, err := r.db.NamedExecContext(ctx, query, account)
	return err
}

func (r *accountRepository) DeleteAccount(ctx context.Context, id string) error {
	query := `DELETE FROM email_accounts WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
