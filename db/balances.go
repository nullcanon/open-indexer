
package db

import (
	"errors"

	"github.com/jinzhu/gorm"
)

// 资产表
// ticks address amount

type UserBalances struct {
	Id          int64 `gorm:"type:int(11) UNSIGNED AUTO_INCREMENT;primary_key" json:"id"`
	Ticks		string `gorm:"column:ticks"`
	Address		string `gorm:"column:address"`
	Amount		string `gorm:"column:amount; default:'0'"`
}

func (u UserBalances) CreateUserBalances(userinfo UserBalances) error {
	return db.Create(&userinfo).Error
}

func (u UserBalances) Update(args map[string]interface{}) error {
	var userinfo UserBalances
	result := db.First(&userinfo, "self = ?", u.Self)

	if result.Error == nil {
		db.Model(&UserTable{}).Where("self = ?", u.Self).Update(args)
	}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return u.CreateUser(u)
	} else {
		return result.Error
	}
}