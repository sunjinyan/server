package trip

import (
	"context"
	"coolcar/proto/gen/go"
)

/**
type TripServiceServer interface {
	GetTrip(context.Context, *GetTripRequest) (*GetTripResponse, error)
	mustEmbedUnimplementedTripServiceServer()
}
/
 */

// server is used to implement trippb.TripServiceServer
type Service struct {
	trippb.UnimplementedTripServiceServer
}


func (*Service)GetTrip(c context.Context,req *trippb.GetTripRequest) (*trippb.GetTripResponse, error) {
	_ = c.Value("test")
	return &trippb.GetTripResponse{
		Id:   req.Id,
		Trip: &trippb.Trip{
			Start:       "abc",
			End:         "def",
			DurationSec: 3600,
			FeeCent:     10000,
			StartPos: &trippb.Location{
				Latitude:  30,
				Longitude: 120,
			},
			EndPos: &trippb.Location{
				Latitude:  35,
				Longitude: 115,
			},
			PathLocations: []*trippb.Location{
				{
					Latitude:  31,
					Longitude: 119,
				},
				{
					Latitude:  32,
					Longitude: 118,
				},
			},
			Status: trippb.TripStatus_FINISHED,
		},
	},nil
}