package blob

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	"coolcar/blob/dao"
	"coolcar/shared/id"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type Storage interface {
	SignURL(c context.Context,method,path string,timeout time.Duration) (string,error)
	Get(c context.Context,path string)(io.ReadCloser,error)
}

type Service struct {
	Mongo	*dao.Mongo
	Logger	*zap.Logger
	blobpb.UnimplementedBlobServiceServer
	Storage Storage
}

func (s *Service)CreateBlob(ctx context.Context, req *blobpb.CreateBlobRequest) (resp *blobpb.CreateBlobResponse,err error)  {
	aid := id.AccountId(req.AccountId)
	br,err :=  s.Mongo.CreateBlob(ctx,aid)
	if err != nil {
		s.Logger.Error("cannot create bolb",zap.Error(err))
		return nil,status.Error(codes.Internal,"")
	}
	url, err := s.Storage.SignURL(ctx, http.MethodPut, br.Path, secToDuration(req.UploadUrlTimeoutSec))
	if err != nil {
		return nil,status.Errorf(codes.Aborted,"cannot sign url: %v",err)
	}
	return &blobpb.CreateBlobResponse{
		Id:        br.ID.Hex(),
		UploadUrl: url,
	},nil
	//return nil, status.Error(codes.Unimplemented, "")
}

func (s *Service)GetBlob(ctx context.Context, req *blobpb.GetBlobRequest) (resp *blobpb.GetBlobResponse,err error)  {
	br,err := s.getBlobRecord(ctx,id.BlobID(req.Id))
	if err != nil {
		return nil,err
	}
	get, err := s.Storage.Get(ctx, br.Path)

	if get != nil {
		defer get.Close()
	}
	if err != nil {
		return nil,status.Errorf(codes.Aborted,"cannot get storage :%v",err)
	}
	all, err := ioutil.ReadAll(get)
	if err != nil {
		return nil,status.Errorf(codes.Aborted,"cannot read from response:%v",err)
	}
	return &blobpb.GetBlobResponse{
		Data: all,
	},nil
	//return nil, status.Error(codes.Unimplemented, "")
}

func (s *Service)GetBlobURL(ctx context.Context, req *blobpb.GetBlobURLRequest) (resp *blobpb.GetBlobURLResponse,err error)  {
	br,err := s.getBlobRecord(ctx,id.BlobID(req.Id))
	if err != nil {
		return nil,status.Errorf(codes.NotFound,"cannot read from response:%v",err)
	}

	url, err := s.Storage.SignURL(ctx, http.MethodGet, br.Path, secToDuration(req.TimeoutSec))
	if err != nil {
		return nil,status.Errorf(codes.Aborted,"cannot read from response:%v",err)
	}
	return &blobpb.GetBlobURLResponse{Url: url},nil
	//return nil, status.Error(codes.Unimplemented, "")
}

func (s *Service)getBlobRecord(c context.Context,bid id.BlobID)(*dao.BlobRecord,error)  {
	blob, err := s.Mongo.GetBlob(c, bid)
	if err == mongo.ErrNoDocuments {
		return nil,status.Error(codes.NotFound,"")
	}

	if err != nil {
		return nil, status.Error(codes.InvalidArgument,err.Error())
	}
	return blob,nil
}

func secToDuration(sec int32) time.Duration {
	return  time.Duration(sec) * time.Second
}