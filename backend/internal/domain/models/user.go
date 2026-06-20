package models

import (
	"time"
)

type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	CompanyID    *string   `json:"company_id" db:"company_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type EmailAccount struct {
	ID            string     `json:"id" db:"id"`
	UserID        string     `json:"user_id" db:"user_id"`
	EmailAddress  string     `json:"email_address" db:"email_address"`
	AccountName   string     `json:"account_name" db:"account_name"`
	IMAPHost      string     `json:"imap_host" db:"imap_host"`
	IMAPPort      int        `json:"imap_port" db:"imap_port"`
	IMAPUser      string     `json:"imap_user" db:"imap_user"`
	IMAPPassword  string     `json:"imap_password,omitempty" db:"imap_password"`
	LastSyncDate  *time.Time `json:"last_sync_date" db:"last_sync_date"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}
