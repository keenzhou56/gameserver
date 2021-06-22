package server

import (
	"net"
	"time"
)

type Response struct {
	user        *User
	conn        *net.TCPConn
	messageType uint16
	fromType    uint16
	body        []byte
	startTime   int64
}

func NewResponse() *Response {
	ctx := new(Response)
	ctx.startTime = time.Now().UnixNano()
	return ctx
}

func (ctx *Response) GetBody() []byte {
	return ctx.body
}

func (ctx *Response) GetConn() *net.TCPConn {
	return ctx.conn
}

func (ctx *Response) GetMessageType() uint16 {
	return ctx.messageType
}

func (ctx *Response) GetFromType() uint16 {
	return ctx.fromType
}

func (ctx *Response) GetUser() *User {
	return ctx.user
}
