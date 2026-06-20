package services

import (
	"crypto/tls"
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
)

type MailerService interface {
	SendEmailAlert(toEmails []string, subject string, htmlBody string, host string, port int, user string, pass string) error
}

type mailerService struct{}

func NewMailerService() MailerService {
	return &mailerService{}
}

func (s *mailerService) SendEmailAlert(toEmails []string, subject string, htmlBody string, host string, port int, user string, pass string) error {
	if len(toEmails) == 0 {
		return nil
	}

	m := gomail.NewMessage()
	m.SetHeader("From", user)
	
	// Convert array of emails to variadic string arguments
	m.SetHeader("To", toEmails...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	// InsecureSkipVerify is needed because the self-hosted mail server might have self-signed certs
	d := gomail.NewDialer(host, port, user, pass)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send email to %v: %v", toEmails, err)
		return fmt.Errorf("failed to send email alert: %v", err)
	}

	log.Printf("Successfully sent email alert to %d recipients", len(toEmails))
	return nil
}
