package entity

import (
	"time"
)

type JWT struct {
	AccessToken           string    `bson:"access_token" json:"access_token"`
	RefreshToken          string    `bson:"refresh_token" json:"refresh_token"`
	AccessTokenExpiredAt  time.Time `bson:"access_token_expired_at" json:"access_token_expired_at"`
	RefreshTokenExpiredAt time.Time `bson:"refresh_token_expired_at" json:"refresh_token_expired_at"`
}
