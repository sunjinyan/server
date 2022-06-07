package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"time"
)

func main() {
	c, err := client.NewClientWithOpts()
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	create, err := c.ContainerCreate(ctx, &container.Config{
		Image: "mongo:4.4",
		ExposedPorts: nat.PortSet{
			"27017/tcp": {},
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
	err = c.ContainerStart(ctx, create.ID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("container started")

	inspect, err := c.ContainerInspect(ctx, create.ID)
	if err != nil {
		panic(err)
	}
	fmt.Println(inspect.NetworkSettings.Ports["27017/tcp"][0])
	time.Sleep(20*time.Second)
	fmt.Println("kill container")
	err = c.ContainerRemove(ctx, create.ID, types.ContainerRemoveOptions{
		Force: true,
	})
	if err != nil {
		panic(err)
	}
}