package main

import (
	"fmt"

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

	state := gamelogic.NewGameState(username)

replLoop:
	for {
		cmds := gamelogic.GetInput()

		switch cmds[0] {
		case "spawn":
			err = state.CommandSpawn(cmds)

			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Spawn successful")
			}
		case "move":
			_, err = state.CommandMove(cmds)

			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Move successful")
			}
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
