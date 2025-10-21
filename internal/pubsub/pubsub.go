package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	durable   SimpleQueueType = iota // 0
	transient                        // 1
)

type AckType int

const (
	Ack         AckType = iota // 0
	NackRequeue                // 1
	NackDiscard                // 2
)

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return ch, amqp.Queue{}, err
	}

	var durable bool
	var exclusive bool
	var autoDelete bool

	//Check if queueType is durable (0), or transient (1)
	if queueType == 0 {
		durable = true
	} else {
		durable = false
		exclusive = true
		autoDelete = true
	}

	table := make(amqp.Table)
	table["x-dead-letter-exchange"] = "peril_dlx"

	queue, err := ch.QueueDeclare(queueName, durable, autoDelete, exclusive, false, table)
	if err != nil {
		return ch, queue, err
	}

	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return ch, queue, err
	}

	return ch, queue, nil

}

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	m := amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	}

	err = ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		m)
	if err != nil {
		return err
	}

	return nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an "enum" to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	amqpCh, newQueue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	newCh, err := amqpCh.Consume(newQueue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for msg := range newCh {
			var data T
			err = json.Unmarshal(msg.Body, &data)
			if err != nil {
				fmt.Println(err)
			}
			ackType := handler(data)
			switch ackType {
			case Ack:
				msg.Ack(false)
				fmt.Println("ack'd message")
			case NackRequeue:
				msg.Nack(false, true)
				fmt.Println("nack'd message with requeue")
			case NackDiscard:
				msg.Nack(false, false)
				fmt.Println("nack'd message without requeing")
			}
		}
	}()

	return nil
}
