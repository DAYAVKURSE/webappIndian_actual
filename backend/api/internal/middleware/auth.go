package middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

func GetTokenFromAuthorizationHeader(c *gin.Context) (string, error) {
	authorizationHeader := c.Request.Header.Get("Authorization")
	if len(authorizationHeader) == 0 {
		return "", errors.New("authorization header is empty")
	}

	token := strings.Split(string(authorizationHeader[:]), " ")
	if len(token) != 2 || token[0] != "Bearer" {
		return "", errors.New("authorization not Bearer format")
	}

	return token[1], nil
}

func TokenCheck(tokenStr string, hmacSecretKey string) (uint64, string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("nexpected signing method: %v", token.Header["alg"])
		}

		return []byte(hmacSecretKey), nil
	})
	if err != nil {
		if !errors.Is(err, jwt.ErrTokenExpired) {
			return 0, "", err
		}
	}

	if claims, ok := token.Claims.(jwt.MapClaims); !ok {
		return 0, "", errors.New("error get jwt.MapClaims")
	} else {
		var userId uint64
		var tokenType string

		if v, ok := claims["userId"]; !ok {
			return 0, "", errors.New("no userId field")
		} else {
			userId = uint64(v.(float64))
		}

		if v, ok := claims["type"]; !ok {
			return 0, "", errors.New("no type field")
		} else {
			tokenType = v.(string)
		}

		// передаём err чтобы ещё раз можно было проверить устарел токен или нет, другие ошибки сюда не попадут
		return userId, tokenType, err
	}
}

func ComparePasswords(passHash string, pass string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passHash), []byte(pass))

	return err == nil
}

func TokenNew(hmacSecretKey string, userId int64, expAccess int64, tokenType string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"type":   tokenType,
		"userId": userId,
		"exp":    expAccess,
	})
	tokenStr, err := token.SignedString([]byte(hmacSecretKey))
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func HashAndSalt(pass string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
