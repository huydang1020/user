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
	req.UpdatedAt = time.Now().Unix()
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
	log.Println("ListPartner", rq)
	list, err := u.Db.ListPartner(rq)
	if err != nil {
		return nil, err
	}
	count, err := u.Db.CountPartner(rq)
	if err != nil {
		return nil, err
	}
	return &pb.Partners{Partners: list, Total: int32(count)}, nil
}

func (u *User) GetPartner(ctx context.Context, rq *pb.PartnerRequest) (*pb.Partner, error) {
	log.Println("GetPartner", rq)
	partner, err := u.Db.GetPartner(rq)
	if err != nil {
		return nil, err
	}
	if partner != nil {
		plan, err := u.Db.GetPlan(&pb.PlansRequest{Id: partner.GetPlanId()})
		if err != nil {
			return nil, err
		}
		if plan != nil {
			partner.Plan = plan
		}
	}
	return partner, nil
}

func (u *User) CountPartner(ctx context.Context, rq *pb.PartnerRequest) (*common.Count, error) {
	log.Println("CountPartner", rq)
	count, err := u.Db.CountPartner(rq)
	if err != nil {
		return nil, err
	}
	return &common.Count{Count: count}, nil
}
