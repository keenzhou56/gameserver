package main

import (
	"flag"
	"gameserver/internal/logic/conf"
	"gameserver/internal/logic/grpc"
	"gameserver/internal/logic/logic"
	"os"
	"os/signal"
	"syscall"

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
	glog.Infof("im-logic [version: %s env: %+v] start", 1, conf.Conf.Env)
	// logic server
	logicSrv := logic.New(conf.Conf)
	// grpc server
	rpcSrv := grpc.New(conf.Conf.RPCServer, logicSrv)
	// http server
	// httpSrv := http.NewHTTPServer(conf.Conf.HTTPServer, logicSrv)
	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		glog.Errorf("server get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			glog.Errorf("server exit")
			logicSrv.Close()
			// httpSrv.Close()
			rpcSrv.GracefulStop()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
