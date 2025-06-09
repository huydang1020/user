package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/huyshop/header/common"
	pb "github.com/huyshop/header/user"
	"github.com/huyshop/user/utils"
)

const REDIS_KEY_ORDER_PLAN = "order_plan_"

func (u *User) CreateOrderPlan(ctx context.Context, req *pb.OrderPlan) (*pb.OrderPlan, error) {
	if req.PlanId == "" {
		return nil, errors.New(utils.E_invalid_plan_id)
	}
	if req.UserId == "" {
		return nil, errors.New(utils.E_invalid_user_id)
	}
	if req.Type == "" {
		return nil, errors.New(utils.E_invalid_plan_type)
	}
	user, err := u.Db.GetUser(&pb.UserRequest{Id: req.UserId})
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New(utils.E_not_found_user)
	}
	if user.RoleId != "roled0di17m9ipf12jq5ndlg" {
		return nil, errors.New(utils.E_already_partner)
	}
	plan, err := u.Db.GetPlan(&pb.PlansRequest{Id: req.PlanId})
	if err != nil {
		log.Println("get plan by id error:", err)
		return nil, errors.New(utils.E_internal_error)
	}
	if plan == nil {
		return nil, errors.New(utils.E_plan_not_found)
	}
	req.Plan = plan
	var totalMoney int64
	for _, price := range plan.Prices {
		if price.Type == req.Type {
			totalMoney = price.Price
			break
		}
	}
	if totalMoney == 0 {
		return nil, errors.New(utils.E_not_found_plam_type)
	}
	log.Println("totalMoney:", totalMoney, "plan:", plan.GetId(), "type:", req.Type)
	log.Println("123: ", fmt.Sprintf("%d", totalMoney))
	randNumber := rand.Intn(99999999999999-10000000000000) + 10000000000000
	req.OrderCode = fmt.Sprint(randNumber)
	req.CreatedAt = time.Now().Unix()
	vnpUrl := os.Getenv("VNP_URL")
	vnpSecret := os.Getenv("VNP_HASH_SECRET")
	vnpTmnCode := os.Getenv("VNP_TMNCODE")
	createdDate, err := utils.ConvertUnixToDateTime("20060102150405", req.CreatedAt)
	if err != nil {
		log.Println("convert time err:", err)
		return nil, errors.New(utils.E_internal_error)
	}
	vnpParams := url.Values{}
	vnpParams.Set("vnp_Version", "2.1.0")
	vnpParams.Set("vnp_Command", "pay")
	vnpParams.Set("vnp_TmnCode", vnpTmnCode)
	vnpParams.Set("vnp_Locale", "vn")
	vnpParams.Set("vnp_CurrCode", "VND")
	vnpParams.Set("vnp_TxnRef", fmt.Sprint(randNumber))
	vnpParams.Set("vnp_OrderInfo", "Thanh toán cho giao dịch: "+req.OrderCode)
	vnpParams.Set("vnp_OrderType", "billpayment")
	vnpParams.Set("vnp_Amount", strconv.FormatInt(totalMoney*100, 10))
	vnpParams.Set("vnp_ReturnUrl", req.VnpayReturnUrl)
	vnpParams.Set("vnp_IpAddr", req.IpAddress)
	vnpParams.Set("vnp_CreateDate", createdDate)
	vnpParams.Set("vnp_BankCode", "VNBANK")

	sortedParams := utils.SortParams(vnpParams)
	signData := sortedParams.Encode()
	hmacSecret := hmac.New(sha512.New, []byte(vnpSecret))
	hmacSecret.Write([]byte(signData))
	signature := fmt.Sprintf("%x", hmacSecret.Sum(nil))
	vnpParams.Set("vnp_SecureHash", signature)
	vnpRedirectURL := vnpUrl + "?" + vnpParams.Encode()
	byteOrder, err := json.Marshal(req)
	if err != nil {
		log.Println("marshal order err:", err)
		return nil, errors.New(utils.E_internal_error)
	}
	exprOderRedis, _ := strconv.Atoi(os.Getenv("TIME_LIVE_ORDER_REDIS"))
	keyRedis := REDIS_KEY_ORDER_PLAN + req.OrderCode
	if err := u.cache.Set(ctx, keyRedis, string(byteOrder), time.Duration(exprOderRedis)*time.Second).Err(); err != nil {
		log.Println("set data redis error:", err)
		return nil, errors.New(utils.E_internal_error)
	}
	return &pb.OrderPlan{VnpayReturnUrl: vnpRedirectURL}, nil
}

