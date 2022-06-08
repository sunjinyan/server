package main

import (
	"context"
	authpb "coolcar/auth/api/gen/v1"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/server"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"net/http"
)

func main() {
	logger,err := server.NewZapLogger()

	if err != nil {
		log.Fatalf("can not create zap logger:%v",err)
	}

	c,cancel := context.WithCancel(context.Background())//只能用带有cancel的上下文，不能用超时的，会报错client close
	defer cancel()

	mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard,&runtime.JSONPb{
		MarshalOptions:   protojson.MarshalOptions{
			Multiline:         false,
			Indent:            "",
			AllowPartial:      false,
			UseProtoNames:     true,
			UseEnumNumbers:    true,
			EmitUnpopulated:   true,
			Resolver:          nil,
		},
		UnmarshalOptions:protojson.UnmarshalOptions{
			AllowPartial:      false,
			DiscardUnknown:    false,
			Resolver:          nil,
		},
	}))

	serverConfig := []struct{
		name string
		addr string
		registerFunc func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)
	}{
		{
			name:"auth",
			addr: "localhost:8081",
			registerFunc: authpb.RegisterAuthServiceHandlerFromEndpoint,
		},
		{
			name: "trip",
			addr:"localhost:8082",
			registerFunc: rentalpb.RegisterTripServiceHandlerFromEndpoint,
		},{
			name: "profile",
			addr:"localhost:8082",
			registerFunc: rentalpb.RegisterProfileServiceHandlerFromEndpoint,
		},
	}

	for _,s := range serverConfig{
		err := s.registerFunc(c,mux,s.addr,
			[]grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			})
		if err != nil {
			logger.Sugar().Fatalf("can not register service %s: %v",s.name,err)
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
	addr := ":8080"
	logger.Sugar().Infof("grpc gateway started at %s",addr)
	logger.Sugar().Fatal(http.ListenAndServe(addr,mux))
}