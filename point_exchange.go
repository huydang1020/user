package main

import (
	"context"
	"errors"
	"log"
	"time"

	pb "github.com/huyshop/header/user"
	"github.com/huyshop/user/utils"
)

func (u *User) CreatePointExchange(ctx context.Context, req *pb.PointExchange) (*pb.PointExchange, error) {
	log.Println("req: ", req)
	if req.ReceiverId == "" {
		return nil, errors.New(utils.E_invalid_receiver_id)
	}

	req.Id = utils.MakePointExchangeId()
	req.CreatedAt = time.Now().Unix()
	if err := u.Db.TranCreatePointExchange(req); err != nil {
		log.Println("err: ", err)
		return nil, err
	}
	return nil, nil
}

func (u *User) GetPointExchange(ctx context.Context, req *pb.PointExchange) (*pb.PointExchange, error) {
	log.Println("req: ", req)
	pointExchange, err := u.Db.GetPointExchange(req)
	if err != nil {
		log.Println("GetPointExchange err: ", err)
		return nil, err
	}
	return pointExchange, nil
}

func (u *User) ListPointExchange(ctx context.Context, req *pb.PointExchangeRequest) (*pb.PointExchanges, error) {
	log.Println("req: ", req)
	list, err := u.Db.ListPointExchange(req)
	if err != nil {
		log.Println("ListPointExchange err: ", err)
		return nil, err
	}
	for _, v := range list {
		user, err := u.Db.GetUser(&pb.UserRequest{Id: v.ReceiverId})
		if err != nil {
			log.Println("GetUser err: ", err)
			return nil, err
		}
		v.Receiver = user
	}
	count, err := u.Db.CountPointExchange(req)
	if err != nil {
		log.Println("CountPointExchange err: ", err)
		return nil, err
	}
	return &pb.PointExchanges{PointExchanges: list, Total: int32(count)}, nil
}
