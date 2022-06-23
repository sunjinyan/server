package main

import (
	"context"
	"coolcar/car/amqpclt"
	carpb "coolcar/car/api/gen/v1"
	"coolcar/car/car"
	"coolcar/car/dao"
	"coolcar/car/sim"
	"coolcar/shared/server"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func main() {
	logger, err := server.NewZapLogger()

	if err != nil {
		log.Fatalf("cannot create logger:%v",err)
	}

	c := context.Background()

	mongoClient, err := mongo.Connect(c, options.Client().ApplyURI("mongodb://47.93.20.75:27017/coolcar?readPreference=primary&ssl=false"))

	if err != nil {
		logger.Fatal("cannot connect mongodb",zap.Error(err))
	}

	db := mongoClient.Database("coolcar")

	conn, err := amqp.Dial("amqp://guest:guest@47.93.20.75:5672/")
	if err != nil {
		logger.Fatal("cannot connect rabbit",zap.Error(err))
	}
	exchange := "coolcar"
	pub ,err := amqpclt.NewPublisher(conn,exchange)
	if err != nil {
		logger.Fatal("cannot create publisher ",zap.Error(err))
	}

	carConn,err :=  grpc.Dial("localhost:8085",grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect car  client ",zap.Error(err))
	}

	sub,err := amqpclt.NewSubscriber(conn,exchange,logger)
	if err != nil {
		logger.Fatal("cannot create subscriber",zap.Error(err))
	}


	simController := &sim.Controller{
		CarService: carpb.NewCarServiceClient(carConn),
		Logger: logger,
		Subscriber: sub,
	}
	go simController.RunSimulations(c)
	logger.Sugar().Fatal(server.RunGRPCServer(&server.GRPCConfig{
		Logger:            logger,
		Addr:              ":8085",
		Name:              "car",
		RegisterFunc: func(s *grpc.Server) {
			carpb.RegisterCarServiceServer(s,&car.Service{
				Logger:    logger,
				Mongo:     dao.NewMongo(db),
				Publisher: pub,
			})
		},
	}))

}
