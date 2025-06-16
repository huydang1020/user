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
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/huyshop/header/common"
	pb "github.com/huyshop/header/user"
	"github.com/huyshop/user/utils"
)

const (
	REDIS_KEY_ORDER_PLAN = "order_plan_"
	ROLE_SELLER          = "roled0vspl69ipf5ueqvq6v0"
	ROLE_CUSTOMER        = "roled0di17m9ipf12jq5ndlg"
	ACCEPT               = "accept"
	REJECT               = "rejected"
)

func (u *User) CreateOrderPlan(ctx context.Context, req *pb.OrderPlan) (*pb.OrderPlan, error) {
	if req.PlanId == "" {
		return nil, errors.New(utils.E_invalid_plan_id)
	}
	if req.UserId == "" {
		return nil, errors.New(utils.E_invalid_user_id)
	}
	if req.Type != "tháng" && req.Type != "năm" {
		return nil, errors.New(utils.E_invalid_plan_type)
	}
	if req.Action == "" {
		return nil, errors.New(utils.E_invalid_action_type)
	}
	if req.VnpayReturnUrl == "" {
		return nil, errors.New(utils.E_invalid_vnpay_return_url)
	}
	user, err := u.Db.GetUser(&pb.UserRequest{Id: req.UserId})
	if err != nil || user == nil {
		log.Println("get user by id error:", err)
		return nil, err
	}
	if user.State != pb.User_active.String() {
		return nil, errors.New(utils.E_user_not_active)
	}
	plan, err := u.Db.GetPlan(&pb.PlansRequest{Id: req.PlanId})
	if err != nil {
		log.Println("get plan by id error:", err)
		return nil, errors.New(utils.E_internal_error)
	}
	if plan == nil {
		return nil, errors.New(utils.E_plan_not_found)
	}
	totalMoney, uPlan, err := u.calculatePayment(user, plan, req.Action, req.Type)
	if err != nil {
		log.Println("calculate payment error:", err)
		return nil, err
	}
	if totalMoney == 0 {
		return nil, errors.New(utils.E_invalid_prices)
	}
	req.UpdatePlan = uPlan
	log.Println("totalMoney:", totalMoney, "plan:", plan.GetId(), "type:", req.Type)
	log.Println("123: ", fmt.Sprintf("%d", totalMoney))
	randNumber := rand.Intn(99999999999999-10000000000000) + 10000000000000
	req.OrderCode = fmt.Sprint(randNumber)
	req.CreatedAt = time.Now().Unix()
	req.PlanPrice = totalMoney
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
	return &pb.OrderPlan{VnpRedirectUrl: vnpRedirectURL, UpdatePlan: uPlan}, nil
}

func (u *User) calculatePayment(user *pb.User, newPlan *pb.Plan, action, newPlanType string) (int64, *pb.UpdatePlan, error) {
	if newPlan == nil {
		return 0, nil, errors.New(utils.E_plan_not_found)
	}

	newPrice := getPrice(newPlan.Prices, newPlanType)
	if newPrice == 0 {
		return 0, nil, errors.New(utils.E_invalid_plan_type)
	}

	// Xử lý theo từng action type
	switch action {
	case pb.OrderPlan_create.String():
		return u.handleCreateNew(user, newPrice, newPlanType)

	case pb.OrderPlan_renew.String():
		return u.handleRenew(user, newPlan, newPrice, newPlanType)

	default:
		return 0, nil, errors.New(utils.E_invalid_action_type)
	}
}

func (u *User) handleCreateNew(user *pb.User, newPrice int64, planType string) (int64, *pb.UpdatePlan, error) {
	if user.RoleId != ROLE_CUSTOMER {
		return 0, nil, errors.New(utils.E_already_partner)
	}

	// Kiểm tra planType hợp lệ (tháng hoặc năm)
	if planType != "tháng" && planType != "năm" {
		return 0, nil, errors.New(utils.E_invalid_plan_type)
	}

	return newPrice, nil, nil
}

func (u *User) handleRenew(user *pb.User, newPlan *pb.Plan, newPrice int64, newPlanType string) (int64, *pb.UpdatePlan, error) {
	partner, currentPlan, err := u.getCurrentPlanInfo(user.PartnerId)
	if err != nil {
		return 0, nil, err
	}

	// Kiểm tra planType hợp lệ (tháng hoặc năm)
	if newPlanType != "tháng" && newPlanType != "năm" {
		return 0, nil, errors.New(utils.E_invalid_plan_type)
	}

	// Lấy planType hiện tại từ partner
	currentPlanType := partner.Type
	currentPrice := getPrice(currentPlan.Prices, currentPlanType)
	if currentPrice == 0 {
		return 0, nil, errors.New(utils.E_invalid_plan_type)
	}

	planUpdate := &pb.UpdatePlan{
		PlanOld:         currentPlan,
		PlanNew:         newPlan,
		RemainingAmount: newPrice,
	}

	// Nếu renew cùng gói hoặc gói rẻ hơn
	if newPrice <= currentPrice && newPlanType == currentPlanType {
		return newPrice, planUpdate, nil
	}

	// Nếu renew lên gói cao hơn hoặc thay đổi loại gói (tính toán chênh lệch)
	adjustedPrice, err := u.calculatePriceDifference(
		currentPrice,
		newPrice,
		partner.PlanExpiredAt,
		currentPlanType,
		newPlanType,
	)
	if err != nil {
		return 0, nil, err
	}

	return adjustedPrice, planUpdate, nil
}

