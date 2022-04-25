package main

import (
	game_frame2 "go-game-sdk"
	"go-game-sdk/lib/recover"
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	conf "github.com/kubegames/kubegames-games/pkg/battle/960208/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960208/game"

	"os"
	"os/signal"
	"syscall"
)

func main() {
	defer recover.Recover()

	//系统中断捕获
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		rand.Seed(time.Now().UnixNano())
		room := game_frame2.NewRoom(&game.ThreeDollRoom{})
		room.Run()
	}()

	// 加载游戏配置；时间配置；控制配置
	conf.ThreeDollConf.LoadThreeDollCfg()

	// 记载机器人配置
	conf.RobotConf.LoadRobotCfg()

	sig := <-sigs
	log.Tracef("###SIGNAL::%v,   PID:%d", sig, os.Getpid())
}
