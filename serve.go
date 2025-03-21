package main

import (
	"log"
	"net"
	"time"

	pb "github.com/huyshop/header/user"
	"github.com/huyshop/user/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type User struct {
	Db IDatabase
}

type IDatabase interface {
	ListUsers(rq *pb.UserRequest) ([]*pb.User, error)
	CountUsers(rq *pb.UserRequest) (int64, error)
	GetUser(rq *pb.UserRequest) (*pb.User, error)
	CreateUser(rq *pb.User) error
	UpdateUser(updator, selector *pb.User) error
	DeleteUser(id string) error
	IsUserExisted(u *pb.User) bool
}

func NewUser(cf *Configs) (*User, error) {
	dbase := &db.DB{}
	if err := dbase.ConnectDb(cf.DBPath, cf.DBName); err != nil {
		return nil, err
	}
	log.Println("Connect db successful")
	return &User{
		Db: dbase,
	}, nil
}

func startGRPCServe(port string, p *User) error {
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionAge: 15 * time.Second,
		}),
	}
	serve := grpc.NewServer(opts...)
	pb.RegisterUserServiceServer(serve, p)
	reflection.Register(serve)
	return serve.Serve(listen)
}
