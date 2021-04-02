#!/bin/bash

# 杀掉server进程
for i in `ps aux|grep "gopath/gameserver/bin/server" | grep -v "grep" |awk '{print $2}'`;do
	echo "kill server process $i..."
	kill $i
done