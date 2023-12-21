package db

import (
	"errors"

	"github.com/jinzhu/gorm"
)

// 资产表
// ticks address amount

type UserBalances struct {
	Ticks   string `gorm:"column:ticks;primary_key"`
	Address string `gorm:"column:address;primary_key;index"`
	Amount  string `gorm:"column:amount; default:'0'"`
}

func (u UserBalances) CreateUserBalances(userinfo UserBalances) error {
	return db.Create(&userinfo).Error
}

func (u UserBalances) Update(args map[string]interface{}) error {
	var userinfo UserBalances
	result := db.First(&userinfo, "ticks = ? and address = ?", u.Ticks, u.Address)

	if result.Error == nil {
		db.Model(&UserBalances{}).Where("ticks = ? and address = ?", u.Ticks, u.Address).Update(args)
	}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return u.CreateUserBalances(u)
	} else {
		return result.Error
	}
	return nil
}

func (u UserBalances) FetchUserBalances(userbalance *[]UserBalances) {
	db.Find(&userbalance)
}
