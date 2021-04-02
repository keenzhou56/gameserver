// 聊天服务器压力测试工具
package main

import (
	"flag"
	"fmt"
	"gameserver/pkg/common"
	"gameserver/pkg/config"
	"gameserver/pkg/json"
	"gameserver/pkg/protocal"
	"io"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	pb "gameserver/api/protocol"

	"google.golang.org/protobuf/proto"
)

// url参数
var (
	file = flag.String("f", "D:\\go_project\\src\\gameserver\\cmd\\server\\127.0.0.1_8899.conf", "conf file")
)

// 用来接受host和port参数
var (
	host     = flag.String("host", "127.0.0.1", "im server host")
	port     = flag.String("port", "8899", "im server port")
	startNum = flag.Int("sn", 10000, "open count")
	count    = flag.Int("n", 15000, "open count")
	step     = flag.Int("s", 3000, "send step")
	testTime = flag.Int("t", 600, "send step")
	all      = flag.Int("all", 0, "send step")
	msgLen   = flag.Int("l", 2026, "send step")
	rwLock   sync.RWMutex
	configs  map[string]string
)

var totalMsg int64
var totalReplyMsg int64
var online int64
var diffTimes []int64
var c100Times int64
var c200Times int64
var c400Times int64
var c500Times int64
var c600Times int64
var c800Times int64
var msg string

// 随机发言的内容
var randMsgArr = [45]string{
	"大家好，我是机器人",
	"你好，请问你是机器人么",
	"你才是机器人，你全家都是机器人",
	"Do one thing at a time, and do well.",
	"Never forget to say &ldquo;thanks&rdquo;.",
	"Keep on going never give up.",
	"Whatever is worth doing is worth doing well.",
	"Believe in yourself.",
	"I can because i think i can.",
	"Action speak louder than words.",
	"Never say die.",
	"Never put off what you can do today until tomorrow.",
	"The best preparation for tomorrow is doing your best today.",
	"You cannot improve your past, but you can improve your future. Once time is wasted, life is wasted.",
	"Knowlegde can change your fate and English can accomplish your future.",
	"Don't aim for success if you want it; just do what you love and believe in, and it will come naturally.",
	"Jack of all trades and master of none.",
	"Judge not from appearances.",
	"Justice has long arms.",
	"Keep good men company and you shall be of the number.",
	"Kill two birds with one stone.",
	"Kings go mad, and the people suffer for it.",
	"Kings have long arms.",
	"Knowledge is power.",
	"Knowledge makes humble, ignorance makes proud.",
	"Learn and live.",
	"Learning makes a good man better and ill man worse.",
	"Learn not and know not.",
	"Learn to walk before you run.",
	"Let bygones be bygones.",
	"Let sleeping dogs lie.",
	"Let the cat out of the bag.",
	"Lies have short legs.",
	"Life is but a span.",
	"Life is half spent before we know what it is.",
	"Life is not all roses.",
	"Life without a friend is death.",
	"Like a rat in a hole.",
	"Like author, like book.",
	"Like father, like son.",
	"Like for like.",
	"Like knows like.",
	"Like mother, like daughter.",
	"Like teacher, like pupil.",
	"Like tree, like fruit.",
}

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

func init() {
	// 解析参数
	flag.Parse()
	configs = loadConfig(*file)
	fmt.Println("im server, host:"+*host+":"+*port+", count:", *count, ", step:", *step)
	msg = GetRandomString(*msgLen)
}

func main() {
	// 开启多核模式
	// runtime.GOMAXPROCS(runtime.NumCPU())
	// var wg sync.WaitGroup
	go func() {
		fmt.Println("testTime:", *testTime)
		for t := 1; t <= *testTime; t++ {
			time.Sleep(time.Second)
			idx := totalMsg
			idx2 := online
			idx3 := totalReplyMsg
			// c := int64(len(diffTimes))
			// pt := int64(0)
			// if c > 0 {
			// 	totalTime := int64(0)
			// 	for _, v := range diffTimes {
			// 		totalTime += v
			// 	}
			// 	pt = int64(totalTime / c)
			// }
			fmt.Println("time:", t, "totalMsg:", idx, "totalReplyMsg:", totalReplyMsg, "online:", idx2,
				"qps:", (int(idx))/t, (int(idx3))/t,
				"relay<100", c100Times, "relay<200", c200Times, "relay<400", c400Times, "relay<500", c500Times,
				"relay<600", c600Times, "relay>600", c800Times)
		}
	}()

	for i := 1; i <= *count; i++ {
		go startClient(i)
		time.Sleep(time.Second / 1000)
	}

	for t := 1; t <= *testTime; t++ {
		time.Sleep(time.Second)
	}
	os.Exit(0)
}

