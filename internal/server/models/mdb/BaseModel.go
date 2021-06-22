package mdb

import (
	"gameserver/internal/server/mysql"
	"sync"
)

var (
	Once sync.Once
	MDB  = mysql.DBMain
)
