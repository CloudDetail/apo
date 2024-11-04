package util

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/dgrijalva/jwt-go"
	"strings"
	"time"
)

// 7days 10.31-11.7
// accToken
// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwidXNlcklEIjowLCJpc1JlZnJlc2giOmZhbHNlLCJleHAiOjE3MzA5NzAwOTYsImlhdCI6MTczMDM2NTI5NiwiaXNzIjoiYXBvIn0.MgdmgjSqs-YlUJGCc8yylEKYIb7_CCdSQzPFw0BYjXs
// refToken
// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwidXNlcklEIjowLCJpc1JlZnJlc2giOnRydWUsImV4cCI6MTczMDk3MDA5NiwiaWF0IjoxNzMwMzY1Mjk2LCJpc3MiOiJhcG8ifQ.QntyKxam4mhSiX94IWr_3U4fQp41zkZA0RBC7LOtj6w
var secret = []byte("APO@2024")
var accessExpireTime = 30 * time.Minute
var refreshExpireTime = 48 * time.Hour
var testExpireTime = 7 * 24 * time.Hour

var (
	TokenInvalid = errors.New("invalid token")
	TokenExpired = errors.New("token is expired")
)

type Claims struct {
	Username  string `json:"username"`
	UserID    int64  `json:"userID"`
	IsRefresh bool   `json:"isRefresh"`
	jwt.StandardClaims
}

func GenerateTokens(username string) (string, string, error) {
	issuedAt := time.Now()
	accessClaims := Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  issuedAt.Unix(),
			ExpiresAt: issuedAt.Add(accessExpireTime).Unix(),
			Issuer:    "apo",
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(secret)
	if err != nil {
		return "", "", err
	}
	refreshClaims := Claims{
		Username:  username,
		IsRefresh: true,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  issuedAt.Unix(),
			ExpiresAt: issuedAt.Add(refreshExpireTime).Unix(),
			Issuer:    "apo",
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(secret)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func RefreshToken(rawToken string) (string, error) {
	token := ParseRawToken(rawToken)
	if len(token) == 0 {
		return "", TokenInvalid
	}
	claims, err := ParseRefreshToken(token)
	if err != nil {
		return "", err
	}
	issuedAt := time.Now()
	accessClaims := Claims{
		Username: claims.Username,
		UserID:   claims.UserID,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  issuedAt.Unix(),
			ExpiresAt: issuedAt.Add(accessExpireTime).Unix(),
			Issuer:    "apo",
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(secret)
	if err != nil {
		return "", err
	}
	return accessToken, nil
}

func ParseRefreshToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !claims.IsRefresh {
		return nil, model.NewErrWithMessage(TokenInvalid, code.InValidToken)
	}
	if token.Valid {
		return claims, nil
	}
	return nil, TokenExpired
}

func ParseAccessToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims.IsRefresh {
		return nil, model.NewErrWithMessage(TokenInvalid, code.InValidToken)
	}
	if token.Valid {
		return claims, nil
	}
	return nil, TokenExpired
}

func ParseRawToken(rawToken string) string {
	if len(rawToken) == 0 {
		return ""
	}
	parts := strings.Split(rawToken, " ")
	if !(len(parts) == 2 && parts[0] != "Bearer ") {
		return ""
	}
	return parts[1]
}

func IsExpire(token string) bool {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return true
	}

	return !parsedToken.Valid
}