func startClient(i int) {
	userID := *startNum + i
	platformID := "AI_" + strconv.Itoa(userID)
	platformName := "AI_" + strconv.Itoa(userID)

	var tcpAddr *net.TCPAddr
	tcpAddr, _ = net.ResolveTCPAddr("tcp", *host+":"+*port)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return
	}
	atomic.AddInt64(&online, 1)
	defer func() {
		// 捕获异常
		if err := recover(); err != nil {
			common.Println("tcpPipe defer recover error:", err, userID)
		}
		online--
		conn.Close()
	}()
	// login
	lgTime := common.GetTime()
	body := &pb.LoginMsg{
		UserID:       int64(userID),
		PlatformID:   platformID,
		PlatformName: platformName,
		LoginTime:    lgTime,
		LoginToken:   getLoginToken(int64(userID), lgTime),
	}
	b, err := proto.Marshal(body)
	if err != nil {
		return
	}
	_, err = protocal.SendProto(conn, config.ImLogin, config.ImFromTypeAi, b)
	if err != nil {
		fmt.Println("login Error:", err, userID)
		return
	}
	// common.Println(body)
	atomic.AddInt64(&totalMsg, 1)

	// go onMessageRecived(conn, strconv.Itoa(userID))

	if *all == 1 {
		// send a message step sencond
		for {
			// 随机休息时间、防止脚本启动消息时，同时发送的消息过多
			rand.Seed(time.Now().UnixNano() - int64(userID*userID*userID))
			randSecond := int(rand.Int31n(int32(*step)))
			for j := 1; j <= randSecond+1; j++ {
				time.Sleep(time.Second)
			}

			messageBody := make(map[string]interface{})

			// 获取随机消息
			rand.Seed(time.Now().UnixNano() - int64(userID*userID*userID))
			randKey := int(rand.Int31n(45))

			idx := atomic.AddInt64(&totalMsg, 1)

			messageBody["msg"] = fmt.Sprintf("[ %d ]: [ %s ]", idx-1, randMsgArr[randKey])
			messageBody["start"] = time.Now().UnixNano()
			// 发送消息
			protocal.Send(conn, config.ImChatBoradcast, config.ImFromTypeAi, messageBody)

			if idx%10 == 0 {
				fmt.Println("["+common.GetTimestamp()+"]"+"totalMsg:", idx, "imPacket.len")
			}

		}
	} else {
		// send a message step sencond

		for {
			// time.Sleep(time.Second / 4)
			// messageBody := make(map[string]interface{})
			// messageBody["msg"] = msg
			// messageBody["startTime"] = strconv.FormatInt((time.Now().UnixNano() / 1e6), 10)

			body := &pb.ImApiMsg{
				Msg:       msg,
				StartTime: strconv.FormatInt((time.Now().UnixNano() / 1e6), 10),
			}
			b, err := proto.Marshal(body)
			if err != nil {
				return
			}
			// 发送消息
			_, err = protocal.SendProto(conn, config.ImChatTestReply, config.ImFromTypeAi, b)
			// fmt.Println("imPacket proto.Marshal Len:", im.GetLength())
			// _, err := protocal.Send(conn, config.ImChatTestReply, config.ImFromTypeAi, messageBody)
			if err != nil {
				fmt.Println("Send Error:", err, userID)
				return
			}
			atomic.AddInt64(&totalMsg, 1)

			// 收消息
			imPacket, err := protocal.ReadPacket(conn)
			// fmt.Println("imPacket.GetLength:", imPacket.GetLength())
			if err != nil {
				if err == io.EOF {
					fmt.Println("Disconnected")
				} else {
					fmt.Println("ReadPacket Error:", err)
				}
				return
			}

			messageType := imPacket.GetType()
			// messageBody, err = json.Decode(string(imPacket.GetBody()))
			// if err != nil {
			// 	fmt.Println("imPacket JsonDecode error:", err)
			// 	return
			// }

			atomic.AddInt64(&totalReplyMsg, 1)

			switch messageType {
			case config.ImKickUser:
				kickMsg := new(pb.KickUserMsgReply)
				err := proto.Unmarshal(imPacket.GetBody(), kickMsg)
				if err != nil {
					return
				}
				common.Println("ImKickUser: msg:", kickMsg.Msg)
				return
			case config.ImChatTestReply:

				imApiMsgReply := new(pb.ImApiMsgReply)
				err := proto.Unmarshal(imPacket.GetBody(), imApiMsgReply)
				if err != nil {
					return
				}
				responseCode := imApiMsgReply.GetCode()
				switch responseCode {
				case config.ImResponseCodeSuccess:
					// msg := imApiMsgReply.GetMsg()
					st := imApiMsgReply.GetStartTime()
					// st, _ := protocal.GetBodyString(messageBody, "startTime")
					startTime, _ := strconv.ParseInt(st, 10, 64)
					endTime := int64(time.Now().UnixNano() / 1e6)

					diffTime := endTime - startTime
					if diffTime <= 100 {
						atomic.AddInt64(&c100Times, 1)
					} else if diffTime < 200 {
						atomic.AddInt64(&c200Times, 1)
					} else if diffTime < 400 {
						atomic.AddInt64(&c400Times, 1)
					} else if diffTime < 500 {
						atomic.AddInt64(&c500Times, 1)
					} else if diffTime < 600 {
						atomic.AddInt64(&c600Times, 1)
					} else {
						atomic.AddInt64(&c800Times, 1)
					}
				default:
					common.Println("Response: code:", responseCode, ", imType:", messageType)
				}

			default:
				// fmt.Println("cannot supported messageType:", messageType)
			}
		}
	}

}

