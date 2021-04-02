package server

import (
	"context"
	"gameserver/internal/server/conf"
	"gameserver/pkg/common"
	"net"
	"time"

	glog "github.com/golang/glog"
)

const (
	maxInt = 1<<31 - 1
)

var receivedAiMsgCount uint64
var sendedAiMsgCount uint64

// InitTCP listen all tcp.bind and start accept connections.
func InitTCP(s *Server, bind string, accept int) (err error) {
	var (
		listener *net.TCPListener
		addr     *net.TCPAddr
	)
	if addr, err = net.ResolveTCPAddr("tcp", bind); err != nil {
		glog.Errorf("net.ResolveTCPAddr(tcp, %s) error(%v)", bind, err)
		return
	}
	if listener, err = net.ListenTCP("tcp", addr); err != nil {
		glog.Errorf("net.ListenTCP(tcp, %s) error(%v)", bind, err)
		return
	}
	glog.Infof("start tcp listen: %s", bind)
	common.Println("start tcp listen:", bind)
	// split N core accept
	for i := 0; i < accept; i++ {
		go acceptTCP(s, listener)
	}
	go s.broadcaster()
	return
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.  Accept blocks; the caller typically
// invokes it in a go statement.
func acceptTCP(s *Server, lis *net.TCPListener) {
	var (
		conn *net.TCPConn
		err  error
		// r    int
	)

	for {
		if conn, err = lis.AcceptTCP(); err != nil {
			// if listener close then return
			glog.Errorf("listener.Accept(\"%s\") error(%v)", lis.Addr().String(), err)
			return
		}
		go s.dispatchTCP(conn)
		// if r++; r == maxInt {
		// 	r = 0
		// }
	}
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (s *Server) dispatchTCP(conn *net.TCPConn) {
	// 当前连接的用户id
	user := NewUser()
	defer func() {
		// 捕获异常
		if err := recover(); err != nil {
			common.Println("dispatchTCP defer recover error:", err)
		}
		// 清除用户数据
		if user.UserID > 0 {
			s.removeUser(user.UserID, conn)
			common.Println("dispatchTCP defer conn.close, clientIP:"+conn.RemoteAddr().String(), "userID:", user.UserID)
		}
		conn.Close()
		// runtime.Goexit()
	}()

	go user.handleLoop(s, conn)
	user.readLoop(conn)

}

// auth for goim handshake with client, use rsa & aes.
func (s *Server) authTCP(ctx context.Context) (mid int64, key, rid string, accepts []int32, hb time.Duration, err error) {
	return
}

// 计算登录token
func (s *Server) getLoginToken(userID int64, time int64) string {
	return common.GetToken(conf.Conf.TCPServer.LoginKey, userID, time)
}

// 创建Api token
func (s *Server) generateToken(userID int64) string {
	return common.GetToken(conf.Conf.TCPServer.ChatKey, userID, common.GetTime())
}

// 计算gmtoken
func (s *Server) getGmToken(userID int64, time int64) string {
	return common.GetToken(conf.Conf.TCPServer.SystemKey, userID, time)
}

// 移除用户，此操作会从mapUser移除用户，并且会从所有Group中移除用户
func (s *Server) removeUser(userID int64, conn *net.TCPConn) {
	user, err := s.bucket.GetUser(userID)
	if err != nil {
		common.Println(err)
		return
	}
	// 如果取得的用户连接，和当前连接不一样，表示已经被重新登录，则直接退出，不处理别的
	if user.Conn != conn {
		return
	}

	// 将用户从所有加入的频道移除
	if len(user.GroupIDs) > 0 {
		s.mapGroup.BatchDelUser(user.GroupIDs, userID)
	}
	// 状态更改
	user.ClosedSig <- true
	user.Closed = true
	// 将用户移除mapUser
	s.bucket.DelUser(userID)

	if conf.Conf.TCPServer.Debug {
		common.Println("removeUser disconnected :", userID)
	}

}

func (s *Server) broadcaster() {
	for {
		select {
		case imPacket := <-s.globalMq: // <-time.After(time.Second * 5)
			dst := make([]*User, 0)
			s.bucket.mapUser.Range(func(key, value interface{}) bool {
				if !value.(*User).GmFlag {
					dst = append(dst, value.(*User))
				}
				return true
			})

			for _, v := range dst {
				v.outChan <- imPacket
			}

			time.Sleep(time.Second * 5)
		}
	}
}
