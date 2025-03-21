package utils

import (
	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
)

func MakeUserId() string {
	return "user" + xid.New().String()
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
