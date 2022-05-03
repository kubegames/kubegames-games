package main

import (
	recover_handle "common/recover"
	game_frame2 "game_frame_v2/game/logic"
	"game_poker/ddzall/config"
	"game_poker/ddzall/game"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/log"
)

func main() {
	defer recover_handle.RecoverHandle()
	log.Traceln("************************************************")
	log.Traceln("*                                              *")
	log.Traceln("*              Dou Di Zhu System !             *")
	log.Traceln("*                                              *")
	log.Traceln("**********************************************\n")

	log.Traceln("### VER: ", "0.0.9")
	log.Traceln("### PID: ", os.Getpid())

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
