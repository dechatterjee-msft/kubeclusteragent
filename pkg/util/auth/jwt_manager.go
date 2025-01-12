package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

type JWTManager struct {
	secretKey string
}

type UserClaims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
	Role     string `json:"role"`
}

func CreateJwtManager(secretKey string) *JWTManager {
	return &JWTManager{secretKey}
}

func (manager *JWTManager) VerifyToken(accessToken string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(accessToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				logger.Info("incorrect token signing method, returning error")
				return nil, fmt.Errorf("incorrect token signing method")
			}
			return []byte(manager.secretKey), nil
		},
	)
	if err != nil {
		logger.Error(err, "invalid token")
		return nil, fmt.Errorf("invalid token: %v", err)
	}
	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		logger.Info("Invalid token claims")
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil

}
