package service

import (
	"gameserver/internal/server/models/mdb"
	"gameserver/internal/server/models/udb"
	"gameserver/pkg/util"
	"time"
)

type UserService struct {
}

var (
	userService *UserService
)

func GetUserService() *UserService {
	mdb.Once.Do(func() {
		userService = &UserService{}
	})
	return userService
}

func (s *UserService) GetPlayerID(userID string, serverID uint32) (*mdb.UserModel, error) {
	userData, err := mdb.NewUserModel().GetPlayerID(userID, serverID)
	return userData, err
}

func (s *UserService) GetUserDetailByUserID(userID string, serverID uint32) *mdb.UserModel {
	userData, _ := mdb.NewUserModel().DetailByUUID(userID, serverID)
	return userData
}

func (s *UserService) GetUserDetailByPlayerID(playerID uint64) *mdb.UserModel {
	userData, _ := mdb.NewUserModel().Detail(playerID)
	return userData
}

func (*UserService) AutoUpdateUserData(playerID uint64) error {
	user := mdb.UserModel{
		LastTime: time.Now().Unix(),
	}

	return user.Update(playerID)
}

func (*UserService) UpdateDryRun(playerID uint64) string {
	user := mdb.UserModel{
		LastTime: time.Now().Unix(),
	}
	return user.UpdateDryRun(playerID)
}

func (*UserService) ForUpdateLock(playerID uint64) error {
	user := mdb.UserModel{
		LastTime: time.Now().Unix(),
	}
	return user.ForUpdateLock(playerID)
}

func (s *UserService) RegData(userID string, serverID uint32) (playerID uint64, err error) {

	userData := mdb.NewUserModel()
	timeStamp := time.Now().Unix()
	userData.AccountName = userID
	userData.AccountUID = userID
	userData.PlatformID = 2
	userData.ChannelID = 24
	userData.ServerID = 1
	userData.Sex = uint32(util.RandInterval(1, 2))
	userData.RegTime = timeStamp
	userData.LastTime = timeStamp

	mTx := mdb.MDB.Begin()
	playerID, err = userData.Insert(mTx)
	err = udb.RegUserData(playerID, serverID)
	if err != nil {
		mTx.Rollback()
		return
	}
	mTx.Commit()

	return playerID, err
}
