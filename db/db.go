package db

import (
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	pb "github.com/huyshop/header/user"
	"github.com/huyshop/user/utils"
	"xorm.io/xorm"
)

type DB struct {
	engine *xorm.Engine
}

func (d *DB) ConnectDb(sqlPath, dbName string) error {
	sqlConnStr := fmt.Sprintf("%s/%s", sqlPath, dbName)
	engine, err := xorm.NewEngine("mysql", sqlConnStr)
	if err != nil {
		return err
	}
	tickPingSql := time.NewTicker(15 * time.Minute)
	go func() {
		for {
			select {
			case <-tickPingSql.C:
				if err := engine.Ping(); err != nil {
					log.Print("sql can not ping")
				}
			}
		}
	}()
	d.engine = engine
	d.engine.ShowSQL(false)
	return err
}

func (d *DB) listUsersQuery(rq *pb.UserRequest) *xorm.Session {
	ss := d.engine.Table(tblUser)
	if rq.GetUsername() != "" {
		ss.And("username = ?", rq.GetUsername())
	}
	if rq.GetFullName() != "" {
		ss.And("full_name like ?", "%"+rq.GetFullName()+"%")
	}
	if rq.GetPhoneNumber() != "" {
		ss.And("phone_number = ?", rq.GetPhoneNumber())
	}
	if rq.GetId() != "" {
		ss.And("id = ?", rq.GetId())
	}
	if rq.GetRoleId() != "" {
		ss.And("role_id = ?", rq.GetRoleId())
	}
	if rq.GetState() != "" {
		ss.And("state = ?", rq.GetState())
	}
	if rq.GetEmail() != "" {
		ss.And("email = ?", rq.GetEmail())
	}
	if rq.GetFrom() != 0 {
		ss.And("created_at >= ?", rq.GetFrom())
	}
	if rq.GetTo() != 0 {
		ss.And("created_at <= ?", rq.GetTo())
	}
	return ss
}

