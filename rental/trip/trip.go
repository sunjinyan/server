package trip

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/trip/dao"
	"coolcar/shared/auth"
	"coolcar/shared/id"
	"coolcar/shared/mongo/objid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/rand"
	"time"
)

//Trip service
type Service struct {
	Logger *zap.Logger
	Mongo *dao.Mongo
	rentalpb.UnimplementedTripServiceServer
	ProfileManager ProfileManager
	CarManager CarManager
	POIManager POIManager
	DistanceCalc DistanceCalc
}

//ACL层  防止入侵层(Anti Corruption Layer)
type ProfileManager interface {
	Verify(context.Context, id.AccountId) (id.IdentityID,error)
}

//Car Manager defines the ACL for car management
type CarManager interface {
	Verify(context.Context,id.CarID,*rentalpb.Location) error
	Unlock(context.Context,id.CarID,id.AccountId,id.TripId, string) error
	//Unlock(context.Context,id.CarID) error
	Lock(context.Context,id.CarID) error
}

//POIManager Point of Interest
type POIManager interface {
	Resolve(context.Context,*rentalpb.Location)(string,error)
}



type DistanceCalc interface {
	DistanceKm(context.Context,*rentalpb.Location,*rentalpb.Location)(float64,error)
}


//Create Trip Service
func (s *Service)CreateTrip(ctx context.Context,res *rentalpb.CreateTripRequest) (resp *rentalpb.TripEntity,err  error)  {
	aid,err := auth.AccountIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if res.CarId == "" || res.Start == nil {
		return nil, status.Error(codes.InvalidArgument, "")
	}

	//验证驾驶者身份
	iID, err := s.ProfileManager.Verify(ctx, aid)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	//检测车辆状态
	carId := id.CarID(res.CarId)
	err = s.CarManager.Verify(ctx,carId,res.Start)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	//创建行程:写入数据库，开始计费


	ls := s.calcCurrentStatus(ctx,&rentalpb.LocationStatus{
		Location:     res.Start,
		TimestampSec: nowFunc(),
	},res.Start)


	//获取POI
	//poi, err := s.POIManager.Resolve(ctx,res.Start)
	//if err != nil {
	//	s.Logger.Info("cnanot resolve poi",zap.Stringer("location",res.Start),zap.Error(err))
	//	//可以不返回错误，继续进行下去，因为poi只是一个名称
	//	//return nil, status.Error(codes.FailedPrecondition, err.Error())
	//}
	//
	//ls := &rentalpb.LocationStatus{
	//	Location: res.Start,
	//	PoiName:  poi,
	//}
	tr, err := s.Mongo.CreateTrip(ctx, &rentalpb.Trip{
		AccountId:  aid.String(),
		CarId:      carId.String(),
		Start:      ls,
		Current:    ls,
		Status:     rentalpb.TripStatus_IN_PROGRESS,
		IdentityId: iID.String(),
	})
	if err != nil {
		s.Logger.Warn("cannot create trip",zap.Error(err))
		return nil, status.Error(codes.AlreadyExists, "")
	}

	//车辆开锁
	go func() {
		//形成已经创建成功了
		//希望立刻让用户知道形成已经成功了，开始计费，同时进行开锁动作
		//所以需要在后台开锁
		err := s.CarManager.Unlock(context.Background(), carId,aid,objid.ToTripId(tr.ID),res.AvatarUrl)
		if err != nil {
			s.Logger.Error("cannot unlock car",zap.Error(err))
		}
	}()

	//aid,err := auth.AccountIdFromContext(ctx)
	//if err != nil {
	//	return nil,err
	//}
	//s.Logger.Debug("create a trip ",zap.String("code",res.Start),zap.String("account_id",aid.String()))
	//return &rentalpb.CreateTripResponse{
	//
	//},nil
	//验证驾驶者身份
	//车辆开锁
	//创建形成：写入数据库，开始计费
	return &rentalpb.TripEntity{
		Id:   tr.ID.Hex(),
		Trip: tr.Trip,
	},nil
	//return nil, status.Error(codes.Unimplemented, "")
}

