package db

import (
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
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
