package server

import (
	"gameserver/pkg/config"
)

func (s *Server) RegisterHandler() {
	s.InitRouter()
	s.AddRouter(config.ImLogin, s.LoginHandler)
	s.AddRouter(config.ImChatTestReply, s.ApitestHandler)
	s.AddRouter(config.ImKickUser, s.KickUserHandler)
	s.AddRouter(config.ImKickAll, s.KickAllHandler)
	s.AddRouter(config.ImStat, s.StatHandler)
	s.AddRouter(config.ImJoinGroup, s.JoinGroupHandler)
	s.AddRouter(config.ImQuitGroup, s.QuitGroupHandler)
	s.AddRouter(config.ImChatGroup, s.ChatGroupHandler)
	s.AddRouter(config.ImChatBoradcast, s.ChatBoradcastHandler)
	s.AddRouter(config.ImCheckOnline, s.GetUserHandler)
	s.AddRouter(config.ImGroupUserList, s.GroupUserHandler)
}
