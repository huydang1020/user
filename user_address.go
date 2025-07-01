package main

import (
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/huyshop/header/common"
	pb "github.com/huyshop/header/user"
	"github.com/huyshop/user/utils"
)

func (u *User) CreateUserAddress(ctx context.Context, rq *pb.UserAddress) (*common.Empty, error) {
	if rq.GetUserId() == "" {
		return nil, errors.New(utils.E_invalid_user_id)
	}
	if rq.GetProvince() == "" {
		return nil, errors.New(utils.E_invalid_province)
	}
	if rq.GetDistrict() == "" {
		return nil, errors.New(utils.E_invalid_district)
	}
	if rq.GetWard() == "" {
		return nil, errors.New(utils.E_invalid_ward)
	}
	if rq.GetAddress() == "" {
		return nil, errors.New(utils.E_invalid_address)
	}
	maxUserAddress, err := strconv.Atoi(config.MaxUserAddress)
	if err != nil {
		return nil, err
	}
	rq.Id = utils.MakeUserAddressId()
	if err := u.Db.TranCreateUserAddress(rq, maxUserAddress); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) UpdateUserAddress(ctx context.Context, rq *pb.UserAddress) (*common.Empty, error) {
	log.Println("rq:", rq)
	if rq.GetId() == "" {
		return nil, errors.New(utils.E_invalid_id)
	}
	if err := u.Db.UpdateUserAddress(rq, &pb.UserAddress{Id: rq.GetId()}); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) DeleteUserAddress(ctx context.Context, rq *pb.UserAddress) (*common.Empty, error) {
	if rq.GetId() == "" {
		return nil, errors.New(utils.E_invalid_id)
	}
	if err := u.Db.DeleteUserAddress(rq.Id); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) ListUserAddress(ctx context.Context, rq *pb.UserAddressRequest) (*pb.UserAddresses, error) {
	log.Println("rq:", rq)
	if rq.GetUserId() == "" {
		return nil, errors.New(utils.E_invalid_user_id)
	}
	addresses, err := u.Db.ListUserAddress(rq)
	if err != nil {
		return nil, err
	}
	return &pb.UserAddresses{UserAddresses: addresses}, nil
}

func (u *User) GetUserAddress(ctx context.Context, rq *pb.UserAddress) (*pb.UserAddress, error) {
	if rq.GetId() == "" {
		return nil, errors.New(utils.E_invalid_id)
	}
	address, err := u.Db.GetUserAddress(rq)
	if err != nil {
		return nil, err
	}

	return address, nil
}
