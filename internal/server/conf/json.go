package conf

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var Server struct {
	LogLevel            string
	LogPath             string
	WSAddr              string
	CertFile            string
	KeyFile             string
	TCPAddr             string
	MaxConnNum          int
	ConsolePort         int
	ProfilePath         string
	RedisUser           string
	DBUser              string
	RedisPlayerInstance []string
	RedisPwd            string
	DBPlayerInstance    []string
	DBPwd               string
	DBDNS               map[string]int
}

var RedisServer struct {
	Addr     string
	Password string
	DB       int
}

func InitJson() {
	data, err := ioutil.ReadFile(DevOpen + "/config/server.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &Server)
	if err != nil {
		panic(err)
	}

	data, err = ioutil.ReadFile(DevOpen + "/config/redis.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &RedisServer)
	if err != nil {
		panic(err)
	}
	log.Println(RedisServer.Addr)
}
