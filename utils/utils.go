package utils

import "github.com/rs/xid"

func MakeUserId() string {
	return "user" + xid.New().String()
}
