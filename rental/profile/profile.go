package profile

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/profile/dao"
	"coolcar/shared/auth"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type Service struct {
	Mongo *dao.Mongo
	Logger *zap.Logger
	rentalpb.UnimplementedProfileServiceServer
}

func (s *Service)GetProfile(c context.Context,req *rentalpb.GetProfileRequest)(resp *rentalpb.Profile ,err error)  {

	aid,err := auth.AccountIdFromContext(c)
	if err != nil {
		return nil,err
	}
	p,err := s.Mongo.GetProfile(c,aid)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &rentalpb.Profile{},nil
		}
		s.Logger.Error("cannot get profile",zap.Error(err))
		return nil,status.Error(codes.Internal,"")
	}
	return p,nil
}

func (s *Service)SubmitProfile(c context.Context,i *rentalpb.Identity)(resp *rentalpb.Profile ,err error){
	aid,err := auth.AccountIdFromContext(c)
	if err != nil {
		return nil,err
	}

	//只有 rentalpb.IdentityStatus_UNSUBMITTED状态时候才可以修改，其他状态不可以修改，所以需要使用乐观锁的方式，去先查询，再确认是否可以修改
	//使用timestamp 来保证，update只是针对的是当前这个请求中GetProfile出来的信息，而不是与其他信息互斥
	//p,err := s.Mongo.GetProfile(c,aid)

	//也可以使用rentalpb.IdentityStatus_UNSUBMITTED来限制条件

	p := &rentalpb.Profile{
		Identity:       i,
		IdentityStatus: rentalpb.IdentityStatus_PENDING,
	}
	err = s.Mongo.UpdateProfile(c,aid,rentalpb.IdentityStatus_UNSUBMITTED,p)
	if err != nil {
		s.Logger.Error("cannot get profile",zap.Error(err))
		return nil,status.Error(codes.Internal,"")
	}

	go func() {
		time.Sleep(3 * time.Second)
		err = s.Mongo.UpdateProfile(context.Background(), aid, rentalpb.IdentityStatus_PENDING, &rentalpb.Profile{
		//err = s.Mongo.UpdateProfile(c, aid, rentalpb.IdentityStatus_PENDING, &rentalpb.Profile{
			Identity:       i,
			IdentityStatus: rentalpb.IdentityStatus_VERIFIED,
		})
		if err != nil {
			s.Logger.Error("cannot get profile",zap.Error(err))
			//return nil,status.Error(codes.Internal,"")
		}
	}()

	return p,nil
}

func (s *Service)ClearProfile(c context.Context,req *rentalpb.ClearProfileRequest)  (resp *rentalpb.Profile ,err error){

	aid,err := auth.AccountIdFromContext(c)
	if err != nil {
		return nil,err
	}

	p := &rentalpb.Profile{}
	err = s.Mongo.UpdateProfile(c,aid,rentalpb.IdentityStatus_VERIFIED,p)
	if err != nil {
		s.Logger.Error("cannot get profile",zap.Error(err))
		return nil,status.Error(codes.Internal,"")
	}
	return p,nil
}