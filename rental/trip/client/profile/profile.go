package profile

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/id"
	"encoding/base64"
	"fmt"
	"github.com/golang/protobuf/proto"
)


//不确定是直接调用函数还是调用GRPC，所以抽象出来，建立一个接口让其他实现方去实现
type Fetcher interface {
	GetProfile(c context.Context,req *rentalpb.GetProfileRequest)(resp *rentalpb.Profile ,err error)
}
//我定义一个接口规范，如过需要使用，你可以定义类型时将我定义在类型中并去使用我来进行操作，操作的具体流程步骤，由谁继承了我以及我包括的方法去实现，你你这里只管调用就可以。

type Manager struct {
	Fetcher Fetcher
}

func (m *Manager) Verify(ctx context.Context,aid id.AccountId) (id.IdentityID,error)  {

	nilID := id.IdentityID("")
	profile, err := m.Fetcher.GetProfile(ctx, &rentalpb.GetProfileRequest{})
	if err != nil {
		return nilID, fmt.Errorf("cannot get profile: %v",err)
	}
	if profile.IdentityStatus != rentalpb.IdentityStatus_VERIFIED {
		return nilID, fmt.Errorf("invalid identity status")
	}
	b,err := proto.Marshal(profile.Identity)
	if err != nil {
		return nilID, fmt.Errorf("cannot marshal identity: %v",err)
	}

	//toString := base64.StdEncoding.EncodeToString(b)

	return id.IdentityID(base64.StdEncoding.EncodeToString(b)), nil
}