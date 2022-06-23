package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

func main() {
	dial, err := amqp.Dial("amqp://guest:guest@47.93.20.75:5672/")
	//amqp.Config{
	//	SASL:            nil,
	//	Vhost:           "",
	//	ChannelMax:      0,
	//	FrameSize:       0,
	//	Heartbeat:       0,
	//	TLSClientConfig: nil,
	//	Properties:      nil,
	//	Locale:          "",
	//	Dial:            nil,
	//}
	if err != nil {
		panic(err)
	}
	ch,err := dial.Channel()
	if err != nil {
		panic(err)
	}

	q, err := ch.QueueDeclare(
		"go_q1",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	go consumer(dial,q.Name)

	i := 0

	for  {
		i++
		err = ch.Publish(
			"",
			q.Name,
			false,
			false,
			amqp.Publishing{
				Body: []byte(fmt.Sprintf("message %d", i)),
			},
		)
		if err != nil {
			fmt.Println(err.Error())
		}
		time.Sleep(3*time.Second)
	}
}

func consumer(conn *amqp.Connection, q string) {
	ch,err := conn.Channel()
	if err != nil {
		panic(err)
	}
	msgs, err := ch.Consume(
		q,
		"c1",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	for msg := range  msgs {
		fmt.Printf("%s\n",msg.Body)
	}
}