package logic

import (
	"context"
	pb "gameserver/api/logic"
	"gameserver/internal/server/conf"
	"gameserver/internal/server/dao"
	"net"
	"time"
)

const (
	_onlineTick     = time.Second * 10
	_onlineDeadline = time.Minute * 5
)

// Logic struct
type Logic struct {
	c   *conf.Config
	dao *dao.Dao
	// online
	totalIPs   int64
	totalConns int64
	roomCount  map[string]int32
}

// New init
func New(c *conf.Config) (l *Logic) {
	l = &Logic{
		c:   c,
		dao: dao.New(c),
	}
	return l
}

// Ping ping resources is ok.
func (l *Logic) Ping(c context.Context) (err error) {
	return l.dao.Ping(c)
}

// Close close resources.
func (l *Logic) Close() {
	l.dao.Close()
}

// DispatchTCP ...
func (l *Logic) DispatchTCP(conn *net.TCPConn, imType uint16, fromType uint16, body map[string]interface{}) (err error) {
	return nil
}

// Test ...
func (l *Logic) Test(c context.Context) (err error) {
	err = l.dao.AddMapping(c, 11, "11", "22")
	return
}

// Receive receive a message.
func (l *Logic) Receive(c context.Context, userID int64, proto *pb.Proto) (err error) {
	return
}
