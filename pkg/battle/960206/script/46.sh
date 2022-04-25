#!/usr/bin/env bash
echo ON
cd ~/go/src/game_buyu/rob_red/
ls
export GOOS=linux
go build -o rob_red_new;
scp ~/go/src/game_buyu/rob_red/rob_red_new  root@192.168.0.46:/home/servers/rob_red/game/rob_red_new
scp -r ~/go/src/game_buyu/rob_red/config  root@192.168.0.46:/home/servers/rob_red/game/

ssh root@192.168.0.46 "pkill rob_red;cd /home/servers/rob_red/game/;rm -rf rob_red;rm -rf nohup.out;mv rob_red_new rob_red;nohup ./rob_red >>nohup.out 2>&1 &"
ssh root@192.168.0.46 "cd /home/servers/rob_red/game/;tail nohup.out -n 500 |grep \"### VER\""