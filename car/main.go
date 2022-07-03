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
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	"os"
)


var addr string
var mongoURI string
var wsAddr string
var amqpURL string
var carAddr string
var tripAddr string
var aiAddr string


func init() {

	app := &cli.App{
		Usage:                  "please inter gateway options",
		Flags:                  []cli.Flag{
			&cli.StringFlag{
				Name:        "addr",
				Usage:       "service addr port",
				//Required:    true,
				Value:       ":8086",
				Destination: &addr,
				Aliases:     []string{"a"},
				EnvVars:     []string{"ADDR"},
			},
			&cli.StringFlag{
				Name:        "mongo_url",
				Usage:       "mongo URI addr port",
				//Required:    true,
				Value:       "mongodb://47.93.20.75:27017/coolcar?readPreference=primary&ssl=false",
				Destination: &mongoURI,
				Aliases:     []string{"mu"},
				EnvVars:     []string{"MONGO_URI"},
			},
			&cli.StringFlag{
				Name:        "ws_addr",
				Usage:       "web socket addr",
				//Required:    true,
				Value:       ":9091",
				Destination: &wsAddr,
				Aliases:     []string{"wa"},
				EnvVars:     []string{"WX_ADDR"},
			},
			&cli.StringFlag{
				Name:        "amqp_url",
				Usage:       "amqp url",
				//Required:    true,
				Value:       "amqp://guest:guest@47.93.20.75:5672/",
				Destination: &amqpURL,
				Aliases:     []string{"au"},
				EnvVars:     []string{"AMQP_URL"},
			},
			&cli.StringFlag{
				Name:        "car_addr",
				Usage:       "car  addr",
				//Required:    true,
				Value:       "localhost:8086",
				Destination: &carAddr,
				Aliases:     []string{"ca"},
				EnvVars:     []string{"CAR_ADDR"},
			},
			&cli.StringFlag{
				Name:        "trip_addr",
				Usage:       "trip addr",
				//Required:    true,
				Value:       "localhost:8082",
				Destination: &tripAddr,
				Aliases:     []string{"ta"},
				EnvVars:     []string{"TRIP_ADDR"},
			},&cli.StringFlag{
				Name:        "ai_addr",
				Usage:       "ai addr",
				//Required:    true,
				Value:       "47.93.20.75:18001",
				Destination: &aiAddr,
				Aliases:     []string{"aa"},
				EnvVars:     []string{"AI_ADDR"},
			},
		},
		EnableBashCompletion:   true,
		Action: func(c *cli.Context) error {
			return nil
		},
	}
	if err  := app.Run(os.Args); err != nil {
		panic(err)
	}
}



func main() {
	logger, err := server.NewZapLogger()

	if err != nil {
		log.Fatalf("cannot create logger:%v",err)
	}

	c := context.Background()

	mongoClient, err := mongo.Connect(c, options.Client().ApplyURI(mongoURI))

	if err != nil {
		logger.Fatal("cannot connect mongodb",zap.Error(err))
	}

	db := mongoClient.Database("coolcar")

	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		logger.Fatal("cannot connect rabbit",zap.Error(err))
	}
	exchange := "coolcar"
	pub ,err := amqpclt.NewPublisher(conn,exchange)
	if err != nil {
		logger.Fatal("cannot create publisher ",zap.Error(err))
	}

	carConn,err :=  grpc.Dial(carAddr,grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect car  client ",zap.Error(err))
	}

	sub,err := amqpclt.NewSubscriber(conn,exchange,logger)
	if err != nil {
		logger.Fatal("cannot create subscriber",zap.Error(err))
	}


	aiConn,err := grpc.Dial(aiAddr,grpc.WithTransportCredentials(insecure.NewCredentials()))
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
		addr  := wsAddr
		logger.Info("HTTP  server started.",zap.String("addr",addr))
		logger.Sugar().Fatal(http.ListenAndServe(addr,nil))
	}()


	tripConn,err := grpc.Dial(tripAddr,grpc.WithTransportCredentials(insecure.NewCredentials()))
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
