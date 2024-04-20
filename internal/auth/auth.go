package auth

import (
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/my_errors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"os"
	"time"
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
		return int(claims["id"].(float64)), nil
	} else {
		return 0, my_errors.AuthenticationError
	}
}

func GenerateTokenFromId(userId int) (string, error) {
	hmac := os.Getenv("HMAC")
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  userId,
		"nbf": now.Unix(),
		"exp": now.Add(5 * 24 * time.Hour).Unix(),
		"iat": now.Unix(),
	})

	tokenString, err := token.SignedString([]byte(hmac))
	if err != nil {
		return "", err
	}
	return tokenString, nil

}

func ComparePasswordWithHashed(password, hashedPassword string) error {
	incoming := []byte(password)
	existing := []byte(hashedPassword)
	return bcrypt.CompareHashAndPassword(existing, incoming)
}

func GenerateHashedPassword(password string) (string, error) {
	saltedBytes := []byte(password)
	hashedBytes, err := bcrypt.GenerateFromPassword(saltedBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	hash := string(hashedBytes)
	return hash, nil
}
