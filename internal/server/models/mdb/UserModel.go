package mdb

import (
	"errors"
	"gameserver/internal/server/gameconst"

	"gorm.io/gorm"
)

// 用户表
type UserModel struct {
	PlayerID     uint64 `gorm:"primary_key"`                     // 用户游戏自增id
	PlayerName   string `gorm:"column:player_name"`              // 玩家名字
	RoleName     string `gorm:"column:role_name"`                // 玩家角色名称
	PlatformID   uint32 `gorm:"column:platform_id"`              // 平台id 1.IOS 2.Bili 3.CPS 4.UO 5.WEB
	ChannelID    uint32 `gorm:"column:channel_id"`               // 渠道id
	AccountName  string `gorm:"column:account_name"`             // 通行证名称appname
	AccountUID   string `gorm:"column:account_uid"`              // 通行证uuid
	Sex          uint32 `gorm:"column:sex"`                      // 性别 1：男 2：女 0:其他
	Level        uint32 `gorm:"column:level"`                    // 等级
	FirstPayTime int64  `gorm:"column:first_pay_time"`           // 首冲时间
	RegTime      int64  `gorm:"column:reg_time;autoCreateTime"`  // 注册时间
	LastTime     int64  `gorm:"column:last_time;autoUpdateTime"` // 最后登陆时间
	PayTotal     uint64 `gorm:"column:pay_total"`                // 充值总值(单位分)
	ServerID     uint32 `gorm:"column:server_id"`                // 区服id
	RegDeviceID  string `gorm:"column:reg_device_id"`            // 注册设备id
	LastDeviceID string `gorm:"column:last_device_id"`           // 最后登陆设备ID
	IsThaw       uint32 `gorm:"column:is_thaw"`                  // 是否冻结：0 冻结，1 未冻结
	ThawAt       int64  `gorm:"column:thaw_at"`                  // 冻结时间
}

func NewUserModel() *UserModel {
	userModel := &UserModel{}
	return userModel
}
func (u *UserModel) BeforeCreate(tx *gorm.DB) (err error) {

	return
}

func (u *UserModel) AfterCreate(tx *gorm.DB) (err error) {

	return
}

// 设置 `User` 的表名为 `user_all`
func (u *UserModel) TableName() string {
	return gameconst.TBLMDBUserAll
}

func (u *UserModel) Insert(tx *gorm.DB) (PlayerID uint64, err error) {
	result := tx.Create(u)
	if result.Error != nil {
		return 0, err
	}
	// user := User{Name: "Jinzhu", Age: 18, Birthday: time.Now()}
	// result := db.Create(&user) // 通过数据的指针来创建
	// user.ID             // 返回插入数据的主键
	// result.Error        // 返回 error
	// result.RowsAffected // 返回插入记录的条数
	return u.PlayerID, nil
}

func (u *UserModel) GetPlayerID(userID string, serverID uint32) (*UserModel, error) {

	err := MDB.Select("player_id").Where("account_uid = ?", userID).Where("server_id = ?", serverID).First(u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return u, err
}

func (u *UserModel) DetailByUUID(userID string, serverID uint32) (*UserModel, error) {

	err := MDB.Where("account_uid = ?", userID).Where("server_id = ?", serverID).First(u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return u, err
}

func (u *UserModel) Detail(playerID uint64) (*UserModel, error) {
	// 锁住指定 player_id 的 User 记录
	// if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&u, player_id).Error; err != nil {
	//  tx.Rollback()
	// 	return u, err
	// }
	err := MDB.Where("player_id = ?", playerID).First(u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return u, err
}

// 使用 `struct` 更新多个属性，只会更新那些被修改了的和非空的字段
func (u *UserModel) Update(playerID uint64) error {

	if err := MDB.Model(u).Where("player_id = ?", playerID).Updates(u).Error; err != nil {
		return err
	}
	return nil
}

// Save 方法在执行 SQL 更新操作时将包含所有字段，即使这些字段没有被修改
func (u *UserModel) Save(playerID uint64) error {
	if err := MDB.Debug().Where("player_id = ?", playerID).Save(u).Error; err != nil {
		return err
	}
	return nil
}

func (u *UserModel) UpdateDryRun(playerID uint64) string {
	stmt := MDB.Session(&gorm.Session{DryRun: true}).Model(u).Where("player_id = ?", playerID).Updates(u).Statement
	return MDB.Dialector.Explain(stmt.SQL.String(), stmt.Vars...)
}

// ForUpdateLock ...
func (u *UserModel) ForUpdateLock(playerID uint64) error {
	// 创建事务
	tx := MDB.Begin()

	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Set("gorm:query_option", "FOR UPDATE").Select("player_id").First(&u, playerID).Error; err != nil {
		tx.Rollback()
		return err
	}
	// 此时指定 id 的记录被锁住.如果表中无符合记录的数据,则排他锁不生效
	// 执行其他数据库操作
	if err := tx.Model(u).Where("player_id = ?", playerID).Updates(UserModel{LastTime: u.LastTime}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
