package conf

import (
	"flag"
	"os"
	"strconv"
	"time"

	xtime "gameserver/pkg/time"

	"github.com/BurntSushi/toml"
)

const (
	// 定义当前服务器
	// dev: dev服
	DevOpen           = "dev"
	DB10TABLE100 bool = false // 分库分表开启设置
)

var (
	confPath  string
	deployEnv string
	host      string
	weight    int64

	// Conf config
	Conf *Config
)

func init() {
	var (
		defHost, _   = os.Hostname()
		defWeight, _ = strconv.ParseInt(os.Getenv("WEIGHT"), 10, 32)
	)
	flag.StringVar(&confPath, "conf", "server-example.toml", "default config path")
	flag.StringVar(&deployEnv, "deploy.env", os.Getenv("DEPLOY_ENV"), "deploy env. or use DEPLOY_ENV env variable, value: dev/fat1/uat/pre/prod etc.")
	flag.StringVar(&host, "host", defHost, "machine hostname. or use default machine hostname.")
	flag.Int64Var(&weight, "weight", defWeight, "load balancing weight, or use WEIGHT env variable, value: 10 etc.")
}

// Init init config.
func Init() (err error) {
	Conf = Default()
	_, err = toml.DecodeFile(confPath, &Conf)
	return
}

// Default new a config with specified defualt value.
func Default() *Config {
	return &Config{
		Env: &Env{DeployEnv: deployEnv, Host: host, Weight: weight},
		TCPServer: &TCPServer{
			Network:      "tcp",
			Addr:         "3119",
			ReadTimeout:  xtime.Duration(time.Second),
			WriteTimeout: xtime.Duration(time.Second),
		},
		HTTPServer: &HTTPServer{
			Network:      "tcp",
			Addr:         "3111",
			ReadTimeout:  xtime.Duration(time.Second),
			WriteTimeout: xtime.Duration(time.Second),
		},
		Bucket: &Bucket{
			Size:          32,
			Channel:       1024,
			Room:          1024,
			RoutineAmount: 32,
			RoutineSize:   1024,
		},
		RPCClient: &RPCClient{Dial: xtime.Duration(time.Second), Timeout: xtime.Duration(time.Second), SrvAddr: ":3119"},
		RPCServer: &RPCServer{
			Network:           "tcp",
			Addr:              "3119",
			Timeout:           xtime.Duration(time.Second),
			IdleTimeout:       xtime.Duration(time.Second * 60),
			MaxLifeTime:       xtime.Duration(time.Hour * 2),
			ForceCloseWait:    xtime.Duration(time.Second * 20),
			KeepAliveInterval: xtime.Duration(time.Second * 60),
			KeepAliveTimeout:  xtime.Duration(time.Second * 20),
		},
	}
}

// Config config.
type Config struct {
	Env        *Env
	TCPServer  *TCPServer
	HTTPServer *HTTPServer
	Kafka      *Kafka
	Redis      *Redis
	Bucket     *Bucket
	RPCClient  *RPCClient
	RPCServer  *RPCServer
}

// Env is env config.
type Env struct {
	DeployEnv string
	Host      string
	Weight    int64
}

// Redis .
type Redis struct {
	Network      string
	Addr         string
	Auth         string
	Active       int
	Idle         int
	DialTimeout  xtime.Duration
	ReadTimeout  xtime.Duration
	WriteTimeout xtime.Duration
	IdleTimeout  xtime.Duration
	Expire       xtime.Duration
}

// Kafka .
type Kafka struct {
	Topic   string
	Brokers []string
}

// TCPServer is http server config.
type TCPServer struct {
	Debug        bool
	SystemKey    string
	LoginKey     string
	ChatKey      string
	Network      string
	Addr         string
	ReadTimeout  xtime.Duration
	WriteTimeout xtime.Duration
}

// Bucket is bucket config.
type Bucket struct {
	Size          int
	Channel       int
	Room          int
	RoutineAmount uint64
	RoutineSize   int
}

// HTTPServer is http server config.
type HTTPServer struct {
	Network      string
	Addr         string
	ReadTimeout  xtime.Duration
	WriteTimeout xtime.Duration
}

// RPCClient is RPC client config.
type RPCClient struct {
	Dial    xtime.Duration
	Timeout xtime.Duration
	SrvAddr string
}

// RPCServer is RPC server config.
type RPCServer struct {
	Network           string
	Addr              string
	Timeout           xtime.Duration
	IdleTimeout       xtime.Duration
	MaxLifeTime       xtime.Duration
	ForceCloseWait    xtime.Duration
	KeepAliveInterval xtime.Duration
	KeepAliveTimeout  xtime.Duration
}
