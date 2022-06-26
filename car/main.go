package main

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	"coolcar/car/car"
	"coolcar/car/dao"
	"coolcar/car/mq/amqpclt"
	"coolcar/car/sim"
	"coolcar/car/sim/pos"
	"coolcar/car/trip"
	"coolcar/car/wx"
	rentalpb "coolcar/rental/api/gen/v1"
	coolenvpb "coolcar/shared/coolenv"
	"coolcar/shared/server"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
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

	carConn,err :=  grpc.Dial("localhost:8086",grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect car  client ",zap.Error(err))
	}

	sub,err := amqpclt.NewSubscriber(conn,exchange,logger)
	if err != nil {
		logger.Fatal("cannot create subscriber",zap.Error(err))
	}


	aiConn,err := grpc.Dial("47.93.20.75:18001",grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect ai service",zap.Error(err))
	}
	posSub,err := amqpclt.NewSubscriber(conn,"pos_sim",logger)
	if err != nil {
		logger.Fatal("cannot create pos subscriber",zap.Error(err))
	}

	simController := &sim.Controller{
		CarService: carpb.NewCarServiceClient(carConn),
		Logger: logger,
		CarSubscriber: sub,
		AIService:coolenvpb.NewAIServiceClient(aiConn),
		PosSubscriber: &pos.Subscriber{Sub: posSub,Logger: logger},
	}
	go simController.RunSimulations(c)

	u := &websocket.Upgrader{
		HandshakeTimeout:  0,//握手超时
		ReadBufferSize:    0,
		WriteBufferSize:   0,
		WriteBufferPool:   nil,
		Subprotocols:      nil,//子协议
		Error:             nil,
		CheckOrigin: func(r *http.Request) bool {
			fmt.Println(r.Header.Get("Origin"))
			return true
		},//检查是否同源，跨域问题
		EnableCompression: false,//是否压缩
	}


	http.HandleFunc("/ws",wx.Handler(u,sub,logger))
	go func() {
		addr  := ":9091"
		logger.Info("HTTP  server started.",zap.String("addr",addr))
		logger.Sugar().Fatal(http.ListenAndServe(addr,nil))
	}()


	tripConn,err := grpc.Dial("localhost:8082",grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot create trip  conn",zap.Error(err))
	}
	go trip.RunUpdater(sub,rentalpb.NewTripServiceClient(tripConn),logger)

	logger.Sugar().Fatal(server.RunGRPCServer(&server.GRPCConfig{
		Logger:            logger,
		Addr:              ":8086",
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
