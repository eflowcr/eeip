package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/eprac/eeip-backend/internal/domain/models"
	"github.com/sashabaranov/go-openai"
)

type ResponseEngine interface {
	SuggestResponse(ctx context.Context, email *models.Email, responseType string) (string, error)
}

type responseEngine struct {
	client *openai.Client
}

func NewResponseEngine(apiKey string) ResponseEngine {
	return &responseEngine{
		client: openai.NewClient(apiKey),
	}
}

type SuggestedResponse struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func (s *responseEngine) SuggestResponse(ctx context.Context, email *models.Email, responseType string) (string, error) {
	prompt := fmt.Sprintf(`Generate a %s response to the following email from %s with subject '%s'.
Body: %s

Please return the suggested response as a JSON object with 'subject' and 'body' fields.
Never commit to exact deadlines unless stated in the context. Keep the tone professional and executive.`, responseType, email.SenderEmail, *email.Subject, *email.BodyText)

	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are the Executive Email Intelligence Platform AI. You draft professional emails on behalf of executives.",
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
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
