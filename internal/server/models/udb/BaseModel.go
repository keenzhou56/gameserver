package udb

import (
	"container/heap"
	"errors"
	"gameserver/internal/server/conf"
	"gameserver/internal/server/gameconst"
	"gameserver/internal/server/models/mdb"
	"gameserver/internal/server/mysql"
	g "gameserver/pkg/go"
	"gameserver/pkg/timer"
	"strconv"

	"time"

	"gorm.io/gorm"
)

// 玩家player_id
type BaseModel struct {
	PlayerID uint64 `gorm:"column:player_id;default:0"`
}

func (bm *BaseModel) getUDB() *gorm.DB {
	dbSuffix, _ := GetSuffixByUid(bm.PlayerID)
	udb := mysql.GetDBUser(dbSuffix)
	return udb
}

func (bm *BaseModel) getTableSuffix() string {
	_, tableSuffix := GetSuffixByUid(bm.PlayerID)
	return tableSuffix
}

// 玩家player_id
type PrimaryModel struct {
	PlayerID uint64 `gorm:"column:player_id;primary_key"`
}

func (pm *PrimaryModel) getUDB() *gorm.DB {
	dbSuffix, _ := GetSuffixByUid(pm.PlayerID)
	udb := mysql.GetDBUser(dbSuffix)
	return udb
}

func (pm *PrimaryModel) getTableSuffix() string {
	_, tableSuffix := GetSuffixByUid(pm.PlayerID)
	return tableSuffix
}

type UserData struct {
	mPlayer           *mdb.UserModel
	uPlayer           *PlayerModel
	uItem             map[uint32]*ItemModel
	udb               *gorm.DB
	dbSuffix          int
	tableSuffix       string
	timers            []*TimerEntry
	TaskTimerEntryMap map[string]*TimerEntry
	UserCron          *timer.Cron
	// 每个请求协议中，涉及需要修改的表，协议处理完后，触发处理此切片进行数据入库
	UpdateTableList []map[string]interface{}
}

func NewPlayerData(playerID uint64) *UserData {
	userData := &UserData{}
	return userData
}

func (u *UserData) Render() (err error) {

	//不涉及空值(如：0、false、"")的更新，可以使用结构体或者map
	u.uPlayer.Level = 20
	u.uPlayer.LastAcTime = time.Now().Unix()
	opMapPlayer := map[string]interface{}{
		"opType": gameconst.OperationDBUpdateType,
		"opObj":  u.uPlayer,
		"opData": map[string]interface{}{"level": u.mPlayer.Level, "last_ac_time": u.uPlayer.LastAcTime},
		//"opData": u.PlayerData,
	}
	u.UpdateTableList = append(u.UpdateTableList, opMapPlayer)

	// 新增插入
	u.uItem = make(map[uint32]*ItemModel)
	u.uItem[1] = &ItemModel{
		PrimaryModel: PrimaryModel{PlayerID: u.mPlayer.PlayerID},
		ItemId:       1,
		UpdateAt:     time.Now().Unix(),
	}

	opMapItem1 := map[string]interface{}{
		"opType": gameconst.OperationDBUpdateType,
		"opObj":  u.uItem[1],
	}
	u.UpdateTableList = append(u.UpdateTableList, opMapItem1)

	// u.uItem[2] = &ItemModel{
	// 	PrimaryModel: PrimaryModel{PlayerID: u.mPlayer.PlayerID},
	// 	ItemId:       2,
	// 	CreateAt:     time.Now().Unix(),
	// }

	// opMapItem2 := map[string]interface{}{
	// 	"opType": gameconst.OperationDBInsertType,
	// 	"opObj":  u.uItem[2],
	// }
	// u.UpdateTableList = append(u.UpdateTableList, opMapItem2)

	err = u.handlerUpdateTable()
	return
}

func (u *UserData) handlerUpdateTable() (err error) {
	if len(u.UpdateTableList) < 1 {
		return nil
	}

	d := g.New(1)
	d.Go(func() {
		tx := u.udb.Begin()
		// 处理异常
		defer func() {
			if r := recover(); r != nil {
				err = tx.Rollback().Error
				return
			}
		}()

		if err := tx.Error; err != nil {
			return
		}

		for _, opMap := range u.UpdateTableList {
			switch opMap["opType"] {
			case gameconst.OperationDBUpdateType:
				err = tx.Model(opMap["opObj"]).UpdateColumns(opMap["opData"]).Error
			case gameconst.OperationDBInsertType:
				err = tx.Create(opMap["opObj"]).Error
			default:
				tx.Rollback()
				return
			}
			if err != nil {
				tx.Rollback()
				return
			}
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return
		}
	}, func() {
		// 清空更新表切片
		u.UpdateTableList = u.UpdateTableList[:0]
	})

	d.Cb(<-d.ChanCb)

	return nil
}

type TimerEntry struct {
	runTime  time.Time    // 到期时间
	callback CallbackFunc // 回调方法
	taskId   string       // 业务ID
}

// 初始化用户连接
func (u *UserData) InitUserData(playerID uint64) (err error) {
	dbSuffix, tableSuffix := GetSuffixByUid(playerID)
	udb := mysql.GetDBUser(dbSuffix)
	u.udb = udb
	u.dbSuffix = dbSuffix
	u.tableSuffix = tableSuffix

	u.mPlayer = mdb.NewUserModel()
	u.mPlayer.Detail(playerID)

	u.uPlayer = NewPlayerModel(playerID)
	u.uPlayer = u.uPlayer.Detail()
	if u.uPlayer == nil {
		err = errors.New("not found user")
		return
	}
	u.uItem, err = NewItemModel(playerID).GetAllData()
	if err == nil {
		heap.Init(u)
	}
	return
}

// 注册用户数据
func RegUserData(playerID uint64, serverID uint32) error {

	userData := new(UserData)

	dbSuffix, tableSuffix := GetSuffixByUid(playerID)
	userData.udb = mysql.GetDBUser(dbSuffix)
	userData.dbSuffix = dbSuffix
	userData.tableSuffix = tableSuffix

	heap.Init(userData)

	userData.TaskTimerEntryMap = make(map[string]*TimerEntry)
	userData.mPlayer = mdb.NewUserModel()
	userData.mPlayer.Detail(playerID)

	userData.uPlayer = NewPlayerModel(playerID)
	userData.uPlayer.ChannelID = 1
	userData.uPlayer.PlatformID = 1
	userData.uPlayer.ServerID = serverID

	userData.uItem = make(map[uint32]*ItemModel)
	item := NewItemModel(playerID)
	item.ItemId = 1
	item.ItemType = 1
	userData.uItem[uint32(item.ItemId)] = item

	tx := userData.udb.Begin()

	err := tx.Create(userData.uPlayer).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Create(item).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

// 根据player_id获取玩家库和表的后缀
func GetSuffixByUid(playerID uint64) (int, string) {
	dbSuffix := "0"
	tableSuffix := "00"
	if conf.DB10TABLE100 {
		if playerID >= 10 {
			uidStr := strconv.FormatUint(playerID, 10)
			dbSuffix = uidStr[len(uidStr)-1:]
			tableSuffix = uidStr[len(uidStr)-2:]
		}
	}
	dbSuffixInt, _ := strconv.Atoi(dbSuffix)
	return dbSuffixInt, tableSuffix
}
