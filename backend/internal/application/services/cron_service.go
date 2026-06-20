package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/eprac/eeip-backend/internal/domain/models"
	"github.com/eprac/eeip-backend/internal/infrastructure/database"
	"github.com/robfig/cron/v3"
)

type CronService interface {
	Start()
	Stop()
}

type cronService struct {
	cronInst       *cron.Cron
	emailRepo      database.EmailRepository
	accRepo        database.AccountRepository
	stakeholders   *database.StakeholderRepository
	telegramSvc    TelegramService
	mailerSvc      MailerService
	emailCollector EmailCollector
}

func NewCronService(emailRepo database.EmailRepository, accRepo database.AccountRepository, stakeholders *database.StakeholderRepository, telegramSvc TelegramService, mailerSvc MailerService, emailCollector EmailCollector) CronService {
	// Create cron with standard parser and local timezone
	c := cron.New(cron.WithLocation(time.Local))
	
	s := &cronService{
		cronInst:       c,
		emailRepo:      emailRepo,
		accRepo:        accRepo,
		stakeholders:   stakeholders,
		telegramSvc:    telegramSvc,
		mailerSvc:      mailerSvc,
		emailCollector: emailCollector,
	}

	// Run at 08:00, 11:00, 14:00, 17:00 every day
	_, err := c.AddFunc("0 8,11,14,17 * * *", s.runAlertsJob)
	if err != nil {
		log.Printf("Failed to add cron job: %v", err)
	}

	// Run IMAP sync every 5 minutes
	_, errSync := c.AddFunc("*/5 * * * *", s.runSyncJob)
	if errSync != nil {
		log.Printf("Failed to add sync cron job: %v", errSync)
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
	
	var summary string
	for i, em := range emails {
		if i >= 5 {
			msg += fmt.Sprintf("\n...y %d más. Revisa la plataforma para más detalles.", len(emails)-5)
			summary += "<li>...y más correos. Revisa la plataforma.</li>"
			break
		}
		priority := "N/A"
		if em.Priority != nil { priority = *em.Priority }
		summary += fmt.Sprintf("<li><b>[%s]</b> %s (De: %s)</li>", priority, *em.Subject, em.SenderEmail)
	}
	
	msg = fmt.Sprintf("🚨 <b>RESUMEN 3H: CORREOS CRÍTICOS</b>\n\nSe han detectado %d correos críticos en las últimas 3 horas.\n\n%s", len(emails), summary)
	
	htmlMsg := fmt.Sprintf("<h2>🚨 RESUMEN 3H: CORREOS CRÍTICOS</h2><p>Se han detectado %d correos críticos en las últimas 3 horas.</p><ul>%s</ul>", len(emails), summary)


	var emailRecipients []string
	for _, sh := range shs {
		if sh.TelegramChatID != "" {
			err := s.telegramSvc.SendMessage(sh.TelegramChatID, msg)
			if err != nil {
				log.Printf("Failed to send cron telegram to %s: %v", sh.Name, err)
			}
		}
		if sh.Email != "" {
			emailRecipients = append(emailRecipients, sh.Email)
		}
	}

	if len(emailRecipients) > 0 {
		// Fetch an account to send the email from
		accounts, err := s.accRepo.GetAccounts(ctx)
		if err == nil && len(accounts) > 0 {
			var acc models.EmailAccount
			for _, a := range accounts {
				if a.EmailAddress == "eitel.rodriguez@eprac.com" {
					acc = a
					break
				}
			}
			if acc.ID == "" {
				acc = accounts[0]
			}
			
			smtpPort := 587
			if acc.IMAPPort == 993 {
				smtpPort = 465
			}
			s.mailerSvc.SendEmailAlert(emailRecipients, "🚨 Resumen de Correos Críticos EEIP", htmlMsg, acc.IMAPHost, smtpPort, acc.IMAPUser, acc.IMAPPassword)
		}
	}
}

func (s *cronService) runSyncJob() {
	log.Println("Running 5-minute IMAP sync cron job...")
	ctx := context.Background()

	accounts, err := s.accRepo.GetAccounts(ctx)
	if err != nil {
		log.Printf("Sync Job failed to fetch accounts: %v", err)
		return
	}

	for _, acc := range accounts {
		if !acc.IsActive {
			continue // Skip inactive accounts
		}
		
		log.Printf("Cron Syncing account: %s", acc.EmailAddress)
		err := s.emailCollector.CollectEmails(ctx, &acc)
		if err != nil {
			log.Printf("Cron Sync failed for %s: %v", acc.EmailAddress, err)
		} else {
			log.Printf("Cron Sync successful for %s", acc.EmailAddress)
		}
	}
}
