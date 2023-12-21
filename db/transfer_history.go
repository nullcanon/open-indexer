// 转账记录表
// ticks status from to amount time hash
package db

import (
	"errors"

	"github.com/jinzhu/gorm"
)

type TradeHistory struct {
	Id     int64  `gorm:"type:int(64) UNSIGNED AUTO_INCREMENT;primary_key" json:"id"`
	Ticks  string `gorm:"column:ticks"`
	Status string `gorm:"column:status"`
	From   string `gorm:"column:from_address"`
	To     string `gorm:"column:to_address"`
	Hash   string `gorm:"column:hash;index"`
	Amount string `gorm:"column:amount"`
	Time   uint64 `gorm:"column:time"`
	Number uint64 `gorm:"column:number"`
}

func (u TradeHistory) GetInscriptionNumber() uint64 {
	var history TradeHistory
	err := db.Order("number desc").First(&history).Error
	if err != nil {
		return 0
	}
	return history.Number
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
