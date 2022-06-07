package token

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"testing"
	"time"
)

const privateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEoQIBAAKCAQBlar2QFxQILPZWMJPL/YXw8uzfDwLOErj9tZRqhTKeaGQi71dT
z5eaJiyy6siRGxpAOC21w3K1odFh2NYQisJjUzH9mbD9Zrsxn80GNiy7qoO49BQN
vKR+r9BpG7tp8kllvZUbukAsua+Ltu6L1//Du/vnZ/o+QZK46yTBliQiR2ZR1LwG
7/oDBrkNRdiL0PEOiUu1yjjBGJu+QH3z7QV+xFqgxuA5qwRJKr8qFx2jI+XGbQ+T
Cnmc0qFlG5UwXQdhQVmRPCKIv9RAPQkwFJWYgm2gLMFE/BF1HGiXqSSzoiKm34UZ
JOfaCK0DgAlD8lXPTU4luou+fvARZUIt0dNrAgMBAAECggEARZMUuOUcOP+ff0GU
Iz2pxcLh/aSup/DwOB151BnMwB+dktnfbv/bYUUPJS8sqV+OgsAHm8qZx0FcA2Re
drq64KkSgogNg8oWYdTfMoO37IbuFtRbbZEcyEyVDYaY6/XrpICK6pq7q/M3GwJE
ZGuqav1rszUB1/PkVDf831HBOEIllcXqMD8tPydKlF/j9B0/8RzEsO0q/yKIBHv5
lKVnwwtzo2kplKf0HFuyaIHyhPkzwP6SADFC98kPozccPCHueHxYNoJeFRZ+d6p9
m7iKxtF3ZGEN6i7H+dTbLz2hjBZGVUyyQaItYs+e9SFIysR8gTxLIeyPi6EsUcDF
PgwuAQKBgQCt65Y+gnWPwZ3fs5s5EciPSg6QnD8XkMb1JAgqA3niRN8fne1Bpq85
o6lAVar1Q1N6IZ2M/lSFvEB5pIsNOsfRlI4lRRT+koVP2UMtgsvU5oiUN6Ihr8pD
NxFJL62f3QRIAtlJGwfN4CWap7MweiSuJhQW6zv8Mw1eSDkzfIzXowKBgQCVR5Fk
+cQKbKhiNkJDKpV6SJkJAZKfkeeQ4UH+77oJTgLNDljoYVLWGdiBifdkR+DcR8fu
dEJqcSyfwZ6QaddKa7RieHW8hGok2HCJGh4UVOPP/c/B9LCElPlUOQOfYpkDBk61
n/qUNU2nAxB1DYIuGMZa83sueSC3NmvhgxJxmQKBgELqNHlCenbf+Jz8Hom7lN3E
kYwEAaZQdqwUu+tmQPvUocApZAJxXlOf61usWkVZLQH9cv+vTtfRkUD8FN/3MLAr
JvGh/azgtNf+8IDPJRt5wyy7pu6tEvX/dvAgnv217JLEPdOJydvlFLLHOQM6y2gR
kIMs6HxlLAyNXyy3C/1fAoGARVKA4GVkdnrTDxinWM2TsL+54YbLcjKtWIhiv2LP
/7YsctEw1GktsKM7+Zv6OKVwdJsY61Et7oQz9tRRMDsWcUbm87uA4VSUfhvM1r48
LbDmQhZZvnZE6qzLxCLd3cxJxi/NqWZrVRwTvXUc1V66j3GN23qqP4CcgrhRDN5z
WDECgYBSoAlUWzl9UjIyzKh4Lxo+NNq9ksVmYOI1Yip2aJbqPXGGX0AFAXEqQVU+
sJ6VRtPBrq9mfQ0VHOlLt/2ZYG5rPuBWGxwJC9OF+nWew2uWphLxqha2o1bot4Sh
n1CyPlTZDM9IrUSFJ44iFgx08tUdgvxFX3N8jfgQl3k8I5hvVQ==
-----END RSA PRIVATE KEY-----`

func TestGenerateToken(t *testing.T) {
	pem, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		t.Fatalf("can not parse private key:%v",err)
	}
	g := NewJWTTokenGen("coolcar/auth",pem)
	g.nowFunc = func() time.Time {
		return time.Unix(1516239022,0)
	}
	token, err := g.GenerateToken("5f7c3168e2283aa722e351a3", 2*time.Hour)
	if err != nil {
		t.Errorf("can not make token : %v",err)
	}
	fmt.Println(token)
}
