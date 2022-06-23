package profile

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/rental/profile/dao"
	"coolcar/shared/auth"
	"coolcar/shared/id"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"net/http/httptest"
	"time"
)

type Service struct {
	BlobClient blobpb.BlobServiceClient
	Mongo *dao.Mongo
	Logger *zap.Logger
	PhotoGetExpire time.Duration
	PhotoUploadExpire time.Duration
	rentalpb.UnimplementedProfileServiceServer
}

func (s *Service) logAndConvertProfileErr(err error) codes.Code {
	if err == mongo.ErrNoDocuments {
		return codes.NotFound
	}
	s.Logger.Error("cannot get profile",zap.Error(err))
	return codes.Internal
}

func (s *Service)GetProfile(c context.Context,req *rentalpb.GetProfileRequest)(resp *rentalpb.Profile ,err error)  {

	aid,err := auth.AccountIdFromContext(c)
	if err != nil {
		return nil,err
	}
	p,err := s.Mongo.GetProfile(c,aid)
	if err != nil {
		code := s.logAndConvertProfileErr(err)
		if code == codes.NotFound {
			return &rentalpb.Profile{},nil
		}
		//if err == mongo.ErrNoDocuments {
		//	return &rentalpb.Profile{},nil
		//}
		//s.Logger.Error("cannot get profile",zap.Error(err))
		return nil,status.Error(code,"")
	}
	if p.Profile == nil {
		return &rentalpb.Profile{},nil
	}
	return p.Profile,nil
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

func (s *Service)GetProfilePhoto(c context.Context, req *rentalpb.GetProfilePhotoRequest) (resp *rentalpb.GetProfilePhotoResponse, err error) {
	aid,err := auth.AccountIdFromContext(c)
	if err != nil {
		return nil,err
	}

	pr,err := s.Mongo.GetProfile(c,aid)
	if err != nil {
		return nil, status.Error(s.logAndConvertProfileErr(err),"")
	}
	if pr.PhotoBlobID == "" {
		return nil, status.Error(codes.NotFound, "")
	}
	url, err := s.BlobClient.GetBlobURL(c, &blobpb.GetBlobURLRequest{
		Id:         pr.PhotoBlobID,
		TimeoutSec: int32(s.PhotoGetExpire.Seconds()),
	})
	if err != nil {
		s.Logger.Error("cannot get url",zap.Error(err))
		return nil, status.Error(codes.NotFound, "")
	}
	return &rentalpb.GetProfilePhotoResponse{Url: url.Url}, nil
	//return nil, status.Errorf(codes.Unimplemented, "method GetProfilePhoto not implemented")
}


func (s *Service) CreateProfilePhoto(c context.Context, req *rentalpb.CreateProfilePhotoRequest) (resp *rentalpb.CreateProfilePhotoResponse, err error) {
	aid,err := auth.AccountIdFromContext(c)
	if err != nil {
		return nil,err
	}
	blob, err := s.BlobClient.CreateBlob(c, &blobpb.CreateBlobRequest{
		AccountId:           aid.String(),
		UploadUrlTimeoutSec: int32(s.PhotoUploadExpire.Seconds()),
	})

	if err != nil {
		s.Logger.Error("cannot get blob",zap.Error(err))
		return nil, status.Error(codes.Aborted, "")
	}

	err = s.Mongo.UpdateProfilePhoto(c, aid, id.BlobID(blob.Id))
	if err != nil {
		s.Logger.Error("cannot update  profile photo",zap.Error(err))
		return nil, status.Error(codes.Aborted, "")
	}
	return &rentalpb.CreateProfilePhotoResponse{UploadUrl: blob.UploadUrl}, nil
	//return nil, status.Errorf(codes.Unimplemented, "method CreateProfilePhoto not implemented")
}

func (s *Service) CompleteProfilePhoto(c context.Context, req *rentalpb.CompleteProfilePhotoRequest) (resp *rentalpb.Identity, err error) {
	aid,err := auth.AccountIdFromContext(c)
	if err != nil {
		return nil,err
	}

	pr,err := s.Mongo.GetProfile(c,aid)
	if err != nil {
		return nil, status.Error(s.logAndConvertProfileErr(err),"")
	}
	if pr.PhotoBlobID == "" {
		return nil, status.Error(codes.NotFound, "")
	}

	br,err := s.BlobClient.GetBlob(c,&blobpb.GetBlobRequest{Id: pr.PhotoBlobID})
	if err != nil {
		s.Logger.Error("cannot get blob",zap.Error(err))
		return nil, status.Error(codes.Aborted, "")
	}

	s.Logger.Info("got profile photo",zap.Int("size",len(br.Data)))
	return &rentalpb.Identity{
		LicNumber:       "黑A88888",
		Name:            "李四",
		Gender:          rentalpb.Gender_FEMALE,
		BirthDateMillis: 631152000000,
	}, nil
	//return nil, status.Errorf(codes.Unimplemented, "method CompleteProfilePhoto not implemented")
}

func (s *Service) ClearProfilePhoto(c context.Context, req *rentalpb.ClearProfilePhotoRequest) (resp *rentalpb.ClearProfilePhotoResponse,err  error) {

	return nil, status.Errorf(codes.Unimplemented, "method ClearProfilePhoto not implemented")
}

func (s *Service) UploadFilePhoto(c context.Context, req *rentalpb.UploadFilePhotoRequest) (resp *rentalpb.UploadFilePhotoResponse,err  error) {
	//file, header, err := http.Request.FormFile("a")
	//if _, _, err := r.FormFile("media"); err != nil {
	//	t.Fatal(err)
	//
	//}
	//fmt.Println("==============================",file,header)
	//if err != nil {
	//	s.Logger.Error("cannot get blob",zap.Error(err))
	//	return nil, status.Errorf(codes.Unimplemented, "method UploadFilePhoto not implemented")
	//}
	httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			s.Logger.Error("Expect 'POST' get '%s'",zap.Error(errors.New(r.Method)))
			//t.Fatalf("Expect 'POST' get '%s'", r.Method)
			return
		}

		//path := r.URL.EscapedPath()
		//if path != apiUploadTemplateMedia {
		//	t.Fatalf("Except to path '%s',get '%s'", apiUploadTemplateMedia, path)
		//}

		//if err := r.ParseForm(); err != nil {
		//	t.Fatal(err)
		//}
		//
		//queries := []string{"access_token", "type"}
		//for _, v := range queries {
		//	content := r.Form.Get(v)
		//	if content == "" {
		//		t.Fatalf("%v can not be empty", v)
		//	}
		//}

		file, header, err := r.FormFile("testing_upload")
		if err != nil {
			s.Logger.Error("get file fail '%s'",zap.Error(err))
			//t.Fatalf("Expect 'POST' get '%s'", r.Method)
			return
		}

		fmt.Println("============================22222222222222===========================",file,header)
		//w.WriteHeader(http.StatusOK)
		//w.Header().Set("Content-Type", "application/json")
		//raw := `{
		//	"errcode": 0,
		//	"errmsg": "ok",
		//	"type": "image",
		//	"media_id": "MEDIA_ID",
		//	"created_at": 1234567890
		//  }`
		//if _, err := w.Write([]byte(raw)); err != nil {
		//	s.Logger.Error("write fail '%s'",zap.Error(err))
		//	//t.Fatalf("Expect 'POST' get '%s'", r.Method)
		//	return
		//}
	}))
	return resp, nil
	//return nil, status.Errorf(codes.Unimplemented, "method ClearProfilePhoto not implemented")
}





















