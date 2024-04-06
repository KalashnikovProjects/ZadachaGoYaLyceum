package auth

import (
	"Zadacha/pkg/my_errors"
	"github.com/golang-jwt/jwt/v5"
	"os"
)

func LoadUserIdFromToken(token string) (int, error) {
	tokenFromString, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return 0, my_errors.AuthenticationError
		}

		return []byte(os.Getenv("HMAC")), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := tokenFromString.Claims.(jwt.MapClaims); ok {
		return claims["id"].(int), nil
	} else {
		return 0, my_errors.AuthenticationError
	}
}
