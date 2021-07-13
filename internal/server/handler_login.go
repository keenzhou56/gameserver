package server

import (
	"errors"
	"fmt"
	pb "gameserver/api/protocol"
	"gameserver/internal/server/models/udb"
	"gameserver/internal/server/service"
	"gameserver/pkg/config"
	"gameserver/pkg/protocal"
	"net"
	"strconv"

	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
)

func (srv *Server) LoginHandler(ctx *Request) (int, error) {
	code, err := srv.LoginService(ctx.conn, ctx.user, ctx.body)
	loginMsgReply := new(pb.LoginMsgReply)
	if err != nil {
		log.Errorln(code, err)
		loginMsgReply.Code = int32(code)
		loginMsgReply.Msg = err.Error()
		loginMsgReply.LastToken = ""
		body, _ := proto.Marshal(loginMsgReply)
		protocal.SendProto(ctx.conn, config.ImLogin, ctx.fromType, body)
		return 0, err
	}

	loginMsgReply.Code = int32(config.ImResponseCodeSuccess)
	loginMsgReply.Msg = ""
	loginMsgReply.LastToken = ctx.user.LastToken
	body, _ := proto.Marshal(loginMsgReply)
	_, err = protocal.SendProto(ctx.conn, ctx.messageType, ctx.fromType, body)
	return 0, err
}

func (srv *Server) LoginService(conn *net.TCPConn, newUser *User, body []byte) (int, error) {

	LoginMsg := new(pb.LoginMsg)
	err := proto.Unmarshal(body, LoginMsg)
	if err != nil {
		return 0, err
	}

	loginUserID := LoginMsg.GetUserID()
	platformID := LoginMsg.GetPlatformID()
	platformName := LoginMsg.GetPlatformName()
	loginTime := LoginMsg.GetLoginTime()

	// 验证登录数据是否完整
	if loginUserID <= 0 {
		errMsg := fmt.Sprintf("Error: login data:[userID], given : %d", loginUserID)
		return config.ImErrorCodeUserID, errors.New(errMsg)
	}
	if len(platformID) == 0 {
		return config.ImErrorCodePlatformID, errors.New("Error: login data:[platformID]")
	}
	if len(platformName) == 0 {
		platformName = "user" + platformID
	}
	// TODO 验证loginToken 可以http,或者直接读取redis数据
	player_id := uint64(0)
	userIDStr := strconv.FormatInt(loginUserID, 10)
	m_player, err := service.GetUserService().GetPlayerID(userIDStr, 1)
	if err != nil {
		return 0, err
	}

	if m_player == nil {
		player_id, err = service.GetUserService().RegData(userIDStr, 1)
		if err != nil {
			return 0, err
		}
	} else {
		player_id = m_player.PlayerID
	}

	if player_id < 1 {
		log.Errorln("palyer_id_error:", userIDStr)
		return 0, errors.New("palyer_id_error")
	}

	loginToken := LoginMsg.GetLoginToken()
	// 用户登入验证
	if loginToken != srv.getLoginToken(loginUserID, loginTime) {
		// GM用户登入验证
		if loginToken != srv.getGmToken(loginUserID, loginTime) {
			return config.ImErrorCodeLoginTokenNotMatched, errors.New("Error:login token not matched")
		}
		// log.Errorln(loginUserID, loginTime, loginToken, srv.getGmToken(loginUserID, loginTime))
		newUser.GmFlag = true
	}

	// 若用户已经登录，则关闭以前的连接，以这次登录的为准
	user, err := srv.GetBucket().GetUser(loginUserID)
	if err == nil && user.Conn != conn {
		// 将用户从所有加入的频道移除
		if len(user.GroupIDs) > 0 {
			srv.GetGroups().BatchDelUser(user.GroupIDs, loginUserID)
		}
		// 给已经连接的用户发送被顶下线的消息
		protocal.SendError(user.Conn, config.ImErrorCodeRelogin, "other login")
		// 关闭连接
		user.Closed = true
		user.Conn.Close()
		// 删除被踢用户
		srv.GetBucket().DelUser(loginUserID)
	}

	user = newUser
	user.UserID = loginUserID
	user.PlatformID = platformID
	user.PlatformName = platformName
	user.Conn = conn
	user.LastToken = srv.generateToken(loginUserID)
	srv.GetBucket().AddUser(user)

	user.PlayerID = player_id
	user.PlayerData = &udb.UserData{}
	err = user.PlayerData.InitUserData(user.PlayerID)
	if err != nil {
		return 0, err
	}

	go user.SendMessage()

	return config.ImResponseCodeSuccess, nil
}
