package db

import (
	"database/sql/driver"
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB


const (
	sqlType = "mysql"
)

func Setup() {
	var err error
	// config.Global.MysqlInfo.User,
	// config.Global.MysqlInfo.Password,
	// config.Global.MysqlInfo.Host,
	// config.Global.MysqlInfo.Db

	user : = ""
	password : = ""
	host : = ""
	db_name : = ""
	db, err = gorm.Open(sqlType, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		user,
		password,
		host,
		db_nameb))

	if err != nil {
		log.Printf(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
			user,
			password,
			host,
			db_nameb))
		log.Fatalf("models.Setup  err: %v", err)
	}
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return defaultTableName
	}
	db.SingularTable(true)
	//db.Callback().Create().Replace("gorm:insert_option", updateTimeStampForCreateCallback)
	//db.Callback().Update().Replace("gorm:update_option", updateTimeStampForUpdateCallback)
	//db.Callback().Delete().Replace("gorm:delete", deleteCallback)
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	//db.LogMode(true)
	//创建表
	db.AutoMigrate(
		UserBalances{},
		BlockScan{},
		InscriptionInfo{},
		TradeHistory{},
	)
	//CreateTableByHash()
}