// Helper functions
func (u *User) getCurrentPlanInfo(partnerId string) (*pb.Partner, *pb.Plan, error) {
	if partnerId == "" {
		return nil, nil, errors.New(utils.E_not_found_partner_id)
	}

	partner, err := u.Db.GetPartner(&pb.PartnerRequest{Id: partnerId})
	if err != nil || partner == nil {
		return nil, nil, errors.New(utils.E_not_found_partner_id)
	}

	currentPlan, err := u.Db.GetPlan(&pb.PlansRequest{Id: partner.PlanId})
	if err != nil || currentPlan == nil {
		return nil, nil, errors.New(utils.E_plan_not_found)
	}

	return partner, currentPlan, nil
}

func (u *User) calculatePriceDifference(currentPrice, newPrice int64, expiredAt int64, currentPlanType, newPlanType string) (int64, error) {
	remainingDays := time.Unix(expiredAt, 0).Sub(time.Now()).Hours() / 24
	if remainingDays <= 0 {
		return newPrice, nil
	}

	// Tính toán số ngày còn lại theo đơn vị của gói cũ
	var remainingCredit int64
	if currentPlanType == "tháng" {
		remainingCredit = (currentPrice * int64(remainingDays)) / 30 // 1 tháng = 30 ngày
	} else { // "năm"
		remainingCredit = (currentPrice * int64(remainingDays)) / 365 // 1 năm = 365 ngày
	}

	// Nếu cùng loại gói (tháng-tháng hoặc năm-năm)
	if currentPlanType == newPlanType {
		return max(newPrice-remainingCredit, 0), nil
	}
	return newPrice, nil
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func getPrice(prices []*pb.Price, planType string) int64 {
	for _, price := range prices {
		if price.Type == planType {
			return price.Price
		}
	}
	return 0
}

func (u *User) CreateOrderPlanVNpay(ctx context.Context, req *pb.OrderPlan) (*common.Empty, error) {
	log.Println("req: ", req)
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
	user, err := u.Db.GetUser(&pb.UserRequest{Id: order.UserId})
	if err != nil {
		log.Println("get user by id error:", err)
		return nil, errors.New(utils.E_internal_error)
	}
	if user == nil {
		log.Println("user not found:", order.UserId)
		return nil, errors.New(utils.E_not_found_user)
	}

	order.CreatedAt = time.Now().Unix()
	if err := u.Db.CreateOrderPlan(order); err != nil {
		log.Println("trans insert order err:", err)
		return nil, errors.New(utils.E_internal_error)
	}

	newPlan, err := u.Db.GetPlan(&pb.PlansRequest{Id: order.PlanId})
	if err != nil || newPlan == nil {
		log.Println("get plan by id error:", err)
		return nil, errors.New(utils.E_plan_not_found)
	}

	// Kiểm tra planType hợp lệ
	if order.Type != "tháng" && order.Type != "năm" {
		return nil, errors.New(utils.E_invalid_plan_type)
	}

	// Xử lý thời gian hết hạn
	var planExpiredAt int64
	now := time.Now()

	if order.Action == pb.OrderPlan_create.String() {
		// Xử lý tạo mới: thời gian hết hạn tính từ now
		switch order.Type {
		case "tháng":
			planExpiredAt = now.AddDate(0, 1, 0).Unix()
		case "năm":
			planExpiredAt = now.AddDate(1, 0, 0).Unix()
		}

		partner := &pb.Partner{
			Id:                  utils.MakePartnerId(),
			Name:                user.GetFullName(),
			Type:                pb.Partner_seller.String(),
			State:               pb.Partner_active.String(),
			PlanId:              order.GetPlanId(),
			PlanExpiredAt:       planExpiredAt,
			MaxStoresAllowed:    newPlan.GetMaxStoresAllowed(),
			MaxProductsPerStore: newPlan.GetMaxProductsPerStore(),
			PlanType:            order.Type,
			CreatedAt:           now.Unix(),
		}

		if err := u.Db.CreatePartner(partner); err != nil {
			log.Println("create partner error:", err)
			return nil, errors.New(utils.E_internal_error)
		}

		// Cập nhật thông tin user
		user.PartnerId = partner.Id
		user.RoleId = ROLE_SELLER
		if err := u.Db.UpdateUser(user, &pb.User{Id: user.Id}); err != nil {
			log.Println("update user error:", err)
			return nil, errors.New(utils.E_internal_error)
		}

	} else if order.Action == pb.OrderPlan_renew.String() {
		// Xử lý renew: thời gian hết hạn tính từ ngày hết hạn hiện tại hoặc now nếu đã hết hạn
		partner, err := u.Db.GetPartner(&pb.PartnerRequest{Id: user.PartnerId})
		if err != nil || partner == nil {
			log.Println("get partner by id error:", err)
			return nil, errors.New(utils.E_internal_error)
		}

		oldPrice := getPrice(partner.Plan.Prices, partner.PlanType)
		newPrice := getPrice(newPlan.Prices, order.Type)

		currentExpiredAt := time.Unix(partner.PlanExpiredAt, 0)
		if currentExpiredAt.Before(now) {
			currentExpiredAt = now
		}

		// Logic xử lý thời gian
		if newPrice <= oldPrice && order.Type == partner.PlanType {
			// TRƯỜNG HỢP 1: Giữ nguyên gói hoặc gói rẻ hơn (cùng loại)
			// Cộng dồn thời gian từ ngày hết hạn cũ
			switch order.Type {
			case "tháng":
				planExpiredAt = currentExpiredAt.AddDate(0, 1, 0).Unix()
			case "năm":
				planExpiredAt = currentExpiredAt.AddDate(1, 0, 0).Unix()
			}
		} else {
			// TRƯỜNG HỢP 2: Nâng cấp gói hoặc đổi loại gói
			// Tính từ thời điểm hiện tại
			switch order.Type {
			case "tháng":
				planExpiredAt = now.AddDate(0, 1, 0).Unix()
			case "năm":
				planExpiredAt = now.AddDate(1, 0, 0).Unix()
			}
		}

		if err = u.Db.UpdatePartner(&pb.Partner{
			Id:                  user.PartnerId,
			PlanId:              order.PlanId,
			MaxStoresAllowed:    newPlan.MaxStoresAllowed,
			MaxProductsPerStore: newPlan.MaxProductsPerStore,
			PlanExpiredAt:       planExpiredAt,
			PlanType:            order.Type,
			UpdatedAt:           now.Unix(),
		}, &pb.Partner{Id: user.PartnerId}); err != nil {
			log.Println("update partner error:", err)
			return nil, errors.New(utils.E_internal_error)
		}
	} else {
		return nil, errors.New(utils.E_invalid_action_type)
	}

	// Gửi email thông báo
	req.UserId = order.UserId
	if err := u.SendEmailCreatePartner(req, ACCEPT, ""); err != nil {
		log.Println("send email create partner error:", err)
		return nil, errors.New(utils.E_internal_error)
	}

	// Xóa dữ liệu redis sau khi đã lưu vào database
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
	log.Println("ListOrderPlan request:", req)
	orderPlans, err := u.Db.ListOrderPlan(req)
	if err != nil {
		return nil, err
	}

	includes := req.GetIncludes()
	if includes != nil {
		if utils.Include(includes, "user") {
			for i, orderPlan := range orderPlans {
				user, err := u.Db.GetUser(&pb.UserRequest{Id: orderPlan.UserId})
				if err != nil {
					log.Println("Error getting user by id:", err)
					return nil, errors.New(utils.E_internal_error)
				}
				orderPlans[i].User = user
			}
		}
		if utils.Include(includes, "plan") {
			for i, orderPlan := range orderPlans {
				plan, err := u.Db.GetPlan(&pb.PlansRequest{Id: orderPlan.PlanId})
				if err != nil {
					log.Println("Error getting plan by id:", err)
					return nil, errors.New(utils.E_internal_error)
				}
				orderPlans[i].Plan = plan
			}
		}
	}

	return &pb.OrderPlans{
		OrderPlans: orderPlans,
		Total:      int32(len(orderPlans)),
	}, nil
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

func (u *User) SendEmailCreatePartner(re *pb.OrderPlan, action, reasonReject string) error {
	user, err := u.Db.GetUser(&pb.UserRequest{Id: re.GetUserId()})
	if err != nil {
		log.Println("get user err:", err)
		return err
	}
	var bin []byte
	var subject string
	log.Println("Gửi mail:", user.GetEmail())
	switch action {
	case ACCEPT:
		subject = "🎉 Chúc mừng! Đơn đăng ký bán hàng của bạn đã được duyệt"
		bin, err = os.ReadFile("assets/approved_partner.html")
		if err != nil {
			log.Println("read file err:", err)
			return err
		}
	case REJECT:
		subject = "😔 Thông báo về đơn đăng ký bán hàng của bạn"
		bin, err = os.ReadFile("assets/rejected_partner.html")
		if err != nil {
			log.Println("read file err:", err)
			return err
		}
	default:
		return errors.New("invalid state for sending email")
	}
	sending_time, _ := utils.ConvertUnixToDateTime("2006-01-02 15:04:05", time.Now().Unix())
	bodyMail := string(bin)
	metric := map[string]string{
		"fullname":      user.GetFullName(),
		"createdAt":     sending_time,
		"reason_reject": reasonReject,
	}
	for k, v := range metric {
		bodyMail = strings.Replace(bodyMail, "{{"+k+"}}", v, -1)
	}
	err = utils.SendEmailPartnerRegistration(
		config.MailKey,
		config.MailUrl,
		user.GetEmail(),
		action,
		subject,
		bodyMail,
	)
	return err
}
