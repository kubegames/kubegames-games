package main

import (
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-games/pkg/battle/960101/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/game"
	room "github.com/kubegames/kubegames-sdk/pkg/room/poker"
)

func main() {

	// 加载游戏配置；时间配置；控制配置
	config.BlackJackConf.LoadBlackjackCfg()

	// 记载机器人配置
	config.RobotConf.LoadRobotCfg()

	rand.Seed(time.Now().UnixNano())

	//初始化房间
	room := room.NewRoom(&game.BlackjackRoom{})
	room.Run()
}
