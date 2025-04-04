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
		return nil, errors.New(utils.E_not_found)
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
		return nil, errors.New(utils.E_not_found)
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
		return nil, errors.New(utils.E_not_found)
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
		return nil, errors.New(utils.E_not_found)
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

func (d *DB) IsExistUserPoint(id string) bool {
	ss := d.engine.Where("id = ?", id).Table(tblUserPoint)
	any, err := ss.Exist()
	if err != nil {
		return false
	}
	return any
}

func (d *DB) listUserPointQuery(rq *pb.UserPointRequest) *xorm.Session {
	ss := d.engine.Table(tblUser)
	if len(rq.GetIds()) > 0 {
		ss.In("id", rq.GetIds())
	} else if rq.Id != "" {
		ss.And("id = ?", rq.GetId())
	}
	return ss
}

func (d *DB) GetUserPoint(rq *pb.UserPoint) (*pb.UserPoint, error) {
	ishas, err := d.engine.Get(rq)
	if err != nil {
		return nil, err
	}
	if !ishas {
		return nil, errors.New(utils.E_not_found)
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
	if d.IsUserExisted(&pb.User{Id: req.ReceiverId}) {
		ss.Rollback()
		return errors.New(utils.E_user_not_existed)
	}
	up, err := d.GetUserPoint(&pb.UserPoint{UserId: req.ReceiverId})
	if err != nil {
		ss.Rollback()
		return err
	}
	if req.GetPoints() != 0 {
		up.OldPoints = up.Points
		up.Points += req.GetPoints()
		if up.Points > 0 {
			up.TotalPoints += req.GetPoints()
		}
		up.UpdateAt = time.Now().Unix()
		if err := d.UpdateUserPoint(up, &pb.UserPoint{UserId: req.ReceiverId}); err != nil {
			ss.Rollback()
			return err
		}
	}
	if _, err := ss.Insert(req); err != nil {
		ss.Rollback()
		return err
	}
	if err := ss.Commit(); err != nil {
		return err
	}
	return nil
}