func onMessageRecived(conn *net.TCPConn, userID string) {
	for {
		imPacket, err := protocal.ReadPacket(conn)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Disconnected")
			} else {
				fmt.Println("ReadPacket Error:", err)
			}
		} else {
			atomic.AddInt64(&totalReplyMsg, 1)
		}
		messageType := imPacket.GetType()
		// 包体 map[string]interface{}
		messageBody, err := json.Decode(string(imPacket.GetBody()))
		if err != nil {
			fmt.Println("imPacket JsonDecode error:", err)
			return
		}
		// fmt.Println("body:", messageBody)
		switch messageType {
		case config.ImResponse:
			responseCode, _ := protocal.GetBodyInt(messageBody, "code")
			responseImType, _ := protocal.GetBodyUint16(messageBody, "imType")
			switch responseCode {
			case config.ImResponseCodeSuccess:
				common.Println("Response: success.", "code:", responseCode, "imType:", responseImType)
			case config.ImResponseCodeReceiverOffline:
				common.Println("private receiver is offline.")
			default:
				common.Println("Response: code:", responseCode, ", imType:", responseImType)
			}
		case config.ImChatTestReply:
			st, _ := protocal.GetBodyString(messageBody, "startTime")
			startTime, _ := strconv.ParseInt(st, 10, 64)
			endTime := int64(time.Now().UnixNano() / 1e6)

			diffTime := endTime - startTime
			if diffTime <= 50 {
				// atomic.AddInt64(&c50Times, 1)
			} else if diffTime < 100 {
				atomic.AddInt64(&c100Times, 1)
			} else if diffTime < 200 {
				atomic.AddInt64(&c200Times, 1)
			} else {
				atomic.AddInt64(&c500Times, 1)
			}

			// t2 := strconv.FormatInt(endTime, 10)
			// fmt.Println("time:", userID, st, t2)

		default:
			// fmt.Println("cannot supported messageType:", messageType)
		}
	}
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

// 计算登录token
func getLoginToken(userID int64, time int64) string {
	return common.GetToken(configs["login_key"], userID, time)
}
