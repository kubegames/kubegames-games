package main

import (
	"os"
	"os/signal"
	"syscall"

	conf "github.com/kubegames/kubegames-games/pkg/battle/960203/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960203/game"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	room "github.com/kubegames/kubegames-sdk/pkg/room/poker"
)

func main() {
	log.Traceln("### VER: ", "0.0.9")
	log.Traceln("### PID: ", os.Getpid())

	//系统中断捕获
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		room := room.NewRoom(&game.WatchBankerRoom{})
		room.Run()
	}()

	// 加载游戏配置；时间配置；控制配置
	conf.WatchBankerConf.LoadWatchBankerCfg()

	// 记载机器人配置
	conf.RobotConf.LoadRobotCfg()

	<-sigs
}
