package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/huyshop/header/common"
	pb "github.com/huyshop/header/user"
	"github.com/huyshop/user/jwt"
	"github.com/huyshop/user/utils"
)

const DEFAULT_LIMIT = 20

func (u *User) SignIn(ctx context.Context, req *pb.User) (*pb.SignInResponse, error) {
	if req.GetUsername() == "" {
		return nil, errors.New(utils.E_not_found_username)
	}
	if req.GetPassword() == "" {
		return nil, errors.New(utils.E_not_found_password)
	}
	user, err := u.Db.GetUser(&pb.UserRequest{PhoneNumber: req.GetUsername()})
	if err != nil {
		user, err = u.Db.GetUser(&pb.UserRequest{Email: req.GetUsername()})
		if err != nil {
			user, err = u.Db.GetUser(&pb.UserRequest{Username: req.GetUsername()})
			if err != nil {
				return nil, errors.New(utils.E_not_found_username)
			}
		}
	}
	if err := utils.ComparePassword(user.Password, req.Password); err != nil {
		return nil, errors.New(utils.E_password_is_incorrect)
	}
	partner, err := u.Db.GetPartner(&pb.PartnerRequest{Id: user.GetPartnerId()})
	if err != nil {
		return nil, err
	}
	exprAct, _ := strconv.Atoi(config.JwtExpireAccessToken)
	exprRft, _ := strconv.Atoi(config.JwtExpireRefreshToken)
	access_token, err := jwt.GenerateAccessToken(user, partner, time.Duration(exprAct), config.JwtSecretKey)
	if err != nil {
		log.Println("generate access token error:", err)
		return nil, err
	}
	refresh_token, err := jwt.GenerateRefreshToken(user, partner, time.Duration(exprRft), config.JwtSecretKey)
	if err != nil {
		log.Println("generate refresh token error:", err)
		return nil, err
	}
	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	keyRedis := fmt.Sprintf("refresh_token_user_id_%s", user.GetId())
	if err := u.cache.Set(c, keyRedis, refresh_token, time.Duration(exprRft)*time.Minute).Err(); err != nil {
		log.Println("set data redis error:", err)
		return nil, err
	}
	user.Password = ""
	// err = utils.SendEmail(config.MailKey, config.MailUrl, "dangquanghuy@media-one.vn", "Huy Shop - Verify Email", fmt.Sprintf("Your account:<br>Username: %s<br>Password: %s", req.GetUsername(), req.GetPassword()))
	// if err != nil {
	// 	log.Println("send email error:", err)
	// 	return nil, err
	// }
	return &pb.SignInResponse{
		User:        user,
		AccessToken: access_token,
	}, nil
}

func (u *User) SignOut(ctx context.Context, req *pb.User) (*common.Empty, error) {
	if req.GetId() == "" {
		return nil, errors.New(utils.E_not_found_user_id)
	}
	if err := u.cache.Del(ctx, fmt.Sprintf("refresh_token_user_id_%s", req.GetId())).Err(); err != nil {
		log.Println("delete data redis error:", err)
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) CreateUser(ctx context.Context, req *pb.User) (*common.Empty, error) {
	// if u.Db.IsUserExisted(&pb.User{Username: req.GetUsername(), PhoneNumber: req.GetPhoneNumber(), Email: req.GetEmail()}) {
	// 	return nil, errors.New(utils.E_user_existed)
	// }
	if req.RoleId == "" {
		return nil, errors.New(utils.E_not_found_role_id)
	}
	hashPw, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Println("hash pw err:", err)
		return nil, err
	}
	req.Password = hashPw
	req.Id = utils.MakeUserId()
	req.State = pb.User_active.String()
	req.CreatedAt = time.Now().Unix()
	if err := u.Db.CreateUser(req); err != nil {
		log.Println("create user err:", err)
		return nil, err
	}
	return &common.Empty{}, nil
}

func (u *User) GetUser(ctx context.Context, req *pb.UserRequest) (*pb.User, error) {
	if req.GetId() == "" && req.GetUsername() == "" && req.GetPhoneNumber() == "" {
		return nil, errors.New(utils.E_internal_error)
	}
	user, err := u.Db.GetUser(req)
	if err != nil {
		log.Println("get user error:", err)
		return nil, err
	}
	user.Password = ""
	up, err := u.Db.GetUserPoint(&pb.UserPoint{UserId: user.GetId()})
	if err != nil {
		log.Println("get user point error:", err)
		return nil, err
	}
	user.Point = &pb.UserPoint{
		Points: up.Points,
	}
	return user, nil
}

func (u *User) ListUsers(ctx context.Context, req *pb.UserRequest) (*pb.Users, error) {
	if req.GetLimit() == 0 || req.GetLimit() > DEFAULT_LIMIT {
		req.Limit = DEFAULT_LIMIT
	}
	users, err := u.Db.ListUsers(req)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return &pb.Users{}, nil
	}
	for _, u := range users {
		u.Password = ""
	}
	count, _ := u.Db.CountUsers(req)
	return &pb.Users{
		Users: users,
		Total: count,
	}, nil
}

