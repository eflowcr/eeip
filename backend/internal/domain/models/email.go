package models

import (
	"encoding/json"
	"time"
)

type Email struct {
	ID               string          `json:"id" db:"id"`
	AccountID        string          `json:"account_id" db:"account_id"`
	MessageID        string          `json:"message_id" db:"message_id"`
	ThreadID         *string         `json:"thread_id" db:"thread_id"`
	SenderEmail      string          `json:"sender_email" db:"sender_email"`
	SenderName       *string         `json:"sender_name" db:"sender_name"`
	RecipientEmails  json.RawMessage `json:"recipient_emails" db:"recipient_emails"`
	Subject          *string         `json:"subject" db:"subject"`
	BodyText         *string         `json:"body_text" db:"body_text"`
	BodyHTML         *string         `json:"body_html" db:"body_html"`
	ReceivedAt       time.Time       `json:"received_at" db:"received_at"`

	Category         *string         `json:"category" db:"category"`
	Priority         *string         `json:"priority" db:"priority"`
	RequiresAction   bool            `json:"requires_action" db:"requires_action"`
	RequiresApproval bool            `json:"requires_approval" db:"requires_approval"`
	IsDelegable      bool            `json:"is_delegable" db:"is_delegable"`
	Deadline         *time.Time      `json:"deadline" db:"deadline"`

	Sentiment             *string `json:"sentiment" db:"sentiment"`
	SentimentScore        *int    `json:"sentiment_score" db:"sentiment_score"`
	DissatisfactionScore  *int    `json:"dissatisfaction_score" db:"dissatisfaction_score"`
	EscalationRiskScore   *int    `json:"escalation_risk_score" db:"escalation_risk_score"`
	CustomerRiskScore     *int    `json:"customer_risk_score" db:"customer_risk_score"`
	DetectedTone          *string `json:"detected_tone" db:"detected_tone"`
	RecommendedAction     *string `json:"recommended_action" db:"recommended_action"`
	AIConfidenceScore     *float64 `json:"ai_confidence_score" db:"ai_confidence_score"`
	ClassificationExpl    *string `json:"classification_explanation" db:"classification_explanation"`

	Status            string    `json:"status" db:"status"`
	SuggestedAssignee *string   `json:"suggested_assignee" db:"suggested_assignee"`
	MonitoredAccount  *string   `json:"monitored_account" db:"monitored_account"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

type Commitment struct {
	ID          string     `json:"id" db:"id"`
	EmailID     string     `json:"email_id" db:"email_id"`
	Description string     `json:"description" db:"description"`
	Responsible *string    `json:"responsible" db:"responsible"`
	Deadline    *time.Time `json:"deadline" db:"deadline"`
	Status      string     `json:"status" db:"status"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

type Risk struct {
	ID          string    `json:"id" db:"id"`
	EmailID     string    `json:"email_id" db:"email_id"`
	Description string    `json:"description" db:"description"`
	RiskLevel   *string   `json:"risk_level" db:"risk_level"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type Client struct {
	ID        string    `json:"id" db:"id"`
	Domain    string    `json:"domain" db:"domain"`
	Name      *string   `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
