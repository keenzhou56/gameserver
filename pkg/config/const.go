package config

// 消息头预定义常量
const (
	LenStackBuf = 4096
	// 自定义信息长度限制
	ImRegisterExtInfoLengthLimt = 1048

	// 来源类型
	ImFromTypeUser  = uint16(0) // 用户
	ImFromTypeSytem = uint16(1) // 系统
	ImFromTypeAi    = uint16(2) // 机器人

	// 协议类型
	ImError      = uint16(1) // 给client发送一条错误消息，一般来说，client收到ImError，都需要断开当前连接并重新连接，如果是重复登录，则只断开、不重连
	ImResponse   = uint16(2) // 返回消息给client
	ImHeartbeat  = uint16(3) // 心跳包
	ImQuitUserMq = uint16(4) // go心跳退出
	// GM协议类型
	ImStat          = uint16(101) // 统计服务器状态
	ImCheckOnline   = uint16(102) // 判断用户是否在线
	ImKickUser      = uint16(103) // 踢某用户下线
	ImKickAll       = uint16(104) // 踢所有用户下线
	ImGroupUserList = uint16(105) // 获取频道用户列表
	// 协议类型
	ImLogin           = uint16(201) // 登录
	ImLogout          = uint16(202) // 退出
	ImRegisterExtInfo = uint16(203) // 注册附加信息
	ImJoinGroup       = uint16(301) // 加入频道
	ImQuitGroup       = uint16(302) // 退出频道
	ImChatBoradcast   = uint16(401) // 世界聊天
	ImChatGroup       = uint16(402) // 频道聊天
	ImChatPrivate     = uint16(403) // 私聊
	ImChatTestReply   = uint16(404) // 测试回包

	// 通知类型
	ImNoticeChat = 0 // 聊天

	// 错误消息
	ImErrorCodeRelogin              = 1  // 重复登录
	ImErrorCodeNoLogin              = 2  // 未登录
	ImErrorCodePacketRead           = 3  // 读取协议包错误
	ImErrorCodePacketBody           = 4  // 解析协议包内容错误
	ImErrorCodeNotAllowedImType     = 5  // 没有权限发送协议
	ImErrorCodePrivateKeyNotMatched = 6  // 私钥不匹配
	ImErrorCodeLoginTokenNotMatched = 7  // 登录token不匹配
	ImErrorCodeTokenNotMatched      = 8  // 消息token不匹配
	ImErrorCodeMsgEmpty             = 9  // 聊天内容为空
	ImErrorCodeUserID               = 10 // 用户id错误，小于=0
	ImErrorCodePlatformID           = 11 // 平台id错误
	ImErrorCodePlatformName         = 12 // 平台名称错误
	ImErrorCodeGroupID              = 13 // 频道id错误，小于=0
	ImErrorCodeUserInfo             = 14 // 读取用户登录信息错误
	ImErrorCodeGroupInfo            = 15 // 读取频道消息错误
	ImErrorCodeExtInfoLength        = 16 // 附加信息长度超出限制
	ImErrorCodeQuitGroup            = 17 // 退出频道错误
	ImErrorCodeLogin                = 18 // 登入错误

	// 系统返回类型、默认是0
	ImResponseCodeSuccess         = 0 // 默认code值
	ImResponseCodeReceiverOffline = 1 // 私聊对象不在线

	GroupIDLengthMin = 1   // 组队id长度最小值
	GroupIDLengthMax = 100 // 组队id长度最大值
)
