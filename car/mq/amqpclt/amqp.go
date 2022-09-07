package amqpclt

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type Publisher struct {
	ch *amqp.Channel
	exchange string
}

func NewPublisher(conn *amqp.Connection,exchange string) (*Publisher, error) {
	channel, err := conn.Channel()

	if err != nil {
		return nil, fmt.Errorf("cannot allocate channel: %v",err)
	}

	err = declareExchange(channel,exchange)

	//channel.ExchangeDeclare(exchange, amqp.ExchangeFanout, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot declare exchange: %v",err)
	}

	return &Publisher{
		ch:       channel,
		exchange: exchange,
	},nil

}

func (p *Publisher)Publish(c context.Context,car *carpb.CarEntity) error {
	b, err := json.Marshal(car)

	if err != nil {
		return fmt.Errorf("cannot marshal: %v",err)
	}

	return p.ch.Publish(
		p.exchange,
		"",
		false,
		false,
		amqp.Publishing{
			Body: b,
			Expiration: "30000",//过期时间
		},
	)
}


type Subscriber struct {
	conn *amqp.Connection
	exchange string
	logger  *zap.Logger
}

func NewSubscriber(conn *amqp.Connection, exchange string, logger *zap.Logger) (*Subscriber,error) {
	channel, err := conn.Channel()

	if err != nil {
		return nil, fmt.Errorf("cannot allocate channel: %v",err)
	}
	defer channel.Close()
	err = declareExchange(channel,exchange)

	//channel.ExchangeDeclare(exchange, amqp.ExchangeFanout, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot declare exchange: %v",err)
	}
	return &Subscriber{
		conn:     conn,
		exchange: exchange,
		logger:   logger,
	},nil
}

func (s *Subscriber)SubscribeRaw(ctx context.Context) (ad <-chan amqp.Delivery,fn func(),err error) {

	//ch := make(chan *carpb.CarEntity)
	ch,err := s.conn.Channel()
	if err != nil {
		return nil, func() {

		},fmt.Errorf("cannot connection: %v",err)
	}
	//defer ch.Close()
	closeCh := func() {
		err = ch.Close()
		if err != nil {
			s.logger.Error("cannot close ch ",zap.Error(err))
		}
	}

	q, err := ch.QueueDeclare("", true, false, false, false, nil)

	if err != nil {
		return nil,closeCh,fmt.Errorf("cannot Queue  Declare: %v",err)
	}


	cleanUp := func() {
		_, err := ch.QueueDelete(
			q.Name,
			false,
			false,
			false,
		)
		if err != nil {
			s.logger.Error("cannot close ch ",zap.String("name",q.Name),zap.Error(err))
		}
		closeCh()
	}

	err = ch.QueueBind(q.Name, "", s.exchange, false, nil)

	if err != nil {
		return nil,cleanUp,fmt.Errorf("cannot Queue   Bind: %v",err)
	}

	msgs,err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,nil,
		)
	if err != nil {
		return nil,cleanUp,fmt.Errorf("cannot Consume  : %v",err)
	}
	return msgs,cleanUp, nil
}

func (s *Subscriber)Subscribe(ctx context.Context) (cc chan *carpb.CarEntity,fn func(),err error)	 {
	raws,cleanUp, err := s.SubscribeRaw(ctx)
	if err != nil {
		return nil,cleanUp, err
	}
	carCh := make(chan *carpb.CarEntity)
	go func() {
		for msg := range raws {
			var car carpb.CarEntity
			err := json.Unmarshal(msg.Body,&car)
			if err != nil {
				s.logger.Error("cannot unmarshal",zap.Error(err))
			}
			carCh <- &car
		}
		close(carCh)
	}()

	return carCh,cleanUp, nil
}

func declareExchange(ch *amqp.Channel, exchange string) error {
	/**延迟队列
	err := ch.ExchangeDeclare(exchange, "x-delayed-message",
		false, false, false, false,
		map[string]interface{}{"x-delayed-type": "direct"})
	if err != nil {
		log.Fatal(err)
	}*/

	return  ch.ExchangeDeclare(exchange, amqp.ExchangeFanout, true, false, false, false, nil)
}