package server

import (
	"gameserver/pkg/common"
)

// Stat 统计数据结构
type Stat struct {
	StartTime                int64  // 启动时间，时间戳
	RunTime                  int64  // 运行时间，单位秒
	ConnectCount             int32  // 当前连接数用户数
	MaxConnectCount          int32  // 最大连接用户数
	GroupCount               int32  // 当前频道数
	MaxGroupCount            int32  // 最大频道数
	SysBoradcastMessageCount uint64 // 广播系统消息数
	SysPrivateMessageCount   int32  // 私聊系统消息数
	SysGroupMessageCount     int32  // 频道系统消息数
	BoradcastMessageCount    uint64 // 广播消息数
	PrivateMessageCount      int32  // 私聊消息数
	GroupMessageCount        int32  // 频道消息数
	LoginTimes               int32  // 总登录次数
	SvrGoroutineCount        int32
}

// NewStat ...
func NewStat() *Stat {
	// 初始化stat
	stat := &Stat{}
	stat.StartTime = common.GetTime()
	return stat
}

// Get 获取统计信息
func (stat *Stat) Get() *Stat {
	// 计算运行时间
	stat.RunTime = common.GetTime() - stat.StartTime
	return stat
}
