// 发送系统消息

package main

import (
	"bufio"
	"flag"
	"fmt"
	pb "gameserver/api/protocol"
	"gameserver/pkg/common"
	"gameserver/pkg/config"
	"gameserver/pkg/protocal"
	"io"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
)

var quitSemaphore chan bool

var configs map[string]string

var lastToken string

// url参数
var (
	file = flag.String("f", "D:\\go_project\\src\\im\\cmd\\server\\127.0.0.1_8899.conf", "conf file")
)

func loadConfig(file string) map[string]string {
	// 读取配置文件
	configs, err := common.LoadConf(file)
	if err != nil {
		common.Println("Load Config file error:", err)
		os.Exit(0)
	}

	// 检测配置文件是否完整
	checkKeys := []string{"host", "port", "system_key"}
	for _, key := range checkKeys {
		if _, exists := configs[key]; !exists {
			common.Println("config [", key, "] not found in config file:", file)
			os.Exit(0)
		}
	}

	// 设置默认值
	defaultKeys := map[string]string{
		"debug": "0",
	}
	for key, val := range defaultKeys {
		if _, exists := configs[key]; !exists {
			configs[key] = val
		}
	}

	return configs
}

func init() {
	// 解析参数
	flag.Parse()

	// 读取配置文件
	configs = loadConfig(*file)

	fmt.Println("remote:", configs["host"]+":"+configs["port"])
}

func main() {
	// 开启多核模式
	runtime.GOMAXPROCS(runtime.NumCPU())

	var tcpAddr *net.TCPAddr
	tcpAddr, _ = net.ResolveTCPAddr("tcp", configs["host"]+":"+configs["port"])
	conn, _ := net.DialTCP("tcp", nil, tcpAddr)
	defer conn.Close()
	fmt.Println("connected!")

	go onMessageRecived(conn)
	var userID int
	userID = 1
	// 控制台聊天功能加入
	for {
		var msg string

		msgReader := bufio.NewReader(os.Stdin)
		msg, _ = msgReader.ReadString('\n')
		msg = strings.TrimSuffix(msg, "\n")

		if len(msg) == 0 {
			fmt.Println("请输入协议")
			continue
		}

		if msg == "quit" {
			// logout
			break
		}

		messageBody := make(map[string]interface{})
		messageBody["systemKey"] = configs["system_key"]

		// 分隔消息
		receiveMessageSplit := strings.SplitN(string(msg), "|", 3)
		messageTypeReceive, _ := strconv.Atoi(receiveMessageSplit[0])
		messageType := uint16(messageTypeReceive)
		common.Println("messageType:", messageType)
		switch messageType {
		case config.ImLogin:
			loginUserID := "1"
			userID, _ = strconv.Atoi(loginUserID)
			time := common.GetTime()

			body := &pb.LoginMsg{
				UserID:       int64(userID),
				PlatformID:   "1",
				PlatformName: "1",
				LoginTime:    time,
				LoginToken:   getGmToken(int64(userID), time),
			}
			b, err := proto.Marshal(body)
			if err != nil {
				return
			}
			_, err = protocal.SendProto(conn, messageType, config.ImFromTypeSytem, b)
		case config.ImLogout:
			messageBody["token"] = lastToken
		case config.ImStat: // 统计服务器状态
			body := &pb.StatMsg{}
			b, err := proto.Marshal(body)
			if err != nil {
				return
			}
			protocal.SendProto(conn, messageType, config.ImFromTypeSytem, b)

		case config.ImCheckOnline: // 判断用户是否在线
			UserID, _ := strconv.ParseInt(receiveMessageSplit[1], 10, 64)
			body := &pb.UserMsg{
				UserID: UserID,
			}
			b, err := proto.Marshal(body)
			if err != nil {
				return
			}

			protocal.SendProto(conn, messageType, config.ImFromTypeSytem, b)

		case config.ImKickUser: // 踢某用户下线
			KickUserID, _ := strconv.ParseInt(receiveMessageSplit[1], 10, 64)
			body := &pb.KickUserMsg{
				KickUserID: KickUserID,
				Msg:        receiveMessageSplit[2],
			}
			b, err := proto.Marshal(body)
			if err != nil {
				return
			}
			protocal.SendProto(conn, messageType, config.ImFromTypeSytem, b)

		case config.ImKickAll: // 踢所有用户下线
			body := &pb.KickAllMsg{
				Msg: receiveMessageSplit[1],
			}
			b, err := proto.Marshal(body)
			if err != nil {
				return
			}
			protocal.SendProto(conn, messageType, config.ImFromTypeSytem, b)

		case config.ImJoinGroup:
			body := &pb.JoinGroupMsg{
				GroupID:   receiveMessageSplit[1],
				LastToken: lastToken,
			}
			b, err := proto.Marshal(body)
			if err != nil {
				return
			}
			protocal.SendProto(conn, messageType, config.ImFromTypeUser, b)

		case config.ImQuitGroup:
			body := &pb.QuitGroupMsg{
				GroupID:   receiveMessageSplit[1],
				LastToken: lastToken,
			}
			b, err := proto.Marshal(body)
			if err != nil {
				return
			}
			protocal.SendProto(conn, messageType, config.ImFromTypeUser, b)

		case config.ImGroupUserList: // 获取频道用户列表
			body := &pb.GroupUserList{
				GroupID: receiveMessageSplit[1],
			}
			b, err := proto.Marshal(body)
			if err != nil {
				return
			}
			protocal.SendProto(conn, messageType, config.ImFromTypeSytem, b)

		case config.ImChatBoradcast: // 世界聊天
			body := &pb.ChatBoradcastMsg{
				Msg: receiveMessageSplit[1],
			}
			b, err := proto.Marshal(body)
			if err != nil {
				return
			}
			_, err = protocal.SendProto(conn, messageType, config.ImFromTypeSytem, b)
			if err != nil {
				common.Println(err.Error())
			}

		case config.ImChatGroup: // 频道聊天
			body := &pb.ChatGroupMsg{
				GroupID: receiveMessageSplit[1],
				Msg:     receiveMessageSplit[2],
			}
			b, err := proto.Marshal(body)
			if err != nil {
				return
			}
			protocal.SendProto(conn, messageType, config.ImFromTypeSytem, b)

		case config.ImChatPrivate: // 私聊
			ReceiverId, _ := strconv.ParseInt(receiveMessageSplit[1], 10, 64)
			body := &pb.ChatPrivateMsg{
				ReceiverId: ReceiverId,
				Msg:        receiveMessageSplit[2],
			}
			b, err := proto.Marshal(body)
			if err != nil {
				return
			}
			protocal.SendProto(conn, messageType, config.ImFromTypeSytem, b)
			continue

		default:
			fmt.Println("do not supported message type:", messageType)
		}
	}
}

