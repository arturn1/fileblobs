package service

import "log"

type MessageService struct{}

func NewMessageService() *MessageService {
	return &MessageService{}
}

func (s *MessageService) ProcessMessage(msg string) {
	log.Printf("✅ Processando mensagem: %s", msg)
}
