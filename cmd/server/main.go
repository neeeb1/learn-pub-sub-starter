package main

import (
	"fmt"
	"os"
	"os/signal"

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

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("\nshutting down server and closing rabbitmq connection")
}
