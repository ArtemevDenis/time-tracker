package auth

import (
	"errors"
	"fmt"
	"github.com/ArtemevDenis/time-tracker/internal/app/entity"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"time"
)

const (
	jwtSecretKey        = "secret"
	jwtRefreshSecretKey = "some-refresh-secret-key"
)

func GetJWTSecret() string {
	return jwtSecretKey
}

type JwtCustomClaims struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	jwt.RegisteredClaims
}

func generateAccessToken(user *entity.User) (string, time.Time, error) {
	expirationTime := time.Now().Add(1 * time.Hour)

	return generateToken(user, expirationTime, []byte(GetJWTSecret()))
}

func generateToken(user *entity.User, expirationTime time.Time, secret []byte) (string, time.Time, error) {
	claims := &JwtCustomClaims{
		user.Name,
		user.ID.Hex(),
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", time.Now(), err
	}

	return tokenString, expirationTime, nil
}

func GetRefreshJWTSecret() string {
	return jwtRefreshSecretKey
}

func GenerateTokens(user *entity.User) (jwt *entity.JWT, error error) {
	accessToken, accessTokenExp, err := generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshTokenExp, err := generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	jwt = &entity.JWT{
		AccessTokenExpiredAt:  accessTokenExp,
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		RefreshTokenExpiredAt: refreshTokenExp,
	}

	return jwt, nil
}

func generateRefreshToken(user *entity.User) (string, time.Time, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	return generateToken(user, expirationTime, []byte(GetRefreshJWTSecret()))
}

func GetClaims(token *jwt.Token) (cl *JwtCustomClaims, err error) {
	fmt.Println("token", token)
	claims, ok := token.Claims.(*JwtCustomClaims)
	if !ok {
		return nil, errors.New("failed to cast claims as jwt.MapClaims")
	}

	return claims, nil
}

func GetTokenFromCtx(ctx echo.Context, filed string) (token *jwt.Token, err error) {
	token, ok := ctx.Get(filed).(*jwt.Token) // by default token is stored under `user` key
	fmt.Println()
	if !ok {
		return nil, errors.New("JWT token missing or invalid")
	}
	return token, nil
}

func GetConfig() (config *echojwt.Config) {
	config = &echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			fmt.Println(c)
			return new(JwtCustomClaims)
		},
		SigningKey: []byte(GetJWTSecret()),
	}

	return config
}
