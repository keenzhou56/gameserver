package main

import (
	"flag"
	"gameserver/internal/server"
	"gameserver/internal/server/conf"
	"gameserver/internal/server/models/mdb"
	"gameserver/internal/server/mysql"
	"gameserver/pkg/common"
	"time"

	"os"
	"os/signal"
	"syscall"

	"net/http"
	_ "net/http/pprof"
	"runtime"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"
)

var (
	logFileName = flag.String("gamelog", "gamelog", "Log file name")
)

func initLog() {
	path := *logFileName
	/* 日志轮转相关函数
	`WithLinkName` 为最新的日志建立软连接
	`WithRotationTime` 设置日志分割的时间，隔多久分割一次
	WithMaxAge 和 WithRotationCount二者只能设置一个
	  `WithMaxAge` 设置文件清理前的最长保存时间
	  `WithRotationCount` 设置文件清理前最多保存的个数
	*/
	// 下面配置日志每隔 1 分钟轮转一个新文件，保留最近 3 分钟的日志文件，多余的自动清理掉。
	writer, _ := rotatelogs.New(
		path+"_%Y-%m-%d-%H-%M.log",
		rotatelogs.WithLinkName(path),
		// rotatelogs.WithMaxAge(time.Duration(180)*time.Second),
		rotatelogs.WithRotationTime(time.Duration(60)*time.Second),
	)
	log.SetOutput(writer)
	log.SetFormatter(&log.JSONFormatter{})
}
func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}

	initLog()

	conf.InitJson()
	mdb.MDB = mysql.GetDBMain()

	// logFile, logErr := os.OpenFile(*logFileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	// if logErr != nil {
	// 	common.Println("logErr", logErr)
	// 	os.Exit(1)
	// }
	// os.Stdout or logFile
	// logger := stdlog.NewLogger(stdlog.Writer(logFile))
	// logger := log.NewStdLogger(stdlog.Writer(logFile))
	// helper := log.NewHelper(logger)

	// 开启多核模式
	runtime.GOMAXPROCS(runtime.NumCPU())
	// ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if err := recover(); err != nil {
			common.Println("main defer recover error:", err)
		}
		// glog.Flush()
		// logger.Close()
		// cancel()
		common.Println("server closed")
	}()

	// 远程获取pprof数据
	go func() {
		common.Println(http.ListenAndServe("localhost:8080", nil))
	}()

	// tcp server
	tcpSrv := server.NewServer(conf.Conf)
	tcpSrv.RegisterHandler()
	// tcpSrv.RegLog(log)
	if err := server.InitTCP(tcpSrv, conf.Conf.TCPServer.Addr, runtime.NumCPU()); err != nil {
		panic(err)
	}

	// xLogic := logic.New(conf.Conf)
	// // xLogic.Ping(ctx)
	// err := xLogic.Pub("mdb", "xx")
	// if err != nil {
	// 	panic(err)
	// }
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// xLogic.Ping(ctx)

	// // logic server
	// logicSrv := logic.New(conf.Conf)
	// // grpc server
	// rpcSrv := grpc.New(conf.Conf.RPCServer, logicSrv)
	// // http server
	// httpSrv := http.NewHTTPServer(conf.Conf.HTTPServer, logicSrv, tcpSrv)
	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		// glog.Errorf("server get a signal %s", s.String())
		log.Errorf("server get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Errorf("server exit")
			tcpSrv.Close()
			// logicSrv.Close()
			// httpSrv.Close()
			// rpcSrv.GracefulStop()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
