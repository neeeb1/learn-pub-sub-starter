package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rabbitmq, err := dialRabbitMQ()
	if err != nil {
		fmt.Println(err)
		return
	}

	defer rabbitmq.Close()

	rabbitCh, err := rabbitmq.Channel()
	if err != nil {
		fmt.Println(err)
		return
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(username)

	state := gamelogic.NewGameState(username)

	_, _, err = pubsub.DeclareAndBind(
		rabbitmq,
		routing.ExchangePerilDirect,
		fmt.Sprintf("pause.%s", username),
		"pause",
		pubsub.Transient)
	if err != nil {
		fmt.Println(err)
		return
	}

	pubsub.SubscribeJSON(
		rabbitmq,
		routing.ExchangePerilDirect,
		fmt.Sprintf("pause.%s", username),
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(state),
	)

	_, _, err = pubsub.DeclareAndBind(
		rabbitmq,
		routing.ExchangePerilTopic,
		fmt.Sprintf("army_moves.%s", username),
		"army_moves.*",
		pubsub.Transient,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = pubsub.SubscribeJSON(
		rabbitmq,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
		fmt.Sprintf("%s.*", routing.ArmyMovesPrefix),
		pubsub.Transient,
		handlerMove(state, rabbitCh),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, _, err = pubsub.DeclareAndBind(
		rabbitmq,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		fmt.Sprintf("%s.*", routing.WarRecognitionsPrefix),
		pubsub.Durable,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = pubsub.SubscribeJSON(
		rabbitmq,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		fmt.Sprintf("%s.*", routing.WarRecognitionsPrefix),
		pubsub.Durable,
		handlerWar(state, rabbitCh),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

replLoop:
	for {
		cmds := gamelogic.GetInput()

		switch cmds[0] {
		case "spawn":
			err = state.CommandSpawn(cmds)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Spawn successful")

		case "move":
			move, err := state.CommandMove(cmds)
			if err != nil {
				fmt.Println(err)
				continue
			}

			pubsub.PublishJSON(
				rabbitCh,
				routing.ExchangePerilTopic,
				fmt.Sprintf("army_moves.%s", username),
				move)

			fmt.Println("Move successful")

		case "status":
			state.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			break replLoop
		default:
			fmt.Println("Command not recognized, try one of these?")
			gamelogic.PrintClientHelp()
		}
	}

	fmt.Println("exiting peril client... goodbye!")
}

func dialRabbitMQ() (*amqp.Connection, error) {
	fmt.Println("Starting Peril client...")
	const conn_string = "amqp://guest:guest@localhost:5672"

	rabbitmq, err := amqp.Dial(conn_string)
	if err != nil {
		return rabbitmq, err
	}
	fmt.Println("connection to rabbitmq successful")
	return rabbitmq, nil
}
