package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishGob[T any](
	ch *amqp.Channel,
	exchange, key string,
	val T) error {

	var data bytes.Buffer
	enc := gob.NewEncoder(&data)
	err := enc.Encode(val)
	if err != nil {
		return err
	}

	m := amqp.Publishing{
		ContentType: "application/gob",
		Body:        data.Bytes(),
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
