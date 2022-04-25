#!/usr/bin/env bash
echo ON
cd ~/go/src/game_buyu/crazy_red/
ls
export GOOS=linux
go build -o crazy_red_new;
scp ~/go/src/game_buyu/crazy_red/crazy_red_new  root@192.168.0.46:/home/servers/crazy_red/game/crazy_red_new
scp -r ~/go/src/game_buyu/crazy_red/config.json  root@192.168.0.46:/home/servers/crazy_red/game/

ssh root@192.168.0.46 "pkill crazy_red;cd /home/servers/crazy_red/game/;rm -rf crazy_red;mv crazy_red_new crazy_red;nohup ./crazy_red >>nohup.out 2>&1 &"
ssh root@192.168.0.46 "cd /home/servers/crazy_red/game/;tail nohup.out -n 500 |grep \"### VER\""