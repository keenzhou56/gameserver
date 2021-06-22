package gameconst

// 这里统一定义全局所用常量
const (
	// PUBLIC
	StatusSuccess                 int32 = 1
	StatusFailed                  int32 = 0
	TableLastSuffixUnderline            = "_"
	GameCSVTableStartLineIndex          = 4        // 策划游戏配表csv文件数据起始行
	HeartbeatCheckIntervalSeconds       = 5        // 心跳检查间隔秒数：每5秒检查一次
	HeartbeatTimeoutSeconds             = 30       // 心跳超时秒数：时间是客户端发心跳包间隔时间的3倍
	OperationDBUpdateType               = "update" // 操作数据库类型：更新
	OperationDBInsertType               = "insert" // 操作数据库类型：插入

	// MDB TABLES
	TBLMDBUserAll = "m_player"
	// UDB TABLES
	TBLUDBPlayer = "u_player" // 玩家表
	TBLUDBItem   = "u_item"   // 物品

)
