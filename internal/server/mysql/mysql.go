package mysql

import (
	"bytes"
	"fmt"
	"gameserver/pkg/log"
	"strconv"

	"github.com/jinzhu/gorm"
)

var (
	DBMain *gorm.DB
	dbUser []*gorm.DB
)

func openDB(user string, pwd string, host string, port string, dbname string) (*gorm.DB, error) {
	var buffer bytes.Buffer
	buffer.WriteString(user)
	buffer.WriteString(":")
	buffer.WriteString(pwd)
	buffer.WriteString("@(")
	buffer.WriteString(host)
	buffer.WriteString(":")
	buffer.WriteString(port)
	buffer.WriteString(")/")
	buffer.WriteString(dbname)
	buffer.WriteString("?charset=utf8mb4&parseTime=true&loc=Local")
	//buffer.WriteString("?charset=utf8mb4&parseTime=true&loc=Asia/Shanghai") // loc 统一时区

	dbConnStr := buffer.String()
	log.Debug("db conn: %s", dbConnStr)
	db, err := gorm.Open("mysql", dbConnStr)
	if err != nil {
		fmt.Println(err)
		panic("connect db error")
	}

	// 启用Logger，显示详细日志，打印输出所有执行的具体sql语句
	db.LogMode(true)

	// 设置全局表名禁用复数
	db.SingularTable(true)

	db.DB().SetMaxOpenConns(5) //设置数据库连接池最大连接数
	db.DB().SetMaxIdleConns(2)
	return db, err
}

func GetDBMain() *gorm.DB {
	if DBMain != nil {
		return DBMain
	}
	DBIndex := conf.Server.DBDNS["DBMain"]
	var err error
	DBMain, err = openDB(conf.Server.DBUser, conf.Server.DBPwd, conf.Server.DBPlayerInstance[DBIndex], "3306", "aico_mdb")
	if err != nil {
		panic("connect db error")
	}
	return DBMain
}

func GetDBUser(dbSuffix int) *gorm.DB {
	var buffer bytes.Buffer
	buffer.WriteString("aico_udb_")
	buffer.WriteString(strconv.Itoa(dbSuffix))
	dbName := buffer.String()

	if dbUser[dbSuffix] != nil {
		return dbUser[dbSuffix]
	}

	var dbDNSBuffer bytes.Buffer
	dbDNSBuffer.WriteString("USERDB")
	dbDNSBuffer.WriteString(strconv.Itoa(dbSuffix))
	dbDNS := dbDNSBuffer.String()
	DBIndex := conf.Server.DBDNS[dbDNS]

	db1, err := openDB(conf.Server.DBUser, conf.Server.DBPwd, conf.Server.DBPlayerInstance[DBIndex], "3306", dbName)
	if err != nil {
		panic("connect db error")
	}

	dbUser[dbSuffix] = db1
	return dbUser[dbSuffix]
}

func init() {
	GetDBMain()
	dbUser = make([]*gorm.DB, 10)
	InitRedis()
	pong, err := DBRedisClient.Ping().Result()
	log.Debug("redis ping: %s; error: %v", pong, err)
}
