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

	_, _, err = pubsub.DeclareAndBind(
		rabbitmq,
		routing.ExchangePerilDirect,
		fmt.Sprintf("pause.%s", username),
		"pause",
		1,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	state := gamelogic.NewGameState(username)

	pubsub.SubscribeJSON(
		rabbitmq,
		routing.ExchangePerilDirect,
		fmt.Sprintf("pause.%s", username),
		routing.PauseKey,
		1,
		handlerPause(state),
	)

	_, _, err = pubsub.DeclareAndBind(
		rabbitmq,
		routing.ExchangePerilTopic,
		fmt.Sprintf("army_moves.%s", username),
		"army_moves.*",
		1,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	pubsub.SubscribeJSON(
		rabbitmq,
		routing.ExchangePerilTopic,
		fmt.Sprintf("army_moves.%s", username),
		"army_moves.*",
		1,
		handlerMove(state),
	)

replLoop:
	for {
		cmds := gamelogic.GetInput()

		switch cmds[0] {
		case "spawn":
			err = state.CommandSpawn(cmds)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Spawn successful")
		case "move":
			move, err := state.CommandMove(cmds)
			if err != nil {
				fmt.Println(err)
			}

			pubsub.PublishJSON(
				rabbitCh,
				routing.ExchangePerilTopic,
				fmt.Sprintf("army_moves.%s", username),
				move,
			)

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

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")

		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")

		outcome := gs.HandleMove(move)

		switch outcome {
		case gamelogic.MoveOutcomeMakeWar:
			return pubsub.NackRequeue
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		default:
			return pubsub.NackDiscard
		}
	}
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
