package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](
	ch *amqp.Channel,
	exchange, key string,
	val T) error {

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
	exchange, queueName, key string,
	queueType SimpleQueueType, // an "enum" to represent "durable" or "transient"
	handler func(T) AckType) error {

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
