package server

import (
	"errors"
	"gameserver/internal/server/models/udb"
	"gameserver/pkg/config"
	"gameserver/pkg/protocal"
	"io"
	"math/rand"
	"net"
	"reflect"
	"runtime"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// User 用户信息
type User struct {
	UserID         int64
	serverID       uint32
	PlatformID     string
	PlatformName   string
	GroupIDs       map[string]string
	ExtInfo        string
	LastActionTime int64
	Conn           *net.TCPConn
	inChan         chan *protocal.ImPacket
	outChan        chan *protocal.ImPacket
	ClosedSig      chan bool
	Closed         bool
	LastHbTime     time.Time
	LastToken      string
	GmFlag         bool
	// ctx            context.Context
	// cancel         context.CancelFunc
	DBInstance sync.Map
	PlayerID   uint64
	PlayerData *udb.UserData
}

// UserList ...
type UserList []interface{}

// NewUser 生成新用户
func NewUser() *User {
	user := &User{}
	user.GroupIDs = make(map[string]string)
	user.outChan = make(chan *protocal.ImPacket, 1024)
	user.inChan = make(chan *protocal.ImPacket, 1024)
	user.ClosedSig = make(chan bool, 1)
	user.Closed = false
	user.UserID = 0
	user.GmFlag = false
	user.PlayerID = 0
	return user
}

// NewUserList ...
func NewUserList() UserList {
	userList := make([]interface{}, 0)
	return userList
}

// SendMessage 检查用户消息队列，并给用户发消息
func (user *User) SendMessage() {
	user.LastHbTime = time.Now()
	var serverHeartbeat = user.RandServerHearbeat()
	for {

		select {
		case imPacket := <-user.outChan:
			if user.Closed {
				return
			}
			_, err := user.Conn.Write(imPacket.Serialize())
			if err != nil {
				log.Errorln("SendMessage conn.Write error:", err.Error(), "userID:", user.UserID)
				goto user_quit
			}
			user.LastHbTime = time.Now()

		case close := <-user.ClosedSig:
			if close {
				goto user_quit
			}

		case <-time.After(time.Second * 300): // 5分钟检测一次心跳
			if user.Closed {
				return
			}
			if time.Now().Sub(user.LastHbTime) > serverHeartbeat {
				log.Errorln("SendMessage.user.lastHb userId:", user.UserID)
				goto user_quit
			}

		}
	}

user_quit:
	user.Closed = true
	user.inChan <- user.ImQuitUserMqPacket()
	runtime.Goexit()

}

// RandServerHearbeat ...
func (user *User) RandServerHearbeat() time.Duration {
	return (minServerHeartbeat + time.Duration(rand.Int63n(int64(maxServerHeartbeat-minServerHeartbeat))))
}

func (user *User) readLoop(conn *net.TCPConn) error {
	var (
		count    int64
		lastTime int64
	)

	for {
		// 读取包内容
		imPacket, err := protocal.ReadPacket(conn)
		if err != nil {
			if err != io.EOF {
				// Error: 解析协议错误
				protocal.SendError(conn, config.ImErrorCodePacketRead, err.Error())
			}
			log.Errorln("ReadPacket Error:", err)
			return err
		}
		if user.Closed == true {
			return nil
		}
		user.inChan <- imPacket
		user.LastHbTime = time.Now()

		// 固定窗口算法Fixed window
		nowTime := time.Now().Unix()
		diffTime := nowTime - lastTime
		if diffTime == 0 {
			count++
			if count > 5 {
				return errors.New("单用户限流 5 qps")
			}
		} else if diffTime >= 1 {
			count = 1
			lastTime = nowTime
		}

	}
}

func (user *User) handleLoop(srv *Server, conn *net.TCPConn) {
	// var (
	// 	autoID int64
	// )

	for {
		select {
		case imPacket := <-user.inChan:
			// 消息类型
			messageType := imPacket.GetType()
			// 来源类型
			fromType := imPacket.GetFrom()
			// 心跳包处理
			if messageType == config.ImHeartbeat {
				user.LastHbTime = time.Now()
				// TODO
				// 返回系统时间
				continue
			}
			// 退出处理
			if messageType == config.ImQuitUserMq {
				goto handleLoopQuit
			}
			// 用户主动退出
			if messageType == config.ImLogout {
				goto handleLoopQuit
			}

			if user.UserID > 0 && messageType == config.ImLogin {
				// 重复登入消息，强制退出
				protocal.SendError(conn, config.ImErrorCodeRelogin, "Repeat login")
				goto handleLoopQuit
			} else if user.UserID < 1 && messageType != config.ImLogin {
				// 未发登入消息，不能发其他消息
				protocal.SendError(conn, config.ImErrorCodeNoLogin, "No login")
				goto handleLoopQuit
			}

			// 预处理如果是gm协议，必须验证user.GmFlag 或 来源IP
			if messageType != config.ImLogin && messageType > 100 && messageType < 200 && user.GmFlag != true {
				protocal.SendError(conn, config.ImErrorCodeNotAllowedImType, "No gm user")
				goto handleLoopQuit
			}

			// 内容分发
			handlerFuncName := srv.FindRouter(messageType)
			if handlerFuncName == "" {
				protocal.SendError(conn, config.ImErrorCodeNotAllowedImType, "Unknown messageType")
				goto handleLoopQuit
			}
			req := NewRequest()
			req.user = user
			req.conn = conn
			req.messageType = messageType
			req.fromType = fromType
			req.body = imPacket.GetBody()
			in := make([]reflect.Value, 1)
			in[0] = reflect.ValueOf(req)
			values := reflect.ValueOf(srv).MethodByName(handlerFuncName).Call(in)
			// 返回结果为数组，[int, error]
			errCode := values[0].Interface().(int)
			if values[1].Interface() != nil || errCode != 0 {

				// user.DBRollBack() // 出错回滚

				errMsg := values[1].Interface().(error).Error()
				protocal.SendError(conn, errCode, errMsg)

				log.Errorln(handlerFuncName, "error:", errMsg, "errcode:", errCode, "userID:", user.UserID)

				goto handleLoopQuit
			}

			// 更新列表提交
			// if err := user.PlayerData.Render(); err != nil {
			// 	log.Errorln("user.PlayerData.Render", "error:", err.Error(), "userID:", user.UserID)
			// 	goto handleLoopQuit
			// }

			// user.DBCommit() // 事务提交
			// 日志提交

		}
	}
handleLoopQuit:
	user.Closed = true
	runtime.Goexit()
}

func (user *User) ImQuitUserMqPacket() *protocal.ImPacket {
	headerBytes := protocal.NewHeader(config.ImQuitUserMq, config.ImFromTypeSytem)
	bodyBytes := []byte("")
	imPacket := protocal.NewImPacket(headerBytes, bodyBytes)
	return imPacket
}

// AddUser ...
func (user *User) AddMDB(db *gorm.DB) error {
	if user.UserID <= 0 {
		return errors.New("Bucket.AddUser :User.UserId must larger than 0")
	}
	user.DBInstance.Store("mdb", db)
	return nil
}

func (user *User) AddUDB(db *gorm.DB) error {
	if user.UserID <= 0 {
		return errors.New("Bucket.AddUser :User.UserId must larger than 0")
	}
	user.DBInstance.Store("udb", db)
	return nil
}

func (user *User) DBRollBack() {
	user.DBInstance.Range(func(key, value interface{}) bool {
		if value.(*gorm.DB) != nil {
			value.(*gorm.DB).Rollback()
		}
		return true
	})
}

func (user *User) DBCommit() {
	user.DBInstance.Range(func(key, value interface{}) bool {
		if value.(*gorm.DB) != nil {
			value.(*gorm.DB).Commit()
		}
		return true
	})
}
