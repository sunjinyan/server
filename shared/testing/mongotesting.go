package mongotesting

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"testing"
	"time"
)

const (
	image = "mongo:4.4"
	containerPort = "27017/tcp"
)

var mongoURI string

const defaultMongoURI = "mongodb://47.93.20.75:27017/coolcar?readPreference=primary&ssl=false"
//func RunWithMongoInDocker(m *testing.M,mongoURI *string) int {
func RunWithMongoInDocker(m *testing.M) int {
	c, err := client.NewClientWithOpts()
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	create, err := c.ContainerCreate(ctx, &container.Config{
		Image: image,
		ExposedPorts: nat.PortSet{
			containerPort: {},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			"27017/tcp": []nat.PortBinding{
				{
					HostIP:   "127.0.0.1",
					HostPort: "0",
				},
			},
		},
	}, nil,nil, "")
	if err != nil {
		panic(err)
	}

	defer func() {
		err = c.ContainerRemove(ctx, create.ID, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			log.Fatalf("error removing container:%v",err)
		}
	}()

	err = c.ContainerStart(ctx, create.ID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("container started")

	inspect, err := c.ContainerInspect(ctx, create.ID)
	if err != nil {
		panic(err)
	}
	hostPort := inspect.NetworkSettings.Ports["27017/tcp"][0]
	mongoURI = fmt.Sprintf("mongodb://%s:%s",hostPort.HostIP,hostPort.HostPort)
	// fmt.Println(inspect.NetworkSettings.Ports["27017/tcp"][0])
	time.Sleep(20*time.Second)
	fmt.Println("kill container")
	//err = c.ContainerRemove(ctx, create.ID, types.ContainerRemoveOptions{
	//	Force: true,
	//})
	//if err != nil {
	//	panic(err)
	//}
	return m.Run()
}

func NewClient(c context.Context) (*mongo.Client,error){
	return mongo.Connect(c,options.Client().ApplyURI(mongoURI))
}

func NewDefaultClient(c context.Context) (*mongo.Client,error){
	return mongo.Connect(c,options.Client().ApplyURI(defaultMongoURI))
}

func SetupIndexes(c context.Context, d *mongo.Database) error {
	_, err := d.Collection("account").Indexes().CreateOne(c, mongo.IndexModel{
		Keys: bson.D{
			{Key: "open_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	_,err = d.Collection("trip").Indexes().CreateOne(c, mongo.IndexModel{
		Keys: bson.D{
			{Key: "trip.accountid", Value: 1},
			{Key: "trip.status", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{
			"trip.status":1,
		}),
	})
	if err != nil {
		return err
	}
	_, err = d.Collection("profile").Indexes().CreateOne(c, mongo.IndexModel{
		Keys: bson.D{
			{Key: "accountid", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})

	return err
}