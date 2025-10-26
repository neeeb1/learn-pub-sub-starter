package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	const conn_string = "amqp://guest:guest@localhost:5672"

	rabbitmq, err := amqp.Dial(conn_string)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer rabbitmq.Close()
	fmt.Println("connection to rabbitmq was successful")
	gamelogic.PrintServerHelp()

	rabbitCh, err := rabbitmq.Channel()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = pubsub.PublishJSON(
		rabbitCh,
		routing.ExchangePerilDirect,
		routing.PauseKey,
		routing.PlayingState{IsPaused: true},
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = pubsub.SubscribeGob(rabbitmq,
		string(routing.ExchangePerilTopic),
		string(routing.GameLogSlug),
		fmt.Sprintf("%s.*", routing.GameLogSlug),
		pubsub.Durable,
		handlerLog(),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

replLoop:
	for {
		cmds := gamelogic.GetInput()

		switch cmds[0] {
		case "pause":
			fmt.Println("Sending pause msg..")
			err = pubsub.PublishJSON(
				rabbitCh,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: true})
			if err != nil {
				fmt.Println(err)
			}
		case "resume":
			fmt.Println("Sending resume msg...")
			err = pubsub.PublishJSON(
				rabbitCh,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: false})
			if err != nil {
				fmt.Println(err)
			}
		case "quit":
			fmt.Println("Exiting peril server... Goodbye!")
			break replLoop
		default:
			fmt.Println("not a recognized command, try one of these instead:")
			gamelogic.PrintServerHelp()
		}
	}

	fmt.Println("\nshutting down server and closing rabbitmq connection")
}
