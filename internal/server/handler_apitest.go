package server

import (
	"context"
	"fmt"
	lpb "gameserver/api/logic"
	pb "gameserver/api/protocol"
	"gameserver/pkg/common"
	"gameserver/pkg/config"
	"gameserver/pkg/protocal"
	"net"
	"sync/atomic"

	"github.com/golang/protobuf/proto"

	log "github.com/sirupsen/logrus"
)

var mysqlUpdateCount uint64
var mysqlUpdateCountErr uint64

func (srv *Server) ApitestHandler(req *Request) (int, error) {
	// 解析
	imApiMsg := new(pb.ImApiMsg)
	err := proto.Unmarshal(req.body, imApiMsg)
	if err != nil {
		return 0, err
	}
	// rpc调用
	// err = srv.ApitestClient(int64(req.user.UserID))
	// if err != nil {
	// 	fmt.Println("["+common.GetTimestamp()+"]:srv.ApiClient:", err.Error())
	// 	return 0, err
	// }

	// lua 调用

	// mysql 调用

	// err = service.GetUserService().AutoUpdateUserData(req.user.PlayerID)
	// err = service.GetUserService().ForUpdateLock(req.user.PlayerID)
	// sql := service.GetUserService().UpdateDryRun(uint64(req.user.PlayerID))

	// sql := fmt.Sprintf("UPDATE `m_player` SET `last_time`= %d WHERE player_id = %d LIMIT 1", time.Now().Unix(), req.user.PlayerID)
	// tableSuffix := ""
	// if req.user.PlayerID < 10 {
	// 	tableSuffix = "00" + strconv.FormatUint(uint64(req.user.PlayerID), 10)
	// } else if req.user.PlayerID < 100 {
	// 	tableSuffix = "0" + strconv.FormatUint(uint64(req.user.PlayerID), 10)
	// } else {
	// 	uidStr := strconv.FormatUint(req.user.PlayerID, 10)
	// 	tableSuffix = uidStr[len(uidStr)-3:]
	// }
	// err = srv.xLogic.Pub("mdb_"+tableSuffix, sql)
	// if err != nil {
	// 	err = srv.xLogic.Pub("mdb_"+tableSuffix, sql)
	// 	if err != nil {
	// 		return 0, err
	// 	}
	// }
	data := make(map[string]string)
	data["k"] = "k"
	data["v"] = "v"
	log.WithFields(log.Fields{"log": data, "user_ip": req.conn.RemoteAddr().String()}).Info()
	if err == nil {
		// idx := atomic.AddUint64(&mysqlUpdateCount, 1)
		// if idx%100 == 0 {
		// 	fmt.Println("["+common.GetTimestamp()+"]:mysqlUpdateCount:", idx)
		// }
	} else {
		idx := atomic.AddUint64(&mysqlUpdateCountErr, 1)
		if idx%100 == 0 {
			fmt.Println("db update err:", err.Error())
			fmt.Println("["+common.GetTimestamp()+"]:mysqlUpdateCountErr:", idx)
		}
	}
	// sql := service.GetUserService().UpdateDryRun(uint64(req.user.UserID))
	// fmt.Println("db resqlgister:", sql)

	// }

	// 回包
	imApiMsgReply := new(pb.ImApiMsgReply)
	imApiMsgReply.Code = int32(config.ImResponseCodeSuccess)
	imApiMsgReply.Msg = imApiMsg.GetMsg()
	imApiMsgReply.StartTime = imApiMsg.GetStartTime()
	body, _ := proto.Marshal(imApiMsgReply)
	_, err = protocal.SendProto(req.conn, req.messageType, req.fromType, body)

	if req.fromType == config.ImFromTypeAi {
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
	err := srv.Receive(ctx, userID, &lpb.Proto{})
	return err
}
