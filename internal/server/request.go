package server

import (
	"net"
	"time"
)

type Request struct {
	user        *User
	conn        *net.TCPConn
	messageType uint16
	fromType    uint16
	body        []byte
	startTime   int64
}

func NewRequest() *Request {
	ctx := new(Request)
	ctx.startTime = time.Now().UnixNano()
	return ctx
}

func (ctx *Request) GetBody() []byte {
	return ctx.body
}

func (ctx *Request) GetConn() *net.TCPConn {
	return ctx.conn
}

func (ctx *Request) GetMessageType() uint16 {
	return ctx.messageType
}

func (ctx *Request) GetFromType() uint16 {
	return ctx.fromType
}

func (ctx *Request) GetUser() *User {
	return ctx.user
}
