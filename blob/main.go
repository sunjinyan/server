package main

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	"coolcar/blob/blob"
	"coolcar/blob/dao"
	"coolcar/blob/oss"
	"coolcar/shared/server"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
)

func main() {
	logger,err := server.NewZapLogger()
	if err != nil {
		log.Fatalf("cannot create logger :%v",err)
	}

	c := context.Background()

	mogClient,err := mongo.Connect(c,options.Client().ApplyURI("mongodb://47.93.20.75:27017/coolcar?readPreference=primary&ssl=false"))

	if err != nil {
		logger.Fatal("can not connect mongodb",zap.Error(err))
	}
	db := mogClient.Database("coolcar")

	st,err := oss.NewService("oss-cn-beijing.aliyuncs.com","LTAI5tNHZ4euwgkaJ6gQ6cSx","wnsMMf6GP1YhH3lK5NwoJK02soNTdI")

	if err != nil {
		log.Fatalf("cannot create storage  :%v",err)
	}

	logger.Sugar().Fatal(server.RunGRPCServer(&server.GRPCConfig{
		Logger:            logger,
		Addr:              ":8083",
		Name:              "blob",
		//AuthPublicKeyFile: "shared/auth/pub.key",
		RegisterFunc: func(s *grpc.Server) {
			blobpb.RegisterBlobServiceServer(s,&blob.Service{
				Mongo:                          dao.NewMongo(db),
				Logger:                         logger,
				Storage: st,
			})
		},
	}))

}
