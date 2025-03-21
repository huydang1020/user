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
		Username:    rq.GetUsername(),
		PhoneNumber: rq.GetPhoneNumber(),
		Email:       rq.GetEmail(),
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