func (d *DB) ListUsers(rq *pb.UserRequest) ([]*pb.User, error) {
	ss := d.listUsersQuery(rq)
	if rq.GetLimit() != 0 {
		ss.Limit(int(rq.GetLimit()), int(rq.GetSkip()*rq.GetLimit()))
	}
	users := make([]*pb.User, 0)
	err := ss.Desc("created_at").Find(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (d *DB) CountUsers(rq *pb.UserRequest) (int64, error) {
	ss := d.listUsersQuery(rq)
	return ss.Count()
}

func (d *DB) GetUser(rq *pb.UserRequest) (*pb.User, error) {
	user := &pb.User{
		Id:          rq.GetId(),
		PhoneNumber: rq.GetPhoneNumber(),
		Email:       rq.GetEmail(),
		Username:    rq.GetUsername(),
	}
	ishas, err := d.engine.Get(user)
	if err != nil {
		return nil, err
	}
	if !ishas {
		return nil, errors.New(utils.E_not_found_user)
	}
	return user, nil
}

func (d *DB) FindUserWithUsername(username string) (*pb.User, error) {
	user := &pb.User{Username: username}
	ishas, err := d.engine.Table(tblUser).Get(user)
	if err != nil {
		return nil, err
	}
	if !ishas {
		return nil, errors.New(utils.E_not_found_user)
	}
	return user, nil
}

func (d *DB) FindUserWithPhone(phone string) (*pb.User, error) {
	user := &pb.User{PhoneNumber: phone}
	ishas, err := d.engine.Table(tblUser).Get(user)
	if err != nil {
		return nil, err
	}
	if !ishas {
		return nil, errors.New(utils.E_not_found_user)
	}
	return user, nil
}

func (d *DB) FindUserWithEmail(email string) (*pb.User, error) {
	user := &pb.User{Email: email}
	ishas, err := d.engine.Table(tblUser).Get(user)
	if err != nil {
		return nil, err
	}
	if !ishas {
		return nil, errors.New(utils.E_not_found_user)
	}
	return user, nil
}

func (d *DB) IsUserExisted(u *pb.User) bool {
	ss := d.engine.Table(tblUser)
	if u.GetUsername() != "" {
		ss = ss.Or("username = ?", u.GetUsername())
	}
	if u.GetPhoneNumber() != "" {
		ss = ss.Or("phone_number = ?", u.GetPhoneNumber())
	}
	if u.GetEmail() != "" {
		ss = ss.Or("email = ?", u.GetEmail())
	}
	any, err := ss.Exist()
	if err != nil {
		return false
	}
	return any
}

func (d *DB) CreateUser(user *pb.User) error {
	log.Println("create user:", user)
	c, err := d.engine.Insert(user)
	if err != nil {
		return err
	}
	if c == 0 {
		return errors.New(utils.E_can_not_insert)
	}
	return nil
}

func (d *DB) UpdateUser(updator, selector *pb.User) error {
	c, err := d.engine.Update(updator, selector)
	if err != nil {
		return err
	}
	if c == 0 {
		return errors.New(utils.E_can_not_update)
	}
	return nil
}

func (d *DB) DeleteUser(id string) error {
	_, err := d.engine.Table(tblUser).Delete(&pb.User{Id: id})
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) TranCreateNewUser(user *pb.User) error {
	ss := d.engine.NewSession()
	defer ss.Close()
	if err := ss.Begin(); err != nil {
		return err
	}
	if d.IsUserExisted(user) {
		ss.Rollback()
		return errors.New(utils.E_user_existed)
	}
	if _, err := ss.Insert(user); err != nil {
		ss.Rollback()
		return err
	}
	if _, err := ss.Insert(&pb.UserPoint{UserId: user.Id, CreatedAt: time.Now().Unix()}); err != nil {
		ss.Rollback()
		return err
	}
	if err := ss.Commit(); err != nil {
		return err
	}
	return nil
}

func (d *DB) CreateUserPoint(req *pb.UserPoint) error {
	c, err := d.engine.Insert(req)
	if err != nil {
		return err
	}
	if c == 0 {
		return errors.New(utils.E_can_not_insert)
	}
	return nil
}

func (d *DB) UpdateUserPoint(updator, selector *pb.UserPoint) error {
	c, err := d.engine.Update(updator, selector)
	if err != nil {
		return err
	}
	if c == 0 {
		return errors.New(utils.E_can_not_update)
	}
	return nil
}

func (d *DB) DeleteUserPoint(id string) error {
	if id == "" {
		return errors.New(utils.E_not_found_id)
	}
	_, err := d.engine.Table(tblUserPoint).Delete(&pb.UserPoint{UserId: id})
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) IsExistUserPoint(id string) (bool, error) {
	ss := d.engine.Where("user_id = ?", id).Table(tblUserPoint)
	return ss.Exist()
}

func (d *DB) listUserPointQuery(rq *pb.UserPointRequest) *xorm.Session {
	ss := d.engine.Table(tblUserPoint)
	if len(rq.GetUserIds()) > 0 {
		ss.In("user_id", rq.GetUserIds())
	} else if rq.GetUserId() != "" {
		ss.And("user_id = ?", rq.GetUserId())
	}
	return ss
}

func (d *DB) GetUserPoint(rq *pb.UserPoint) (*pb.UserPoint, error) {
	ishas, err := d.engine.Get(rq)
	if err != nil {
		return nil, err
	}
	if !ishas {
		return nil, errors.New(utils.E_not_found_user_point)
	}
	return rq, nil
}

func (d *DB) ListUserPoint(rq *pb.UserPointRequest) ([]*pb.UserPoint, error) {
	ss := d.listUserPointQuery(rq)
	up := make([]*pb.UserPoint, 0)
	if rq.GetLimit() != 0 {
		ss.Limit(int(rq.GetLimit()), int(rq.GetSkip()*rq.GetLimit()))
	}
	err := ss.Desc("created_at").Find(&up)
	if err != nil {
		return nil, err
	}
	return up, nil
}

func (d *DB) CreatePointExchange(req *pb.PointExchange) error {
	c, err := d.engine.Insert(req)
	if err != nil {
		return err
	}
	if c == 0 {
		return errors.New(utils.E_can_not_insert)
	}
	return nil
}

func (d *DB) GetPointExchange(req *pb.PointExchange) (*pb.PointExchange, error) {
	ishas, err := d.engine.Get(req)
	if err != nil {
		return nil, err
	}
	if !ishas {
		return nil, errors.New(utils.E_not_found)
	}
	return req, nil
}

func (d *DB) CountPointExchange(rq *pb.PointExchangeRequest) (int64, error) {
	ss := d.listPointExchangeQuery(rq)
	return ss.Count()
}

func (d *DB) listPointExchangeQuery(rq *pb.PointExchangeRequest) *xorm.Session {
	ss := d.engine.Table(tblPointExchange)
	if len(rq.GetIds()) > 0 {
		ss.In("id", rq.GetIds())
	} else if rq.Id != "" {
		ss.And("id = ?", rq.GetId())
	}
	if len(rq.GetReceiverId()) > 0 {
		ss.In("receiver_id", rq.GetReceiverId())
	} else if rq.GetReceiverId() != "" {
		ss.And("receiver_id = ?", rq.GetReceiverId())
	}
	return ss
}

func (d *DB) ListPointExchange(req *pb.PointExchangeRequest) ([]*pb.PointExchange, error) {
	ss := d.listPointExchangeQuery(req)
	if req.GetLimit() != 0 {
		ss.Limit(int(req.GetLimit()), int(req.GetSkip()*req.GetLimit()))
	}
	up := make([]*pb.PointExchange, 0)
	err := ss.Desc("created_at").Find(&up)
	if err != nil {
		return nil, err
	}
	return up, nil
}

func (d *DB) TranCreatePointExchange(req *pb.PointExchange) error {
	ss := d.engine.NewSession()
	defer ss.Close()

	if err := ss.Begin(); err != nil {
		return err
	}

	// Kiểm tra user có tồn tại không
	if _, err := d.GetUser(&pb.UserRequest{Id: req.ReceiverId}); err != nil {
		ss.Rollback()
		return errors.New(utils.E_user_not_existed)
	}

	now := time.Now().Unix()

	// Kiểm tra user point có tồn tại không
	up, err := d.GetUserPoint(&pb.UserPoint{UserId: req.ReceiverId})
	if err != nil {
		log.Println("err:", err)
		// Nếu không tồn tại và là giao dịch trừ điểm -> lỗi
		if req.GetPoints() < 0 {
			ss.Rollback()
			return errors.New(utils.E_not_enough_points)
		}

		// Nếu là giao dịch cộng điểm thì tạo mới
		up = &pb.UserPoint{
			UserId:      req.ReceiverId,
			Points:      req.GetPoints(),
			OldPoints:   0,
			TotalPoints: req.GetPoints(),
			CreatedAt:   now,
			UpdateAt:    now,
		}
		if _, err := ss.Insert(up); err != nil {
			ss.Rollback()
			return err
		}
	} else {
		log.Println("up:", up)
		log.Println("req.GetPoints():", req.GetPoints())
		log.Println("up.Points:", up.Points)
		// Nếu là giao dịch trừ điểm, kiểm tra đủ điểm không
		if req.GetPoints() < 0 && up.Points < -req.GetPoints() {
			ss.Rollback()
			return errors.New(utils.E_not_enough_points)
		}

		// Cập nhật điểm
		up.OldPoints = up.Points
		up.Points += req.GetPoints()

		// Chỉ cộng vào TotalPoints nếu là giao dịch cộng điểm
		if req.GetPoints() > 0 {
			up.TotalPoints += req.GetPoints()
		}
		up.UpdateAt = now
		if _, err := ss.Where("user_id = ?", req.ReceiverId).Update(up); err != nil {
			ss.Rollback()
			return err
		}
	}

	// Tạo giao dịch point exchange
	if _, err := ss.Insert(req); err != nil {
		ss.Rollback()
		return err
	}

	if err := ss.Commit(); err != nil {
		return err
	}
	return nil
}

func (d *DB) CreateStore(store *pb.Store) error {
	c, err := d.engine.Insert(store)
	if err != nil {
		return err
	}
	if c == 0 {
		return errors.New(utils.E_can_not_insert)
	}
	return nil
}

func (d *DB) UpdateStore(updator, selector *pb.Store) error {
	_, err := d.engine.Update(updator, selector)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) DeleteStore(id string) error {
	if id == "" {
		return errors.New(utils.E_not_found_id)
	}
	_, err := d.engine.Table(tblStore).Delete(&pb.Store{Id: id})
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) CountStore(rq *pb.StoreRequest) (int64, error) {
	ss := d.listStoreQuery(rq)
	return ss.Count()
}

func (d *DB) listStoreQuery(rq *pb.StoreRequest) *xorm.Session {
	ss := d.engine.Table(tblStore)
	if len(rq.GetIds()) > 0 {
		ss.In("id", rq.GetIds())
	} else if rq.GetId() != "" {
		ss.And("id = ?", rq.GetId())
	}
	if rq.GetName() != "" {
		ss.And("name like ?", "%"+rq.GetName()+"%")
	}
	if rq.GetPhoneNumber() != "" {
		ss.And("phone_number = ?", rq.GetPhoneNumber())
	}
	if rq.GetState() != "" {
		ss.And("state = ?", rq.GetState())
	}
	if rq.GetProvince() != "" {
		ss.And("province = ?", rq.GetProvince())
	}
	if rq.GetDistrict() != "" {
		ss.And("district = ?", rq.GetDistrict())
	}
	if rq.GetWard() != "" {
		ss.And("ward = ?", rq.GetWard())
	}
	if rq.GetAddress() != "" {
		ss.And("address like ?", "%"+rq.GetAddress()+"%")
	}
	if len(rq.GetPartnerIds()) > 0 {
		ss.In("partner_id", rq.GetPartnerIds())
	} else if rq.GetPartnerId() != "" {
		ss.And("partner_id = ?", rq.GetPartnerId())
	}
	if rq.GetSlug() != "" {
		ss.And("slug = ?", rq.GetSlug())
	}
	return ss
}

func (d *DB) ListStore(rq *pb.StoreRequest) ([]*pb.Store, error) {
	ss := d.listStoreQuery(rq)
	if rq.GetLimit() != 0 {
		ss.Limit(int(rq.GetLimit()), int(rq.GetSkip()*rq.GetLimit()))
	}
	stores := make([]*pb.Store, 0)
	err := ss.Desc("created_at").Find(&stores)
	if err != nil {
		return nil, err
	}
	return stores, nil
}

func (d *DB) GetStore(rq *pb.StoreRequest) (*pb.Store, error) {
	store := &pb.Store{
		Id:   rq.GetId(),
		Slug: rq.GetSlug(),
	}
	ishas, err := d.engine.Get(store)
	if err != nil {
		return nil, err
	}
	if !ishas {
		return nil, errors.New(utils.E_not_found)
	}
	return store, nil
}

func (d *DB) IsStoreExisted(u *pb.Store) bool {
	any, err := d.engine.Exist(u)
	if err != nil {
		return false
	}
	return any
}

func (d *DB) CreatePartner(partner *pb.Partner) error {
	c, err := d.engine.Insert(partner)
	if err != nil {
		return err
	}
	if c == 0 {
		return errors.New(utils.E_can_not_insert)
	}
	return nil
}

func (d *DB) UpdatePartner(updator, selector *pb.Partner) error {
	_, err := d.engine.Update(updator, selector)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) DeletePartner(id string) error {
	if id == "" {
		return errors.New(utils.E_not_found_id)
	}
	_, err := d.engine.Table(tblPartner).Delete(&pb.Partner{Id: id})
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) CountPartner(rq *pb.PartnerRequest) (int64, error) {
	ss := d.listPartnerQuery(rq)
	return ss.Count()
}

func (d *DB) listPartnerQuery(rq *pb.PartnerRequest) *xorm.Session {
	ss := d.engine.Table(tblPartner)
	if len(rq.GetIds()) > 0 {
		ss.In("id", rq.GetIds())
	} else if rq.GetId() != "" {
		ss.And("id = ?", rq.GetId())
	}
	if rq.GetName() != "" {
		ss.And("name like ?", "%"+rq.GetName()+"%")
	}
	if rq.GetType() != "" {
		ss.And("type = ?", rq.GetType())
	}
	if rq.GetState() != "" {
		ss.And("state = ?", rq.GetState())
	}
	if rq.GetFrom() > 0 {
		ss.And("created_at >= ?", rq.GetFrom())
	}
	if rq.GetTo() > 0 {
		ss.And("created_at <= ?", rq.GetTo())
	}
	return ss
}

func (d *DB) ListPartner(rq *pb.PartnerRequest) ([]*pb.Partner, error) {
	ss := d.listPartnerQuery(rq)
	if rq.GetLimit() != 0 {
		ss.Limit(int(rq.GetLimit()), int(rq.GetSkip()*rq.GetLimit()))
	}
	partners := make([]*pb.Partner, 0)
	err := ss.Desc("created_at").Find(&partners)
	if err != nil {
		return nil, err
	}
	return partners, nil
}

func (d *DB) GetPartner(rq *pb.PartnerRequest) (*pb.Partner, error) {
	partner := &pb.Partner{
		Id: rq.GetId(),
	}
	ishas, err := d.engine.Get(partner)
	if err != nil {
		return nil, err
	}
	if !ishas {
		return nil, errors.New(utils.E_not_found)
	}
	return partner, nil
}

func (d *DB) IsPartnerExisted(u *pb.Partner) bool {
	any, err := d.engine.Exist(u)
	if err != nil {
		return false
	}
	return any
}

func (d *DB) CreatePlan(plan *pb.Plan) error {
	c, err := d.engine.Insert(plan)
	if err != nil {
		return err
	}
	if c == 0 {
		return errors.New(utils.E_can_not_insert)
	}
	return nil
}

func (d *DB) UpdatePlan(updator, selector *pb.Plan) error {
	c, err := d.engine.Update(updator, selector)
	if err != nil {
		return err
	}
	if c == 0 {
		log.Println("can_not_update")
	}
	return nil
}

func (d *DB) DeletePlan(id string) error {
	if id == "" {
		return errors.New(utils.E_not_found_id)
	}
	_, err := d.engine.Table(tblPlan).Delete(&pb.Plan{Id: id})
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) CountPlan(rq *pb.PlansRequest) (int64, error) {
	ss := d.listPlansQuery(rq)
	return ss.Count()
}

func (d *DB) listPlansQuery(rq *pb.PlansRequest) *xorm.Session {
	ss := d.engine.Table(tblPlan)
	if len(rq.GetIds()) > 0 {
		ss.In("id", rq.GetIds())
	} else if rq.GetId() != "" {
		ss.And("id = ?", rq.GetId())
	}
	if rq.GetName() != "" {
		ss.And("name like ?", "%"+rq.GetName()+"%")
	}
	if rq.GetState() != "" {
		ss.And("state = ?", rq.GetState())
	}
	return ss
}

func (d *DB) ListPlans(rq *pb.PlansRequest) ([]*pb.Plan, error) {
	ss := d.listPlansQuery(rq)
	if rq.GetLimit() != 0 {
		ss.Limit(int(rq.GetLimit()), int(rq.GetSkip()*rq.GetLimit()))
	}
	plans := make([]*pb.Plan, 0)
	err := ss.Asc("created_at").Find(&plans)
	if err != nil {
		return nil, err
	}
	return plans, nil
}

func (d *DB) GetPlan(rq *pb.PlansRequest) (*pb.Plan, error) {
	plan := &pb.Plan{
		Id: rq.GetId(),
	}
	_, err := d.engine.Get(plan)
	if err != nil {
		return nil, err
	}
	return plan, nil
}

func (d *DB) IsPlansExisted(u *pb.Plan) bool {
	any, err := d.engine.Exist(u)
	if err != nil {
		return false
	}
	return any
}

func (d *DB) CreateOrderPlan(req *pb.OrderPlan) error {
	c, err := d.engine.Insert(req)
	if err != nil {
		return err
	}
	if c == 0 {
		return errors.New(utils.E_can_not_insert)
	}
	return nil
}

func (d *DB) GetOrderPlan(req *pb.OrderPlan) (*pb.OrderPlan, error) {
	ishas, err := d.engine.Get(req)
	if err != nil {
		return nil, err
	}
	if !ishas {
		return nil, errors.New(utils.E_not_found)
	}
	return req, nil
}

func (d *DB) listOrderPlanQuery(rq *pb.OrderPlanRequest) *xorm.Session {
	ss := d.engine.Table(tblOrderPlan)
	if len(rq.GetIds()) > 0 {
		ss.In("id", rq.GetIds())
	} else if rq.Id != "" {
		ss.And("id = ?", rq.GetId())
	}
	if rq.GetUserId() != "" {
		ss.And("user_id = ?", rq.GetUserId())
	}
	if rq.GetPlanId() != "" {
		ss.And("plan_id = ?", rq.GetPlanId())
	}
	if rq.GetStartDate() != 0 {
		ss.And("start_date >= ?", rq.GetStartDate())
	}
	if rq.GetEndDate() != 0 {
		ss.And("end_date >= ?", rq.GetEndDate())
	}
	if rq.GetExpiredAt() != 0 {
		ss.And("exprire_at = ?", rq.GetExpiredAt())
	}
	return ss
}

func (d *DB) ListOrderPlan(rq *pb.OrderPlanRequest) ([]*pb.OrderPlan, error) {
	ss := d.listOrderPlanQuery(rq)
	if rq.GetLimit() != 0 {
		ss.Limit(int(rq.GetLimit()), int(rq.GetSkip()*rq.GetLimit()))
	}
	orderPlans := make([]*pb.OrderPlan, 0)
	err := ss.Desc("created_at").Find(&orderPlans)
	if err != nil {
		return nil, err
	}
	return orderPlans, nil
}

func (d *DB) CountOrderPlan(rq *pb.OrderPlanRequest) (int64, error) {
	ss := d.listOrderPlanQuery(rq)
	return ss.Count()
}

func (d *DB) IsOrderPlanExisted(u *pb.OrderPlan) bool {
	any, err := d.engine.Exist(u)
	if err != nil {
		return false
	}
	return any
}

func (d *DB) UpdateOrderPlan(updator, selector *pb.OrderPlan) error {
	c, err := d.engine.Update(updator, selector)
	if err != nil {
		return err
	}
	if c == 0 {
		log.Println("can_not_update")
	}
	return nil
}

func (d *DB) DeleteOrderPlan(id string) error {
	if id == "" {
		return errors.New(utils.E_not_found_id)
	}
	_, err := d.engine.Table(tblOrderPlan).Delete(&pb.OrderPlan{Id: id})
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) IsExistOrderPlan(id string) (bool, error) {
	ss := d.engine.Where("id = ?", id).Table(tblOrderPlan)
	return ss.Exist()
}

func (d *DB) CreateUserAddress(req *pb.UserAddress) error {
	c, err := d.engine.Table(tblUserAddress).Insert(req)
	if err != nil {
		return err
	}
	if c == 0 {
		return errors.New(utils.E_can_not_insert)
	}
	return nil
}

func (d *DB) UpdateUserAddress(updated, selector *pb.UserAddress) error {
	if updated.IsDefault == "true" {
		userAddress := &pb.UserAddress{UserId: selector.UserId, IsDefault: "true"}
		_, err := d.engine.Get(userAddress)
		if err != nil {
			return err
		}
		c, err := d.engine.Update(&pb.UserAddress{IsDefault: "false"}, &pb.UserAddress{Id: userAddress.Id})
		if err != nil {
			return err
		}
		if c == 0 {
			return errors.New(utils.E_can_not_update_user_address)
		}
	}
	c, err := d.engine.Update(updated, selector)
	if err != nil {
		return err
	}
	if c == 0 {
		return errors.New(utils.E_can_not_update_user_address)
	}
	return nil
}

func (d *DB) DeleteUserAddress(id string) error {
	if id == "" {
		return errors.New(utils.E_not_found_id)
	}
	_, err := d.engine.Table(tblUserAddress).Delete(&pb.UserAddress{Id: id})
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) IsExistUserAddress(id string) (bool, error) {
	ss := d.engine.Where("id = ?", id).Table(tblUserAddress)
	return ss.Exist()
}

func (d *DB) listUserAddressQuery(rq *pb.UserAddressRequest) *xorm.Session {
	ss := d.engine.Table(tblUserAddress)
	if rq.GetId() != "" {
		ss.And("id = ?", rq.GetId())
	}
	if rq.GetUserId() != "" {
		ss.And("user_id = ?", rq.GetUserId())
	}
	if rq.GetProvince() != "" {
		ss.And("province = ?", rq.GetProvince())
	}
	if rq.GetDistrict() != "" {
		ss.And("district = ?", rq.GetDistrict())
	}
	if rq.GetWard() != "" {
		ss.And("ward = ?", rq.GetWard())
	}
	if rq.GetAddress() != "" {
		ss.And("address = ?", rq.GetAddress())
	}
	if rq.GetIsDefault() != "" {
		ss.And("is_default = ?", rq.GetIsDefault())
	}
	return ss
}

func (d *DB) ListUserAddress(rq *pb.UserAddressRequest) ([]*pb.UserAddress, error) {
	ss := d.listUserAddressQuery(rq)
	if rq.GetLimit() != 0 {
		ss.Limit(int(rq.GetLimit()), int(rq.GetSkip()*rq.GetLimit()))
	}
	userAddresses := make([]*pb.UserAddress, 0)
	err := ss.OrderBy("is_default DESC, created_at DESC").Find(&userAddresses)
	if err != nil {
		return nil, err
	}
	return userAddresses, nil
}

func (d *DB) GetUserAddress(rq *pb.UserAddress) (*pb.UserAddress, error) {
	userAddress := &pb.UserAddress{
		Id: rq.GetId(),
	}
	ishas, err := d.engine.Get(userAddress)
	if err != nil {
		return nil, err
	}
	if !ishas {
		return nil, errors.New(utils.E_not_found_user_address)
	}
	return userAddress, nil
}

func (d *DB) TranCreateUserAddress(req *pb.UserAddress, maxUserAddress int) error {
	ss := d.engine.NewSession()
	defer ss.Close()
	if err := ss.Begin(); err != nil {
		return err
	}
	listAddress, err := d.ListUserAddress(&pb.UserAddressRequest{UserId: req.GetUserId()})
	if err != nil {
		ss.Rollback()
		return err
	}
	if len(listAddress) >= maxUserAddress {
		ss.Rollback()
		return errors.New(utils.E_max_user_address)
	}
	if len(listAddress) <= 0 {
		req.IsDefault = "true"
	} else {
		req.IsDefault = "false"
	}
	req.CreatedAt = time.Now().Unix()
	if err := d.CreateUserAddress(req); err != nil {
		ss.Rollback()
		return err
	}
	if err := ss.Commit(); err != nil {
		return err
	}
	return nil
}
