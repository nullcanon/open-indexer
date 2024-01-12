package db

// push失败的消息
// ticks address amount

type RocketMsg struct {
	Hash    string `gorm:"column:tx_hash;primary_key"`
	Message string `gorm:"column:message"`
}

func (u RocketMsg) CreateRocketMsg(userinfo RocketMsg) error {
	return db.Create(&userinfo).Error
}

func (u RocketMsg) FetchRocketMsg(userbalance *[]RocketMsg) {
	db.Find(&userbalance)
}

func (u RocketMsg) DelMsg(msg RocketMsg) error {
	return db.Where("tx_hash = ?", msg.Hash).Delete(&msg).Error
}
