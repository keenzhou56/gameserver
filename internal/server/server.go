package server

import (
	"context"
	"gameserver/api/logic"
	"gameserver/internal/server/conf"
	xLogic "gameserver/internal/server/logic"
	"gameserver/pkg/config"
	"gameserver/pkg/json"
	"gameserver/pkg/protocal"
	"gameserver/pkg/rate/limit/bbr"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/keepalive"
)

const (
	minServerHeartbeat = time.Minute * 1
	maxServerHeartbeat = time.Minute * 3
	// grpc options
	grpcInitialWindowSize     = 1 << 24
	grpcInitialConnWindowSize = 1 << 24
	grpcMaxSendMsgSize        = 1 << 24
	grpcMaxCallMsgSize        = 1 << 24
	grpcKeepAliveTime         = time.Second * 10
	grpcKeepAliveTimeout      = time.Second * 3
	grpcBackoffMaxDelay       = time.Second * 3
)

// Server is comet server.
type Server struct {
	c        *conf.Config
	bucket   *Bucket
	mapGroup *Groups
	serverID string
	globalMq chan *protocal.ImPacket
	stat     *Stat
	// log       *log.Helper
	router    *HandlersChain
	rpcClient logic.LogicClient
	xLogic    *xLogic.Logic
	mdbMq     chan string
	LimitBbr  *bbr.Group
}

func newLogicClient(c *conf.RPCClient) logic.LogicClient {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Dial))
	defer cancel()
	conn, err := grpc.DialContext(ctx, c.SrvAddr,
		[]grpc.DialOption{
			grpc.WithInsecure(),
			grpc.WithInitialWindowSize(grpcInitialWindowSize),
			grpc.WithInitialConnWindowSize(grpcInitialConnWindowSize),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMaxCallMsgSize)),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMaxSendMsgSize)),
			grpc.WithBackoffMaxDelay(grpcBackoffMaxDelay),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                grpcKeepAliveTime,
				Timeout:             grpcKeepAliveTimeout,
				PermitWithoutStream: true,
			}),
			grpc.WithBalancerName(roundrobin.Name),
		}...)
	if err != nil {
		panic(err)
	}
	return logic.NewLogicClient(conn)
}

// NewServer returns a new Server.
func NewServer(c *conf.Config) *Server {
	s := &Server{
		c: c,
	}
	// init bucket
	s.bucket = NewBucket(c.Bucket)
	s.serverID = c.Env.Host
	s.globalMq = make(chan *protocal.ImPacket, 1024)
	s.mapGroup = NewGroups()
	s.stat = NewStat()
	s.rpcClient = newLogicClient(c.RPCClient)
	s.mdbMq = make(chan string, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.xLogic = xLogic.New(s.c)
	err := s.xLogic.Ping(ctx)
	if err != nil {
		panic(err)
	}
	//
	cfg := &bbr.Config{
		Window:       time.Second * 10,
		WinBucket:    100,
		CPUThreshold: 800,
	}
	s.LimitBbr = bbr.NewGroup(cfg)

	go s.onlineproc()
	return s
}

// Close close the server.
func (s *Server) Close() (err error) {
	return
}

func (s *Server) GetBucket() *Bucket {
	return s.bucket
}

func (s *Server) GetGroups() *Groups {
	return s.mapGroup
}

// onlineproc 可以推送在线统计数据
func (s *Server) onlineproc() {
	// 使用redis做为数据缓存
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	for {
		select {
		case <-time.After(time.Second * 10):
			// xLogic := logic.New(s.c)
			// if err := xLogic.Ping(ctx); err != nil {
			// 	log.Errorln("xLogic.Ping, err:", err.Error())
			// 	continue
			// }
			// xLogic.SetUserOnline(ctx, uint64(s.bucket.Online))
			// xLogic.Close()
		}
	}
}

// RegLog ...
// func (s *Server) RegLog(log *log.Helper) {
// 	// s.log = log
// }

// BroadcastMsg ...
func (s *Server) BroadcastMsg(msg string) {
	// 生成包头
	headerBytes := protocal.NewHeader(config.ImChatBoradcast, config.ImFromTypeSytem)
	// 生成包体
	// 若是由用户发起的，需要在包体中注入发送者信息
	// 生成完整包数据
	body := make(map[string]interface{})
	body["msg"] = msg
	bodyBytes, _ := json.Encode(body)
	imPacket := protocal.NewImPacket(headerBytes, bodyBytes)

	s.globalMq <- imPacket
}
