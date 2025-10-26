package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")

		outcome := gs.HandleMove(move)

		switch outcome {
		case gamelogic.MoveOutcomeMakeWar:
			rw := gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: gs.Player,
			}

			err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, gs.Player.Username),
				rw,
			)
			if err != nil {
				fmt.Println(err)
				return pubsub.NackRequeue
			}

			//fmt.Println("publishing:", fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, gs.Player.Username))

			return pubsub.Ack
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		default:
			return pubsub.NackDiscard
		}
	}
}
