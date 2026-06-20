CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'Executive Viewer', -- Super Admin, Tenant Admin, Executive Viewer, etc.
    company_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS email_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    email_address VARCHAR(255) NOT NULL,
    account_name VARCHAR(255) DEFAULT '',
    imap_host VARCHAR(255) NOT NULL,
    imap_port INT NOT NULL,
    imap_user VARCHAR(255) NOT NULL,
    imap_password VARCHAR(255) NOT NULL,
    last_sync_date TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS emails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID REFERENCES email_accounts(id) ON DELETE CASCADE,
    message_id VARCHAR(512) NOT NULL,
    thread_id VARCHAR(512),
    sender_email VARCHAR(255) NOT NULL,
    sender_name VARCHAR(255),
    recipient_emails JSONB,
    subject TEXT,
    body_text TEXT,
    body_html TEXT,
    received_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Classification
    category VARCHAR(100),
    priority VARCHAR(50),
    requires_action BOOLEAN DEFAULT FALSE,
    requires_approval BOOLEAN DEFAULT FALSE,
    is_delegable BOOLEAN DEFAULT FALSE,
    deadline TIMESTAMP WITH TIME ZONE,
    
    -- Sentiment Analysis
    sentiment VARCHAR(50),
    sentiment_score INT,
    dissatisfaction_score INT,
    escalation_risk_score INT,
    customer_risk_score INT,
    detected_tone TEXT,
    recommended_action TEXT,
    ai_confidence_score FLOAT,
    classification_explanation TEXT,
    
    -- Status
    status VARCHAR(50) DEFAULT 'Unread', -- Unread, Read, Actioned, Ignored
    suggested_assignee UUID REFERENCES users(id),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS commitments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email_id UUID REFERENCES emails(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    responsible VARCHAR(255),
    deadline TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'Pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS risks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email_id UUID REFERENCES emails(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    risk_level VARCHAR(50),
    status VARCHAR(50) DEFAULT 'Open',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_emails_account_id ON emails(account_id);
CREATE INDEX idx_emails_sender_email ON emails(sender_email);
CREATE INDEX idx_emails_received_at ON emails(received_at);
CREATE INDEX idx_emails_category ON emails(category);
CREATE INDEX idx_emails_priority ON emails(priority);
CREATE INDEX idx_emails_sentiment ON emails(sentiment);
