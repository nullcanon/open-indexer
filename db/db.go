package db

import (
	// "database/sql/driver"
	"fmt"
	"log"

	// "time"
	"open-indexer/config"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

const (
	sqlType = "mysql"
)

func Setup() {
	var err error

	db, err = gorm.Open(sqlType, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		config.Global.MysqlInfo.User,
		config.Global.MysqlInfo.Password,
		config.Global.MysqlInfo.Host,
		config.Global.MysqlInfo.Db))

	if err != nil {
		log.Printf(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
			config.Global.MysqlInfo.User,
			config.Global.MysqlInfo.Password,
			config.Global.MysqlInfo.Host,
			config.Global.MysqlInfo.Db))
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
		RocketMsg{},
	)
	//CreateTableByHash()
}
