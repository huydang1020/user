package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/huyshop/header/common"
	pb "github.com/huyshop/header/user"
	"github.com/huyshop/user/utils"
)

const DEFAULT_LIMIT = 20

func (u *User) SignIn(ctx context.Context, req *pb.User) (*pb.User, error) {
	return &pb.User{}, nil
}

func (u *User) CreateUser(ctx context.Context, req *pb.User) (*common.Empty, error) {
	if u.Db.IsUserExisted(req) {
		return nil, errors.New(utils.E_user_existed)
	}
	if req.RoleId == 0 {
		return nil, errors.New(utils.E_not_found_role_id)
	}
	hashPw, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Println("hash pw err:", err)
		return nil, err
	}
	req.Password = hashPw
	req.Id = utils.MakeUserId()
	req.State = pb.User_active.String()
	req.CreatedAt = time.Now().Unix()
	if err := u.Db.CreateUser(req); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) GetUser(ctx context.Context, req *pb.UserRequest) (*pb.User, error) {
	if req.GetId() == "" && req.GetUsername() == "" && req.GetPhoneNumber() == "" {
		return nil, errors.New(utils.E_internal_error)
	}
	user, err := u.Db.GetUser(req)
	if err != nil {
		return nil, err
	}
	user.Password = ""
	return user, nil
}

func (u *User) ListUsers(ctx context.Context, req *pb.UserRequest) (*pb.Users, error) {
	if req.GetLimit() == 0 || req.GetLimit() > DEFAULT_LIMIT {
		req.Limit = DEFAULT_LIMIT
	}
	users, err := u.Db.ListUsers(req)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return &pb.Users{}, nil
	}
	for _, u := range users {
		u.Password = ""
	}
	count, _ := u.Db.CountUsers(req)
	return &pb.Users{
		Users: users,
		Total: count,
	}, nil
}

func (u *User) UpdateUser(ctx context.Context, req *pb.User) (*common.Empty, error) {
	if req.GetId() == "" {
		return nil, errors.New(utils.E_not_found_user_id)
	}
	_, err := u.Db.GetUser(&pb.UserRequest{Id: req.GetId()})
	if err != nil {
		return nil, errors.New(utils.E_can_not_update)
	}
	req.UpdatedAt = time.Now().Unix()
	if err := u.Db.UpdateUser(req, &pb.User{Id: req.GetId()}); err != nil {
		return nil, errors.New(utils.E_can_not_update)
	}
	return &common.Empty{}, nil
}

func (u *User) DeleteUser(ctx context.Context, req *pb.User) (*common.Empty, error) {
	if req.GetId() == "" {
		return nil, errors.New(utils.E_not_found_user_id)
	}
	_, err := u.Db.GetUser(&pb.UserRequest{Id: req.GetId()})
	if err != nil {
		return nil, errors.New(utils.E_can_not_delete)
	}
	if err := u.Db.DeleteUser(req.GetId()); err != nil {
		return nil, errors.New(utils.E_can_not_delete)
	}
	return &common.Empty{}, nil
}
