package main

import (
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	conf "github.com/kubegames/kubegames-games/pkg/battle/960207/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960207/game"
	room "github.com/kubegames/kubegames-sdk/pkg/room/poker"
)

func main() {
	//系统中断捕获
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		rand.Seed(time.Now().UnixNano())
		room := room.NewRoom(&game.GeneralNiuniuRoom{})
		room.Run()
	}()

	// 加载游戏配置；时间配置；控制配置
	conf.GeneralNiuniuConf.LoadGeneralNiuniuCfg()

	// 记载机器人配置
	conf.RobotConf.LoadRobotCfg()

	<-sigs
}