func (u *User) UpdateUser(ctx context.Context, req *pb.User) (*common.Empty, error) {
	if req.GetId() == "" {
		return nil, errors.New(utils.E_not_found_user_id)
	}
	_, err := u.Db.GetUser(&pb.UserRequest{Id: req.GetId()})
	if err != nil {
		return nil, errors.New(utils.E_can_not_update)
	}
	req.UpdatedAt = time.Now().Unix()
	if err := u.Db.UpdateUser(req, &pb.User{Id: req.GetId()}); err != nil {
		return nil, errors.New(utils.E_can_not_update)
	}
	return &common.Empty{}, nil
}

func (u *User) DeleteUser(ctx context.Context, req *pb.User) (*common.Empty, error) {
	if req.GetId() == "" {
		return nil, errors.New(utils.E_not_found_user_id)
	}
	_, err := u.Db.GetUser(&pb.UserRequest{Id: req.GetId()})
	if err != nil {
		return nil, errors.New(utils.E_can_not_delete)
	}
	if err := u.Db.DeleteUser(req.GetId()); err != nil {
		return nil, errors.New(utils.E_can_not_delete)
	}
	return &common.Empty{}, nil
}

func (u *User) CreateNewUser(ctx context.Context, req *pb.User) (*pb.User, error) {
	if err := u.Db.TranCreateNewUser(req); err != nil {
		return nil, err
	}
	req.Password = ""
	return req, nil
}

func (u *User) IsExistUser(ctx context.Context, req *pb.User) (*common.IsExist, error) {
	c := u.Db.IsUserExisted(req)
	return &common.IsExist{Exist: c}, nil
}

func (u *User) SignInCustomer(ctx context.Context, req *pb.User) (*pb.SignInResponse, error) {
	if req.GetUsername() == "" {
		return nil, errors.New(utils.E_username_is_incorrect)
	}
	if req.GetPassword() == "" {
		return nil, errors.New(utils.E_password_is_incorrect)
	}
	user, err := u.Db.GetUser(&pb.UserRequest{PhoneNumber: req.GetUsername()})
	if err != nil {
		user, err = u.Db.GetUser(&pb.UserRequest{Email: req.GetUsername()})
		if err != nil {
			return nil, errors.New(utils.E_not_found_username)
		}
	}
	if err := utils.ComparePassword(user.Password, req.Password); err != nil {
		return nil, errors.New(utils.E_password_is_incorrect)
	}
	if user.GetState() == pb.User_inactive.String() {
		return nil, errors.New(utils.E_account_not_activated)
	}
	exprAct, _ := strconv.Atoi(config.JwtExpireAccessToken)
	exprRft, _ := strconv.Atoi(config.JwtExpireRefreshToken)
	access_token, err := jwt.GenerateAccessToken(user, &pb.Partner{}, time.Duration(exprAct), config.JwtSecretKey)
	if err != nil {
		log.Println("generate access token error:", err)
		return nil, err
	}
	refresh_token, err := jwt.GenerateRefreshToken(user, &pb.Partner{}, time.Duration(exprRft), config.JwtSecretKey)
	if err != nil {
		log.Println("generate refresh token error:", err)
		return nil, err
	}
	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	keyRedis := fmt.Sprintf("refresh_token_user_id_%s", user.GetId())
	if err := u.cache.Set(c, keyRedis, refresh_token, time.Duration(exprRft)*time.Minute).Err(); err != nil {
		log.Println("set data redis error:", err)
		return nil, err
	}
	user.Password = ""
	up, err := u.GetUserPoint(ctx, &pb.UserPointRequest{Id: user.GetId()})
	if err != nil {
		log.Println("get user point error:", err)
		return nil, err
	}
	user.Point = &pb.UserPoint{
		Points: up.Points,
	}
	return &pb.SignInResponse{
		User:        user,
		AccessToken: access_token,
	}, nil
}

func (u *User) CreateCustomer(ctx context.Context, req *pb.User) (*common.Empty, error) {
	if u.Db.IsUserExisted(&pb.User{PhoneNumber: req.GetPhoneNumber(), Email: req.GetEmail()}) {
		user, err := u.Db.GetUser(&pb.UserRequest{PhoneNumber: req.GetPhoneNumber()})
		if err != nil {
			user, err = u.Db.GetUser(&pb.UserRequest{Email: req.GetEmail()})
			if err != nil {
				return nil, errors.New(utils.E_not_found_username)
			}
		}
		log.Println("user existed:", user)
		if user.GetState() == pb.User_active.String() {
			return nil, errors.New(utils.E_user_existed)
		} else if user.GetState() == pb.User_inactive.String() {
			// u.SendEmail <- req
			return &common.Empty{}, nil
		}
	}
	hashPw, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Println("hash pw err:", err)
		return nil, err
	}
	req.Password = hashPw
	req.State = pb.User_inactive.String()
	req.Id = utils.MakeUserId()
	req.CreatedAt = time.Now().Unix()
	if err := u.Db.CreateUser(req); err != nil {
		log.Println("create user err:", err)
		return nil, err
	}
	// u.SendEmail <- req
	log.Println("123")
	return &common.Empty{}, nil
}

