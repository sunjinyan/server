package token

import (
	"crypto/rsa"
	"fmt"
	"github.com/dgrijalva/jwt-go"
)

type JWTVerifyToken struct {
	PublicKey *rsa.PublicKey
}

//Verify  jwt
func (v *JWTVerifyToken)Verify(token string)(string,error)  {
	claims, err := jwt.ParseWithClaims(token, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return v.PublicKey, nil
	})
	if err != nil {
		return "", fmt.Errorf("can not parse toekn : %v",err)
	}
	if !claims.Valid {
		return "", fmt.Errorf("token not valid %v",err)
	}
	clm,ok := claims.Claims.(*jwt.StandardClaims)
	if !ok {
		return "", fmt.Errorf("token claim is not StandardClaim %v",err)
	}
	if err := clm.Valid(); err != nil{
		return "", fmt.Errorf("token claim is not StandardClaim %v",err)
	}
	return clm.Subject,nil
}