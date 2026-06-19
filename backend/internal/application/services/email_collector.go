package services

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/eprac/eeip-backend/internal/domain/models"
	"github.com/eprac/eeip-backend/internal/infrastructure/database"
)

type EmailCollector interface {
	CollectEmails(ctx context.Context, account *models.EmailAccount) error
}

type emailCollector struct {
	repo       database.EmailRepository
	aiEngine   AIClassificationEngine
}

func NewEmailCollector(repo database.EmailRepository, aiEngine AIClassificationEngine) EmailCollector {
	return &emailCollector{repo: repo, aiEngine: aiEngine}
}

func (s *emailCollector) CollectEmails(ctx context.Context, account *models.EmailAccount) error {
	log.Printf("Connecting to IMAP server %s:%d for account %s...", account.IMAPHost, account.IMAPPort, account.EmailAddress)

	// Connect to server
	var c *client.Client
	var err error
	if account.IMAPPort == 993 {
		tlsConfig := &tls.Config{InsecureSkipVerify: true}
		c, err = client.DialTLS(fmt.Sprintf("%s:%d", account.IMAPHost, account.IMAPPort), tlsConfig)
	} else {
		c, err = client.Dial(fmt.Sprintf("%s:%d", account.IMAPHost, account.IMAPPort))
	}

	if err != nil {
		return err
	}
	defer c.Logout()

	// Login
	if err := c.Login(account.IMAPUser, account.IMAPPassword); err != nil {
		return err
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		return err
	}

	if mbox.Messages == 0 {
		log.Println("No messages in INBOX")
		return nil
	}

	// Fetch last 10 messages for demonstration (in production this would use LastSyncDate)
	from := uint32(1)
	if mbox.Messages > 10 {
		from = mbox.Messages - 10
	}
	to := mbox.Messages
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	var section imap.BodySectionName
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{section.FetchItem()}, messages)
	}()

	for msg := range messages {
		if msg == nil {
			log.Println("Server didn't returned message")
			continue
		}

		r := msg.GetBody(&section)
		if r == nil {
			log.Println("Server didn't returned message body")
			continue
		}

		mr, err := mail.CreateReader(r)
		if err != nil {
			log.Printf("Failed to create mail reader: %v", err)
			continue
		}

		header := mr.Header
		subject, _ := header.Subject()
		fromAddrs, _ := header.AddressList("From")
		toAddrs, _ := header.AddressList("To")
		msgID, _ := header.MessageID()
		date, _ := header.Date()

		var senderEmail, senderName string
		if len(fromAddrs) > 0 {
			senderEmail = fromAddrs[0].Address
			senderName = fromAddrs[0].Name
		}

		recipients := make([]string, 0)
		for _, addr := range toAddrs {
			recipients = append(recipients, addr.Address)
		}
		recipientsJSON, _ := json.Marshal(recipients)

		var bodyText string
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Printf("Failed to read part: %v", err)
				break
			}

			switch h := p.Header.(type) {
			case *mail.InlineHeader:
				b, _ := io.ReadAll(p.Body)
				contentType, _, _ := h.ContentType()
				if contentType == "text/plain" {
					bodyText = string(b)
				}
			}
		}

		email := &models.Email{
			AccountID:       account.ID,
			MessageID:       msgID,
			SenderEmail:     senderEmail,
			SenderName:      &senderName,
			RecipientEmails: recipientsJSON,
			Subject:         &subject,
			BodyText:        &bodyText,
			ReceivedAt:      date,
			Status:          "Unread",
		}

		// AI Classification
		if err := s.aiEngine.ClassifyEmail(ctx, email); err != nil {
			log.Printf("Failed to classify email %s: %v", msgID, err)
		}

		// Save to repository
		if err := s.repo.SaveEmail(ctx, email); err != nil {
			log.Printf("Failed to save email %s: %v", msgID, err)
		} else {
			log.Printf("Processed email: %s", subject)
		}
	}

	if err := <-done; err != nil {
		return err
	}

	// Update last sync date
	now := time.Now()
	account.LastSyncDate = &now
	// Here we should update the account in DB, omitted for brevity

	return nil
}
