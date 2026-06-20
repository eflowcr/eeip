package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/eprac/eeip-backend/internal/infrastructure/database"
	"github.com/robfig/cron/v3"
)

type CronService interface {
	Start()
	Stop()
}

type cronService struct {
	cronInst     *cron.Cron
	emailRepo    database.EmailRepository
	stakeholders *database.StakeholderRepository
	telegramSvc  TelegramService
}

func NewCronService(emailRepo database.EmailRepository, stakeholders *database.StakeholderRepository, telegramSvc TelegramService) CronService {
	// Create cron with standard parser and local timezone
	c := cron.New(cron.WithLocation(time.Local))
	
	s := &cronService{
		cronInst:     c,
		emailRepo:    emailRepo,
		stakeholders: stakeholders,
		telegramSvc:  telegramSvc,
	}

	// Run at 08:00, 11:00, 14:00, 17:00 every day
	_, err := c.AddFunc("0 8,11,14,17 * * *", s.runAlertsJob)
	if err != nil {
		log.Printf("Failed to add cron job: %v", err)
	}

	return s
}

func (s *cronService) Start() {
	s.cronInst.Start()
	log.Println("Cron service started")
}

func (s *cronService) Stop() {
	s.cronInst.Stop()
	log.Println("Cron service stopped")
}

func (s *cronService) runAlertsJob() {
	log.Println("Running 3-hour alert cron job...")
	ctx := context.Background()
	
	// Fetch emails from the last 3 hours
	since := time.Now().Add(-3 * time.Hour)
	emails, err := s.emailRepo.GetAlertEmails(ctx, since)
	if err != nil {
		log.Printf("Error fetching alert emails: %v", err)
		return
	}

	if len(emails) == 0 {
		log.Println("No critical/urgent emails to report in the last 3 hours")
		return
	}

	// Get stakeholders
	shs, err := s.stakeholders.GetAll()
	if err != nil {
		log.Printf("Failed to get stakeholders: %v", err)
		return
	}

	// Build summary message
	msg := fmt.Sprintf("📊 <b>RESUMEN EJECUTIVO (Últimas 3 horas)</b>\nSe encontraron <b>%d</b> correos de alto riesgo que requieren tu atención:\n\n", len(emails))
	
	for i, em := range emails {
		if i >= 5 {
			msg += fmt.Sprintf("\n...y %d más. Revisa la plataforma para más detalles.", len(emails)-5)
			break
		}
		
		priority := "N/A"
		if em.Priority != nil { priority = *em.Priority }
		
		msg += fmt.Sprintf("🔹 <b>[%s]</b> %s (De: %s)\n", priority, *em.Subject, em.SenderEmail)
	}

	msg += "\n\n🌐 <a href=\"http://localhost:4200\">Abrir EEIP Global Inbox</a>"

	// Broadcast
	for _, sh := range shs {
		if sh.TelegramChatID != "" {
			err := s.telegramSvc.SendMessage(sh.TelegramChatID, msg)
			if err != nil {
				log.Printf("Failed to send cron telegram to %s: %v", sh.Name, err)
			} else {
				log.Printf("Sent cron telegram alert to %s", sh.Name)
			}
		}
	}
}
