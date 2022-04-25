#!/usr/bin/env bash
pkill server_po;
cd ~/go/src/server_poker_jh/;
go build;
nohup ./server_poker_jh >> nohup.out &
tail -f nohup.out -n 100