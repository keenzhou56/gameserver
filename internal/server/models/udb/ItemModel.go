package udb

import (
	"errors"
	"gameserver/internal/server/gameconst"

	"gorm.io/gorm"
)

type ItemModel struct {
	PrimaryModel
	ItemId   uint32 `gorm:"column:item_id;primary_key"`      // 物品id
	ItemType uint32 `gorm:"column:item_type;default:0"`      // 物品类别
	Num      uint32 `gorm:"column:num;default:0"`            // 物品数量
	UpdateAt int64  `gorm:"column:update_at;autoUpdateTime"` // 更新时间
	CreateAt int64  `gorm:"column:create_at;autoCreateTime"` // 创建时间
}

func NewItemModel(playerID uint64) *ItemModel {
	itemModel := &ItemModel{}
	itemModel.PlayerID = playerID
	return itemModel
}

// 设置 `ItemModel` 的表名为 `udb_item_00`
func (bp *ItemModel) TableName() string {
	return gameconst.TBLUDBItem + gameconst.TableLastSuffixUnderline + bp.getTableSuffix()
}

func (bp *ItemModel) Insert() *ItemModel {
	if err := bp.getUDB().Create(bp).Error; err != nil {
		return nil
	}
	return bp
}

func (bp *ItemModel) Detail() *ItemModel {
	if errors.Is(bp.getUDB().Model(bp).Take(bp).Error, gorm.ErrRecordNotFound) {
		return nil
	}
	return bp
}

func (bp *ItemModel) Update() error {
	if err := bp.getUDB().Model(bp).Updates(bp).Error; err != nil {
		return err
	}
	return nil
}

func (bp *ItemModel) GetAllData() (uItem map[uint32]*ItemModel, err error) {
	uItem = map[uint32]*ItemModel{}
	data := make([]*ItemModel, 0)

	if err := bp.getUDB().Table(bp.TableName()).Find(&data, "player_id = ?", bp.PlayerID).Error; err != nil {
		return nil, err
	}

	for i := 0; i < len(data); i++ {
		uItem[data[i].ItemId] = data[i]
	}
	return uItem, nil
}
