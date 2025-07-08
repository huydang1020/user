package main

import (
	"context"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	pb "github.com/huyshop/header/user"
	"github.com/huyshop/user/db"
	"github.com/huyshop/user/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type User struct {
	Db        IDatabase
	cache     *redis.Client
	SendEmail chan *pb.User
}

type IDatabase interface {
	ListUsers(rq *pb.UserRequest) ([]*pb.User, error)
	CountUsers(rq *pb.UserRequest) (int64, error)
	GetUser(rq *pb.UserRequest) (*pb.User, error)
	CreateUser(rq *pb.User) error
	UpdateUser(updator, selector *pb.User) error
	DeleteUser(id string) error
	IsUserExisted(u *pb.User) bool
	FindUserWithPhone(phone string) (*pb.User, error)
	FindUserWithEmail(email string) (*pb.User, error)

	CreateUserPoint(req *pb.UserPoint) error
	UpdateUserPoint(updator, selector *pb.UserPoint) error
	DeleteUserPoint(id string) error
	IsExistUserPoint(id string) (bool, error)
	ListUserPoint(rq *pb.UserPointRequest) ([]*pb.UserPoint, error)
	GetUserPoint(rq *pb.UserPoint) (*pb.UserPoint, error)
	TranCreateNewUser(user *pb.User) error

	CreatePointExchange(req *pb.PointExchange) error
	GetPointExchange(req *pb.PointExchange) (*pb.PointExchange, error)
	CountPointExchange(rq *pb.PointExchangeRequest) (int64, error)
	ListPointExchange(req *pb.PointExchangeRequest) ([]*pb.PointExchange, error)
	TranCreatePointExchange(req *pb.PointExchange) error

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

	CreatePlan(req *pb.Plan) error
	GetPlan(rq *pb.PlansRequest) (*pb.Plan, error)
	ListPlans(rq *pb.PlansRequest) ([]*pb.Plan, error)
	UpdatePlan(updator, selector *pb.Plan) error
	DeletePlan(id string) error
	CountPlan(rq *pb.PlansRequest) (int64, error)

	CreateOrderPlan(req *pb.OrderPlan) error
	GetOrderPlan(rq *pb.OrderPlan) (*pb.OrderPlan, error)
	ListOrderPlan(rq *pb.OrderPlanRequest) ([]*pb.OrderPlan, error)
	UpdateOrderPlan(updator, selector *pb.OrderPlan) error
	CountOrderPlan(rq *pb.OrderPlanRequest) (int64, error)
	DeleteOrderPlan(id string) error

	CreateUserAddress(req *pb.UserAddress) error
	UpdateUserAddress(updator, selector *pb.UserAddress) error
	DeleteUserAddress(id string) error
	IsExistUserAddress(id string) (bool, error)
	ListUserAddress(rq *pb.UserAddressRequest) ([]*pb.UserAddress, error)
	GetUserAddress(rq *pb.UserAddress) (*pb.UserAddress, error)
	TranCreateUserAddress(req *pb.UserAddress, maxUserAddress int) error
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
		Db:        dbase,
		cache:     rd,
		SendEmail: make(chan *pb.User, 100),
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

func HTTPServe(p *User) error {
	g := gin.New()
	g.Use(gin.Recovery(), gin.Logger())
	r1 := g.Group("/v1/user")

	r1.POST("/create-point-exchange", func(c *gin.Context) {
		req := &pb.PointExchange{}
		if err := c.BindJSON(req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}
		if req.ReceiverId == "" {
			c.JSON(400, gin.H{"error": utils.E_invalid_receiver_id})
			return
		}
		req.Id = utils.MakePointExchangeId()
		req.CreatedAt = time.Now().Unix()
		if err := p.Db.TranCreatePointExchange(req); err != nil {
			log.Println("err: ", err)
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"code": 0, "message": "success"})

	})

	r1.GET("/test-cron", func(c *gin.Context) {
		log.Println("runn test cron")
		c.JSON(200, gin.H{"message": "test-cron"})
	})

	g.Run(":" + config.HttpPort)
	return nil
}
