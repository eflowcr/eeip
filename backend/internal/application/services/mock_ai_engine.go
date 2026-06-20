package services

import (
	"context"
	"math/rand"
	"time"

	"github.com/eprac/eeip-backend/internal/domain/models"
)

type mockAIEngine struct{}

func NewMockAIEngine() AIClassificationEngine {
	return &mockAIEngine{}
}

func (s *mockAIEngine) ClassifyEmail(ctx context.Context, email *models.Email) error {
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	categories := []string{"Cliente", "Prospecto", "Soporte", "Facturacion", "Proveedor"}
	priorities := []string{"Critical", "High", "Medium", "Low"}
	sentiments := []string{"Neutral", "Positivo", "Preocupado", "Molesto"}

	cat := categories[rand.Intn(len(categories))]
	pri := priorities[rand.Intn(len(priorities))]
	sent := sentiments[rand.Intn(len(sentiments))]

	score := rand.Intn(100)
	tone := "Profesional"
	expl := "Clasificado automáticamente por el motor de prueba (Mock) basado en palabras clave simuladas."

	email.Category = &cat
	email.Priority = &pri
	email.RequiresAction = rand.Float32() > 0.5
	email.RequiresApproval = rand.Float32() > 0.8
	email.IsDelegable = true
	email.Sentiment = &sent
	email.SentimentScore = &score
	email.CustomerRiskScore = &score
	email.DetectedTone = &tone
	email.ClassificationExpl = &expl
	
	conf := 0.99
	email.AIConfidenceScore = &conf

	return nil
}
