package main

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main()  {
	conn,err := grpc.Dial("localhost:8086",grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		panic(err)
	}


	cs := carpb.NewCarServiceClient(conn)

	c := context.Background()

	//Create some cars
	//for i := 0; i < 5; i++ {
	//	car, err := cs.CreateCar(c, &carpb.CreateCarRequest{})
	//
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//	fmt.Printf("create car:%s\n",car.Id)
	//
	//}

	//Reset all cars
	res,err := cs.GetCars(c,&carpb.GetCarsRequest{})
	if err != nil {
		panic(err)
	}

	for _, car := range res.Cars {
		_, err = cs.UpdateCar(c, &carpb.UpdateCarRequest{
			Id:     car.Id,
			Status: carpb.CarStatus_LOCKED,
		})
		if err != nil {
			fmt.Printf("cannot reset car %q:%v",car.Id,err)
		}
	}
	fmt.Printf("%d cars are reset.",len(res.Cars))
}