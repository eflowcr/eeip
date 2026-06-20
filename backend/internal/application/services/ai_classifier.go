package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eprac/eeip-backend/internal/domain/models"
	"github.com/sashabaranov/go-openai"
)

type AIClassificationEngine interface {
	ClassifyEmail(ctx context.Context, email *models.Email) error
}

type aiClassificationEngine struct {
	client *openai.Client
}

func NewAIClassificationEngine(apiKey string) AIClassificationEngine {
	return &aiClassificationEngine{
		client: openai.NewClient(apiKey),
	}
}

type ClassificationResult struct {
	Category              string  `json:"category"`
	Priority              string  `json:"priority"`
	RequiresAction        bool    `json:"requires_action"`
	RequiresApproval      bool    `json:"requires_approval"`
	IsDelegable           bool    `json:"is_delegable"`
	Sentiment             string  `json:"sentiment"`
	SentimentScore        int     `json:"sentiment_score"`
	DissatisfactionScore  int     `json:"dissatisfaction_score"`
	EscalationRiskScore   int     `json:"escalation_risk_score"`
	CustomerRiskScore     int     `json:"customer_risk_score"`
	DetectedTone          string  `json:"detected_tone"`
	RecommendedAction     string  `json:"recommended_action"`
	ClassificationExpl    string  `json:"classification_explanation"`
}

func (s *aiClassificationEngine) ClassifyEmail(ctx context.Context, email *models.Email) error {
	prompt := fmt.Sprintf(`Analyze the following email from '%s' with subject '%s'.
Body: %s

Please classify it according to the EEIP platform rules and provide a JSON response with the following keys:
- category (Cliente, Prospecto, Soporte, Incidente, Comercial, Contrato, Facturacion, Finanzas, RH, Proveedor, Operacion, Despliegue, Seguridad, Informativo, Ruido)
- priority (Critical, High, Medium, Low)
- requires_action (boolean)
- requires_approval (boolean)
- is_delegable (boolean)
- sentiment (Neutral, Positivo, Preocupado, Molesto, Frustrado, Insatisfecho, Critico, Escalado)
- sentiment_score (0-100)
- dissatisfaction_score (0-100)
- escalation_risk_score (0-100)
- customer_risk_score (0-100)
- detected_tone (MUST be exactly one of: Optimista, Confrontativo, Agresivo/violento, Amenazante, Neutral, Profesional, Frustrado, Formal, Comercial, Oportunidad de negocios)
- recommended_action (string, MUST be written in Spanish)
- classification_explanation (string, MUST be written in Spanish)

CRITICAL RULES & BUSINESS CONTEXT:
1. You are auditing employee inboxes on behalf of the company's executive. Your primary goal is to find "dropped balls", angry customers, and neglected sales opportunities (e.g., software licenses, proposals).
2. If the email contains a complaint, frustration, or a customer who is upset, set priority="Critical" and sentiment to an appropriate negative value.
3. If the email is a clear sales opportunity, contract, or license request, set priority="High" and category="Comercial".
4. If the email is a newsletter, promotional, automated marketing, vendor spam, or general news (like from connectab2b.com), you MUST classify it as category="Ruido" and priority="Low", and requires_action=false.
5. Only mark requires_action=true if the executive needs to intervene or if the employee is neglecting an important issue.
6. All textual descriptions (recommended_action, classification_explanation) MUST be written in Spanish.

Return ONLY valid JSON.`, email.SenderEmail, *email.Subject, *email.BodyText)

	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are the Executive Email Intelligence Platform (EEIP) AI Engine. Your goal is to accurately classify corporate emails to reduce executive cognitive load.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)

	if err != nil {
		return err
	}

	var result ClassificationResult
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result)
	if err != nil {
		return err
	}

	email.Category = &result.Category
	email.Priority = &result.Priority
	email.RequiresAction = result.RequiresAction
	email.RequiresApproval = result.RequiresApproval
	email.IsDelegable = result.IsDelegable
	email.Sentiment = &result.Sentiment
	email.SentimentScore = &result.SentimentScore
	email.DissatisfactionScore = &result.DissatisfactionScore
	email.EscalationRiskScore = &result.EscalationRiskScore
	email.CustomerRiskScore = &result.CustomerRiskScore
	email.DetectedTone = &result.DetectedTone
	email.RecommendedAction = &result.RecommendedAction
	email.ClassificationExpl = &result.ClassificationExpl
	
	confidence := 0.95 // Mock confidence for now
	email.AIConfidenceScore = &confidence

	return nil
}
