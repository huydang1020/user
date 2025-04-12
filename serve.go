package main

import (
	"context"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	pb "github.com/huyshop/header/user"
	"github.com/huyshop/user/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type User struct {
	Db    IDatabase
	cache *redis.Client
}

type IDatabase interface {
	ListUsers(rq *pb.UserRequest) ([]*pb.User, error)
	CountUsers(rq *pb.UserRequest) (int64, error)
	GetUser(rq *pb.UserRequest) (*pb.User, error)
	CreateUser(rq *pb.User) error
	UpdateUser(updator, selector *pb.User) error
	DeleteUser(id string) error
	IsUserExisted(u *pb.User) bool

	CreateUserPoint(req *pb.UserPoint) error
	UpdateUserPoint(updator, selector *pb.UserPoint) error
	DeleteUserPoint(id string) error
	IsExistUserPoint(id string) bool
	ListUserPoint(rq *pb.UserPointRequest) ([]*pb.UserPoint, error)
	GetUserPoint(rq *pb.UserPoint) (*pb.UserPoint, error)
	TranCreateNewUser(user *pb.User) error

	CreatePointExchange(req *pb.PointExchange) error
	GetPointExchange(req *pb.PointExchange) (*pb.PointExchange, error)
	ListPointExchange(req *pb.PointExchangeRequest) ([]*pb.PointExchange, error)

	CreateStore(store *pb.Store) error
	UpdateStore(updator, selector *pb.Store) error
	DeleteStore(id string) error
	GetStore(rq *pb.StoreRequest) (*pb.Store, error)
	ListStore(rq *pb.StoreRequest) ([]*pb.Store, error)
	CountStore(rq *pb.StoreRequest) (int64, error)

	CreatePartner(partner *pb.Partner) error
	UpdatePartner(updator, selector *pb.Partner) error
	DeletePartner(id string) error
	GetPartner(rq *pb.PartnerRequest) (*pb.Partner, error)
	ListPartner(rq *pb.PartnerRequest) ([]*pb.Partner, error)
	CountPartner(rq *pb.PartnerRequest) (int64, error)
}

func NewRedisCache(addr, pw string, db int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pw,
		DB:       db,
	})
	tick := time.NewTicker(10 * time.Minute)
	ctx := context.Background()
	go func(client *redis.Client) {
		for {
			select {
			case <-tick.C:
				if err := client.Ping(ctx).Err(); err != nil {
					panic(err)
				}
			}
		}
	}(client)
	return client
}

func NewUser(cf *Configs) (*User, error) {
	dbase := &db.DB{}
	if err := dbase.ConnectDb(cf.DBPath, cf.DBName); err != nil {
		return nil, err
	}
	log.Println("Connect db successful")
	redisDb, _ := strconv.Atoi(config.RedisDb)
	rd := NewRedisCache(config.RedisAddr, config.RedisPassword, redisDb)
	log.Println("Connect redis successful")
	return &User{
		Db:    dbase,
		cache: rd,
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
