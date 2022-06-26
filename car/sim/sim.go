package sim

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
	"coolcar/car/mq"
	coolenvpb "coolcar/shared/coolenv"
	"fmt"
	"go.uber.org/zap"
	"time"
)


type PosSubscriber interface {
	Subscribe(context.Context)(chan *coolenvpb.CarPosUpdate, func(),error)//这里的chan，不要写amqp的ch，需要与业务相关的，不管外边如何送过来的，都需要是这样
}

type Controller struct {
	CarService carpb.CarServiceClient
	Logger *zap.Logger
	CarSubscriber mq.Subscriber
	PosSubscriber PosSubscriber
	AIService coolenvpb.AIServiceClient
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

	msgCh,cleanUp,err := c.CarSubscriber.Subscribe(ctx)
	defer cleanUp()


	if err != nil {
		c.Logger.Error("cannot subscribe",zap.Error(err))
		return
	}


	posCh,posCleanUp,err := c.PosSubscriber.Subscribe(ctx)
	defer posCleanUp()


	if err != nil {
		c.Logger.Error("cannot Pos Subscriber",zap.Error(err))
		return
	}


	res,err := c.CarService.GetCars(ctx,&carpb.GetCarsRequest{})
	if err != nil {
		c.Logger.Error("cannot get cars",zap.Error(err))
		return
	}

	carChans := make(map[string]chan *carpb.Car)
	posChans := make(map[string]chan *carpb.Location)
	for _, car := range res.Cars {
		carFanoutCh := make(chan *carpb.Car)
		carChans[car.Id] = carFanoutCh
		posFanoutCh := make(chan *carpb.Location)
		posChans[car.Id] = posFanoutCh
		go c.SimulateCar(ctx,car,carFanoutCh,posFanoutCh)
	}

	//for carUpdate := range msgCh {
	for  {
		select {
		case carUpdate := <- msgCh:
			ch := carChans[carUpdate.Id]
			if ch != nil {
				go func() {
					ch <- carUpdate.Car
				}()
			}
		case posUpdate := <-posCh:
			ch := posChans[posUpdate.CarId]
			if ch != nil {
				ch <- &carpb.Location{
					Latitude:  posUpdate.Pos.Latitude,
					Longitude: posUpdate.Pos.Longitude,
				}
			}
		}
	}
}

func (c *Controller)SimulateCar(ctx context.Context,initial *carpb.CarEntity,carCh chan *carpb.Car,posCh chan *carpb.Location)  {
	//carID := initial.Id
	car := initial
	c.Logger.Info("Running  car simulations",zap.String("id",car.Id))
	//for update := range carCh {
	for {
		select {
		case update := <-carCh:
			if update.Status == carpb.CarStatus_UNLOCKING {
				//car.Car.Status = carpb.CarStatus_UNLOCKING
				//_, err := c.CarService.UpdateCar(ctx, &carpb.UpdateCarRequest{
				//	Id:     car.Id,
				//	Status: carpb.CarStatus_UNLOCKING,
				//})
				//if err != nil {
				//	c.Logger.Error("cannot unlock car",zap.Error(err))
				//	return
				//}
				//_,err = c.AIService.SimulateCarPos(ctx,&coolenvpb.SimulateCarPosRequest{
				//	CarId:      car.Id,
				//	InitialPos: &coolenvpb.Location{
				//		Latitude:  car.Car.Position.Latitude,
				//		Longitude: car.Car.Position.Longitude,
				//	},
				//	Type: coolenvpb.PosType_RANDOM,
				//})
				//if err != nil {
				//	c.Logger.Error("cannot simulate car pos",zap.Error(err))
				//}
				unlockCar, err := c.unlockCar(ctx, car)
				if err != nil {
					c.Logger.Error("cannot unlock car ",zap.Error(err))
					break
				}
				car = unlockCar
			}else if update.Status == carpb.CarStatus_LOCKING {

				update,err := c.lockCar(ctx,car)
				if err != nil {
					c.Logger.Error("cannot lock car ",zap.Error(err))
					break
				}
				car = update
				//car.Car.Status = carpb.CarStatus_LOCKED
				//_, err := c.CarService.UpdateCar(ctx, &carpb.UpdateCarRequest{
				//	Id:     car.Id,
				//	Status: carpb.CarStatus_LOCKED,
				//})
				//if err != nil {
				//	c.Logger.Error("cannot unlock car",zap.Error(err))
				//	return
				//}
				//_,err = c.AIService.EndSimulateCarPos(ctx,&coolenvpb.EndSimulateCarPosRequest{CarId: car.Id})
				//if err != nil {
				//	c.Logger.Error("cannot end simulate car pos",zap.Error(err))
				//}

			}
		case pos := <-posCh:
			updated,err := c.moveCar(ctx,car,pos)
			if err != nil {
				c.Logger.Error("cannot move car ",zap.Error(err))
				break
			}
			car = updated
		}

	}
}

func (c *Controller)lockCar(ctx context.Context,car *carpb.CarEntity)(*carpb.CarEntity,error)  {

	car.Car.Status = carpb.CarStatus_LOCKED
	_, err := c.CarService.UpdateCar(ctx, &carpb.UpdateCarRequest{
		Id:     car.Id,
		Status: carpb.CarStatus_LOCKED,
	})
	if err != nil {
		c.Logger.Error("cannot unlock car",zap.Error(err))
		return nil,fmt.Errorf("cannot update car state:%v",err)
	}
	_,err = c.AIService.EndSimulateCarPos(ctx,&coolenvpb.EndSimulateCarPosRequest{CarId: car.Id})
	if err != nil {
		c.Logger.Error("cannot end simulate car pos",zap.Error(err))
		return nil, fmt.Errorf("cannot end simulate car pos : %v",err)
	}
	return car,nil
}

func (c *Controller)unlockCar(ctx context.Context,car *carpb.CarEntity)(*carpb.CarEntity,error)  {
	car.Car.Status = carpb.CarStatus_UNLOCKING
	_, err := c.CarService.UpdateCar(ctx, &carpb.UpdateCarRequest{
		Id:     car.Id,
		Status: carpb.CarStatus_UNLOCKING,
	})
	if err != nil {
		c.Logger.Error("cannot unlock car",zap.Error(err))
		return nil, fmt.Errorf("cannot unlock car: %v",err)
	}
	_,err = c.AIService.SimulateCarPos(ctx,&coolenvpb.SimulateCarPosRequest{
		CarId:      car.Id,
		InitialPos: &coolenvpb.Location{
			Latitude:  car.Car.Position.Latitude,
			Longitude: car.Car.Position.Longitude,
		},
		Type: coolenvpb.PosType_NINGBO,
	})
	if err != nil {
		c.Logger.Error("cannot simulate car pos",zap.Error(err))
		return nil, fmt.Errorf("cannot simulate car pos: %v",err)
	}
	return car,nil
}

func (c *Controller)moveCar	(ctx context.Context,car *carpb.CarEntity,pos *carpb.Location)(*carpb.CarEntity,error)  {
	car.Car.Position = pos
	_,err := c.CarService.UpdateCar(ctx,&carpb.UpdateCarRequest{
		Id:       car.Id,
		Position: pos,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot update car: %v",err)
	}
	return car,nil
}