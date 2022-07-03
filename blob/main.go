package main

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	"coolcar/blob/blob"
	"coolcar/blob/dao"
	"coolcar/blob/oss"
	"coolcar/shared/server"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"os"
)



var addr string
var mongoURI string
var ossAddr string
var ossSecID string
var ossSecKey string


func init() {

	app := &cli.App{
		Usage:                  "please inter gateway options",
		Flags:                  []cli.Flag{
			&cli.StringFlag{
				Name:        "addr",
				Usage:       "service addr port",
				//Required:    true,
				Value:       ":8083",
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
				Name:        "oss_addr",
				Usage:       "private  key file path",
				//Required:    true,
				Value:       "oss-cn-beijing.aliyuncs.com",
				Destination: &ossAddr,
				Aliases:     []string{"ossaddr"},
				EnvVars:     []string{"OSS_ADDR"},
			},
			&cli.StringFlag{
				Name:        "oss_access_key_id",
				Usage:       "oss access key id",
				//Required:    true,
				Value:       "LTAI5tNHZ4euwgkaJ6gQ6cSx",
				Destination: &ossSecID,
				Aliases:     []string{"wai"},
				EnvVars:     []string{"OSS_ACCESS_KEY_ID"},
			},
			&cli.StringFlag{
				Name:        "oss_secret_key",
				Usage:       "cos  secret key",
				//Required:    true,
				Value:       "wnsMMf6GP1YhH3lK5NwoJK02soNTdI",
				Destination: &ossSecKey,
				Aliases:     []string{"osk"},
				EnvVars:     []string{"OSS_SECRET_KEY"},
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
	logger,err := server.NewZapLogger()
	if err != nil {
		log.Fatalf("cannot create logger :%v",err)
	}

	c := context.Background()

	mogClient,err := mongo.Connect(c,options.Client().ApplyURI(mongoURI))

	if err != nil {
		logger.Fatal("can not connect mongodb",zap.Error(err))
	}
	db := mogClient.Database("coolcar")

	st,err := oss.NewService(ossAddr,ossSecID,ossSecKey)

	if err != nil {
		log.Fatalf("cannot create storage  :%v",err)
	}

	logger.Sugar().Fatal(server.RunGRPCServer(&server.GRPCConfig{
		Logger:            logger,
		Addr:              addr,
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
