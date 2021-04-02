package server

import (
	"net"
	"time"
)

type Context struct {
	user        *User
	conn        *net.TCPConn
	messageType uint16
	fromType    uint16
	body        []byte
	startTime   int64
}

// NewUser 生成新用户
func NewContext() *Context {
	ctx := new(Context)
	ctx.startTime = time.Now().UnixNano()
	return ctx
}

func (ctx *Context) GetBody() []byte {
	return ctx.body
}

func (ctx *Context) GetConn() *net.TCPConn {
	return ctx.conn
}

func (ctx *Context) GetMessageType() uint16 {
	return ctx.messageType
}

func (ctx *Context) GetFromType() uint16 {
	return ctx.fromType
}

func (ctx *Context) GetUser() *User {
	return ctx.user
}