func (u *User) CreateOrderPlanVNpay(ctx context.Context, req *pb.OrderPlan) (*common.Empty, error) {
	if req.GetOrderCode() == "" {
		return nil, errors.New(utils.E_not_found_order_code)
	}
	keyRedis := REDIS_KEY_ORDER_PLAN + req.OrderCode
	result, err := u.cache.Get(ctx, keyRedis).Result()
	if err == redis.Nil {
		log.Println("redis key does not exist:", keyRedis)
		return nil, errors.New(utils.E_not_found_order_data)
	} else if err != nil {
		log.Println("get data redis error:", err)
		return nil, errors.New(utils.E_internal_error)
	}
	order := &pb.OrderPlan{}
	if err := json.Unmarshal([]byte(result), order); err != nil {
		log.Println("unmarshal err:", err)
		return nil, errors.New(utils.E_internal_error)
	}
	order.Id = utils.MakeOrderPlanId()
	order.StartDate = time.Now().Unix()
	switch order.Type {
	case "1 month":
		order.EndDate = order.StartDate + 30*24*3600
	case "3 month":
		order.EndDate = order.StartDate + 90*24*3600
	case "6 month":
		order.EndDate = order.StartDate + 180*24*3600
	case "1 year":
		order.EndDate = order.StartDate + 365*24*3600
	case "Unlimited":
		order.EndDate = -1
	default:
		return nil, errors.New(utils.E_invalid_plan_type)
	}
	pr := &pb.PartnerRegistration{
		Id:        utils.MakePartnerRegistrationId(),
		UserId:    order.UserId,
		State:     pb.PartnerRegistration_pending.String(),
		CreatedAt: time.Now().Unix(),
		PlanId:    order.PlanId,
		PlanType:  order.Type,
	}
	order.PartnerRegistration = pr
	if err := u.Db.TranCreateOrderPlan(order); err != nil {
		log.Println("trans insert order err:", err)
		return nil, errors.New(utils.E_internal_error)
	}
	if err := u.cache.Del(ctx, keyRedis); err != nil {
		log.Println("del key redis err:", err)
	}
	return &common.Empty{}, nil
}

func (u *User) GetOrderPlan(ctx context.Context, req *pb.OrderPlan) (*pb.OrderPlan, error) {
	if req.Id == "" {
		return nil, errors.New(utils.E_not_found_order_plan_id)
	}
	orderPlan, err := u.Db.GetOrderPlan(req)
	if err != nil {
		return nil, err
	}
	if orderPlan == nil {
		return nil, errors.New(utils.E_order_plan_not_found)
	}
	return orderPlan, nil
}

func (u *User) ListOrderPlan(ctx context.Context, req *pb.OrderPlanRequest) (*pb.OrderPlans, error) {
	log.Println("req: ", req)
	orderPlans, err := u.Db.ListOrderPlan(req)
	if err != nil {
		return nil, err
	}
	
	return &pb.OrderPlans{OrderPlans: orderPlans, Total: int32(len(orderPlans))}, nil
}

func (u *User) UpdateOrderPlan(ctx context.Context, req *pb.OrderPlan) (*common.Empty, error) {
	if req.Id == "" {
		return nil, errors.New(utils.E_not_found_order_plan_id)
	}

	if err := u.Db.UpdateOrderPlan(req, &pb.OrderPlan{Id: req.GetId()}); err != nil {
		return nil, err
	}
	return &common.Empty{}, nil
}