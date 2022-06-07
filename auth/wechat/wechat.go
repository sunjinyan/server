package wechat

import (
	"fmt"
	"github.com/medivhzhan/weapp/v2"
)

//server implements a wechat auth service
type Service struct {
	AppID string
	AppSecret string
}
//resolve auth code to wechat open id
func (s *Service)Resolver(code string)(string,error)  {
	//weapp.Login("wx851020fe449a84e2",secret,code)
	response, err := weapp.Login(s.AppID, s.AppSecret, code)
	if err != nil {
		return "", fmt.Errorf("weapp login : %v",err)
	}
	if err := response.GetResponseError();err != nil {
		return "", fmt.Errorf("get response error : %v",err)
	}
	return response.OpenID,nil
}