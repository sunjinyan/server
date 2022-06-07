package ai

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	coolenvpb "coolcar/shared/coolenv"
)

type Client struct {
	AIClient coolenvpb.AIServiceClient
}

func (c *Client)DistanceKm(ctx context.Context,last *rentalpb.Location,cur *rentalpb.Location)(float64,error)  {
	resp,err := c.AIClient.MeasureDistance(ctx,&coolenvpb.MeasureDistanceRequest{
			From: &coolenvpb.Location{
				Latitude:  last.Latitude,
				Longitude: last.Longitude,
			},
			To: &coolenvpb.Location{
				Latitude:  cur.Latitude,
				Longitude: cur.Longitude,
			},
	})

	//conn,err := grpc.Dial("47.93.20.75:18001",grpc.WithTransportCredentials(insecure.NewCredentials()))
	//if err != nil {
	//	panic(err)
	//}
	//ac := coolenvpb.NewAIServiceClient(conn)
	////c := context.Background()
	//res, err := ac.MeasureDistance(ctx, &coolenvpb.MeasureDistanceRequest{
	//	From: &coolenvpb.Location{
	//		Latitude:  last.Latitude,
	//		Longitude: last.Longitude,
	//	},
	//	To: &coolenvpb.Location{
	//		Latitude:  cur.Latitude,
	//		Longitude: cur.Longitude,
	//	},
	//})
	//
	if err != nil {
		return 0, err
	}
	//
	//fmt.Printf("%+v\n",res)
	return resp.DistanceKm, nil
}