func (s *Service)GetTrip(ctx context.Context,res *rentalpb.GetTripRequest) (resp *rentalpb.Trip,err  error) {
	aid, err := auth.AccountIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tr,err := s.Mongo.GetTrip(ctx,id.TripId(res.Id),aid)
	if err != nil {
		return nil, status.Error(codes.NotFound, "")
	}
	return tr.Trip, nil
	//return nil, status.Error(codes.Unimplemented, "")
}

func (s *Service)GetTrips(ctx context.Context,res *rentalpb.GetTripsRequest) (resp *rentalpb.GetTripsResponse,err  error) {
	aid, err := auth.AccountIdFromContext(ctx)
	if err != nil {
		return nil, err
	}
	trs,err := s.Mongo.GetTrips(ctx,aid,res.Status)
	if err != nil {
		s.Logger.Error("cannot get trips",zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}
	getResp := &rentalpb.GetTripsResponse{}
	for _, tr := range trs {
		getResp.Trips = append(getResp.Trips,&rentalpb.TripEntity{
			Id:   tr.ID.Hex(),
			Trip: tr.Trip,
		})
	}
	return getResp, nil
	//return nil, status.Error(codes.Unimplemented, "")
}

func (s *Service)UpdateTrip(ctx context.Context,res *rentalpb.UpdateTripRequest) (resp *rentalpb.Trip,err  error) {

	aid,err := auth.AccountIdFromContext(ctx)

	if err != nil {
		return nil,status.Error(codes.Unauthenticated,"")
	}


	tid := id.TripId(res.Id)
	tr,err  := s.Mongo.GetTrip(ctx,id.TripId(res.Id),aid)
	if tr == nil {
		return nil, status.Error(codes.Unimplemented, "")
	}

	if tr.Trip.Status == rentalpb.TripStatus_FINISHED {
		return nil, status.Error(codes.FailedPrecondition, "cannot update a finished trip")
	}

	if tr.Trip.Current == nil {
		s.Logger.Error("trip without current set",zap.String("id",tid.String()))
		return nil, status.Error(codes.Internal, "")
	}

	cur := tr.Trip.Current.Location
	if res.Current != nil {
		//tr.Trip.Current.Location = res.Current
		cur = res.Current
	}
	tr.Trip.Current = s.calcCurrentStatus(ctx,tr.Trip.Current,cur)

	if res.EndTrip {
		tr.Trip.End = tr.Trip.Current
		tr.Trip.Status = rentalpb.TripStatus_FINISHED
		err = s.CarManager.Lock(ctx, id.CarID(tr.Trip.CarId))
		if err != nil {
			return nil,status.Errorf(codes.FailedPrecondition,"cannot lock car: %v",err)
		}
	}

	err = s.Mongo.UpdateTrip(ctx, tid, aid, tr.UpdateAt, tr.Trip)
	if err != nil {
		s.Logger.Error("cannot update trip",zap.String("id",tid.String()),zap.String("err",err.Error()))
	}
	return tr.Trip, nil
	//return nil, status.Error(codes.Unimplemented, "")
}

var nowFunc = func() int64 {
	return time.Now().Unix()
}
const centsPerSec = 0.7

func (s *Service) calcCurrentStatus(ctx context.Context,last *rentalpb.LocationStatus,cur *rentalpb.Location) *rentalpb.LocationStatus {

	now := nowFunc()
	elapsedSec := float64(now - last.TimestampSec)

	km, err := s.DistanceCalc.DistanceKm(ctx, last.Location, cur)
	if err != nil {
		s.Logger.Error("cannot calc distance",zap.Error(err))
	}

	//获取POI
	poi, err := s.POIManager.Resolve(ctx,cur)
	if err != nil {
		s.Logger.Info("cnanot resolve poi",zap.Stringer("location",cur),zap.Error(err))
		//可以不返回错误，继续进行下去，因为poi只是一个名称
		//return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return &rentalpb.LocationStatus{
		Location: cur,
		//FeeCent:  last.FeeCent+int32(centsPerSec*elapsedSec),
		FeeCent:  last.FeeCent+int32(centsPerSec*elapsedSec*2*rand.Float64()),//增加点趣味性*2*rand.Float64()
		KmDriven: last.KmDriven + km,
		TimestampSec: now,
		PoiName:poi,
	}
}