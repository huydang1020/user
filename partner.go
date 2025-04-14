package main

import (
	"context"
	"errors"
	"time"

	"github.com/huyshop/header/common"
	pb "github.com/huyshop/header/user"
	"github.com/huyshop/user/utils"
)

func (u *User) CreatePartner(ctx context.Context, partner *pb.Partner) (*common.Empty, error) {
	if partner.GetName() == "" {
		return nil, errors.New(utils.E_not_found_name)
	}
	partner.CreatedAt = time.Now().Unix()
	partner.Id = utils.MakePartnerId()
	if partner.GetState() == "" {
		partner.State = pb.Partner_active.String()
	}
	if err := u.Db.CreatePartner(partner); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) UpdatePartner(ctx context.Context, req *pb.Partner) (*common.Empty, error) {
	if req.GetId() == "" {
		return nil, errors.New(utils.E_not_found_id)
	}
	if err := u.Db.UpdatePartner(req, &pb.Partner{Id: req.GetId()}); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) DeletePartner(ctx context.Context, req *pb.Partner) (*common.Empty, error) {
	if req.GetId() == "" {
		return nil, errors.New(utils.E_not_found_id)
	}
	if err := u.Db.DeletePartner(req.GetId()); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) ListPartner(ctx context.Context, rq *pb.PartnerRequest) (*pb.Partners, error) {
	list, err := u.Db.ListPartner(rq)
	if err != nil {
		return nil, err
	}
	for _, p := range list {
		if p.GetUserId() != "" {
			user, err := u.Db.GetUser(&pb.UserRequest{Id: p.GetUserId()})
			if err != nil {
				return nil, err
			}
			p.User = user
		}
	}
	count, err := u.Db.CountPartner(rq)
	if err != nil {
		return nil, err
	}
	return &pb.Partners{Partners: list, Total: int32(count)}, nil
}

func (u *User) GetPartner(ctx context.Context, rq *pb.PartnerRequest) (*pb.Partner, error) {
	partner, err := u.Db.GetPartner(rq)
	if err != nil {
		return nil, err
	}
	return partner, nil
}
