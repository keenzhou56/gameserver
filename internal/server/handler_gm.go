package server

import (
	"errors"
	pb "gameserver/api/protocol"
	"gameserver/pkg/common"
	"gameserver/pkg/config"
	"gameserver/pkg/protocal"
	"runtime"

	"github.com/golang/protobuf/proto"
)

func (s *Server) HandlerGmBefore(imType, fromType uint16, userID, gmTime int64, gmToken, loginIP string) (int, error) {
	// todo gm ip 验证
	if gmToken != s.getGmToken(userID, gmTime) {
		return config.ImErrorCodePrivateKeyNotMatched, errors.New("Error:gmToken not matched")
	}
	return 0, nil
}

func (srv *Server) KickUserHandler(ctx *Context) (int, error) {

	body := new(pb.KickUserMsg)
	err := proto.Unmarshal(ctx.body, body)
	if err != nil {
		return 0, err
	}

	kickUserID := body.GetKickUserID()
	kickUser, err := srv.bucket.GetUser(kickUserID)
	if err != nil {
		return 0, err
	}
	// 回用户
	temp := new(pb.KickUserMsgReply)
	temp.Msg = body.Msg
	bodyReply, _ := proto.Marshal(temp)
	headerBytes := protocal.NewHeader(config.ImKickUser, config.ImFromTypeSytem)
	imPacket := protocal.NewImPacket(headerBytes, bodyReply)
	kickUser.outChan <- imPacket

	if err != nil {
		// 回gm
		protocal.SendCommon(ctx.conn, ctx.messageType, ctx.fromType, config.ImResponseCodeSuccess, "")
	}

	return 0, err
}

func (srv *Server) KickAllHandler(ctx *Context) (int, error) {

	body := new(pb.KickAllMsg)
	err := proto.Unmarshal(ctx.body, body)
	if err != nil {
		return 0, err
	}

	// 相当于全服消息
	temp := new(pb.KickUserMsgReply)
	temp.Msg = body.Msg
	bodyReply, _ := proto.Marshal(temp)
	headerBytes := protocal.NewHeader(config.ImKickUser, config.ImFromTypeSytem)
	imPacket := protocal.NewImPacket(headerBytes, bodyReply)
	srv.globalMq <- imPacket

	return 0, err
}

func (srv *Server) StatHandler(ctx *Context) (int, error) {

	var statInfo = srv.stat.Get()
	body := new(pb.StatMsgReply)
	body.StartTime = statInfo.StartTime
	body.RunTime = statInfo.RunTime
	body.ConnectCount = int32(srv.bucket.LenUser())

	body.MaxConnectCount = srv.bucket.MaxOnLine
	body.GroupCount = int32(srv.mapGroup.GetOnline())
	body.MaxGroupCount = srv.mapGroup.MaxOnLine
	body.PrivateMessageCount = statInfo.PrivateMessageCount
	body.BoradcastMessageCount = statInfo.BoradcastMessageCount
	body.GroupMessageCount = statInfo.GroupMessageCount
	body.SysBoradcastMessageCount = statInfo.SysBoradcastMessageCount
	body.SysPrivateMessageCount = statInfo.SysPrivateMessageCount
	body.SysPrivateMessageCount = statInfo.SysGroupMessageCount

	if body.ConnectCount > 0 && body.ConnectCount < 1000 {
		srv.bucket.mapUser.Range(func(key, value interface{}) bool {
			common.Println("UserID:", value.(*User).UserID)
			return true
		})
	}
	body.SvrGoroutineCount = int32(runtime.NumGoroutine())

	bodyReply, _ := proto.Marshal(body)

	_, err := protocal.SendProto(ctx.conn, ctx.messageType, ctx.fromType, bodyReply)

	return 0, err
}

func (srv *Server) GetUserHandler(ctx *Context) (int, error) {

	body := new(pb.UserMsg)
	err := proto.Unmarshal(ctx.body, body)
	if err != nil {
		return 1, err
	}
	userID := body.GetUserID()
	user, err := srv.bucket.GetUser(userID)
	if err != nil {
		return 2, err
	}
	// 相当于全服消息
	temp := new(pb.UserMsgReply)
	temp.UserID = user.UserID
	temp.Closed = user.Closed
	temp.GmFlag = user.GmFlag
	temp.PlatformID = user.PlatformID
	temp.PlatformName = user.PlatformName
	bodyReply, _ := proto.Marshal(temp)

	_, err = protocal.SendProto(ctx.conn, ctx.messageType, ctx.fromType, bodyReply)

	return 0, err
}
