package consumer

import (
	"log"

	"github.com/streadway/amqp"
)

type Consumer struct {
	conn *amqp.Connection
}

func NewConsumer(conn *amqp.Connection) *Consumer {
	return &Consumer{conn: conn}
}

func (c *Consumer) Start(queueName string) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	forever := make(chan bool)

	go func() {
		for msg := range msgs {
			log.Printf("📩 Mensagem recebida: %s", msg.Body)
		}
	}()

	log.Printf("🎧 Aguardando mensagens na fila: %s", queueName)
	<-forever

	return nil
}
