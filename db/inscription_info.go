package db

import (
	"errors"

	"github.com/jinzhu/gorm"
)


// 资产信息表
// avas trxs: 21002, total: 21000000, minted: 21000000, holders: 443

type InscriptionInfo struct {
	Id          int64 `gorm:"type:int(11) UNSIGNED AUTO_INCREMENT;primary_key" json:"id"`
	Trxs		int64 `gorm:"column:trxs"`
	Total		int64 `gorm:"column:total"`
	Minted		int64 `gorm:"column:minted"`
	Holders		int64 `gorm:"column:holders"`
	Ticks		string `gorm:"column:ticks"`
}


func (u InscriptionInfo) CreateInscriptionInfo(inscriptionInfo InscriptionInfo) error {
	return db.Create(&inscriptionInfo).Error
}

func (u InscriptionInfo) Update(args map[string]interface{}) error {
	var inscriptionInfo InscriptionInfo
	result := db.First(&inscriptionInfo, "self = ?", u.Self)

	if result.Error == nil {
		db.Model(&InscriptionInfo{}).Where("self = ?", u.Self).Update(args)
	}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return u.CreateInscriptionInfo(u)
	} else {
		return result.Error
	}
}