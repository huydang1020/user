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
