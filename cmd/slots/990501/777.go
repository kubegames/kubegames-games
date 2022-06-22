//拉霸

package main

import (
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/labacom/config"
	"github.com/kubegames/kubegames-games/pkg/slots/990501/gamelogic"
	"github.com/kubegames/kubegames-games/pkg/slots/990501/test"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	room "github.com/kubegames/kubegames-sdk/pkg/room/slots"
)

func maintest() {
	log.Traceln("开始", time.Now())
	var g gamelogic.Game
	g.Init(&config.LBConfig)
	//这里调用测试工具，如果是正式版本需要屏蔽这个结果
	test.Test(&config.LBConfig)
	log.Traceln("结束", time.Now())
}

func main() {
	rand.Seed(time.Now().UnixNano())
	gamelogic.LoadConfig()
	config.LBConfig.LoadLabadCfg()
	maintest()

	room := room.NewRoom(&gamelogic.LaBaRoom{})
	room.Run()
}
