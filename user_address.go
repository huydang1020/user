package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
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
	if rq.GetPhone() == "" {
		return nil, errors.New(utils.E_invalid_phone_number)
	}
	if rq.GetFullName() == "" {
		return nil, errors.New(utils.E_invalid_fullname)
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

type Province struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type District struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	ProvinceId   string `json:"province_id"`
	ProvinceName string `json:"province_name"`
}

type Ward struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	DistrictId   string `json:"district_id"`
	DistrictName string `json:"district_name"`
	ProvinceId   string `json:"province_id"`
	ProvinceName string `json:"province_name"`
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

	for _, address := range addresses {
		provinceName := ""
		districtName := ""
		wardName := ""
		for _, province := range provinces {
			if province.Id == address.Province {
				provinceName = province.Name
			}
		}
		for _, district := range districts {
			if district.Id == address.District {
				districtName = district.Name
			}
		}
		for _, ward := range wards {
			if ward.Id == address.Ward {
				wardName = ward.Name
			}
		}
		address.FullAddress = fmt.Sprintf("%s, %s, %s, %s", address.Address, wardName, districtName, provinceName)
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
		if province.Id == address.Province {
			provinceName = province.Name
		}
	}
	for _, district := range districts {
		if district.Id == address.District {
			districtName = district.Name
		}
	}
	for _, ward := range wards {
		if ward.Id == address.Ward {
			wardName = ward.Name
		}
	}
	address.FullAddress = fmt.Sprintf("%s, %s, %s, %s", address.Address, wardName, districtName, provinceName)
	return address, nil
}
