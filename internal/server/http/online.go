package http

import (
	"github.com/gin-gonic/gin"
)

func (s *ServerHTTP) onlineUser(c *gin.Context) {

	res, err := s.logic.GetUserOnline(c)
	if err != nil {
		result(c, nil, RequestErr)
		return
	}

	// s.server.BroadcastMsg("test")

	result(c, res, OK)
}
