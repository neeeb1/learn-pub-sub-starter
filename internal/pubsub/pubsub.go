package pubsub

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable   SimpleQueueType = iota // 0
	Transient                        // 1
)

type AckType int

const (
	Ack         AckType = iota // 0
	NackRequeue                // 1
	NackDiscard                // 2
)

func DeclareAndBind(conn *amqp.Connection,
	exchange, queueName, key string,
	queueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return ch, amqp.Queue{}, err
	}

	var durable bool
	var exclusive bool
	var autoDelete bool

	//Check if queueType is durable (0), or transient (1)
	if queueType == Durable {
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
