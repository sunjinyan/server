package main

import (
	"context"
	trippb "coolcar/proto/gen/go"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func main() {
	log.SetFlags(log.Lshortfile)
	c, err := grpc.Dial("localhost:8081",grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("can not connect server:%v",err)
	}
	tsClient := trippb.NewTripServiceClient(c)
	trip, err := tsClient.GetTrip(context.Background(), &trippb.GetTripRequest{Id: "trip567"})
	if err != nil {
		log.Fatalf("can not connect server:%v",err)
	}
	fmt.Println(trip)
}
