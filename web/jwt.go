package web

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/zgwit/iot-master/config"
	"time"
)

type Claims struct {
	Id string `json:"id"`
	jwt.RegisteredClaims
}

var JwtKey = "iot-master"
var JwtExpire = time.Hour * 24 * 30

func JwtGenerate(id string) (string, error) {
	var claims Claims
	claims.Id = id
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(JwtExpire))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JwtKey)
}

func JwtVerify(str string) (*Claims, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(str, &claims, func(token *jwt.Token) (any, error) {
		return config.GetString(MODULE, "jwt_key"), nil
	})
	if token.Valid {
		return &claims, nil
	} else {
		return nil, err
	}
}
