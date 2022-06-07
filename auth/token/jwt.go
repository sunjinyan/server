package token

import (
	"crypto/rsa"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type JWTTokenGen struct {
	Issue string
	nowFunc func() time.Time
	privateKey *rsa.PrivateKey
}

func NewJWTTokenGen(issure string,privateKey *rsa.PrivateKey) *JWTTokenGen {
	return &JWTTokenGen{
		Issue:   issure,
		nowFunc: time.Now,
		privateKey:privateKey,
	}
}

func (t *JWTTokenGen)GenerateToken(accountId string,expire time.Duration)(token string,err error)  {
	now := t.nowFunc().Unix()
	claims := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.StandardClaims{
		//Audience:  "",
		ExpiresAt: now + int64(expire),
		//Id:        "",
		IssuedAt: t.nowFunc().Unix(),
		Issuer:   t.Issue,
		//NotBefore: 0,
		Subject: accountId,
	})

	return claims.SignedString(t.privateKey)
}