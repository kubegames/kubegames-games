package main

import (
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kubegames/kubegames-games/pkg/battle/960204/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/game"
	room "github.com/kubegames/kubegames-sdk/pkg/room/poker"
)

func main() {
	//系统中断捕获
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		rand.Seed(time.Now().UnixNano())
		room := room.NewRoom(&game.RunFasterRoom{})
		room.Run()
	}()

	// 加载游戏配置；时间配置；控制配置
	config.RunFasterConf.LoadRunFasterCfg()

	// 加载机器人配置
	config.RobotConf.LoadRobotCfg()

	// 加载牌型顺序表配置
	config.CardsOrderConf.LoadCardsOrderCfg()

	<-sigs
}
