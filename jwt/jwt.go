package jwt

import (
	"errors"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
	pb "github.com/huyshop/header/user"
	"github.com/huyshop/user/utils"
)

type JWTClaim struct {
	UserId      string `json:"user_id"`
	RoleId      string `json:"role_id"`
	PartnerId   string `json:"partner_id"`
	PartnerType string `json:"partner_type"`
	jwt.StandardClaims
}

func GenerateAccessToken(user *pb.User, partner *pb.Partner, expireTime time.Duration, secretKey string) (string, error) {
	expirationTime := time.Now().Add(expireTime * time.Minute)
	claims := &JWTClaim{
		UserId:      user.GetId(),
		RoleId:      user.GetRoleId(),
		PartnerId:   partner.GetId(),
		PartnerType: partner.GetType(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		log.Println("generate access token err:", err)
		return "", err
	}
	return tokenString, nil
}

func GenerateRefreshToken(user *pb.User, partner *pb.Partner, expireTime time.Duration, secretKey string) (string, error) {
	expirationTime := time.Now().Add(expireTime * time.Minute)
	claims := &JWTClaim{
		UserId:      user.GetId(),
		RoleId:      user.GetRoleId(),
		PartnerId:   partner.GetId(),
		PartnerType: partner.GetType(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		log.Println("generate refresh token err:", err)
		return "", err
	}
	return tokenString, nil
}

func ValidateRefreshToken(refreshToken, secretKey string) (*JWTClaim, error) {
	token, err := jwt.ParseWithClaims(refreshToken, &JWTClaim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, errors.New(utils.E_invalid_sigature + err.Error())
		}
		return nil, errors.New(utils.E_invalid_token)
	}

	claims, ok := token.Claims.(*JWTClaim)
	if !ok || !token.Valid {
		return nil, errors.New(utils.E_invalid_token)
	}

	return claims, nil
}
