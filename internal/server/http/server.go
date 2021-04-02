package http

import (
	"gameserver/internal/server"
	"gameserver/internal/server/conf"
	"gameserver/internal/server/logic"
	"gameserver/pkg/common"

	"github.com/gin-gonic/gin"
)

// ServerHTTP is http server.
type ServerHTTP struct {
	engine *gin.Engine
	logic  *logic.Logic
	server *server.Server
}

// NewHTTPServer new a http server.
func NewHTTPServer(c *conf.HTTPServer, l *logic.Logic, svr *server.Server) *ServerHTTP {
	defer func() {
		if err := recover(); err != nil {
			common.Println("NewHTTPServer defer recover error:", err)
		}
	}()
	engine := gin.New()
	engine.Use(loggerHandler, recoverHandler)
	go func() {
		if err := engine.Run(c.Addr); err != nil {
			panic(err)
		}
	}()
	s := &ServerHTTP{
		engine: engine,
		logic:  l,
		server: svr,
	}
	s.initRouter()
	return s
}

func (s *ServerHTTP) initRouter() {
	group := s.engine.Group("/im")
	// group.POST("/push/mids", s.pushMids)
	// group.POST("/push/room", s.pushRoom)
	group.GET("/online", s.onlineUser)

}

// Close close the server.
func (s *ServerHTTP) Close() {

}
