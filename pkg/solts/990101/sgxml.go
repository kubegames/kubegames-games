//拉霸

package main

import (
	game_frame "go-game-sdk"
	"go-game-sdk/example/game_LaBa/990101/gamelogic"
	"go-game-sdk/example/game_LaBa/990101/test"
	"go-game-sdk/example/game_LaBa/labacom/config"
	"go-game-sdk/example/game_LaBa/labacom/xiaomali"
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/log"
)

func maintest() {
	log.Traceln("开始", time.Now())
	var g gamelogic.Game
	g.Init(&config.LBConfig, &xiaomali.XMLConfig)
	//这里调用测试工具，如果是正式版本需要屏蔽这个结果
	test.Test(&config.LBConfig, &xiaomali.XMLConfig)
	log.Traceln("结束", time.Now())
}

func main() {
	rand.Seed(time.Now().UnixNano())
	config.LBConfig.LoadLabadCfg()
	xiaomali.XMLConfig.LoadXiaoMaLiCfg()
	maintest()

	room := game_frame.NewRoom(&gamelogic.LaBaRoom{})
	room.Run()
}
