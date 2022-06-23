package sim

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	"go.uber.org/zap"
	"time"
)


//由于收消息不知道怎么收，所以定义接口，然后让使用的人来传递进来
type Subscriber interface {
	Subscribe(context.Context)(chan *carpb.CarEntity, func(),error)//这里的chan，不要写amqp的ch，需要与业务相关的，不管外边如何送过来的，都需要是这样
}



type Controller struct {
	CarService carpb.CarServiceClient
	Logger *zap.Logger
	Subscriber Subscriber
}


//车辆总控，涉及到控制车辆，收发消息

//收取消息，然后进行分发到具体某一个车辆
func (c *Controller)RunSimulations(ctx context.Context)  {
	// conn,err := grpc.Dial("localhost:8085",grpc.WithTransportCredentials(insecure.NewCredentials()))
	//
	//
	////conn,err := grpc.Dial("localhost:8085",grpc.WithTransportCredentials(insecure.NewCredentials()))
	//
	//if err != nil {
	//	panic(err)
	//}
	//
	//
	//cs := carpb.NewCarServiceClient(conn)
	//
	//c.CarService = cs

	var cars []*carpb.CarEntity
	for  {
		time.Sleep(3 * time.Second)
		res,err := c.CarService.GetCars(ctx,&carpb.GetCarsRequest{})
		if err != nil {
			c.Logger.Error("cannot get cars",zap.Error(err))
		}
		cars = res.Cars
		break
	}
	c.Logger.Info("Running  car simulations",zap.Int("car_count",len(cars)))

	msgCh,cleanUp,err := c.Subscriber.Subscribe(ctx)
	defer cleanUp()


	if err != nil {
		c.Logger.Error("cannot subscribe",zap.Error(err))
		return
	}


	res,err := c.CarService.GetCars(ctx,&carpb.GetCarsRequest{})
	if err != nil {
		c.Logger.Error("cannot get cars",zap.Error(err))
		return
	}

	carChans := make(map[string]chan *carpb.Car)
	for _, car := range res.Cars {
		ch := make(chan *carpb.Car)
		carChans[car.Id] = ch
		go c.SimulateCar(ctx,car,ch)
	}

	for carUpdate := range msgCh {
		ch := carChans[carUpdate.Id]
		if ch != nil {
			ch <- carUpdate.Car
		}
	}
}

func (c *Controller)SimulateCar(ctx context.Context,initial *carpb.CarEntity,ch chan *carpb.Car)  {
	carID := initial.Id
	c.Logger.Info("Running  car simulations",zap.String("id",carID))
	for update := range ch {
		if update.Status == carpb.CarStatus_UNLOCKING {
			_, err := c.CarService.UpdateCar(ctx, &carpb.UpdateCarRequest{
				Id:     carID,
				Status: carpb.CarStatus_UNLOCKING,
			})
			if err != nil {
				c.Logger.Error("cannot unlock car",zap.Error(err))
				return
			}
		}else if update.Status == carpb.CarStatus_LOCKING {
			_, err := c.CarService.UpdateCar(ctx, &carpb.UpdateCarRequest{
				Id:     carID,
				Status: carpb.CarStatus_LOCKED,
			})
			if err != nil {
				c.Logger.Error("cannot unlock car",zap.Error(err))
				return
			}
		}
	}
}