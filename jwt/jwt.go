package jwt

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	Secret []byte
}

type JWT struct {
	secret []byte
	alg    jwt.SigningMethod
}

func New(c Config) *JWT {
	return &JWT{
		secret: c.Secret,
		alg:    jwt.SigningMethodHS256,
	}
}

func (j *JWT) NewWithClaims(claims map[string]any) (string, error) {
	mc := jwt.MapClaims{}
	for k, v := range claims {
		mc[k] = v
	}
	token := jwt.NewWithClaims(j.alg, mc)
	return token.SignedString(j.secret)
}

func (j *JWT) Parse(tokenString string) (map[string]any, error) {
	mc := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &mc, func(t *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token error: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("token invalid")
	}

	return mc, nil
}
