package main

import (
	"context"

	pb "github.com/huyshop/header/user"
)

func (u *User) SignIn(ctx context.Context, req *pb.User) (*pb.User, error) {
	return &pb.User{}, nil
}
