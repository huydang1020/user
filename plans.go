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

func (u *User) CreatePlan(ctx context.Context, req *pb.Plan) (*common.Empty, error) {
	if req.GetName() == "" {
		return nil, errors.New(utils.E_invalid_name)
	}
	if req.GetPrices() == nil {
		return nil, errors.New(utils.E_invalid_prices)
	}
	if req.GetFeatures() == nil {
		return nil, errors.New(utils.E_invalid_features)
	}
	for _, tp := range req.GetPrices() {
		if tp.GetPrice() <= 0 {
			return nil, errors.New(utils.E_invalid_prices)
		}
		if tp.GetType() == "" || (tp.GetType() != "tháng" && tp.GetType() != "năm") {
			return nil, errors.New(utils.E_invalid_plan_type)
		}
	}
	req.Id = utils.MakePlanId()
	req.CreatedAt = time.Now().Unix()
	req.State = pb.Plan_active.String()
	if err := u.Db.CreatePlan(req); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) GetPlan(ctx context.Context, req *pb.Plan) (*pb.Plan, error) {
	if req.Id == "" {
		return nil, errors.New(utils.E_invalid_id)
	}
	plan, err := u.Db.GetPlan(&pb.PlansRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, errors.New(utils.E_not_found_plan)
	}
	return plan, nil
}

func (u *User) ListPlans(ctx context.Context, req *pb.PlansRequest) (*pb.Plans, error) {
	log.Println("req:", req)
	list, err := u.Db.ListPlans(req)
	if err != nil {
		log.Println("ListPlan error:", err)
		return nil, err
	}
	return &pb.Plans{Plans: list, Total: int32(len(list))}, nil
}

func (u *User) UpdatePlan(ctx context.Context, req *pb.Plan) (*common.Empty, error) {
	if req.Id == "" {
		return nil, errors.New(utils.E_invalid_id)
	}
	plan, err := u.Db.GetPlan(&pb.PlansRequest{
		Id: req.Id,
	})
	if plan == nil {
		return nil, errors.New(utils.E_not_found_plan)
	}
	if err != nil {
		return nil, err
	}
	req.UpdatedAt = time.Now().Unix()
	if err := u.Db.UpdatePlan(req, &pb.Plan{Id: req.Id}); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) DeletePlan(ctx context.Context, req *pb.Plan) (*common.Empty, error) {
	if req.Id == "" {
		return nil, errors.New(utils.E_invalid_id)
	}
	if err := u.Db.DeletePlan(req.Id); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}