// 计算登录token
func getGmToken(userID int64, time int64) string {
	return common.GetToken(configs["system_key"], userID, time)
}
func onMessageRecived(conn *net.TCPConn) {
	for {
		imPacket, err := protocal.ReadPacket(conn)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Disconnected")
				os.Exit(0)
			} else {
				fmt.Println("ReadPacket Error:", err)
			}
			return
		}

		// 相关参数读取
		// 消息类型
		messageType := imPacket.GetType()

		switch messageType {
		case config.ImLogin:
			loginMsgReply := new(pb.LoginMsgReply)
			err := proto.Unmarshal(imPacket.GetBody(), loginMsgReply)
			if err != nil {
				return
			}
			responseCode := loginMsgReply.GetCode()
			switch responseCode {
			case config.ImResponseCodeSuccess:
				lastToken = loginMsgReply.GetLastToken()
				common.Println("Response: success.", "code:", responseCode, "imType:", messageType)
			case config.ImResponseCodeReceiverOffline:
				common.Println("private receiver is offline.")
			default:
				common.Println("Response: code:", responseCode, ", imType:", messageType)
			}

		case config.ImChatTestReply:
			imApiMsgReply := new(pb.ImApiMsgReply)
			err := proto.Unmarshal(imPacket.GetBody(), imApiMsgReply)
			if err != nil {
				return
			}
			responseCode := imApiMsgReply.GetCode()
			switch responseCode {
			case config.ImResponseCodeSuccess:
				msg := imApiMsgReply.GetMsg()
				startTime := imApiMsgReply.GetStartTime()
				common.Println("Response: success.", "code:", responseCode, "imType:", messageType, "msg", msg, "startTime", startTime)
			default:
				common.Println("Response: code:", responseCode, ", imType:", messageType)
			}

		case config.ImKickUser:
			kickMsg := new(pb.KickUserMsgReply)
			err := proto.Unmarshal(imPacket.GetBody(), kickMsg)
			if err != nil {
				return
			}
			common.Println("ImKickUser: msg:", kickMsg.Msg)
			os.Exit(0)
		case config.ImError:
			errorMsg := new(pb.CommonMsg)
			err := proto.Unmarshal(imPacket.GetBody(), errorMsg)
			if err != nil {
				return
			}
			common.Println("ImError: code: [", errorMsg.Code, "], msg:", errorMsg.Msg)
			os.Exit(0)
		case config.ImGroupUserList:
			groupListRep := new(pb.GroupUserListReply)
			err := proto.Unmarshal(imPacket.GetBody(), groupListRep)
			if err != nil {
				return
			}

			for _, user := range groupListRep.UserList {
				common.Println("Response: userID.", user.UserID, "Pid:", user.PlatformID, "pname:", user.PlatformName)
			}

		case config.ImCheckOnline: // 判断用户是否在线
			user := new(pb.UserMsgReply)
			err := proto.Unmarshal(imPacket.GetBody(), user)
			if err != nil {
				return
			}

			common.Println("Response: userID.", user.UserID, "Pid:", user.PlatformID, "pname:", user.PlatformName, "closed:", user.Closed, "gmflag:", user.GmFlag)

		case config.ImStat: // 统计信息
			stat := new(pb.StatMsgReply)
			err := proto.Unmarshal(imPacket.GetBody(), stat)
			if err != nil {
				return
			}
			fmt.Println("启动时间:", common.FormatUnixTime(stat.StartTime))
			fmt.Println("运行时长:", common.FormateRunTime(stat.RunTime))
			fmt.Println("当前协程的数量:", stat.SvrGoroutineCount)
			fmt.Println("连接用户数:", stat.ConnectCount)
			fmt.Println("最大连接用户数:", stat.MaxConnectCount)
			fmt.Println("频道数:", stat.GroupCount)
			fmt.Println("最大频道数:", stat.MaxGroupCount)
			fmt.Println("世界消息数:", stat.BoradcastMessageCount)
			fmt.Println("频道消息数:", stat.GroupMessageCount)
			fmt.Println("私聊消息数:", stat.PrivateMessageCount)
			fmt.Println("系统世界消息数:", stat.SysBoradcastMessageCount)
			fmt.Println("系统频道消息数:", stat.SysGroupMessageCount)
			fmt.Println("系统私聊消息数:", stat.SysPrivateMessageCount)
		case config.ImJoinGroup:
			errorMsg := new(pb.CommonMsg)
			err := proto.Unmarshal(imPacket.GetBody(), errorMsg)
			if err != nil {
				return
			}
			common.Println("response: code: [", errorMsg.Code, "], msg:", errorMsg.Msg)

		case config.ImQuitGroup:
			errorMsg := new(pb.CommonMsg)
			err := proto.Unmarshal(imPacket.GetBody(), errorMsg)
			if err != nil {
				return
			}
			common.Println("response: code: [", errorMsg.Code, "], msg:", errorMsg.Msg)

		case config.ImChatBoradcast:
			msg := new(pb.ChatBoradcastMsgReply)
			err := proto.Unmarshal(imPacket.GetBody(), msg)
			if err != nil {
				return
			}
			common.Println("[world][", msg.SenderId, msg.SenderName, "][say]:", msg.Msg)

		case config.ImChatGroup:
			msg := new(pb.ChatGroupMsgReply)
			err := proto.Unmarshal(imPacket.GetBody(), msg)
			if err != nil {
				return
			}
			common.Println("[group][groupID:", msg.GroupID, "][", msg.SenderId, "][say]:", msg.Msg)

		case config.ImChatPrivate:
			msg := new(pb.ChatPrivateReply)
			err := proto.Unmarshal(imPacket.GetBody(), msg)
			if err != nil {
				return
			}
			common.Println("[private]", "[from:", msg.SenderId, msg.SenderName, "]:", msg.Msg)
		default:
			fmt.Println("cannot supported messageType:", messageType)
		}
	}
	// quitSemaphore <- true
}
