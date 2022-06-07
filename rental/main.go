package main

import (
	"context"
	"coolcar/rental/ai"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/trip"
	"coolcar/rental/trip/client/car"
	"coolcar/rental/trip/client/poi"
	"coolcar/rental/trip/client/profile"
	"coolcar/rental/trip/dao"
	coolenvpb "coolcar/shared/coolenv"
	"coolcar/shared/server"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func main() {
	logger,err := server.NewZapLogger()
	if err != nil {
		log.Fatalf("can not create logger,error:%v",err)
	}

	/*lis,err := net.Listen("tcp",":8082")
	if err != nil {
		logger.Fatal("can not listen",zap.Error(err))
		return
	}
	//添加拦截器
	in,err := auth.Interceptor("shared/auth/pub.key")
	if err != nil {
		logger.Fatal("can not Interceptor",zap.Error(err))
		return
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(in))
	rentalpb.RegisterTripServiceServer(s,&trip.Service{
		Logger:                         logger,
	})
	err = s.Serve(lis)*/
	//建立mongodb
	connect, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://47.93.20.75:27017/coolcar?readPreference=primary&ssl=false"))
	if err != nil {
		logger.Fatal("can not connect mongodb",zap.Error(err))
	}
	//logger.Fatal("can not server ",zap.Error(err))
	conn,err := grpc.Dial("47.93.20.75:18001",grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("can not connect AIClient",zap.Error(err))
	}

	logger.Sugar().Fatal(server.RunGRPCServer(&server.GRPCConfig{
		Logger:            logger,
		Addr:              ":8082",
		Name:              "rental",
		AuthPublicKeyFile: "shared/auth/pub.key",
		RegisterFunc: func(server *grpc.Server) {
			rentalpb.RegisterTripServiceServer(server, &trip.Service{
				Logger: logger,
				CarManager: &car.Manager{},
				ProfileManager: &profile.Manager{},
				POIManager: &poi.Manager{},
				DistanceCalc: &ai.Client{
					AIClient:coolenvpb.NewAIServiceClient(conn),
				},
				Mongo: dao.NewMongo(connect.Database("coolcar")),
			})
		},
	}))
}