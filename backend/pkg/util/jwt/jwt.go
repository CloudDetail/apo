// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package jwt

import (
	"errors"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/golang-jwt/jwt/v4"
)

var secret = []byte("APO@2024")
var accessExpireTime = time.Duration(config.Get().Server.AccessTokenExpireMinutes) * time.Minute
var refreshExpireTime = time.Duration(config.Get().Server.RefreshTokenExpireHours) * time.Hour

var (
	TokenInvalid = errors.New("invalid token")
	TokenExpired = errors.New("token is expired")
)

type Claims struct {
	Username  string `json:"username"`
	UserID    int64  `json:"userID"`
	IsRefresh bool   `json:"isRefresh"`
	jwt.RegisteredClaims
}

func GenerateTokens(username string, userID int64) (string, string, error) {
	issuedAt := time.Now()
	accessClaims := Claims{
		Username: username,
		UserID:   userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(accessExpireTime)),
			Issuer:    "apo",
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(secret)
	if err != nil {
		return "", "", err
	}
	refreshClaims := Claims{
		Username:  username,
		UserID:    userID,
		IsRefresh: true,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(refreshExpireTime)),
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
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(accessExpireTime)),
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
		return nil, core.Error(code.InValidToken, TokenInvalid.Error())
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
		return nil, core.Error(code.InValidToken, TokenInvalid.Error())
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
