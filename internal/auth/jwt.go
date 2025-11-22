package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(username string, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})

	return token.SignedString([]byte(secretKey))
}
