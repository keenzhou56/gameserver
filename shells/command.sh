#im服务的相关命令备份
#!/bin/bash

# 启动im服务
/data1/gopath/im/bin/server -f=/data1/gopath/im/conf/127.0.0.1:8899.conf >> /data1/gopath/im/log/127.0.0.1:8899.log &

# 启动im客户端
/data1/gopath/im/bin/client -f=/data1/gopath/im/conf/127.0.0.1:8899.conf

# 启动GM客户端
/data1/gopath/im/bin/system -f=/data1/gopath/im/conf/127.0.0.1:8899.conf

# 启动压力测试工具
/data1/gopath/im/bin/stress -host=127.0.0.1 -port=8899 -n=1000 -s=10 -t=60
