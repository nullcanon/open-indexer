// 转账记录表
// ticks status from to amount time hash



type TradeHistory struct {
	Id          int64 `gorm:"type:int(11) UNSIGNED AUTO_INCREMENT;primary_key" json:"id"`
	Ticks		string `gorm:"column:ticks"`
	Status		string `gorm:"column:status"`
	From		string `gorm:"column:from"`
	To			string `gorm:"column:to"`
	Hash		string `gorm:"column:hash"`
	Time		int64 `gorm:"column:time"`
}


func (u TradeHistory) CreateTradeHistory(tradeHistory TradeHistory) error {
	return db.Create(&tradeHistory).Error
}

func (u TradeHistory) Update(args map[string]interface{}) error {
	var tradeHistory TradeHistory
	result := db.First(&tradeHistory, "self = ?", u.Self)

	if result.Error == nil {
		db.Model(&TradeHistory{}).Where("self = ?", u.Self).Update(args)
	}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return u.CreateTradeHistory(u)
	} else {
		return result.Error
	}
}