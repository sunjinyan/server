package main

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	carpb "coolcar/car/api/gen/v1"
	"coolcar/rental/ai"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/profile"
	profiledao "coolcar/rental/profile/dao"
	"coolcar/rental/trip"
	"coolcar/rental/trip/client/car"
	"coolcar/rental/trip/client/poi"
	profClient "coolcar/rental/trip/client/profile"
	tripdao "coolcar/rental/trip/dao"
	coolenvpb "coolcar/shared/coolenv"
	"coolcar/shared/server"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var addr string
var mongoURI string
var blobAddr string
var carAddr string
var authPublicKeyFile string
var aiAddr string


func init() {

	app := &cli.App{
		Usage:                  "please inter gateway options",
		Flags:                  []cli.Flag{
			&cli.StringFlag{
				Name:        "addr",
				Usage:       "service addr port",
				//Required:    true,
				Value:       ":8082",
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
				Name:        "blob_addr",
				Usage:       "blob addr",
				//Required:    true,
				Value:       "localhost:8083",
				Destination: &blobAddr,
				Aliases:     []string{"ba"},
				EnvVars:     []string{"BLOB_ADDR"},
			},
			&cli.StringFlag{
				Name:        "auth_public_key_file",
				Usage:       "auth public key file",
				//Required:    true,
				Value:       "shared/auth/public.key",
				Destination: &authPublicKeyFile,
				Aliases:     []string{"APK"},
				EnvVars:     []string{"AUTH_PUBLIC_KEY_FILE"},
			},
			&cli.StringFlag{
				Name:        "car_addr",
				Usage:       "car  addr",
				//Required:    true,
				Value:       "localhost:8086",
				Destination: &carAddr,
				Aliases:     []string{"ca"},
				EnvVars:     []string{"CAR_ADDR"},
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
		log.Fatalf("can not create logger,error:%v", err)
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
	connect, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		logger.Fatal("can not connect mongodb", zap.Error(err))
	}
	//logger.Fatal("can not server ",zap.Error(err))
	conn, err := grpc.Dial(aiAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("can not connect AIClient", zap.Error(err))
	}

	db := connect.Database("coolcar")

	blobConn, err := grpc.Dial(blobAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect blob service", zap.Error(err))
	}

	profService := &profile.Service{
		Mongo:             profiledao.NewMongo(db),
		Logger:            logger,
		BlobClient:        blobpb.NewBlobServiceClient(blobConn),
		PhotoGetExpire:    5 * time.Second,
		PhotoUploadExpire: 10 * time.Second,
	}

	carConn, err := grpc.Dial(carAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("cannot connect car service", zap.Error(err))
	}

	logger.Sugar().Fatal(server.RunGRPCServer(&server.GRPCConfig{
		Logger:            logger,
		Addr:              addr,
		Name:              "rental",
		AuthPublicKeyFile: authPublicKeyFile,
		RegisterFunc: func(server *grpc.Server) {
			rentalpb.RegisterTripServiceServer(server, &trip.Service{
				Logger: logger,
				CarManager: &car.Manager{
					CarService: carpb.NewCarServiceClient(carConn),
				},
				ProfileManager: &profClient.Manager{
					Fetcher: profService,
				},
				POIManager: &poi.Manager{},
				DistanceCalc: &ai.Client{
					AIClient: coolenvpb.NewAIServiceClient(conn),
				},
				Mongo: tripdao.NewMongo(db),
			})
			rentalpb.RegisterProfileServiceServer(server, profService)
		},
	}))
}
