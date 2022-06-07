package auth

import (
	"context"
	authpb "coolcar/auth/api/gen/v1" //和proto文件的go_package = "coolcar/auth/api/gen/v1;authpb";对应，其中authpb是包名，后边的是路径
	"coolcar/auth/dao"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

//Service implements auth service
type Service struct {
	Log *zap.Logger
	authpb.UnimplementedAuthServiceServer
	OpenIDResolver OpenIDResolver
	Mongo	*dao.Mongo
	TokenGenerator TokenGenerator
	TokenExpire time.Duration
}

type OpenIDResolver interface {
	Resolver(code string)(string,error)
}

//make a token
type TokenGenerator interface {
	GenerateToken(accountID string,expire time.Duration)(string,error)
}

func (s *Service) Login(c context.Context, res *authpb.LoginRequest) (rpn *authpb.LoginResponse,err error) {
	s.Log.Info("received code",zap.String("code",res.Code))
	//login要使用接口的resolver，那么就需要有我login来自己定义
	openId,err := s.OpenIDResolver.Resolver(res.Code)
	if err != nil {
		return nil,status.Errorf(codes.Unavailable,"can not resolve openid, err:%v",err)
	}

	accountId,err := s.Mongo.ResolveAccountId(c,openId)
	if err != nil{
		s.Log.Error("can not resolve account id",zap.Error(err))
		return nil,status.Error(codes.Internal,"")
	}

	tkn,err := s.TokenGenerator.GenerateToken(accountId.String(),s.TokenExpire)

	if err != nil {
		s.Log.Error("can not token",zap.Error(err))
		return nil,status.Error(codes.Internal,"")
	}

	return &authpb.LoginResponse{
		AccessToken: tkn,
		ExpiresIn:   int32(s.TokenExpire.Seconds()),
	},nil
}
