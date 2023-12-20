// 转账记录表
// ticks status from to amount time hash
package db

import (
	"errors"

	"github.com/jinzhu/gorm"
)


type TradeHistory struct {
	Ticks		string `gorm:"column:ticks"`
	Status		string `gorm:"column:status"`
	From		string `gorm:"column:from"`
	To			string `gorm:"column:to"`
	Hash		string `gorm:"column:hash;primary_key"`
	Time		uint64 `gorm:"column:time"`
}


func (u TradeHistory) CreateTradeHistory(tradeHistory TradeHistory) error {
	return db.Create(&tradeHistory).Error
}

func (u TradeHistory) Update(args map[string]interface{}) error {
	var tradeHistory TradeHistory
	result := db.First(&tradeHistory, "hash = ?", u.Hash)

	if result.Error == nil {
		db.Model(&TradeHistory{}).Where("hash = ?", u.Hash).Update(args)
	}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return u.CreateTradeHistory(u)
	} else {
		return result.Error
	}
	return nil
}