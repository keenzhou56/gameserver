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
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
)

// url参数
var (
	file = flag.String("f", "D:\\go_project\\src\\gameserver\\cmd\\server\\127.0.0.1_8899.conf", "conf file")
)

var (
	userID    int64
	configs   map[string]string
	lastToken string
	debug     bool
)

var quitSemaphore chan bool
var receivedAiMsgCount uint64

// 载入配置文件
func loadConfig(file string) map[string]string {
	// 读取配置文件
	configs, err := common.LoadConf(file)
	if err != nil {
		common.Println("Load Config file error:", err)
		os.Exit(0)
	}

	// 检测配置文件是否完整
	checkKeys := []string{"host", "port", "debug", "system_key", "login_key", "chat_key"}
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

// 初始化
func init() {
	flag.Parse()

	// 读取配置文件
	configs = loadConfig(*file)
	// 调试模式赋值
	if "1" == configs["debug"] {
		debug = true
	} else {
		debug = false
	}
}

func main() {
	var tcpAddr *net.TCPAddr
	tcpAddr, _ = net.ResolveTCPAddr("tcp", configs["host"]+":"+configs["port"])
	conn, _ := net.DialTCP("tcp", nil, tcpAddr)
	defer conn.Close()
	fmt.Println("connected!")

	go onMessageRecived(conn)

	// 控制台聊天功能加入
	for {
		var msg string

		msgReader := bufio.NewReader(os.Stdin)
		msg, _ = msgReader.ReadString('\n')
		msg = strings.TrimSuffix(msg, "\n")
		common.Println(msg)
		if msg == "quit" {
			// logout
			break
		}

		// 分隔消息
		receiveMessageSplit := strings.SplitN(string(msg), "|", 5)
		if len(receiveMessageSplit) < 2 {
			fmt.Println("输入错误，不足2个参数，请重新输入")
			continue
		}

		// 生成包头，2个字节
		imType, _ := strconv.Atoi(receiveMessageSplit[0])
		messageType := uint16(imType)
		// 生成包体
		messageBody := make(map[string]interface{})
		switch messageType {
		case config.ImLogin:
			loginUserID := receiveMessageSplit[1]
			userID, _ = strconv.ParseInt(loginUserID, 10, 64)
			time := common.GetTime()
			messageBody["userID"] = loginUserID
			messageBody["platformID"] = receiveMessageSplit[2]
			messageBody["platformName"] = receiveMessageSplit[3]
			messageBody["time"] = time
			messageBody["loginToken"] = getLoginToken(userID, time)
			body := &pb.LoginMsg{
				UserID:       int64(userID),
				PlatformID:   receiveMessageSplit[2],
				PlatformName: receiveMessageSplit[3],
				LoginTime:    time,
				LoginToken:   getLoginToken(userID, time),
			}
			b, err := proto.Marshal(body)
			if err != nil {
				return
			}
			protocal.SendProto(conn, messageType, config.ImFromTypeUser, b)
			continue
		case config.ImLogout:
			body := &pb.LogoutMsg{}
			b, err := proto.Marshal(body)
			if err != nil {
				return
			}
			protocal.SendProto(conn, messageType, config.ImFromTypeUser, b)
			return
		case config.ImRegisterExtInfo:
			messageBody["token"] = lastToken
			messageBody["extInfo"] = receiveMessageSplit[1]

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

		case config.ImChatBoradcast:
			body := &pb.ChatBoradcastMsg{
				LastToken: lastToken,
				Msg:       receiveMessageSplit[1],
			}
			b, err := proto.Marshal(body)
			if err != nil {
				return
			}
			protocal.SendProto(conn, messageType, config.ImFromTypeUser, b)

		case config.ImChatPrivate:
			ReceiverId, _ := strconv.ParseInt(receiveMessageSplit[1], 10, 64)
			body := &pb.ChatPrivateMsg{
				ReceiverId: ReceiverId,
				LastToken:  lastToken,
				Msg:        receiveMessageSplit[2],
			}
			b, err := proto.Marshal(body)
			if err != nil {
				return
			}
			protocal.SendProto(conn, messageType, config.ImFromTypeUser, b)

		case config.ImChatGroup:
			if len(receiveMessageSplit) != 3 {
				fmt.Println("输入错误，不足3个参数，请重新输入")
				continue
			}
			body := &pb.ChatGroupMsg{
				GroupID:   receiveMessageSplit[1],
				LastToken: lastToken,
				Msg:       receiveMessageSplit[2],
			}
			b, err := proto.Marshal(body)
			if err != nil {
				return
			}
			protocal.SendProto(conn, messageType, config.ImFromTypeUser, b)

		case config.ImChatTestReply:
			// messageBody["token"] = lastToken
			// messageBody["msg"] = receiveMessageSplit[1]

			// messageBody["msg"] = msg
			// messageBody["startTime"] = strconv.FormatInt((time.Now().UnixNano() / 1e6), 10)

			body := &pb.ImApiMsg{
				Msg:       GetRandomString(5),
				StartTime: strconv.FormatInt((time.Now().UnixNano() / 1e6), 10),
			}
			b, err := proto.Marshal(body)
			if err != nil {
				return
			}
			protocal.SendProto(conn, messageType, config.ImFromTypeUser, b)

		default:
			fmt.Println("cannot supported messageType:", receiveMessageSplit[0])

		}
	}
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

		// 调试输出，显示收到的二进制消息
		// 相关参数读取
		// 消息类型
		messageType := imPacket.GetType()
		// fromType := imPacket.GetFrom()

		// 内容分发
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

// 计算登录token
func getLoginToken(userID int64, time int64) string {
	return common.GetToken(configs["login_key"], userID, time)
}

// GetRandomString 生成随机字符串
func GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
