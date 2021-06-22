package udb

import (
	"errors"
	"gameserver/internal/server/gameconst"

	"gorm.io/gorm"
)

// PlayerModel 玩家表
type PlayerModel struct {
	PrimaryModel
	PlatformID uint32 `gorm:"column:platform_id"`                 // 平台id 1.IOS 2.Bili 3.CPS 4.UO 5.WEB
	ChannelID  uint32 `gorm:"column:channel_id"`                  // 渠道id
	Level      uint32 `gorm:"column:level"`                       // 等级
	EventAt    int64  `gorm:"column:event_at"`                    // 下次事件更新时间
	LastAc     string `gorm:"column:last_ac"`                     // 最后一次行为
	LastAcTime int64  `gorm:"column:last_ac_time;autoUpdateTime"` // 最后一次执行时间
	ServerID   uint32 `gorm:"column:server_id"`                   // 区服id
	Regtime    int64  `gorm:"column:reg_time;autoCreateTime"`     // 注册时间

}

func NewPlayerModel(playerID uint64) *PlayerModel {
	playerModel := &PlayerModel{}
	playerModel.PlayerID = playerID
	return playerModel
}

// 设置 `PlayerModel` 的表名为 `udb_player`
func (p *PlayerModel) TableName() string {
	return gameconst.TBLUDBPlayer + gameconst.TableLastSuffixUnderline + p.getTableSuffix()
}

func (p *PlayerModel) Insert() error {
	if err := p.getUDB().Create(p).Error; err != nil {
		return err
	}
	return nil
}

func (p *PlayerModel) Detail() *PlayerModel {
	// Debug 可以输出完整的执行sql
	//if UDB.Debug().Model(p).Where("uid = ?", p.UID).First(p).RecordNotFound() {
	if errors.Is(p.getUDB().Model(p).First(p).Error, gorm.ErrRecordNotFound) {
		return nil
	}
	return p
}

// 使用 `struct` 更新多个属性，只会更新那些被修改了的和非空的字段
func (p *PlayerModel) Update() error {
	if err := p.getUDB().Model(p).Updates(p).Error; err != nil {
		return err
	}
	return nil
}

// Save 方法在执行 SQL 更新操作时将包含所有字段，即使这些字段没有被修改
func (p *PlayerModel) Save() error {
	if err := p.getUDB().Save(p).Error; err != nil {
		return err
	}
	return nil
}
