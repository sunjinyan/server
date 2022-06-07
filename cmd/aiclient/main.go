package main

import (
	"context"
	coolenvpb "coolcar/shared/coolenv"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn,err := grpc.Dial("47.93.20.75:18001",grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	ac := coolenvpb.NewAIServiceClient(conn)
	c := context.Background()
	res, err := ac.MeasureDistance(c, &coolenvpb.MeasureDistanceRequest{
		From: &coolenvpb.Location{
			Latitude:  30,
			Longitude: 120,
		},
		To: &coolenvpb.Location{
			Latitude:  31,
			Longitude: 121,
		},
	})

	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n",res)
}
