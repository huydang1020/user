package main

import (
	"context"
	"log"

	pb "github.com/huyshop/header/user"
)

func (u *User) CreatePointExchange(ctx context.Context, req *pb.PointExchange) (*pb.PointExchange, error) {
	
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
	return &pb.PointExchanges{PointExchanges: list, }, nil
}
