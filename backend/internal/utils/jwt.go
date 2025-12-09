package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT声明
type JWTClaims struct {
	UserID      uint     `json:"user_id"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT Token
func GenerateToken(userID uint, username, role, secret string, expireHour int) (string, error) {
	return GenerateTokenWithPermissions(userID, username, role, nil, secret, expireHour)
}

// GenerateTokenWithPermissions 生成带权限的JWT Token
func GenerateTokenWithPermissions(userID uint, username, role string, permissions []string, secret string, expireHour int) (string, error) {
	claims := jwt.MapClaims{
		"user_id":     userID,
		"username":    username,
		"role":        role,
		"permissions": permissions,
		"exp":         time.Now().Add(time.Hour * time.Duration(expireHour)).Unix(),
		"iat":         time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析JWT Token
func ParseToken(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
