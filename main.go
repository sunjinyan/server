package main

import (
	"context"
	trippb "coolcar/proto/gen/go"
	"coolcar/trip"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"net"
	"net/http"
)

func main() {
	log.SetFlags(log.Lshortfile)

	lis, err := net.Listen("tcp", ":8081")
	go startGrpcGateway()

	if err != nil {
		log.Fatalf("failed to lesten: v% \n", err)
	}
	s := grpc.NewServer()
	trippb.RegisterTripServiceServer(s,&trip.Service{})
	log.Fatal(s.Serve(lis))
}

func startGrpcGateway() {
	c := context.Background()//生成了没有具体内容的上下文
	c,cancel := context.WithCancel(c)//具有cancel能力的上下文
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
	err := trippb.RegisterTripServiceHandlerFromEndpoint(
		c,                     //通过这个context来链接服务
		mux, //mux: multiplexer 就是1 对多的意思，分发器，链接注册在NewServeMux
		":8081",
		[]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials())}, //如果链接，使用http铭文链接，不是用https
	)
	if err != nil{
		log.Fatalf("grpc gateway err: v%",err)
	}
	err = http.ListenAndServe(":8080", mux)
	if err != nil{
		log.Fatalf("grpc gateway err: v%",err)
	}
}