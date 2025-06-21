package models

import (
	"BlessedApi/cmd/db"
	"BlessedApi/pkg/logger"
)

type RefreshToken struct {
	Id           uint64 `json:"id"`
	RefreshToken string `json:"refresh_token"`
	UserId       uint64 `json:"user_id"`
	Expiration   uint64 `json:"expiration"`
	TmCreate     uint64 `json:"tm_create"`
}

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func GetRefreshTokenByRefreshToken(tokenName string) (*RefreshToken, error) {
	var token RefreshToken

	err := db.DB.
		Where("refresh_token = ?", tokenName).
		First(&token).Error
	if err != nil {
		return nil, logger.WrapError(err, "")
	}

	return &token, nil
}

func DeleteRefreshTokenById(tokenId uint64) error {

	err := db.DB.Delete(&RefreshToken{}, "id = ?", tokenId).Error
	if err != nil {
		return logger.WrapError(err, "")
	}

	return nil
}

func CreateRefreshToken(token *RefreshToken) error {
	db.DB.Create(&token)
	return nil
}
