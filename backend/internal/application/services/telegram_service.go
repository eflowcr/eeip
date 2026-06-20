package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type TelegramService interface {
	SendMessage(chatID string, message string) error
}

type telegramService struct {
	botToken string
}

func NewTelegramService(botToken string) TelegramService {
	return &telegramService{botToken: botToken}
}

func (s *telegramService) SendMessage(chatID string, message string) error {
	if s.botToken == "" {
		log.Println("Telegram bot token is empty, skipping message")
		return nil
	}
	if chatID == "" {
		return fmt.Errorf("chat ID is empty")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.botToken)
	
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       message,
		"parse_mode": "HTML",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send telegram message, status: %d", resp.StatusCode)
	}

	return nil
}
