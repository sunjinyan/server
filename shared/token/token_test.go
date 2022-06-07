package token

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"testing"
	"time"
)

const publicKey = `-----BEGIN PUBLIC KEY-----
MIIBITANBgkqhkiG9w0BAQEFAAOCAQ4AMIIBCQKCAQBlar2QFxQILPZWMJPL/YXw
8uzfDwLOErj9tZRqhTKeaGQi71dTz5eaJiyy6siRGxpAOC21w3K1odFh2NYQisJj
UzH9mbD9Zrsxn80GNiy7qoO49BQNvKR+r9BpG7tp8kllvZUbukAsua+Ltu6L1//D
u/vnZ/o+QZK46yTBliQiR2ZR1LwG7/oDBrkNRdiL0PEOiUu1yjjBGJu+QH3z7QV+
xFqgxuA5qwRJKr8qFx2jI+XGbQ+TCnmc0qFlG5UwXQdhQVmRPCKIv9RAPQkwFJWY
gm2gLMFE/BF1HGiXqSSzoiKm34UZJOfaCK0DgAlD8lXPTU4luou+fvARZUIt0dNr
AgMBAAE=
-----END PUBLIC KEY-----`

func TestJWTVerifyToken_Verify(t *testing.T) {
	pem, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
	if err != nil {
		t.Errorf("can not parse public key %v",err)
	}
	v := &JWTVerifyToken{PublicKey: pem}
	tkn := "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjcyMDE2NTIyNjM2ODEsImlhdCI6MTY1MjI2MzY4MSwiaXNzIjoiY29vbGNhci9hdXRoIiwic3ViIjoiNjI3YjhiMDE5MjA0OGVjZmM3ZTRmZjJkIn0.H-P3jVZ93dxtz03LCnm2DFuqXkkqUXNEQ-BXh590UOeTj1Ko7STabGb3ZGkwWM9SAB9UHQ2Y0MQRw81MTY0aJtWReGqCtwqYFhAaO_uyApmeXXFDwt3h1lrUDo9Tso1w62PtryIcyOSCr1C2DDLzmyI9R8nxd4nfiolnwPjQFpzu1PirM8qdWDLDdFZp1Lut6ROwOwjzymKtO4HGNHRqwulYn6cj9wsWcqqfzL5siCavklUb_rr83hhP0KtmrlFbPKTHZ0ol4a02oxqfnBvlRXREZ3GPtShjMgddHl7orRGJK-l09T36fUIQtoEue4eyP95e8Rc6NGOkZp7Fi5GqzA"
	jwt.TimeFunc = func() time.Time {
		//return time.Unix(1517239122,0)
		//return time.Unix(time.Now().Unix() - 100000000,0)
		return time.Now()
	}
	verify, err := v.Verify(tkn)
	if err != nil {
		t.Errorf("verify token failed : %v",err)
	}
	fmt.Println(verify)
}