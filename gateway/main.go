package main

import (
	"context"
	authpb "coolcar/auth/api/gen/v1"
	carpb "coolcar/car/api/gen/v1"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/auth"
	"coolcar/shared/server"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"net/http"
	"net/textproto"
	"os"
)

//var addr = flag.String("addr",":8080","address to listen")

var addr string
var authAddr  string
var tripAddr string
var profileAddr string
var carAddr	string


//var Arg0 string
//var Arg1 string
//var Arg2 string

func init() {

	app := &cli.App{
		Usage:                  "please inter gateway options",
		Flags:                  []cli.Flag{
			&cli.StringFlag{
				Name:        "addr",
				Category:    "gateway",
				Usage:       "gateway addr port",
				//Required:    true,
				Value:       ":8080",
				Destination: &addr,
				Aliases:     []string{"a"},
				EnvVars:     []string{"ADDR"},
			},
			&cli.StringFlag{
				Name:        "auth_addr",
				Category:    "service addr",
				Usage:       "auth service addr and port",
				//Required:    true,
				Value:       "localhost:8081",
				Destination: &authAddr,
				Aliases:     []string{"aa"},
				EnvVars:     []string{"AUTH_ADDR"},
			},
			&cli.StringFlag{
				Name:        "trip_addr",
				Category:    "service addr",
				Usage:       "trip service addr and port",
				//Required:    true,
				Value:       "localhost:8082",
				Destination: &tripAddr,
				Aliases:     []string{"ta"},
				EnvVars:     []string{"TRIP_ADDR"},
			},
			&cli.StringFlag{
				Name:        "profile_addr",
				Category:    "service addr",
				Usage:       "profile service addr and port",
				//Required:    true,
				Value:       "localhost:8083",
				Destination: &profileAddr,
				Aliases:     []string{"pa"},
				EnvVars:     []string{"PROFILE_ADDR"},
			},
			&cli.StringFlag{
				Name:        "car_addr",
				Category:    "service addr",
				Usage:       "car service addr and port",
				//Required:    true,
				Value:       "localhost:8086",
				Destination: &carAddr,
				Aliases:     []string{"ca"},
				EnvVars:     []string{"CAR_ADDR"},
			},
		},
		EnableBashCompletion:   true,
		Action: func(c *cli.Context) error {
			//fmt.Println("===========NArg================",c.NArg())
			//fmt.Println("===========len================",c.Args().Len())
			//fmt.Println("===========Get0================",c.Args().Get(0))
			//fmt.Println("===========Get1================",c.Args().Get(1))
			//fmt.Println("===========Get2================",c.Args().Get(2))
			//fmt.Println("===========Get3================",c.Args().Get(3))
			//fmt.Println("===========Get4================",c.Args().Get(4))
			//fmt.Println("===========Get5================",c.Args().Get(5))


			//fmt.Printf("gateway addr: %q \n",addr)
			//fmt.Printf("auth addr: %q \n",authAddr)
			//fmt.Printf("trip addr: %q \n",tripAddr)
			//fmt.Printf("profile addr: %q \n",profileAddr)
			//fmt.Printf("car addr: %q \n",carAddr)

			//Arg0 = fmt.Sprintf("%s",c.Args().Get(0))
			//Arg1 = fmt.Sprintf("%s",c.Args().Get(1))
			//Arg2 = fmt.Sprintf("%s",c.Args().Get(2))

			return nil
		},
	}

	//fmt.Println("========参数1=======",carAddr,addr,authAddr,profileAddr,tripAddr)

	if err  := app.Run(os.Args); err != nil {
		panic(err)
	}

	//fmt.Println("========参数2=======",carAddr,addr,authAddr,profileAddr,tripAddr)
}


func main() {
	//flag.Parse()

	//fmt.Println("========Args0---------2=======",Arg0,Arg1,Arg2)

	//os.Exit(0)

	logger, err := server.NewZapLogger()

	if err != nil {
		log.Fatalf("can not create zap logger:%v", err)
	}

	c, cancel := context.WithCancel(context.Background()) //只能用带有cancel的上下文，不能用超时的，会报错client close
	defer cancel()

	mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			Multiline:       false,
			Indent:          "",
			AllowPartial:    false,
			UseProtoNames:   true,
			UseEnumNumbers:  true,
			EmitUnpopulated: true,
			Resolver:        nil,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			AllowPartial:   false,
			DiscardUnknown: false,
			Resolver:       nil,
		},
	}),
		runtime.WithIncomingHeaderMatcher(func(s string) (string, bool) {
			if s == textproto.CanonicalMIMEHeaderKey(runtime.MetadataHeaderPrefix+auth.ImpersonteAccountHeader) {
				//如果客户送了一个伪造的头部,那么就把这个扔掉
				return "", false
			}
			return runtime.DefaultHeaderMatcher(s)
		}))

	serverConfig := []struct {
		name         string
		addr         string
		registerFunc func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)
	}{
		{
			name:         "auth",
			//addr:         "localhost:8081",
			addr:         authAddr,
			registerFunc: authpb.RegisterAuthServiceHandlerFromEndpoint,
		},
		{
			name:         "trip",
			//addr:         "localhost:8082",
			addr:         tripAddr,
			registerFunc: rentalpb.RegisterTripServiceHandlerFromEndpoint,
		}, {
			name:         "profile",
			//addr:         "localhost:8082",
			addr:         profileAddr,
			registerFunc: rentalpb.RegisterProfileServiceHandlerFromEndpoint,
		}, {
			name:         "car",
			//addr:         "localhost:8086",
			addr:         carAddr,
			registerFunc: carpb.RegisterCarServiceHandlerFromEndpoint,
		},
	}

	for _, s := range serverConfig {
		err := s.registerFunc(c, mux, s.addr,
			[]grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			})
		if err != nil {
			logger.Sugar().Fatalf("can not register service %s: %v", s.name, err)
		}
	}

	//err := authpb.RegisterAuthServiceHandlerFromEndpoint(c,mux,"localhost:8081",
	//	[]grpc.DialOption{
	//		grpc.WithTransportCredentials(insecure.NewCredentials()),
	//	})
	//if err != nil {
	//	log.Fatalf("can not register service: %v",err)
	//}
	//
	//err = rentalpb.RegisterTripServiceHandlerFromEndpoint(c,mux,"localhost:8082",[]grpc.DialOption{
	//	grpc.WithTransportCredentials(insecure.NewCredentials()),
	//})
	//if err != nil {
	//	log.Fatalf("can not register trip service: %v",err)
	//}
	//addr := ":8090"
	logger.Sugar().Infof("grpc gateway started at %s", addr)
	logger.Sugar().Fatal(http.ListenAndServe(addr, mux))
}
