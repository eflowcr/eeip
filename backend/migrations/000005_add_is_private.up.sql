ALTER TABLE email_accounts ADD COLUMN IF NOT EXISTS account_name VARCHAR(255) DEFAULT '';
ALTER TABLE email_accounts ADD COLUMN IF NOT EXISTS is_private BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN IF NOT EXISTS is_replied BOOLEAN DEFAULT FALSE;

CREATE UNIQUE INDEX IF NOT EXISTS unique_email ON emails (account_id, sender_email, subject, received_at);
