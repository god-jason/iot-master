package web

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims

	//管理员
	Admin bool `json:"admin,omitempty"`

	//租户
	Tenant string `json:"tenant,omitempty"`
}

var JwtKey = []byte("boat")
var JwtExpire = time.Hour * 24 * 30

func JwtGenerate(id string, admin bool, tenant string) (string, error) {
	var claims Claims
	claims.ID = id
	claims.Admin = admin
	claims.Tenant = tenant
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(JwtExpire))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JwtKey))
}

func JwtVerify(str string) (*Claims, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(str, &claims, func(token *jwt.Token) (any, error) {
		return []byte(JwtKey), nil
		//return config.GetString(MODULE, "jwt_key"), nil
	})
	if err != nil {
		return nil, err
	}
	if token.Valid {
		return &claims, nil
	} else {
		return nil, err
	}
}
