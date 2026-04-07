package utils

import (
	"backend/internal/config"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// MyClaims 自定义声明
type MyClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken 生成一对儿 Token
func GenerateToken(userID uint, cfg *config.JwtConfig) (accessToken, refreshToken string, err error) {
	now := time.Now()

	// 1. Access Token
	aClaims := MyClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Second * time.Duration(cfg.AccessExp))),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "ai-friends-backend",
		},
	}
	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, aClaims).SignedString([]byte(cfg.Secret))
	if err != nil {
		return "", "", err
	}

	// 2. Refresh Token (修正 & 为 *，并建议带上 UserID)
	rClaims := MyClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Second * time.Duration(cfg.RefreshExp))),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, rClaims).SignedString([]byte(cfg.Secret))

	return accessToken, refreshToken, err
}

// ParseToken 解析 Access Token
func ParseToken(tokenString string, secret string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