func (u *User) VerifyEmail(ctx context.Context, req *pb.User) (*common.Empty, error) {
	if req.GetEmail() == "" {
		return nil, errors.New(utils.E_invalid_email)
	}
	if req.GetVerifyOtp() == "" {
		return nil, errors.New(utils.E_invalid_otp)
	}
	if u.Db.IsUserExisted(&pb.User{State: pb.User_active.String(), Email: req.GetEmail()}) {
		return nil, errors.New(utils.E_email_activated)
	}
	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	keyRedis := fmt.Sprintf("verify_email:%s", req.GetEmail())
	otp, err := u.cache.Get(c, keyRedis).Result()
	if err == redis.Nil {
		return nil, errors.New(utils.E_verify_otp_is_expired)
	} else if err != nil {
		log.Println("Redis error:", err)
		return nil, errors.New(utils.E_internal_error)
	}
	if otp != req.GetVerifyOtp() {
		return nil, errors.New(utils.E_verify_otp_incorrect)
	}
	user, err := u.Db.GetUser(&pb.UserRequest{Email: req.GetEmail()})
	if err != nil {
		return nil, errors.New(utils.E_not_found_user)
	}
	user.State = pb.User_active.String()
	user.UpdatedAt = time.Now().Unix()
	if err := u.Db.UpdateUser(user, &pb.User{Id: user.GetId()}); err != nil {
		return nil, errors.New(utils.E_can_not_update)
	}
	err = u.Db.CreateUserPoint(&pb.UserPoint{UserId: user.GetId(), CreatedAt: time.Now().Unix()})
	if err != nil {
		log.Println("create user point error:", err)
		return nil, errors.New(utils.E_can_not_insert)
	}
	_ = u.cache.Del(c, keyRedis).Err()
	return &common.Empty{}, nil
}

func (u *User) SendVerifyOtp(ctx context.Context, req *pb.User) (*common.TTL, error) {
	if req.GetEmail() == "" {
		return nil, errors.New(utils.E_invalid_email)
	}
	if u.Db.IsUserExisted(&pb.User{State: pb.User_active.String(), Email: req.GetEmail()}) {
		return nil, errors.New(utils.E_email_activated)
	}
	keyRedis := fmt.Sprintf("verify_email:%s", req.GetEmail())
	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := u.cache.Get(c, keyRedis).Result()
	if err == redis.Nil {
		log.Println("otp not found in redis, sending new otp")
		code := utils.GenerateVerifyOtp()
		err = utils.SendEmail(config.MailKey, config.MailUrl, req.GetEmail(), "Huy Shop - Verify Email", code)
		if err != nil {
			log.Println("send email error:", err)
			return nil, err
		}
		if err := u.cache.Set(c, keyRedis, code, time.Duration(5)*time.Minute).Err(); err != nil {
			log.Println("set data redis error:", err)
			return nil, err
		}
	} else if err != nil {
		log.Println("Redis error:", err)
		return nil, errors.New(utils.E_internal_error)
	}
	ttl, err := u.cache.TTL(ctx, keyRedis).Result()
	if err != nil {
		log.Println("err", err)
		return nil, errors.New(utils.E_internal_error)
	}
	return &common.TTL{
		Ttl: int64(ttl.Seconds()),
	}, nil
}

func (u *User) CheckAndSendEmailVerifyOtp(req *pb.User) error {
	if req.GetEmail() == "" {
		return errors.New(utils.E_invalid_email)
	}
	keyRedis := fmt.Sprintf("verify_email:%s", req.GetEmail())

	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := u.cache.Get(c, keyRedis).Result()
	if err == redis.Nil {
		log.Println("otp not found in redis, sending new otp")
		code := utils.GenerateVerifyOtp()
		err = utils.SendEmail(config.MailKey, config.MailUrl, req.GetEmail(), "Huy Shop - Verify Email", code)
		if err != nil {
			log.Println("send email error:", err)
			return err
		}
		if err := u.cache.Set(c, keyRedis, code, time.Duration(5)*time.Minute).Err(); err != nil {
			log.Println("set data redis error:", err)
			return err
		}
	} else if err != nil {
		log.Println("Redis error:", err)
		return errors.New(utils.E_internal_error)
	}
	return nil
}
