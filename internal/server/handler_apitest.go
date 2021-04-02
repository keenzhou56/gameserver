package server

import (
	"context"
	"fmt"
	"gameserver/api/logic"
	pb "gameserver/api/protocol"
	"gameserver/pkg/common"
	"gameserver/pkg/config"
	"gameserver/pkg/protocal"
	"net"
	"sync/atomic"

	"github.com/golang/protobuf/proto"
)

func (srv *Server) ApitestHandler(ctx *Context) (int, error) {
	// 解析
	imApiMsg := new(pb.ImApiMsg)
	err := proto.Unmarshal(ctx.body, imApiMsg)
	if err != nil {
		return 0, err
	}
	// rpc调用
	// err = srv.ApitestClient(int64(ctx.user.UserID))
	// if err != nil {
	// 	fmt.Println("["+common.GetTimestamp()+"]:srv.ApiClient:", err.Error())
	// 	return 0, err
	// }
	// lua 调用

	// 回包
	imApiMsgReply := new(pb.ImApiMsgReply)
	imApiMsgReply.Code = int32(config.ImResponseCodeSuccess)
	imApiMsgReply.Msg = imApiMsg.GetMsg()
	imApiMsgReply.StartTime = imApiMsg.GetStartTime()
	body, _ := proto.Marshal(imApiMsgReply)
	_, err = protocal.SendProto(ctx.conn, ctx.messageType, ctx.fromType, body)

	if ctx.fromType == config.ImFromTypeAi {
		// 压力测试输出，接收消息数量
		idx := atomic.AddUint64(&receivedAiMsgCount, 1)
		if idx%100 == 0 {
			fmt.Println("["+common.GetTimestamp()+"]:receivedAiMsgCount:", idx)
		}

	}

	return 0, err
}

func (srv *Server) ApitestService(conn *net.TCPConn, user *User, body []byte) (int, error) {
	imApiMsg := new(pb.ImApiMsg)
	err := proto.Unmarshal(body, imApiMsg)
	if err != nil {
		return 0, err
	}
	return config.ImResponseCodeSuccess, nil
}

func (srv *Server) ApitestClient(userID int64) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := srv.Receive(ctx, userID, &logic.Proto{})
	return err
}
