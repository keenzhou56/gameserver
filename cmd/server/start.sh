
# 杀掉server进程
for i in `ps aux|grep "main"|grep "log" | grep -v "grep" |awk '{print $2}'`;do
	echo "kill server process $i..."
	kill $i
done

rm ./main -f
rm ./main.I* -f
rm ./main.W* -f
rm ./main.E* -f
rm ./main.a3* -f
rm ./nohup.out -f

go build ./main.go

nohup ./main --log_dir=./ &

