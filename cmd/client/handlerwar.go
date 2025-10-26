package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(rw)
		//fmt.Println("war received for:", rw.Attacker.Username, "vs", rw.Defender.Username, "outcome:", outcome)

		log := fmt.Sprintf("%s won a war against %s", winner, loser)

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue

		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard

		case gamelogic.WarOutcomeOpponentWon:
			err := pubsub.PublishGob(ch,
				string(routing.ExchangePerilTopic),
				fmt.Sprintf("%s.%s", routing.GameLogSlug, rw.Attacker.Username),
				log)
			if err != nil {
				fmt.Println(err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		case gamelogic.WarOutcomeYouWon:
			err := pubsub.PublishGob(
				ch,
				string(routing.ExchangePerilTopic),
				fmt.Sprintf("%s.%s", routing.GameLogSlug, rw.Attacker.Username),
				log)
			if err != nil {
				fmt.Println(err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		case gamelogic.WarOutcomeDraw:
			log = fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)

			err := pubsub.PublishGob(
				ch,
				string(routing.ExchangePerilTopic),
				fmt.Sprintf("%s.%s", routing.GameLogSlug, rw.Attacker.Username),
				log)
			if err != nil {
				fmt.Println(err)
				return pubsub.NackRequeue
			}

			return pubsub.Ack

		default:
			fmt.Println("err unknown WarOutcome")
			return pubsub.NackDiscard
		}
	}
}
