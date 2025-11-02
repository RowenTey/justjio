package utils

import (
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

func GetUserInfoFromToken(t *jwt.Token, field string) string {
	claims := t.Claims.(jwt.MapClaims)
	if field == "user_id" {
		return fmt.Sprint(claims["user_id"].(float64))
	}
	return claims[field].(string)
}
