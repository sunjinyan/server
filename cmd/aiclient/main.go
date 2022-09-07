package main

import (
	"context"
	"coolcar/car/mq/amqpclt"
	coolenvpb "coolcar/shared/coolenv"
	"coolcar/shared/server"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

func main() {
	conn,err := grpc.Dial("47.93.20.75:18001",grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	ac := coolenvpb.NewAIServiceClient(conn)
	c := context.Background()
	res, err := ac.MeasureDistance(c, &coolenvpb.MeasureDistanceRequest{
		From: &coolenvpb.Location{
			Latitude:  30,
			Longitude: 120,
		},
		To: &coolenvpb.Location{
			Latitude:  31,
			Longitude: 121,
		},
	})

	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n",res)


	idRes,err := ac.LicIdentity(c,&coolenvpb.IdentityRequest{
		Photo:  []byte{1, 2, 3, 4, 5},
	})

	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n",idRes)


	//Car position simuliation

	pos, err := ac.SimulateCarPos(c, &coolenvpb.SimulateCarPosRequest{
		CarId: "car123",
		InitialPos: &coolenvpb.Location{
			Latitude:  30,
			Longitude: 120,
		},
		Type: coolenvpb.PosType_NINGBO,
	})

	if err != nil {
		panic(err)
	}
	fmt.Println(pos.String())


	logger, err := server.NewZapLogger()

	if err != nil {
		log.Fatalf("cannot create logger:%v",err)
	}

	exchange := "pos_sim"
	conns, err := amqp.Dial("amqp://guest:guest@47.93.20.75:5672/")
	if err != nil {
		logger.Fatal("cannot connect rabbit",zap.Error(err))
	}


	//carConn,err :=  grpc.Dial("localhost:8085",grpc.WithTransportCredentials(insecure.NewCredentials()))
	//if err != nil {
	//	logger.Fatal("cannot connect car  client ",zap.Error(err))
	//}
	sub,err := amqpclt.NewSubscriber(conns,exchange,logger)
	if err != nil {
		logger.Fatal("cannot create subscriber",zap.Error(err))
		panic(err)
	}

	raw, fn, err := sub.SubscribeRaw(c)

	if err != nil {
		logger.Fatal("cannot SubscribeRaw",zap.Error(err))
		panic(err)
	}

	defer fn()

	tm := time.After(10 * time.Second)
	//for msg := range raw {
	for  {
		shouldStop := false
		select {
		case msg := <-raw:
			var update coolenvpb.CarPosUpdate
			err := json.Unmarshal(msg.Body, &update)
			if err != nil {
				panic(err)
			}

			fmt.Printf("%+v\n",&update)
		case <-tm:
			shouldStop = true
			//break
		}
		if shouldStop {
			break
		}
	}

	_, err = ac.EndSimulateCarPos(c, &coolenvpb.EndSimulateCarPosRequest{CarId: "car123"})
	if err != nil {
		panic(err)
	}

}
