package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	const conn_string = "amqp://guest:guest@localhost:5672"

	rabbitmq, err := amqp.Dial(conn_string)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer rabbitmq.Close()
	fmt.Println("connection to rabbitmq successful")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(username)

	_, _, err = pubsub.DeclareAndBind(
		rabbitmq,
		"peril_direct",
		fmt.Sprintf("pause.%s", username),
		"pause",
		1)
	if err != nil {
		fmt.Println(err)
		return
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

}
