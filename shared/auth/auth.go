package auth

import (
	"context"
	"coolcar/shared/id"
	"coolcar/shared/token"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"os"
	"strings"
)

type tokenVerifier interface {
	Verify(token string)(string,error)
}

//拦截器
type interceptor struct {
	//publicKey *rsa.PublicKey
	verifier tokenVerifier
}


const (
	authorizationHeader = "authorization"
	bearerPrefix = "Bearer "
	ImpersonteAccountHeader = "impersonate-account-id"
)

func Interceptor(publicKeyFile string) (grpc.UnaryServerInterceptor,error) {

	f,err := os.Open(publicKeyFile)
	if err != nil {
		return nil, fmt.Errorf("can not open public key file:%v",publicKeyFile)
	}
	b,err := ioutil.ReadAll(f)

	if err != nil {
		return nil,fmt.Errorf("can not read public key:v%",err)
	}

	pubKey,err  := jwt.ParseRSAPublicKeyFromPEM(b)
	if err != nil {
		return nil,fmt.Errorf("can not parse publich key %v",err)
	}
	i := &interceptor{
		//publicKey: pubKey,
		verifier: &token.JWTVerifyToken{
			PublicKey:pubKey,
		},
	}

	return i.HandleReq,nil
}

func (i *interceptor) HandleReq(ctx context.Context, req interface {}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	aid := impersonationFromContext(ctx)

	if aid != "" {
		fmt.Printf("impersonationg %q\n",aid)
		return  handler(ContextWithAccountId(ctx,id.AccountId(aid)),req)
	}

	tkn,err := tokenFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated,"can not authenticated")
	}
	aid, err = i.verifier.Verify(tkn)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "token not valid: " + err.Error())
	}

	return  handler(ContextWithAccountId(ctx,id.AccountId(aid)),req)
}

func impersonationFromContext(c context.Context) string {
	m, b := metadata.FromIncomingContext(c)
	if !b {
		return ""
	}
	imp  := m[ImpersonteAccountHeader]
	if len(imp) == 0 {
		return ""
	}

	return imp[0]
}

func tokenFromContext(ctx context.Context)(string,error) {
	m, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "can not authenticated")
	}
	tkn := ""
	for _, v := range m[authorizationHeader] {
		if strings.HasPrefix(v, bearerPrefix) {
			tkn = v[len(bearerPrefix):]
		}
	}
	if tkn == "" {
		return "", status.Error(codes.Unauthenticated, "can not authenticated")
	}
	return tkn, nil
}

type accountIdKey struct {

}

//type AccountId string
//
//func (a AccountId)String()string  {
//	return string(a)
//}

//identifier Type 设计模式
func ContextWithAccountId(ctx context.Context,aid id.AccountId) context.Context  {
	return context.WithValue(ctx,accountIdKey{},aid)
}

func AccountIdFromContext(c context.Context) (id.AccountId, error) {
	v := c.Value(accountIdKey{})
	aid,ok := v.(id.AccountId)
	if !ok {
		return "", status.Error(codes.Unauthenticated,"can not authenticated")
	}
	return aid, nil
}