package main

import (
	"common/log"
	recover_handle "common/recover"
	"fmt"
	game_frame2 "game_frame_v2/game/logic"
	"game_poker/doudizhu/config"
	"game_poker/doudizhu/game"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	defer recover_handle.RecoverHandle()
	fmt.Println("************************************************")
	fmt.Println("*                                              *")
	fmt.Println("*              Dou Di Zhu System !             *")
	fmt.Println("*                                              *")
	fmt.Println("**********************************************\n")

	fmt.Println("### VER: ", "0.0.9")
	fmt.Println("### PID: ", os.Getpid())

	//系统中断捕获
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		rand.Seed(time.Now().UnixNano())
		room := game_frame2.NewRoom(&game.DouDizhuRoom{})
		room.Run()
	}()

	// 加载游戏配置；时间配置；控制配置
	config.DoudizhuConf.LoadDoudizhuCfg()

	// 加载机器人配置
	config.RobotConf.LoadRobotCfg()

	// 加载牌型顺序表配置
	config.PutScoreConf.LoadPutScoreCfg()

	sig := <-sigs
	log.Warnf("###SIGNAL::%v,   PID:%d", sig, os.Getpid())
}
