package main

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn,err := grpc.Dial("localhost:8083", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	c := blobpb.NewBlobServiceClient(conn)

	ctx := context.Background()
	blob, err := c.CreateBlob(ctx, &blobpb.CreateBlobRequest{
		AccountId:           "account_1",
		UploadUrlTimeoutSec: 1000,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n",blob)
}
