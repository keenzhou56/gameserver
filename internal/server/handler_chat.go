package server

import (
	pb "gameserver/api/protocol"
	"gameserver/pkg/config"
	"gameserver/pkg/protocal"

	"github.com/golang/protobuf/proto"
)

func (srv *Server) ChatGroupHandler(ctx *Request) (int, error) {

	msg := new(pb.ChatGroupMsg)
	err := proto.Unmarshal(ctx.GetBody(), msg)
	if err != nil {
		return 0, err
	}
	groupID := msg.GetGroupID()
	// 获取接受消息的频道信息
	group, err := srv.mapGroup.Get(groupID)
	if err != nil {
		return config.ImErrorCodeGroupInfo, err
	}

	// 生成包
	msgReply := new(pb.ChatGroupMsgReply)
	msgReply.Msg = msg.GetMsg()
	msgReply.SenderId = ctx.user.UserID
	msgReply.SenderName = ""
	body, _ := proto.Marshal(msgReply)
	headerBytes := protocal.NewHeader(config.ImChatGroup, ctx.fromType)
	imPacket := protocal.NewImPacket(headerBytes, body)

	// 遍历频道所有用户，发送消息
	for _, receivedUserID := range group.UserIDs {
		receiverInfo, err := srv.bucket.GetUser(receivedUserID)
		if err != nil {
			continue
		}
		receiverInfo.outChan <- imPacket
	}

	// protocal.SendCommon(ctx.conn, ctx.messageType, ctx.fromType, config.ImResponseCodeSuccess, "")

	return config.ImResponseCodeSuccess, nil
}

func (srv *Server) ChatPrivateHandler(ctx *Request) (int, error) {
	// 读取接收者信息
	msg := new(pb.ChatPrivateMsg)
	err := proto.Unmarshal(ctx.GetBody(), msg)
	if err != nil {
		return 0, err
	}
	receivedUserID := msg.ReceiverId
	receiverInfo, err := srv.bucket.GetUser(receivedUserID)
	if err != nil {
		// 对方不在线，给发送方发送对方不在线的notice
		return config.ImResponseCodeReceiverOffline, nil
	}

	// 生成包头
	headerBytes := protocal.NewHeader(config.ImChatPrivate, ctx.fromType)
	// 生成包体
	msgReply := new(pb.ChatPrivateReply)
	msgReply.Msg = msg.GetMsg()
	msgReply.SenderId = ctx.user.UserID
	msgReply.SenderName = ""
	body, _ := proto.Marshal(msgReply)
	imPacket := protocal.NewImPacket(headerBytes, body)

	// 给接收者发送消息
	receiverInfo.outChan <- imPacket

	protocal.SendCommon(ctx.conn, ctx.messageType, ctx.fromType, config.ImResponseCodeSuccess, "")

	return config.ImResponseCodeSuccess, nil
}

func (srv *Server) ChatBoradcastHandler(ctx *Request) (int, error) {
	// 读取接收者信息
	msg := new(pb.ChatBoradcastMsg)
	err := proto.Unmarshal(ctx.GetBody(), msg)
	if err != nil {
		return 0, err
	}

	// 生成包头
	headerBytes := protocal.NewHeader(config.ImChatBoradcast, ctx.fromType)
	// 生成包体
	msgReply := new(pb.ChatBoradcastMsgReply)
	msgReply.Msg = msg.GetMsg()
	msgReply.SenderId = ctx.user.UserID
	msgReply.SenderName = ""
	body, _ := proto.Marshal(msgReply)
	imPacket := protocal.NewImPacket(headerBytes, body)

	// 给接收者发送消息
	srv.globalMq <- imPacket

	// protocal.SendCommon(ctx.conn, ctx.messageType, ctx.fromType, config.ImResponseCodeSuccess, "")

	return config.ImResponseCodeSuccess, nil
}
