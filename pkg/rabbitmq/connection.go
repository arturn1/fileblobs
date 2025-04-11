package rabbitmq

import (
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
)

func Connect() (*amqp.Connection, error) {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao RabbitMQ: %w", err)
	}
	log.Println("✅ Conectado ao RabbitMQ!")
	return conn, nil
}
