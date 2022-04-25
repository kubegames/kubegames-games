#!/usr/bin/env bash
echo ON
cd ~/go/src/game_buyu/crazy_red/
ls
export GOOS=linux
go build -o crazy_red_new;
docker build -t harbor.wonderland.life/wanxiang-local/game-970201:latest .
echo "docker打包ok 扫雷红包"
docker push harbor.wonderland.life/wanxiang-local/game-970201:latest
docker rmi harbor.wonderland.life/wanxiang-local/game-970201:latest -f