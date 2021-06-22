package mysql

import (
	"bytes"
	"fmt"
	"gameserver/internal/server/conf"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang/glog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	buffer.WriteString("@tcp(")
	buffer.WriteString(host)
	buffer.WriteString(":")
	buffer.WriteString(port)
	buffer.WriteString(")/")
	buffer.WriteString(dbname)
	buffer.WriteString("?charset=utf8mb4&parseTime=true&loc=Local")
	//buffer.WriteString("?charset=utf8mb4&parseTime=true&loc=Asia/Shanghai") // loc 统一时区

	dsn := buffer.String()
	glog.Infof("dsn: %s", dsn)
	// dsn := "user:pass@tcp(localhost:9910)/dbname?charset==utf8mb4&parseTime=True&loc=Local"

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             200 * time.Millisecond, // Slow SQL threshold  default: 200 * time.Millisecond, time.Second
			LogLevel:                  logger.Error,           // Log level
			IgnoreRecordNotFoundError: true,                   // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,                   // Disable color
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		fmt.Println(err)
		panic("connect db error")
	}

	sqlDB, err := db.DB()
	// 设置全局表名禁用复数
	// sqlDB.SingularTable(true)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(3000)
	sqlDB.SetConnMaxLifetime(-1) // 根据mysql服务器的超时时间设置 SHOW VARIABLES LIKE '%timeout%';

	// db.Use(prometheus.New(prometheus.Config{
	// 	DBName:          "mdb", // 使用 `DBName` 作为指标 label
	// 	RefreshInterval: 15,    // 指标刷新频率（默认为 15 秒）
	// 	// PushAddr:        "prometheus pusher address", // 如果配置了 `PushAddr`，则推送指标
	// 	StartServer:    true, // 启用一个 http 服务来暴露指标
	// 	HTTPServerPort: 8080, // 配置 http 服务监听端口，默认端口为 8080 （如果您配置了多个，只有第一个 `HTTPServerPort` 会被使用）
	// 	MetricsCollector: []prometheus.MetricsCollector{
	// 		&prometheus.MySQL{
	// 			VariableNames: []string{"Threads_running"},
	// 		},
	// 	}, // 用户自定义指标
	// }))

	return db, err
}

func GetDBMain() *gorm.DB {
	if DBMain != nil {
		return DBMain
	}
	DBIndex := conf.Server.DBDNS["DBMain"]
	var err error
	DBMain, err = openDB(conf.Server.DBUser, conf.Server.DBPwd, conf.Server.DBPlayerInstance[DBIndex], "3306", "mdb")
	if err != nil {
		panic("connect db error")
	}
	return DBMain
}

func GetDBUser(dbSuffix int) *gorm.DB {
	var buffer bytes.Buffer
	buffer.WriteString("udb_")
	buffer.WriteString(strconv.Itoa(dbSuffix))
	dbName := buffer.String()

	if dbUser != nil && len(dbUser) > 0 && dbUser[dbSuffix] != nil {
		return dbUser[dbSuffix]
	}

	var dbDNSBuffer bytes.Buffer
	dbDNSBuffer.WriteString("USERDB")
	dbDNSBuffer.WriteString(strconv.Itoa(dbSuffix))
	dbDNS := dbDNSBuffer.String()
	DBIndex := conf.Server.DBDNS[dbDNS]

	db1, err := openDB(conf.Server.DBUser, conf.Server.DBPwd, conf.Server.DBPlayerInstance[DBIndex], "3306", dbName)
	if err != nil {
		fmt.Println("connect db error", err)
		panic("connect db error")
	}

	dbUser[dbSuffix] = db1
	return dbUser[dbSuffix]
}

func init() {
	// GetDBMain()
	dbUser = make([]*gorm.DB, 10)
	// InitRedis()
	// pong, err := DBRedisClient.Ping().Result()
	// log.Debug("redis ping: %s; error: %v", pong, err)
	//
}
