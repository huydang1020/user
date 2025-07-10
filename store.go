package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/huyshop/header/common"
	pb "github.com/huyshop/header/user"
	"github.com/huyshop/user/utils"
)

func (u *User) CreateStore(ctx context.Context, store *pb.Store) (*common.Empty, error) {
	if store.GetName() == "" {
		return nil, errors.New(utils.E_not_found_name)
	}
	if store.GetAddress() == "" {
		return nil, errors.New(utils.E_not_found_address)
	}
	if store.GetProvince() == "" {
		return nil, errors.New(utils.E_not_found_province)
	}
	if store.GetDistrict() == "" {
		return nil, errors.New(utils.E_not_found_district)
	}
	if store.GetWard() == "" {
		return nil, errors.New(utils.E_not_found_ward)
	}
	store.Id = utils.MakeStoreId()
	store.State = pb.Store_active.String()
	store.CreatedAt = time.Now().Unix()
	store.Slug = utils.ToSlug(store.Name)
	if err := u.Db.CreateStore(store); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) UpdateStore(ctx context.Context, req *pb.Store) (*common.Empty, error) {
	if req.GetId() == "" {
		return nil, errors.New(utils.E_not_found_id)
	}
	req.Slug = utils.ToSlug(req.Name)
	req.UpdatedAt = time.Now().Unix()
	if err := u.Db.UpdateStore(req, &pb.Store{Id: req.GetId()}); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) DeleteStore(ctx context.Context, req *pb.Store) (*common.Empty, error) {
	if req.GetId() == "" {
		return nil, errors.New(utils.E_not_found_id)
	}
	if err := u.Db.DeleteStore(req.GetId()); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) ListStore(ctx context.Context, rq *pb.StoreRequest) (*pb.Stores, error) {
	list, err := u.Db.ListStore(rq)
	if err != nil {
		return nil, err
	}
	for _, l := range list {
		pa, err := u.Db.GetPartner(&pb.PartnerRequest{Id: l.GetPartnerId()})
		if err != nil {
			log.Println("GetPartner", err)
			continue
		}
		l.Partner = pa
	}
	count, err := u.Db.CountStore(rq)
	if err != nil {
		return nil, err
	}
	return &pb.Stores{Stores: list, Total: int32(count)}, nil
}

func (u *User) GetStore(ctx context.Context, rq *pb.StoreRequest) (*pb.Store, error) {
	store, err := u.Db.GetStore(rq)
	if err != nil {
		return nil, err
	}
	provinces := []Province{}
	districts := []District{}
	wards := []Ward{}

	province, err := os.ReadFile("assets/province.json")
	if err != nil {
		return nil, err
	}
	district, err := os.ReadFile("assets/district.json")
	if err != nil {
		return nil, err
	}
	ward, err := os.ReadFile("assets/ward.json")
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(province, &provinces); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(district, &districts); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(ward, &wards); err != nil {
		return nil, err
	}
	provinceName := ""
	districtName := ""
	wardName := ""
	for _, province := range provinces {
		if province.Id == store.Province {
			provinceName = province.Name
		}
	}
	for _, district := range districts {
		if district.Id == store.District {
			districtName = district.Name
		}
	}
	for _, ward := range wards {
		if ward.Id == store.Ward {
			wardName = ward.Name
		}
	}
	store.FullAddress = fmt.Sprintf("%s, %s, %s, %s", store.Address, wardName, districtName, provinceName)
	return store, nil
}

func (u *User) CountStore(ctx context.Context, rq *pb.StoreRequest) (*common.Count, error) {
	count, err := u.Db.CountStore(rq)
	if err != nil {
		return nil, err
	}
	return &common.Count{Count: count}, nil
}
