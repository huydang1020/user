package db

import (
	"log"

	pb "github.com/huyshop/header/user"

	"xorm.io/xorm"
)

const (
	tblUser                   = "user"
	tblUserPoint              = "user_point"
	tblPointExchange          = "point_exchange"
	tblStore                  = "store"
	tblPartner                = "partner"
	tblPlan                   = "plan"
	tblOrderPlan              = "order_plan"
	tblUserAddress            = "user_address"
)

func createTable(model interface{}, tblName string, engine *xorm.Engine) error {
	b, err := engine.IsTableExist(model)
	if err != nil {
		return err
	}
	log.Print(b, " ", tblName)
	if b {
		if err = engine.Sync2(model); err != nil {
			return err
		}
		return nil
	}
	if !b {
		if err := engine.CreateTables(model); err != nil {
			log.Print(err)
			return err
		}
	}
	return nil
}

func (d *DB) CreateDb() error {
	if err := createTable(&pb.User{}, tblUser, d.engine); err != nil {
		return err
	}
	if err := createTable(&pb.Store{}, tblStore, d.engine); err != nil {
		return err
	}
	if err := createTable(&pb.Partner{}, tblPartner, d.engine); err != nil {
		return err
	}
	if err := createTable(&pb.UserPoint{}, tblUserPoint, d.engine); err != nil {
		return err
	}
	if err := createTable(&pb.PointExchange{}, tblPointExchange, d.engine); err != nil {
		return err
	}
	if err := createTable(&pb.Plan{}, tblPlan, d.engine); err != nil {
		return err
	}
	if err := createTable(&pb.OrderPlan{}, tblOrderPlan, d.engine); err != nil {
		return err
	}
	if err := createTable(&pb.UserAddress{}, tblUserAddress, d.engine); err != nil {
		return err
	}
	return nil
}
