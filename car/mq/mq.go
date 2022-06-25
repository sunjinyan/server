package mq

import (
	"context"
	carpb "coolcar/car/api/gen/v1"
)

//由于收消息不知道怎么收，所以定义接口，然后让使用的人来传递进来
type Subscriber interface {
	Subscribe(context.Context)(chan *carpb.CarEntity, func(),error)//这里的chan，不要写amqp的ch，需要与业务相关的，不管外边如何送过来的，都需要是这样
}


type Publisher interface {
	Publish(context.Context,*carpb.CarEntity) error
}