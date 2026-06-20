package services

import (
	"context"
	"fmt"

	"github.com/eprac/eeip-backend/internal/domain/models"
	"github.com/sashabaranov/go-openai"
)

type SummaryEngine interface {
	GenerateExecutiveSummary(ctx context.Context, emails []models.Email) (string, error)
	GenerateEmailSummary(ctx context.Context, emailBody string) (string, error)
}

type summaryEngine struct {
	client *openai.Client
}

func NewSummaryEngine(apiKey string) SummaryEngine {
	return &summaryEngine{
		client: openai.NewClient(apiKey),
	}
}

func (s *summaryEngine) GenerateExecutiveSummary(ctx context.Context, emails []models.Email) (string, error) {
	if len(emails) == 0 {
		return "No emails to summarize.", nil
	}

	emailContext := "Here are the recent important emails:\n"
	for _, e := range emails {
		emailContext += fmt.Sprintf("- Sender: %s | Subject: %s | Priority: %s | Tone: %s\n", e.SenderEmail, *e.Subject, *e.Priority, *e.DetectedTone)
	}

	prompt := fmt.Sprintf(`Based on the following emails, generate a concise Executive Summary.
Highlight critical issues, main risks, and pending actions. Output the summary in Markdown format.
	
%s`, emailContext)

	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an Executive Assistant AI generating summaries for a busy CTO/Executive.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func (s *summaryEngine) GenerateEmailSummary(ctx context.Context, emailBody string) (string, error) {
	if emailBody == "" {
		return "Sin contenido para resumir.", nil
	}
	
	// Trim to save tokens
	if len(emailBody) > 4000 {
		emailBody = emailBody[:4000]
	}

	prompt := fmt.Sprintf("Genera un resumen en un solo párrafo, conciso y directo, del siguiente correo. Resalta lo más crítico y ve directo al grano:\n\n%s", emailBody)

	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "Eres un asistente ejecutivo que resume correos de forma directa y al grano en español.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
