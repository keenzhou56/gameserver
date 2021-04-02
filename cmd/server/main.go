package main

import (
	"flag"
	"gameserver/internal/server"
	"gameserver/internal/server/conf"
	"gameserver/pkg/common"
	"gameserver/pkg/log"
	"gameserver/pkg/log/stdlog"
	"os"
	"os/signal"
	"syscall"

	// _ "net/http/pprof"
	"runtime"

	glog "github.com/golang/glog"
)

var (
	logFileName = flag.String("gamelog", "game-server.log", "Log file name")
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}

	logFile, logErr := os.OpenFile(*logFileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if logErr != nil {
		common.Println("Fail to find", *logFile, "cServer start Failed")
		os.Exit(1)
	}
	// os.Stdout or logFile
	logger := stdlog.NewLogger(stdlog.Writer(logFile))
	log := log.NewHelper("gameserver", logger)
	// 开启多核模式
	runtime.GOMAXPROCS(runtime.NumCPU())
	// ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if err := recover(); err != nil {
			common.Println("main defer recover error:", err)
		}
		glog.Flush()
		logger.Close()
		// cancel()
		common.Println("server closed")
	}()

	// 远程获取pprof数据
	// go func() {
	// 	common.Println(http.ListenAndServe("localhost:8080", nil))
	// }()

	// tcp server
	tcpSrv := server.NewServer(conf.Conf)
	tcpSrv.RegisterHandler()
	tcpSrv.RegLog(log)
	if err := server.InitTCP(tcpSrv, conf.Conf.TCPServer.Addr, runtime.NumCPU()); err != nil {
		panic(err)
	}
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
		glog.Errorf("server get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			glog.Errorf("server exit")
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
