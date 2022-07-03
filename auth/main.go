package main

import (
	"context"
	authpb "coolcar/auth/api/gen/v1"
	"coolcar/auth/auth"
	"coolcar/auth/dao"
	"coolcar/auth/token"
	"coolcar/auth/wechat"
	"coolcar/shared/server"
	"github.com/dgrijalva/jwt-go"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"os"
	"time"
)

const privateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEoQIBAAKCAQBlar2QFxQILPZWMJPL/YXw8uzfDwLOErj9tZRqhTKeaGQi71dT
z5eaJiyy6siRGxpAOC21w3K1odFh2NYQisJjUzH9mbD9Zrsxn80GNiy7qoO49BQN
vKR+r9BpG7tp8kllvZUbukAsua+Ltu6L1//Du/vnZ/o+QZK46yTBliQiR2ZR1LwG
7/oDBrkNRdiL0PEOiUu1yjjBGJu+QH3z7QV+xFqgxuA5qwRJKr8qFx2jI+XGbQ+T
Cnmc0qFlG5UwXQdhQVmRPCKIv9RAPQkwFJWYgm2gLMFE/BF1HGiXqSSzoiKm34UZ
JOfaCK0DgAlD8lXPTU4luou+fvARZUIt0dNrAgMBAAECggEARZMUuOUcOP+ff0GU
Iz2pxcLh/aSup/DwOB151BnMwB+dktnfbv/bYUUPJS8sqV+OgsAHm8qZx0FcA2Re
drq64KkSgogNg8oWYdTfMoO37IbuFtRbbZEcyEyVDYaY6/XrpICK6pq7q/M3GwJE
ZGuqav1rszUB1/PkVDf831HBOEIllcXqMD8tPydKlF/j9B0/8RzEsO0q/yKIBHv5
lKVnwwtzo2kplKf0HFuyaIHyhPkzwP6SADFC98kPozccPCHueHxYNoJeFRZ+d6p9
m7iKxtF3ZGEN6i7H+dTbLz2hjBZGVUyyQaItYs+e9SFIysR8gTxLIeyPi6EsUcDF
PgwuAQKBgQCt65Y+gnWPwZ3fs5s5EciPSg6QnD8XkMb1JAgqA3niRN8fne1Bpq85
o6lAVar1Q1N6IZ2M/lSFvEB5pIsNOsfRlI4lRRT+koVP2UMtgsvU5oiUN6Ihr8pD
NxFJL62f3QRIAtlJGwfN4CWap7MweiSuJhQW6zv8Mw1eSDkzfIzXowKBgQCVR5Fk
+cQKbKhiNkJDKpV6SJkJAZKfkeeQ4UH+77oJTgLNDljoYVLWGdiBifdkR+DcR8fu
dEJqcSyfwZ6QaddKa7RieHW8hGok2HCJGh4UVOPP/c/B9LCElPlUOQOfYpkDBk61
n/qUNU2nAxB1DYIuGMZa83sueSC3NmvhgxJxmQKBgELqNHlCenbf+Jz8Hom7lN3E
kYwEAaZQdqwUu+tmQPvUocApZAJxXlOf61usWkVZLQH9cv+vTtfRkUD8FN/3MLAr
JvGh/azgtNf+8IDPJRt5wyy7pu6tEvX/dvAgnv217JLEPdOJydvlFLLHOQM6y2gR
kIMs6HxlLAyNXyy3C/1fAoGARVKA4GVkdnrTDxinWM2TsL+54YbLcjKtWIhiv2LP
/7YsctEw1GktsKM7+Zv6OKVwdJsY61Et7oQz9tRRMDsWcUbm87uA4VSUfhvM1r48
LbDmQhZZvnZE6qzLxCLd3cxJxi/NqWZrVRwTvXUc1V66j3GN23qqP4CcgrhRDN5z
WDECgYBSoAlUWzl9UjIyzKh4Lxo+NNq9ksVmYOI1Yip2aJbqPXGGX0AFAXEqQVU+
sJ6VRtPBrq9mfQ0VHOlLt/2ZYG5rPuBWGxwJC9OF+nWew2uWphLxqha2o1bot4Sh
n1CyPlTZDM9IrUSFJ44iFgx08tUdgvxFX3N8jfgQl3k8I5hvVQ==
-----END RSA PRIVATE KEY-----`

var addr string
var mongoURI string
var privateKeyFile string
var wechatAppID string
var wechatAppSecret string


func init() {

	app := &cli.App{
		Usage:                  "please inter gateway options",
		Flags:                  []cli.Flag{
			&cli.StringFlag{
				Name:        "addr",
				Usage:       "service addr port",
				//Required:    true,
				Value:       ":8081",
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
				Name:        "private_key_file",
				Usage:       "private  key file path",
				//Required:    true,
				Value:       "auth/prev.key",
				Destination: &privateKeyFile,
				Aliases:     []string{"pkf"},
				EnvVars:     []string{"PRIVATE_KEY_FILE"},
			},
			&cli.StringFlag{
				Name:        "wechat_app_id",
				Usage:       "wechat app ID",
				//Required:    true,
				Value:       "wx851020fe449a84e2",
				Destination: &wechatAppID,
				Aliases:     []string{"wai"},
				EnvVars:     []string{"WECHAT_APP_ID"},
			},
			&cli.StringFlag{
				Name:        "wechat_app_secret",
				Usage:       "wechat app secret",
				//Required:    true,
				Value:       "26c95bc945fd0d7ccf7c473ef5a5e7f8",
				Destination: &wechatAppSecret,
				Aliases:     []string{"was"},
				EnvVars:     []string{"WECHAT_APP_SECRET"},
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
		log.Fatalf("can not create logger, error:%v",err)
		return
	}

	//建立mongodb
	connect, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		logger.Fatal("can not connect mongodb",zap.Error(err))
	}

	pkFile,err := os.Open(privateKeyFile)
	if err != nil {
		log.Fatalf("can not open private key:%v",err)
	}
	pkBytes, err := ioutil.ReadAll(pkFile)
	if err != nil {
		log.Fatalf("can not read private key:%v",err)
	}
	pem,err := jwt.ParseRSAPrivateKeyFromPEM(pkBytes)
	if err != nil {
		log.Fatalf("can not parse private key:%v",err)
	}

	logger.Sugar().Fatal(server.RunGRPCServer(&server.GRPCConfig{
		Logger:            logger,
		//Addr:              ":8081",
		Addr:              addr,
		Name:              "auth",
		RegisterFunc: func(server *grpc.Server) {
			authpb.RegisterAuthServiceServer(server,&auth.Service{
				OpenIDResolver: &wechat.Service{
					AppID: wechatAppID,
					AppSecret: wechatAppSecret,
				},
				Log: logger,
				Mongo: dao.NewMongo(connect.Database("coolcar")),
				TokenExpire: 2 * time.Hour,
				TokenGenerator: token.NewJWTTokenGen("coolcar/auth",pem),
			})
		},
	}))

	//lis, err := net.Listen("tcp", ":8081")
	//if err != nil {
	//	logger.Fatal("can not listen",zap.Error(err))
	//	return
	//}




	//s := grpc.NewServer()
	//authpb.RegisterAuthServiceServer(s,&auth.Service{
	//	OpenIDResolver: &wechat.Service{
	//		AppID: "wx851020fe449a84e2",
	//		AppSecret: "26c95bc945fd0d7ccf7c473ef5a5e7f8",
	//	},
	//	Log: logger,
	//	Mongo: dao.NewMongo(connect.Database("coolcar")),
	//	TokenExpire: 2 * time.Hour,
	//	TokenGenerator: token.NewJWTTokenGen("coolcar/auth",pem),
	//})
	//err = s.Serve(lis)
	//logger.Fatal("can not server ",zap.Error(err))
}