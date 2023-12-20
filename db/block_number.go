package db

import (
	"errors"

	"github.com/jinzhu/gorm"
)

// 扫链配置表
// blockNumber

type BlockScan struct {
	Id          int64 `gorm:"type:int(11) UNSIGNED AUTO_INCREMENT;primary_key" json:"id"`
	BlockNumber int64 `gorm:"type:int(64) UNSIGNED not null COMMENT '同步的区块高度'" json:"block_number"`
}


func (b BlockScan) Create(blockScan BlockScan) error {
	return db.Create(&blockScan).Error
}
func (b *BlockScan) GetNumber() int64 {
	var bscScan BlockScan
	err := db.Order("id desc").First(&bscScan).Error
	if err != nil {
		return 0
	}
	return bscScan.BlockNumber
}

func (b *BlockScan) Edit(data map[string]interface{}) error {
	return db.Model(&b).Updates(data).Error
}

func (b *BlockScan) UptadeBlockNumber(blockNumber uint64) error {
	var blockscan BlockScan
	result := db.First(&blockscan, "id = ?", 1)

	if result.Error == nil {
		db.Model(&BlockScan{}).Where("id = ?", 1).Update(map[string]interface{}{"block_number": blockNumber})
	}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return db.Create(b).Error
	} else {
		return result.Error
	}
	return nil
}
