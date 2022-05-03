//拉霸

package main

import (
	game_frame "go-game-sdk"
	"go-game-sdk/example/game_LaBa/990201/bibei"
	"go-game-sdk/example/game_LaBa/990201/gamelogic"
	"go-game-sdk/example/game_LaBa/990201/test"
	"go-game-sdk/example/game_LaBa/labacom/config"
	"go-game-sdk/example/game_LaBa/labacom/xiaomali"
	"go-game-sdk/sdk/api"
	"go-game-sdk/sdk/global"
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/log"
)

func maintest() {
	log.Traceln("开始", time.Now())
	var g gamelogic.Game
	g.Init(&config.LBConfig, &xiaomali.XMLConfig, &bibei.BBConfig)
	//这里调用测试工具，如果是正式版本需要屏蔽这个结果
	test.Test(&config.LBConfig, &xiaomali.XMLConfig, &bibei.BBConfig)
	log.Traceln("结束", time.Now())
}

func main() {
	//初始化api
	api.InitAPI(global.GConfig.URL)
	rand.Seed(time.Now().UnixNano())
	gamelogic.LoadConfig()
	config.LBConfig.LoadLabadCfg()
	xiaomali.XMLConfig.LoadXiaoMaLiCfg()
	bibei.BBConfig.LoadBiBeiCfg()
	maintest()

	room := game_frame.NewRoom(&gamelogic.LaBaRoom{})
	room.Run()
}
