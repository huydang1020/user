package main

import (
	"context"
	"errors"
	"log"
	"time"

	pb "github.com/huyshop/header/user"
	"github.com/huyshop/user/utils"
)

func (u *User) GetUserPoint(ctx context.Context, req *pb.UserPointRequest) (*pb.UserPoint, error) {
	if req.Id == "" {
		return nil, errors.New(utils.E_not_found_id)
	}
	exits, err := u.Db.IsExistUserPoint(req.Id)
	if err != nil {
		log.Println("IsExistUserPoint error:", err)
		return nil, err
	}
	if !exits {
		err := u.Db.CreateUserPoint(&pb.UserPoint{
			UserId:    req.Id,
			CreatedAt: time.Now().Unix(),
		})
		if err != nil {
			log.Println("CreateUserPoint error:", err)
			return nil, err
		}
	}
	userpoint, err := u.Db.GetUserPoint(&pb.UserPoint{UserId: req.Id})
	if err != nil {
		log.Println("GetUserPoint error:", err)
		return nil, err
	}
	return userpoint, nil
}

func (u *User) ListUserPoint(ctx context.Context, req *pb.UserPointRequest) (*pb.UserPoints, error) {
	listUser, err := u.Db.ListUserPoint(req)
	if err != nil {
		log.Println("ListUserPoint error:", err)
		return nil, err
	}
	return &pb.UserPoints{UserPoints: listUser}, nil
}
