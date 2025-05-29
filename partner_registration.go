package main

import (
	"context"
	"errors"
	"time"

	"github.com/goccy/go-json"
	"github.com/huyshop/header/common"
	pb "github.com/huyshop/header/user"
	"github.com/huyshop/user/utils"
)

func (u *User) CreatePartnerRegistration(ctx context.Context, req *pb.PartnerRegistration) (*common.Empty, error) {
	if req.GetFullName() == "" {
		return nil, errors.New(utils.E_invalid_fullname)
	}
	if req.GetEmail() == "" {
		return nil, errors.New(utils.E_invalid_email)
	}
	if req.GetPhoneNumber() == "" {
		return nil, errors.New(utils.E_invalid_phone_number)
	}
	if req.GetAddress() == "" {
		return nil, errors.New(utils.E_invalid_address)
	}
	if req.GetProvince() == "" {
		return nil, errors.New(utils.E_not_found_province)
	}
	if req.GetDistrict() == "" {
		return nil, errors.New(utils.E_not_found_district)
	}
	if req.GetWard() == "" {
		return nil, errors.New(utils.E_not_found_ward)
	}
	exit := u.Db.IsPartnerRegistrationExisted(&pb.PartnerRegistration{
		Email:       req.GetEmail(),
		PhoneNumber: req.GetPhoneNumber(),
		FullName:    req.GetFullName(),
	})
	if exit {
		return nil, errors.New(utils.E_user_existed)
	}
	req.CreatedAt = time.Now().Unix()
	req.UserId = utils.MakeUserId()
	req.PartnerId = utils.MakePartnerId()
	if err := u.Db.CreatePartnerRegistration(req); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) GetPartnerRegistration(ctx context.Context, req *pb.PartnerRegistrationRequest) (*pb.PartnerRegistration, error) {
	if req.UserId == "" {
		return nil, errors.New(utils.E_invalid_user_id)
	}
	if req.PartnerId == "" {
		return nil, errors.New(utils.E_invalid_partner_id)
	}
	registration, err := u.Db.GetPartnerRegistration(req)
	if err != nil {
		return nil, err
	}
	if registration == nil {
		return nil, errors.New(utils.E_not_found)
	}
	return registration, nil
}
func (u *User) ListPartnerRegistration(ctx context.Context, req *pb.PartnerRegistrationRequest) (*pb.PartnerRegistrations, error) {
	list, err := u.Db.ListPartnerRegistration(req)
	if err != nil {
		return nil, err
	}
	return &pb.PartnerRegistrations{PartnerRegistrations: list}, nil
}

func (u *User) UpdatePartnerRegistration(ctx context.Context, req *pb.PartnerRegistration) (*common.Empty, error) {
	if req.UserId == "" {
		return nil, errors.New(utils.E_invalid_user_id)
	}
	if req.PartnerId == "" {
		return nil, errors.New(utils.E_invalid_partner_id)
	}
	req.UpdatedAt = time.Now().Unix()
	req.State = pb.PartnerRegistration_pending.String()
	registration, err := u.Db.GetPartnerRegistration(&pb.PartnerRegistrationRequest{
		UserId:    req.GetUserId(),
		PartnerId: req.GetPartnerId(),
	})
	if registration == nil {
		return nil, errors.New(utils.E_not_found)
	}
	if registration.GetState() == pb.PartnerRegistration_approved.String() {
		before, err := json.Marshal(registration)
		if err != nil {
			return nil, err
		}
		req.Before = string(before)
	}
	if err != nil {
		return nil, err
	}
	if err := u.Db.UpdatePartnerRegistration(req, &pb.PartnerRegistration{UserId: req.GetUserId(), PartnerId: req.GetPartnerId()}); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) DeletePartnerRegistration(ctx context.Context, req *pb.PartnerRegistration) (*common.Empty, error) {
	if req.UserId == "" {
		return nil, errors.New(utils.E_invalid_user_id)
	}
	if req.PartnerId == "" {
		return nil, errors.New(utils.E_invalid_partner_id)
	}
	if err := u.Db.DeletePartnerRegistration(req); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}
