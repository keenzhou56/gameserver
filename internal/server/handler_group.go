package server

import (
	pb "gameserver/api/protocol"
	"gameserver/pkg/config"
	"gameserver/pkg/protocal"

	"github.com/golang/protobuf/proto"
)

func (srv *Server) JoinGroupHandler(ctx *Context) (int, error) {
	msg := new(pb.JoinGroupMsg)
	err := proto.Unmarshal(ctx.GetBody(), msg)
	if err != nil {
		return 0, err
	}
	groupID := msg.GroupID
	userID := ctx.user.UserID
	if err := CheckGroupIDValid(groupID); err != nil {
		return config.ImErrorCodeGroupID, err
	}

	// 将频道id写入用户数据
	srv.bucket.JoinUserGroupID(userID, groupID)

	// 将用户数据写入group
	srv.mapGroup.JoinGroup(groupID, userID)

	protocal.SendCommon(ctx.conn, ctx.messageType, ctx.fromType, config.ImResponseCodeSuccess, "")

	return config.ImResponseCodeSuccess, nil
}

func (srv *Server) QuitGroupHandler(ctx *Context) (int, error) {
	msg := new(pb.QuitGroupMsg)
	err := proto.Unmarshal(ctx.GetBody(), msg)
	if err != nil {
		return 0, err
	}

	groupID := msg.GetGroupID()
	userID := ctx.user.UserID

	if err := CheckGroupIDValid(groupID); err != nil {
		return config.ImErrorCodeGroupID, err
	}

	// 删除用户所在组
	if err := srv.bucket.DelUserGroupID(userID, groupID); err != nil {
		return config.ImErrorCodeQuitGroup, err
	}

	// 删除组内用户
	if err := srv.mapGroup.DelGroupUserID(groupID, userID); err != nil {
		return config.ImErrorCodeQuitGroup, err
	}

	protocal.SendCommon(ctx.conn, ctx.messageType, ctx.fromType, config.ImResponseCodeSuccess, "")

	return config.ImResponseCodeSuccess, nil
}

func (srv *Server) GroupUserHandler(ctx *Context) (int, error) {
	msg := new(pb.GroupUserList)
	err := proto.Unmarshal(ctx.GetBody(), msg)
	if err != nil {
		return 0, err
	}

	groupID := msg.GetGroupID()

	// 频道成员列表
	groupUserListReply := new(pb.GroupUserListReply)
	if group, err := srv.mapGroup.Get(groupID); err == nil {
		if len(group.UserIDs) > 0 {

			for _, userID := range group.UserIDs {
				if user, err := srv.bucket.GetUser(userID); err == nil {
					groupUser := new(pb.GroupUserMsg)

					groupUser.UserID = user.UserID
					groupUser.PlatformName = user.PlatformName
					groupUser.PlatformID = user.PlatformID

					groupUserListReply.UserList = append(groupUserListReply.UserList, groupUser)
				}
			}
		}
	}

	body, _ := proto.Marshal(groupUserListReply)
	protocal.SendProto(ctx.conn, ctx.messageType, ctx.fromType, body)

	return config.ImResponseCodeSuccess, nil
}
