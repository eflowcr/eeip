ALTER TABLE email_accounts DROP COLUMN IF EXISTS account_name;
ALTER TABLE email_accounts DROP COLUMN IF EXISTS is_private;
ALTER TABLE emails DROP COLUMN IF EXISTS is_replied;

DROP INDEX IF EXISTS unique_email;
