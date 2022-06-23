package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@47.93.20.75:5672/")

	if err != nil {
		panic(err)
	}

	channel, err := conn.Channel()

	if err != nil {
		panic(err)
	}

	err = channel.ExchangeDeclare("pubsub", amqp.ExchangeFanout, true, false, false, false, nil)
	if err != nil {
		panic(err)
	}

	q1, err := channel.QueueDeclare("pubsub_1", true, false, false, false, nil)
	if err != nil {
		panic(err)
	}



	q2, err := channel.QueueDeclare("pubsub_2", true, false, false, false, nil)
	if err != nil {
		panic(err)
	}


	err = channel.QueueBind(q1.Name, "", "pubsub", false, nil)

	if err != nil {
		panic(err)
	}
	err = channel.QueueBind(q2.Name, "", "pubsub", false, nil)

	if err != nil {
		panic(err)
	}


	go consumer(conn,q1.Name,"c1")
	go consumer(conn,q2.Name,"c2")

	i := 0

	for  {
		i++
		err = channel.Publish(
			"pubsub",
			"",
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

func consumer(conn *amqp.Connection, q string,cons string) {
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	msgs, err := ch.Consume(q, cons, true,
		false,
		false,
		false,
		nil)
	if err != nil {
		panic(err)
	}

	for msg := range  msgs {
		fmt.Printf("%s -------     %s     --------- %s\n",q,cons,msg.Body)
	}
}


func subscribe(conn *amqp.Connection, ex string) {
	ch,err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	q1, err := ch.QueueDeclare("pubsub_1", true, false, false, false, nil)

	if err != nil {
		panic(err)
	}

	err = ch.QueueBind(q1.Name, "", "pubsub", false, nil)

	if err != nil {
		panic(err)
	}
}